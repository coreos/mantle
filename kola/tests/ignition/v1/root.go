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

package ignition

import (
	"fmt"

	"github.com/coreos/mantle/kola/register"
	"github.com/coreos/mantle/platform"
)

func init() {
	// Reformat the root as btrfs
	register.Register(&register.Test{
		Name:        "coreos.ignition.btrfsroot",
		Run:         btrfsRoot,
		ClusterSize: 1,
		Platforms:   []string{"aws"},
		UserData: `{
		               "ignitionVersion": 1,
		               "storage": {
		                   "filesystems": [
		                       {
		                           "device": "/dev/disk/by-label/ROOT",
		                           "format": "btrfs",
		                           "create": {
		                               "force": true,
		                               "options": [
		                                   "--label=ROOT"
		                               ]
		                           }
		                       }
		                   ]
		               }
		           }`,
	})

	// Reformat the root as xfs
	register.Register(&register.Test{
		Name:        "coreos.ignition.xfsroot",
		Run:         xfsRoot,
		ClusterSize: 1,
		Platforms:   []string{"aws"},
		UserData: `{
		               "ignitionVersion": 1,
		               "storage": {
		                   "filesystems": [
		                       {
		                           "device": "/dev/disk/by-label/ROOT",
		                           "format": "xfs",
		                           "create": {
		                               "force": true,
		                               "options": [
		                                   "-L", "ROOT"
		                               ]
		                           }
		                       }
		                   ]
		               }
		           }`,
	})
}

func btrfsRoot(c platform.TestCluster) error {
	return testRoot(c, "btrfs")
}

func xfsRoot(c platform.TestCluster) error {
	return testRoot(c, "xfs")
}

func testRoot(c platform.TestCluster, fs string) error {
	m := c.Machines()[0]

	out, err := m.SSH("findmnt --noheadings --output FSTYPE --target /")
	if err != nil {
		return fmt.Errorf("failed to run findmnt: %s: %v", out, err)
	}

	if string(out) != fs {
		return fmt.Errorf("root wasn't correctly reformatted:\n%s", out)
	}

	return nil
}
