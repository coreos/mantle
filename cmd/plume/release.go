// Copyright 2016 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/api/compute/v1"
	gs "google.golang.org/api/storage/v1"

	"github.com/coreos/mantle/auth"
	"github.com/coreos/mantle/platform/api/aws"
	"github.com/coreos/mantle/platform/api/azure"
	"github.com/coreos/mantle/platform/api/gcloud"
	"github.com/coreos/mantle/platform/api/oci"
	"github.com/coreos/mantle/storage"
	"github.com/coreos/mantle/storage/index"
)

var (
	releaseDryRun bool
	cmdRelease    = &cobra.Command{
		Use:   "release [options]",
		Short: "Publish a new CoreOS release.",
		Run:   runRelease,
		Long:  `Publish a new CoreOS release.`,
	}
)

func init() {
	cmdRelease.Flags().StringVar(&awsCredentialsFile, "aws-credentials", "", "AWS credentials file")
	cmdRelease.Flags().StringVar(&azureProfile, "azure-profile", "", "Azure Profile json file")
	cmdRelease.Flags().StringVar(&ociCredentialsFile, "oci-credentials", "", "OCI credentials file")
	cmdRelease.Flags().BoolVarP(&releaseDryRun, "dry-run", "n", false,
		"perform a trial run, do not make changes")
	AddSpecFlags(cmdRelease.Flags())
	root.AddCommand(cmdRelease)
}

func runRelease(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		plog.Fatal("No args accepted")
	}

	spec := ChannelSpec()
	ctx := context.Background()
	client, err := getGoogleClient()
	if err != nil {
		plog.Fatalf("Authentication failed: %v", err)
	}

	src, err := storage.NewBucket(client, spec.SourceURL())
	if err != nil {
		plog.Fatal(err)
	}
	src.WriteDryRun(releaseDryRun)

	if err := src.Fetch(ctx); err != nil {
		plog.Fatal(err)
	}

	// Sanity check!
	if vertxt := src.Object(src.Prefix() + "version.txt"); vertxt == nil {
		verurl := src.URL().String() + "version.txt"
		plog.Fatalf("File not found: %s", verurl)
	}

	// Register GCE image if needed.
	doGCE(ctx, client, src, &spec)

	// Make Azure images public.
	doAzure(ctx, client, src, &spec)

	// Make AWS images public.
	doAWS(ctx, client, src, &spec)

	// Make OCI images public.
	doOCI(ctx, client, src, &spec)

	for _, dSpec := range spec.Destinations {
		dst, err := storage.NewBucket(client, dSpec.BaseURL)
		if err != nil {
			plog.Fatal(err)
		}
		dst.WriteDryRun(releaseDryRun)

		// Fetch parent directories non-recursively to re-index it later.
		for _, prefix := range dSpec.ParentPrefixes() {
			if err := dst.FetchPrefix(ctx, prefix, false); err != nil {
				plog.Fatal(err)
			}
		}

		// Fetch and sync each destination directory.
		for _, prefix := range dSpec.FinalPrefixes() {
			if err := dst.FetchPrefix(ctx, prefix, true); err != nil {
				plog.Fatal(err)
			}

			sync := index.NewSyncIndexJob(src, dst)
			sync.DestinationPrefix(prefix)
			sync.DirectoryHTML(dSpec.DirectoryHTML)
			sync.IndexHTML(dSpec.IndexHTML)
			sync.Delete(true)
			if dSpec.Title != "" {
				sync.Name(dSpec.Title)
			}
			if err := sync.Do(ctx); err != nil {
				plog.Fatal(err)
			}
		}

		// Now refresh the parent directory indexes.
		for _, prefix := range dSpec.ParentPrefixes() {
			parent := index.NewIndexJob(dst)
			parent.Prefix(prefix)
			parent.DirectoryHTML(dSpec.DirectoryHTML)
			parent.IndexHTML(dSpec.IndexHTML)
			parent.Recursive(false)
			parent.Delete(true)
			if dSpec.Title != "" {
				parent.Name(dSpec.Title)
			}
			if err := parent.Do(ctx); err != nil {
				plog.Fatal(err)
			}
		}
	}
}

func sanitizeVersion() string {
	v := strings.Replace(specVersion, ".", "-", -1)
	return strings.Replace(v, "+", "-", -1)
}

func gceWaitForImage(pending *gcloud.Pending) {
	plog.Infof("Waiting for image creation to finish...")
	pending.Interval = 3 * time.Second
	pending.Progress = func(_ string, _ time.Duration, op *compute.Operation) error {
		status := strings.ToLower(op.Status)
		if op.Progress != 0 {
			plog.Infof("Image creation is %s: %s % 2d%%", status, op.StatusMessage, op.Progress)
		} else {
			plog.Infof("Image creation is %s. %s", status, op.StatusMessage)
		}
		return nil
	}
	if err := pending.Wait(); err != nil {
		plog.Fatal(err)
	}
	plog.Info("Success!")
}

func gceUploadImage(spec *channelSpec, api *gcloud.API, obj *gs.Object, name, desc string) string {
	plog.Noticef("Creating GCE image %s", name)
	op, pending, err := api.CreateImage(&gcloud.ImageSpec{
		SourceImage: obj.MediaLink,
		Family:      spec.GCE.Family,
		Name:        name,
		Description: desc,
		Licenses:    spec.GCE.Licenses,
	}, false)
	if err != nil {
		plog.Fatalf("GCE image creation failed: %v", err)
	}

	gceWaitForImage(pending)

	return op.TargetLink
}

func doGCE(ctx context.Context, client *http.Client, src *storage.Bucket, spec *channelSpec) {
	if spec.GCE.Project == "" || spec.GCE.Image == "" {
		plog.Notice("GCE image creation disabled.")
		return
	}

	api, err := gcloud.New(&gcloud.Options{
		Project:     spec.GCE.Project,
		JSONKeyFile: gceJSONKeyFile,
	})
	if err != nil {
		plog.Fatalf("GCE client failed: %v", err)
	}

	nameVer := fmt.Sprintf("%s-%s-v", spec.GCE.Family, sanitizeVersion())
	date := time.Now().UTC()
	name := nameVer + date.Format("20060102")
	desc := fmt.Sprintf("%s, %s, %s published on %s", spec.GCE.Description,
		specVersion, specBoard, date.Format("2006-01-02"))

	images, err := api.ListImages(ctx, spec.GCE.Family+"-")
	if err != nil {
		plog.Fatal(err)
	}

	var conflicting, oldImages []*compute.Image
	for _, image := range images {
		if strings.HasPrefix(image.Name, nameVer) {
			conflicting = append(conflicting, image)
		} else {
			oldImages = append(oldImages, image)
		}
	}
	sort.Slice(oldImages, func(i, j int) bool {
		getCreation := func(image *compute.Image) time.Time {
			stamp, err := time.Parse(time.RFC3339, image.CreationTimestamp)
			if err != nil {
				plog.Fatalf("Couldn't parse timestamp %q: %v", image.CreationTimestamp, err)
			}
			return stamp
		}
		return getCreation(oldImages[i]).After(getCreation(oldImages[j]))
	})

	// Check for any with the same version but possibly different dates.
	var imageLink string
	if len(conflicting) > 1 {
		plog.Fatalf("Duplicate GCE images found: %v", conflicting)
	} else if len(conflicting) == 1 {
		image := conflicting[0]
		name = image.Name
		imageLink = image.SelfLink

		if image.Status == "FAILED" {
			plog.Fatalf("Found existing GCE image %q in state %q", name, image.Status)
		}

		plog.Noticef("GCE image already exists: %s", name)

		if releaseDryRun {
			return
		}

		if image.Status == "PENDING" {
			pending, err := api.GetPendingForImage(image)
			if err != nil {
				plog.Fatalf("Couldn't wait for image creation: %v", err)
			}
			gceWaitForImage(pending)
		}
	} else {
		obj := src.Object(src.Prefix() + spec.GCE.Image)
		if obj == nil {
			plog.Fatalf("GCE image not found %s%s", src.URL(), spec.GCE.Image)
		}

		if releaseDryRun {
			plog.Noticef("Would create GCE image %s", name)
			return
		}

		imageLink = gceUploadImage(spec, api, obj, name, desc)
	}

	if spec.GCE.Publish != "" {
		obj := gs.Object{
			Name:        src.Prefix() + spec.GCE.Publish,
			ContentType: "text/plain",
		}
		media := strings.NewReader(
			fmt.Sprintf("projects/%s/global/images/%s\n",
				spec.GCE.Project, name))
		if err := src.Upload(ctx, &obj, media); err != nil {
			plog.Fatal(err)
		}
	} else {
		plog.Notice("GCE image name publishing disabled.")
	}

	var pendings []*gcloud.Pending
	for _, old := range oldImages {
		if old.Deprecated != nil && old.Deprecated.State != "" {
			continue
		}
		plog.Noticef("Deprecating old image %s", old.Name)
		pending, err := api.DeprecateImage(old.Name, gcloud.DeprecationStateDeprecated, imageLink)
		if err != nil {
			plog.Fatal(err)
		}
		pending.Interval = 1 * time.Second
		pending.Timeout = 0
		pendings = append(pendings, pending)
	}

	if spec.GCE.Limit > 0 && len(oldImages) > spec.GCE.Limit {
		plog.Noticef("Pruning %d GCE images.", len(oldImages)-spec.GCE.Limit)
		for _, old := range oldImages[spec.GCE.Limit:] {
			if old.Name == "coreos-alpha-1122-0-0-v20160727" {
				plog.Noticef("%v: not deleting: hardcoded solution to hardcoded problem", old.Name)
				continue
			}
			plog.Noticef("Deleting old image %s", old.Name)
			pending, err := api.DeleteImage(old.Name)
			if err != nil {
				plog.Fatal(err)
			}
			pending.Interval = 1 * time.Second
			pending.Timeout = 0
			pendings = append(pendings, pending)
		}
	}

	plog.Infof("Waiting on %d operations.", len(pendings))
	for _, pending := range pendings {
		err := pending.Wait()
		if err != nil {
			plog.Fatal(err)
		}
	}
}

func doAzure(ctx context.Context, client *http.Client, src *storage.Bucket, spec *channelSpec) {
	if spec.Azure.StorageAccount == "" {
		plog.Notice("Azure image creation disabled.")
		return
	}

	prof, err := auth.ReadAzureProfile(azureProfile)
	if err != nil {
		plog.Fatalf("failed reading Azure profile: %v", err)
	}

	// channel name should be caps for azure image
	imageName := fmt.Sprintf("%s-%s-%s", spec.Azure.Offer, strings.Title(specChannel), specVersion)

	for _, environment := range spec.Azure.Environments {
		opt := prof.SubscriptionOptions(environment.SubscriptionName)
		if opt == nil {
			plog.Fatalf("couldn't find subscription %q", environment.SubscriptionName)
		}

		api, err := azure.New(opt)
		if err != nil {
			plog.Fatalf("failed to create Azure API: %v", err)
		}

		if releaseDryRun {
			// TODO(bgilbert): check that the image exists
			plog.Printf("Would share %q on %v", imageName, environment.SubscriptionName)
			continue
		} else {
			plog.Printf("Sharing %q on %v...", imageName, environment.SubscriptionName)
		}

		if err := api.ShareImage(imageName, "public"); err != nil {
			plog.Fatalf("failed to share image %q: %v", imageName, err)
		}
	}
}

func doAWS(ctx context.Context, client *http.Client, src *storage.Bucket, spec *channelSpec) {
	if spec.AWS.Image == "" {
		plog.Notice("AWS image creation disabled.")
		return
	}

	imageName := fmt.Sprintf("%v-%v-%v", spec.AWS.BaseName, specChannel, specVersion)
	imageName = regexp.MustCompile(`[^A-Za-z0-9()\\./_-]`).ReplaceAllLiteralString(imageName, "_")

	for _, part := range spec.AWS.Partitions {
		for _, region := range part.Regions {
			if releaseDryRun {
				plog.Printf("Checking for images in %v %v...", part.Name, region)
			} else {
				plog.Printf("Publishing images in %v %v...", part.Name, region)
			}

			api, err := aws.New(&aws.Options{
				CredentialsFile: awsCredentialsFile,
				Profile:         part.Profile,
				Region:          region,
			})
			if err != nil {
				plog.Fatalf("creating client for %v %v: %v", part.Name, region, err)
			}

			publish := func(imageName string) {
				imageID, err := api.FindImage(imageName)
				if err != nil {
					plog.Fatalf("couldn't find image %q in %v %v: %v", imageName, part.Name, region, err)
				}

				if !releaseDryRun {
					err := api.PublishImage(imageID)
					if err != nil {
						plog.Fatalf("couldn't publish image in %v %v: %v", part.Name, region, err)
					}
				}
			}
			if aws.RegionSupportsPV(region) {
				publish(imageName)
			}
			publish(imageName + "-hvm")
		}
	}
}

func doOCI(ctx context.Context, client *http.Client, src *storage.Bucket, spec *channelSpec) {
	if spec.OCI.Image == "" {
		plog.Notice("OCI image creation disabled.")
		return
	}

	imageName := fmt.Sprintf("%v-%v-%v.qcow", spec.OCI.BaseName, specChannel, specVersion)
	imageName = regexp.MustCompile(`[^A-Za-z0-9()\\./_-]`).ReplaceAllLiteralString(imageName, "_")

	if releaseDryRun {
		plog.Print("Checking for images...")
	} else {
		plog.Printf("Publishing images...")
	}

	api, err := oci.New(&oci.Options{
		ConfigPath: ociCredentialsFile,
		Region:     spec.OCI.BucketRegion,
	})
	if err != nil {
		plog.Fatalf("creating client: %v", err)
	}

	imagePath, err := getImageFile(client, src, spec.OCI.Image)
	if err != nil {
		plog.Fatalf("retrieving image: %v", err)
	}

	f, err := os.Open(imagePath)
	if err != nil {
		plog.Fatalf("opening image: %v", err)
	}
	defer f.Close()

	err = api.UploadObject(f, spec.OCI.Bucket, imageName, false)
	if err != nil {
		plog.Fatalf("uploading image: %v", err)
	}
}
