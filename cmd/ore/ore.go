// Copyright 2014 CoreOS, Inc.
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
	"github.com/spf13/cobra"

	"github.com/coreos/mantle/cli"
	"github.com/coreos/mantle/platform"
	"github.com/coreos/mantle/platform/api/gce"

	"github.com/coreos/pkg/capnslog"
)

var (
	root = &cobra.Command{
		Use:   "ore [command]",
		Short: "gce image creation and upload tools",
	}

	opts = gce.Options{Options: &platform.Options{}}

	plog = capnslog.NewPackageLogger("github.com/coreos/mantle", "cmd/ore")

	api *gce.API
)

func preauth(cmd *cobra.Command, args []string) {
	a, err := gce.New(&opts)
	if err != nil {
		plog.Fatalf("Unable to create api client: %v", err)
	}

	api = a
}

func main() {
	sv := root.PersistentFlags().StringVar

	sv(&opts.Image, "image", "", "image name")
	sv(&opts.Project, "project", "coreos-gce-testing", "project")
	sv(&opts.Zone, "zone", "us-central1-a", "zone")
	sv(&opts.MachineType, "machinetype", "n1-standard-1", "machine type")
	sv(&opts.DiskType, "disktype", "pd-ssd", "disk type")
	sv(&opts.BaseName, "basename", "kola", "instance name prefix")
	sv(&opts.Network, "network", "default", "network name")
	root.PersistentFlags().BoolVar(&opts.ServiceAuth, "service-auth", false, "use non-interactive auth when running within GCE")

	cli.Execute(root)
}
