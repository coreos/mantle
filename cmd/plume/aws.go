// Copyright 2015 CoreOS, Inc.
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
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/coreos/mantle/platform/amazon"
	"github.com/coreos/mantle/sdk"

	"github.com/spf13/cobra"
)

var (
	cmdAWS = &cobra.Command{
		Use:   "aws",
		Short: "manage aws ec2 and s3",
	}

	// options shared by aws commands
	coreosVersion  string
	coreosGroup    string
	awsRegion      amazon.Region = amazon.Region("us-west-1")
	s3Bucket       amazon.Bucket = amazon.Bucket("mischief-testing")
	diskSize       int64
	forceOverwrite bool

	cmdAWSOneShot = &cobra.Command{
		Use:   "one-shot [file]",
		Short: "one-shot ami creation",
		Long:  "create an ami from a local file in one shot",
		Run:   runAWSOneShot,
	}

	cmdAWSUpload = &cobra.Command{
		Use:   "s3-upload [file]",
		Short: "upload a file to s3",
		Run:   runAWSUpload,
	}

	cmdAWSManifest = &cobra.Command{
		Use:   "s3-manifest [disk-url]",
		Short: "create a import volume manifest",
		Run:   runAWSManifest,
	}

	cmdAWSImportVolume = &cobra.Command{
		Use:   "import-volume [manifest-url]",
		Short: "import a ec2 volume from s3 manifest",
		Run:   runAWSImportVolume,
	}

	cmdAWSCreateSnapshot = &cobra.Command{
		Use:   "create-snapshot [volume-id]",
		Short: "create a snapshot from a volume",
		Run:   runAWSCreateSnapshot,
	}

	cmdAWSRegisterImage = &cobra.Command{
		Use:   "register-image [snap-id]",
		Short: "create an ami from a snapshot",
		Run:   runAWSRegisterImage,
	}

	createHVM bool
)

func init() {
	sv := cmdAWS.PersistentFlags().StringVar

	sv(&coreosVersion, "version", "master", "AMI version")
	sv(&coreosGroup, "group", "master", "update group")

	cmdAWS.PersistentFlags().Var(&awsRegion, "region", "aws region")
	cmdAWS.PersistentFlags().Var(&s3Bucket, "bucket", "aws s3 bucket")

	cmdAWS.PersistentFlags().Int64Var(&diskSize, "size", 8, "size of disk images in GB")

	cmdAWS.PersistentFlags().BoolVar(&forceOverwrite, "overwrite", false, "force overwriting")

	root.AddCommand(cmdAWS)

	cmdAWS.AddCommand(cmdAWSOneShot)
	cmdAWS.AddCommand(cmdAWSUpload)
	cmdAWS.AddCommand(cmdAWSManifest)
	cmdAWS.AddCommand(cmdAWSImportVolume)
	cmdAWS.AddCommand(cmdAWSCreateSnapshot)

	cmdAWSRegisterImage.Flags().BoolVar(&createHVM, "hvm", false, "create a hvm instead of pv ami")
	cmdAWS.AddCommand(cmdAWSRegisterImage)
}

func runAWSOneShot(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		os.Exit(1)
	}

	awsapi, err := amazon.NewAWSAPI(awsRegion)
	if err != nil {
		plog.Fatalf("Failed to create AWS API client: %v", err)
	}

	image := args[0]

	u, err := url.Parse(image)
	if err != nil {
		plog.Fatalf("invalid image %q: %v", image, err)
	}

	// TODO(mischief): accept http:// or gs://
	if u.Scheme != "" {
		plog.Fatalf("invalid image %q: must be a local file path", image)
	}

	// check if it exists
	fi, err := os.Stat(u.Path)
	if err != nil {
		plog.Fatalf("invalid image %q: %v", image, err)
	}

	plog.Printf("using %d-byte image %q", fi.Size(), u.Path)

	plog.Printf("epoch %+v", sdk.GetEpoch())

	if coreosVersion == "master" {
		coreosVersion = fmt.Sprintf("%d-%s", sdk.GetEpoch(), time.Now().Format("15:04"))
		coreosGroup = "master"
	}

	// in GB
	diskSize := 8

	arch := "x86_64"
	description := fmt.Sprintf("CoreOS %s %s", coreosGroup, coreosVersion)

	// open image
	var file *os.File

	switch u.Scheme {
	case "":
		file, err = os.Open(u.Path)
		if err != nil {
			plog.Fatalf("failed to open image file: %v", err)
		}
		defer file.Close()
	case "http":
		fallthrough
	case "https":
	case "gs":
		plog.Fatal("unsupported")
	}

	// if it's bzipped, transparently unzip it.
	// XXX: boo, s3 will only take a ReadSeeker.
	//if path.Ext(u.Path) == ".bz2" {
	//	reader = bzip2.NewReader(reader)
	//}

	typ, err := amazon.ImageType(u.Path)
	if err != nil {
		plog.Errorf("failed to detect image type: %v", err)
		return
	}

	plog.Infof("size = %dG arch = %s description = %q typ = %s", diskSize, arch, description, typ)

	// upload image to S3
	s3url, err := awsapi.S3Put(file, s3Bucket, fmt.Sprintf("%d/%s", sdk.GetEpoch(), path.Base(u.Path)), false)
	if err != nil {
		plog.Errorf("failed to upload image to S3: %v", err)
		return
	}

	plog.Debugf("s3 url is %v", s3url)

	mf, err := amazon.GenerateManifest(typ, s3url, fi.Size())
	if err != nil {
		plog.Errorf("Failed generating manifest: %v", err)
		return
	}

	murl, err := awsapi.S3Put(mf, s3Bucket, fmt.Sprintf("%d/%s.manifest.xml", sdk.GetEpoch(), path.Base(u.Path)), false)
	if err != nil {
		plog.Errorf("Failed uploading manifest: %v", err)
		return
	}

	plog.Debugf("manifest is %v", murl)

	// Create a EC2 volume from a manifest on S3
	plog.Infof("Importing volume to EC2")

	// create 8gb volume
	vid, err := awsapi.ImportVolume(murl.String(), typ, 8)
	if err != nil {
		plog.Errorf("Failed to import volume: %v", err)
		return
	}
	plog.Infof("Volume %s ready", vid)

	// TODO(mischief): delete disk from S3

	// Create a EC2 snapshot from EC2 volume
	plog.Info("Creating snapshot")

	snap, err := awsapi.CreateSnapshot(vid)
	if err != nil {
		plog.Errorf("Failed to create snapshot: %v", err)
		return
	}
	plog.Infof("Created snapshot %s", snap)

	// TODO(mischief): delete volume

	// Register AMIs from snapshot
	name := fmt.Sprintf("CoreOS-%s-%d", coreosGroup, sdk.GetEpoch())

	plog.Info("Registering HVM AMI")
	hvmid, err := awsapi.RegisterImage(name+"-hvm", snap, true)
	if err != nil {
		plog.Errorf("Failed to create HVM AMI: %v", err)
		return
	}
	plog.Infof("HVM AMI is %s", hvmid)

	plog.Info("Registering PV AMI")
	pvid, err := awsapi.RegisterImage(name, snap, false)
	if err != nil {
		plog.Errorf("Failed to create PV AMI: %v", err)
		return
	}

	plog.Infof("PV AMI is %s", pvid)
}

func runAWSUpload(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		os.Exit(1)
	}

	awsapi, err := amazon.NewAWSAPI(awsRegion)
	if err != nil {
		plog.Errorf("Failed to create AWS API client: %v", err)
		os.Exit(1)
	}

	file, err := os.Open(args[0])
	if err != nil {
		plog.Errorf("Failed to open image file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	// upload image to S3
	s3url, err := awsapi.S3Put(file, s3Bucket, fmt.Sprintf("%d/%s", sdk.GetEpoch(), path.Base(file.Name())), forceOverwrite)
	if err != nil {
		plog.Errorf("Failed to upload image to S3: %v", err)
		os.Exit(1)
	}

	fmt.Println(s3url)
}

func runAWSManifest(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		os.Exit(1)
	}

	u, err := url.Parse(args[0])
	if err != nil {
		plog.Fatalf("Failed to parse disk URL: %v", err)
	}

	typ, err := amazon.ImageType(u.Path)
	if err != nil {
		plog.Fatalf("Failed to determine image type: %v", err)
	}

	awsapi, err := amazon.NewAWSAPI(awsRegion)
	if err != nil {
		plog.Fatalf("Failed to create AWS API client: %v", err)
	}

	size, err := urlSize(u.String())
	if err != nil {
		plog.Fatal("Failed to determine disk image size: %v", err)
	}

	mf, err := amazon.GenerateManifest(typ, u, size)
	if err != nil {
		plog.Errorf("Failed generating manifest: %v", err)
		return
	}

	murl, err := awsapi.S3Put(mf, s3Bucket, fmt.Sprintf("%d/%s.manifest.xml", sdk.GetEpoch(), path.Base(u.Path)), false)
	if err != nil {
		plog.Errorf("Failed uploading manifest: %v", err)
		return
	}

	fmt.Println(murl)
}

func runAWSImportVolume(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		os.Exit(1)
	}

	u, err := url.Parse(args[0])
	if err != nil {
		plog.Fatalf("Failed to parse disk URL: %v", err)
	}

	typ, err := amazon.ImageType(strings.TrimSuffix(u.Path, ".manifest.xml"))
	if err != nil {
		plog.Fatalf("Failed to determine image type: %v", err)
	}

	awsapi, err := amazon.NewAWSAPI(awsRegion)
	if err != nil {
		plog.Fatalf("Failed to create AWS API client: %v", err)
	}

	vid, err := awsapi.ImportVolume(u.String(), typ, diskSize)
	if err != nil {
		plog.Fatalf("Failed to import volume: %v", err)
	}

	fmt.Println(vid)
}

func runAWSCreateSnapshot(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		os.Exit(1)
	}

	var vid amazon.VolumeID

	if err := vid.Set(args[0]); err != nil {
		plog.Fatal("Failed to parse volume ID: %v", err)
	}

	awsapi, err := amazon.NewAWSAPI(awsRegion)
	if err != nil {
		plog.Fatalf("Failed to create AWS API client: %v", err)
	}

	snap, err := awsapi.CreateSnapshot(&vid)
	if err != nil {
		plog.Errorf("Failed to create snapshot: %v", err)
		return
	}

	fmt.Println(snap)
}

func runAWSRegisterImage(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		os.Exit(1)
	}

	var sid amazon.SnapshotID

	if err := sid.Set(args[0]); err != nil {
		plog.Fatalf("Failed to parse snapshot ID: %v", err)
	}

	awsapi, err := amazon.NewAWSAPI(awsRegion)
	if err != nil {
		plog.Fatalf("Failed to create AWS API client: %v", err)
	}

	name := fmt.Sprintf("CoreOS-%s-%d", coreosGroup, sdk.GetEpoch())

	if createHVM {
		name += "-hvm"
	}

	amiid, err := awsapi.RegisterImage(name, &sid, createHVM)
	if err != nil {
		plog.Fatalf("Failed to register EC2 AMI: %v", err)
	}

	fmt.Println(amiid)
}

func urlSize(url string) (int64, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, err
	}

	resp.Body.Close()

	return resp.ContentLength, nil
}
