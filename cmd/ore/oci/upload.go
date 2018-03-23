// Copyright 2017 CoreOS, Inc.
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

package oci

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/coreos/mantle/sdk"
	"github.com/spf13/cobra"
)

var (
	cmdUpload = &cobra.Command{
		Use:   "upload",
		Short: "Upload OCI image",
		Long:  "Upload OCI image from a local file",
		Run:   runUploadImage,
	}

	uploadImageName   string
	uploadImageBucket string
	uploadImageFile   string
	uploadImageForce  bool
)

func init() {
	build := sdk.BuildRoot()
	cmdUpload.Flags().StringVar(&uploadImageName, "name", "", "Image name")
	cmdUpload.Flags().StringVar(&uploadImageBucket, "bucket", "image-upload", "OCI storage bucket name")
	cmdUpload.Flags().StringVar(&uploadImageFile, "file",
		build+"/images/amd64-usr/latest/coreos_production_oracle_oci_qcow_image.img",
		"Image file")
	cmdUpload.Flags().BoolVar(&uploadImageForce, "force", false, "overwrite existing s3 object without prompt")
	OCI.AddCommand(cmdUpload)
}

func runUploadImage(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		fmt.Fprintf(os.Stderr, "Unrecognized args in ore upload cmd: %v\n", args)
		os.Exit(2)
	}

	if uploadImageName == "" {
		ver, err := sdk.VersionsFromDir(filepath.Dir(uploadImageFile))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to get version from image directory, provide a -name flag or include a version.txt in the image directory: %v\n", err)
			os.Exit(1)
		}
		uploadImageName = ver.Version
	}

	f, err := os.Open(uploadImageFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open image file %v: %v\n", uploadImageFile, err)
		os.Exit(1)
	}
	defer f.Close()

	err = API.UploadObject(f, uploadImageBucket, uploadImageName, uploadImageForce)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed uploading image: %v\n", err)
		os.Exit(1)
	}

	img, err = API.CreateImage(uploadImageBucket, uploadImageName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "creating image: %v\n", err)
		os.Exit(1)
	}

	err = API.DeleteObject(uploadImageBucket, uploadImageName)
	if err != nil {
		// only warn on delete failures
		fmt.Fprintf(os.Stderr, "deleting object: %v\n", err)
	}

	fmt.Println(img)
}
