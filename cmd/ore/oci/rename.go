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

	"github.com/spf13/cobra"
)

var (
	cmdRename = &cobra.Command{
		Use:   "rename",
		Short: "Rename OCI image",
		Long:  "Rename OCI image in objectstorage",
		Run:   runRenameImage,
	}

	renameOldName string
	renameNewName string
	renameBucket  string
)

func init() {
	cmdRename.Flags().StringVar(&renameOldName, "old-name", "", "Current image name")
	cmdRename.Flags().StringVar(&renameNewName, "new-name", "", "New image name")
	cmdRename.Flags().StringVar(&renameBucket, "bucket", "", "OCI storage bucket name")
	OCI.AddCommand(cmdRename)
}

func runRenameImage(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		fmt.Fprintf(os.Stderr, "Unrecognized args in ore rename cmd: %v\n", args)
		os.Exit(2)
	}

	err := API.RenameImage(renameBucket, renameOldName, renameNewName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed renaming image: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Image %v successfully renamed in OCI\n", renameNewName)
}
