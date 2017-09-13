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

package misc

import (
	"encoding/json"

	"github.com/coreos/go-semver/semver"

	"github.com/coreos/mantle/kola/cluster"
	"github.com/coreos/mantle/kola/register"
	"github.com/coreos/mantle/platform"
	"github.com/coreos/mantle/platform/conf"
	"github.com/coreos/mantle/platform/machine/qemu"
)

var (
	raidRootUserData = conf.ContainerLinuxConfig(`storage:
  disks:
    - device: /dev/vdb
      partitions:
       - label: root1
         number: 1
         size: 256MiB
         type_guid: be9067b9-ea49-4f15-b4f6-f36f8c9e1818
       - label: root2
         number: 2
         size: 256MiB
         type_guid: be9067b9-ea49-4f15-b4f6-f36f8c9e1818
  raid:
    - name: "rootarray"
      level: "raid1"
      devices:
        - "/dev/vdb1"
        - "/dev/vdb2"
  filesystems:
    - name: "ROOT"
      mount:
        device: "/dev/md/rootarray"
        format: "ext4"
        create:
          options:
            - "-L"
            - "ROOT"
    - name: "NOT_ROOT"
      mount:
        device: "/dev/vda9"
        format: "ext4"
        create:
          options:
            - "-L"
            - "wasteland"
          force: true`)
	nestedRaidRootUserData = conf.ContainerLinuxConfig(`storage:
  disks:
    - device: /dev/vdb
      partitions:
       - label: root11
         number: 1
         size: 256MiB
         type_guid: be9067b9-ea49-4f15-b4f6-f36f8c9e1818
       - label: root12
         number: 2
         size: 256MiB
         type_guid: be9067b9-ea49-4f15-b4f6-f36f8c9e1818
       - label: root21
         number: 3
         size: 256MiB
         type_guid: be9067b9-ea49-4f15-b4f6-f36f8c9e1818
       - label: root22
         number: 4
         size: 256MiB
         type_guid: be9067b9-ea49-4f15-b4f6-f36f8c9e1818
    - device: /dev/md/inner1
      partitions:
       - label: inner_part1
         number: 1
         type_guid: be9067b9-ea49-4f15-b4f6-f36f8c9e1818
    - device: /dev/md/inner2
      partitions:
       - label: inner_part2
         number: 1
         type_guid: be9067b9-ea49-4f15-b4f6-f36f8c9e1818
    - device: /dev/md/outer
      partitions:
       - label: ROOT
         number: 1
  raid:
    - name: "inner1"
      level: "raid1"
      devices:
        - "/dev/vdb1"
        - "/dev/vdb2"
    - name: "inner2"
      level: "raid1"
      devices:
        - "/dev/vdb3"
        - "/dev/vdb4"
    - name: "outer"
      level: "raid1"
      devices:
        - "/dev/md/inner1p1"
        - "/dev/md/inner2p1"
  filesystems:
    - name: "ROOT"
      mount:
        device: "/dev/md/outer1"
        format: "ext4"
        create:
          options:
            - "-L"
            - "ROOT"
    - name: "NOT_ROOT"
      mount:
        device: "/dev/vda9"
        format: "ext4"
        create:
          options:
            - "-L"
            - "wasteland"
          force: true`)
)

func init() {
	register.Register(&register.Test{
		// This test needs additional disks which is only supported on qemu since Ignition
		// does not support deleting partitions without wiping the partition table and the
		// disk doesn't have room for new partitions.
		// TODO(ajeddeloh): change this to delete partition 9 and replace it with 9 and 10
		// once Ignition supports it.
		Run:         RootOnRaid,
		ClusterSize: 0,
		Platforms:   []string{"qemu"},
		MinVersion:  semver.Version{Major: 1520},
		Name:        "coreos.disk.raid.root",
	})
	register.Register(&register.Test{
		// This test needs additional disks which is only supported on qemu since Ignition
		// does not support deleting partitions without wiping the partition table and the
		// disk doesn't have room for new partitions.
		// TODO(ajeddeloh): change this to delete partition 9 and replace it with 9 and 10
		// once Ignition supports it.
		Run:         NestedRootOnRaid,
		ClusterSize: 0,
		Platforms:   []string{"qemu"},
		MinVersion:  semver.Version{Major: 1535},
		Name:        "coreos.disk.raid.nestedroot",
	})
	register.Register(&register.Test{
		Run:         DataOnRaid,
		ClusterSize: 1,
		Name:        "coreos.disk.raid.data",
		UserData: conf.ContainerLinuxConfig(`storage:
  raid:
    - name: "DATA"
      level: "raid1"
      devices:
        - "/dev/disk/by-partlabel/OEM-CONFIG"
        - "/dev/disk/by-partlabel/USR-B"
  filesystems:
    - name: "DATA"
      mount:
        device: "/dev/md/DATA"
        format: "ext4"
        create:
          options:
            - "-L"
            - "DATA"
systemd:
  units:
    - name: "var-lib-data.mount"
      enable: true
      contents: |
          [Mount]
          What=/dev/md/DATA
          Where=/var/lib/data
          Type=ext4
          
          [Install]
          WantedBy=local-fs.target`),
	})
}

func RootOnRaid(c cluster.TestCluster) {
	options := qemu.MachineOptions{
		AdditionalDisks: []qemu.Disk{
			{Size: "520M"},
		},
	}
	m, err := c.Cluster.(*qemu.Cluster).NewMachineWithOptions(raidRootUserData, options)
	if err != nil {
		c.Fatal(err)
	}

	checkIfMountpointIsType(c, m, "/", "raid1")

	// reboot it to make sure it comes up again
	err = m.Reboot()
	if err != nil {
		c.Fatalf("could not reboot machine: %v", err)
	}

	checkIfMountpointIsType(c, m, "/", "raid1")
}

func NestedRootOnRaid(c cluster.TestCluster) {
	options := qemu.MachineOptions{
		AdditionalDisks: []qemu.Disk{
			{Size: "1100M"},
		},
	}
	m, err := c.Cluster.(*qemu.Cluster).NewMachineWithOptions(nestedRaidRootUserData, options)
	if err != nil {
		c.Fatal(err)
	}

	checkIfMountpointIsType(c, m, "/", "md")

	// reboot it to make sure it comes up again
	err = m.Reboot()
	if err != nil {
		c.Fatalf("could not reboot machine: %v", err)
	}

	checkIfMountpointIsType(c, m, "/", "md")
}

func DataOnRaid(c cluster.TestCluster) {
	m := c.Machines()[0]

	checkIfMountpointIsType(c, m, "/var/lib/data", "raid1")

	// reboot it to make sure it comes up again
	err := m.Reboot()
	if err != nil {
		c.Fatalf("could not reboot machine: %v", err)
	}

	checkIfMountpointIsType(c, m, "/var/lib/data", "raid1")
}

type lsblkOutput struct {
	Blockdevices []blockdevice `json:"blockdevices"`
}

type blockdevice struct {
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	Mountpoint *string       `json:"mountpoint"`
	Children   []blockdevice `json:"children"`
}

// checkIfMountpointIsRaid will check if a given machine has a device of type
// raid1 mounted at the given mountpoint. If it does not, the test is failed.
func checkIfMountpointIsType(c cluster.TestCluster, m platform.Machine, mountpoint, mount_type string) {
	output, err := m.SSH("lsblk --json")
	if err != nil {
		c.Fatalf("couldn't list block devices: %v", err)
	}

	l := lsblkOutput{}
	err = json.Unmarshal(output, &l)
	if err != nil {
		c.Fatalf("couldn't unmarshal lsblk output: %v", err)
	}

	foundRoot := checkIfMountpointIsTypeWalker(c, l.Blockdevices, mountpoint, mount_type)
	if !foundRoot {
		c.Fatalf("didn't find root mountpoint in lsblk output")
	}
}

// checkIfMountpointIsRaidWalker will iterate over bs and recurse into its
// children, looking for a device mounted at / with type raid1. true is returned
// if such a device is found. The test is failed if a device of a different type
// is found to be mounted at /.
func checkIfMountpointIsTypeWalker(c cluster.TestCluster, bs []blockdevice, mountpoint, mount_type string) bool {
	for _, b := range bs {
		if b.Mountpoint != nil && *b.Mountpoint == mountpoint {
			if b.Type != mount_type {
				c.Fatalf("device %q is mounted at %q with type %q (was expecting raid1)", b.Name, mountpoint, b.Type)
			}
			return true
		}
		foundRoot := checkIfMountpointIsTypeWalker(c, b.Children, mountpoint, mount_type)
		if foundRoot {
			return true
		}
	}
	return false
}
