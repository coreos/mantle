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
	"path/filepath"
	"strings"
	"time"

	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/coreos/mantle/sdk"
	"github.com/coreos/mantle/util"
)

var (
	downloadImageCmd = &cobra.Command{
		Use:   "download-image",
		Short: "Download and verify CoreOS images",
		Long:  "Download and verify current CoreOS images to a local cache.",
		Run:   runDownloadImage,
	}
	downloadImageRoot         string
	downloadImageCacheDir     string
	downloadImagePrefix       string
	downloadImageVerify       bool
	downloadImageRetries      int
	downloadImagePlatformList platformList
)

func init() {
	downloadImageCmd.Flags().StringVar(&downloadImageRoot,
		"root", "http://storage.core-os.net/coreos/amd64-usr/master/", "base URL of images")
	downloadImageCmd.Flags().StringVar(&downloadImageCacheDir,
		"cache-dir", filepath.Join(sdk.RepoCache(), "images"), "local dir for image cache")
	downloadImageCmd.Flags().StringVar(&downloadImagePrefix,
		"image-prefix", "coreos_production", "image filename prefix")
	downloadImageCmd.Flags().BoolVar(&downloadImageVerify,
		"verify", true, "verify image signature")
	downloadImageCmd.Flags().IntVar(&downloadImageRetries,
		"retries", 0, "re-attempt download this many times on failure")
	downloadImageCmd.Flags().Var(&downloadImagePlatformList,
		"platform", "Choose qemu, gce, or aws. Multiple platforms can be specified by repeating the flag")

	root.AddCommand(downloadImageCmd)
}

type platformList []string // satisfies pflag.Value interface

func (platforms *platformList) String() string {
	return fmt.Sprintf("%v", *platforms)
}

// not sure what this is for, but won't compile without it
func (platforms *platformList) Type() string {
	return "platformList"
}

// Set will append additional platform for each flag set. Comma
// separated flags without spaces will also be parsed correctly.
func (platforms *platformList) Set(value string) error {

	// maps names of platforms to filename prefix
	platformMap := map[string]string{
		"qemu": "_image.bin.bz2",
		"gce":  "_gce.tar.gz",
		"aws":  "_ami_image.bin.bz2",
	}

	values := strings.Split(value, ",")

	for _, platform := range values {
		prefix, ok := platformMap[platform]
		if !ok {
			plog.Fatalf("platform not supported: %v", platform)
		}
		*platforms = append(*platforms, prefix)
	}
	return nil
}

func runDownloadImage(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		plog.Fatalf("Unrecognized arguments: %v", args)
	}

	if downloadImageCacheDir == "" {
		plog.Fatal("Missing --cache-dir=FILEPATH")
	}
	if len(downloadImagePlatformList) == 0 {
		plog.Fatal("Must specify 1 or more platforms to download")
	}
	if downloadImageVerify == false {
		plog.Notice("Warning: image verification turned off")
	}

	versionFile := filepath.Join(downloadImageCacheDir, "version.txt")
	versionURL := strings.TrimRight(downloadImageRoot, "/") + "/" + "version.txt"
	if err := sdk.UpdateFile(versionFile, versionURL); err != nil {
		plog.Fatalf("downloading version.txt: %v", err)
	}

	for _, suffix := range downloadImagePlatformList {
		fileName := downloadImagePrefix + suffix
		filePath := filepath.Join(downloadImageCacheDir, fileName)

		// path.Join doesn't work with urls
		url := strings.TrimRight(downloadImageRoot, "/") + "/" + fileName

		download := func() error {
			if downloadImageVerify {
				plog.Noticef("Verifying and updating to latest image %v", fileName)
				err := sdk.UpdateSignedFile(filePath, url)
				if err != nil {
					plog.Noticef("updating signed file: %v", err)
					return err
				}
				return nil
			} else {
				plog.Noticef("Starting non-verified image update %v", fileName)
				if err := sdk.UpdateFile(filePath, url); err != nil {
					plog.Noticef("downloading image: %v", err)
					return err
				}
				return nil
			}
		}
		if err := util.Retry(downloadImageRetries+1, time.Second, download); err != nil {
			plog.Fatalf("updating image: %v", err)
		}
	}
}
