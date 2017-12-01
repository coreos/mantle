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

package misc

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/coreos/mantle/kola/cluster"
	"github.com/coreos/mantle/kola/register"
	"github.com/coreos/mantle/platform"
	"github.com/coreos/mantle/platform/conf"
)

func init() {
	register.Register(&register.Test{
		Run:         Filesystem,
		ClusterSize: 1,
		Name:        "coreos.filesystem",
	})
}

func Filesystem(c cluster.TestCluster) {
	c.Run("deadlinks", DeadLinks)
	c.Run("suid", SUIDFiles)
	c.Run("sgid", SGIDFiles)
	c.Run("writablefiles", WritableFiles)
	c.Run("writabledirs", WritableDirs)
	c.Run("stickydirs", StickyDirs)
	c.Run("blacklist", Blacklist)
	c.Run("wiperoot", WipeRoot)
}

func sugidFiles(c cluster.TestCluster, validfiles []string, mode string) {
	m := c.Machines()[0]
	badfiles := make([]string, 0, 0)

	output := c.MustSSH(m, fmt.Sprintf("sudo find / -ignore_readdir_race -path /sys -prune -o -path /proc -prune -o -path /var/lib/rkt -prune -o -type f -perm -%v -print", mode))

	if string(output) == "" {
		return
	}

	files := strings.Split(string(output), "\n")
	for _, file := range files {
		var valid bool

		for _, validfile := range validfiles {
			if file == validfile {
				valid = true
			}
		}
		if valid != true {
			badfiles = append(badfiles, file)
		}
	}

	if len(badfiles) != 0 {
		c.Fatalf("Unknown SUID or SGID files found: %v", badfiles)
	}
}

func DeadLinks(c cluster.TestCluster) {
	m := c.Machines()[0]

	ignore := []string{
		"/dev",
		"/proc",
		"/run/udev/watch",
		"/sys",
		"/var/lib/docker",
		"/var/lib/rkt",
	}

	output := c.MustSSH(m, fmt.Sprintf("sudo find / -ignore_readdir_race -path %s -prune -o -xtype l -print", strings.Join(ignore, " -prune -o -path ")))

	if string(output) != "" {
		c.Fatalf("Dead symbolic links found: %v", strings.Split(string(output), "\n"))
	}
}

func SUIDFiles(c cluster.TestCluster) {
	validfiles := []string{
		"/usr/bin/chage",
		"/usr/bin/chfn",
		"/usr/bin/chsh",
		"/usr/bin/expiry",
		"/usr/bin/gpasswd",
		"/usr/bin/ksu",
		"/usr/bin/man",
		"/usr/bin/mandb",
		"/usr/bin/mount",
		"/usr/bin/newgidmap",
		"/usr/bin/newgrp",
		"/usr/bin/newuidmap",
		"/usr/bin/passwd",
		"/usr/bin/pkexec",
		"/usr/bin/umount",
		"/usr/bin/su",
		"/usr/bin/sudo",
		"/usr/lib/polkit-1/polkit-agent-helper-1",
		"/usr/lib64/polkit-1/polkit-agent-helper-1",
		"/usr/libexec/dbus-daemon-launch-helper",
		"/usr/sbin/mount.nfs",
		"/usr/sbin/unix_chkpwd",
	}

	sugidFiles(c, validfiles, "4000")
}

func SGIDFiles(c cluster.TestCluster) {
	validfiles := []string{}

	sugidFiles(c, validfiles, "2000")
}

func WritableFiles(c cluster.TestCluster) {
	m := c.Machines()[0]

	output := c.MustSSH(m, "sudo find / -ignore_readdir_race -path /sys -prune -o -path /proc -prune -o -path /var/lib/rkt -prune -o -type f -perm -0002 -print")

	if string(output) != "" {
		c.Fatalf("Unknown writable files found: %s", output)
	}
}

func WritableDirs(c cluster.TestCluster) {
	m := c.Machines()[0]

	output := c.MustSSH(m, "sudo find / -ignore_readdir_race -path /sys -prune -o -path /proc -prune -o -path /var/lib/rkt -prune -o -type d -perm -0002 -a ! -perm -1000 -print")

	if string(output) != "" {
		c.Fatalf("Unknown writable directories found: %s", output)
	}
}

// The default permissions for the root of a tmpfs are 1777
// https://github.com/coreos/bugs/issues/1812
func StickyDirs(c cluster.TestCluster) {
	m := c.Machines()[0]

	ignore := []string{
		// don't descend into these
		"/proc",
		"/sys",
		"/var/lib/docker",
		"/var/lib/rkt",

		// should be sticky, and may have sticky children
		"/dev/mqueue",
		"/dev/shm",
		"/media",
		"/tmp",
		"/var/tmp",
	}

	output := c.MustSSH(m, fmt.Sprintf("sudo find / -ignore_readdir_race -path %s -prune -o -type d -perm /1000 -print", strings.Join(ignore, " -prune -o -path ")))

	if string(output) != "" {
		c.Fatalf("Unknown sticky directories found: %s", output)
	}
}

func Blacklist(c cluster.TestCluster) {
	m := c.Machines()[0]

	skip := []string{
		// Directories not to descend into
		"/proc",
		"/sys",
		"/var/lib/docker",
		"/var/lib/rkt",
	}

	blacklist := []string{
		// Things excluded from the image that might slip in
		"/usr/bin/perl",
		"/usr/bin/python",
		"/usr/share/man",

		// net-tools "make install" copies binaries from
		// /usr/bin/{} to /usr/bin/{}.old before overwriting them.
		// This sometimes produced an extraneous set of {}.old
		// binaries due to make parallelism.
		// https://github.com/coreos/coreos-overlay/pull/2734
		"/usr/bin/*.old",

		// Control characters in filenames
		"*[\x01-\x1f]*",
		// Space
		"* *",
		// DEL
		"*\x7f*",
	}

	output := c.MustSSH(m, fmt.Sprintf("sudo find / -ignore_readdir_race -path %s -prune -o -path '%s' -print", strings.Join(skip, " -prune -o -path "), strings.Join(blacklist, "' -print -o -path '")))

	if string(output) != "" {
		c.Fatalf("Blacklisted files or directories found:\n%s", output)
	}
}

func getAllFiles(c cluster.TestCluster, m platform.Machine) map[string]string {
	out := string(c.MustSSH(m, "sudo find / -ignore_readdir_race -mount -path /var -prune -o -type f -print0 | xargs -0 sudo md5sum"))
	ret := map[string]string{}
	for _, line := range strings.Split(out, "\n") {
		line := strings.TrimSpace(line)
		key := strings.Split(line,"  ")[1]
		value := strings.Split(line, " ")[0]
		ret[key] = value
	}
	return ret
}

// WipeRoot tests Ignition wiping the root fs actually performs a "factory reset"
func WipeRoot(c cluster.TestCluster) {
	wipeConf := conf.ContainerLinuxConfig(`storage:
  filesystems:
    - mount:
        device: /dev/disk/by-label/ROOT
        format: ext4
        wipe_filesystem: true
        label: ROOT
`)

	wipedMachine, err := c.NewMachine(wipeConf)
	if err != nil {
		c.Fatalf("Cluster.NewMachine: %s", err)
	}

	defer wipedMachine.Destroy()

	ignore := []string {
		// ignore generated host keys
		"/etc/ssh/ssh_host_dsa_key",
		"/etc/ssh/ssh_host_ecdsa_key",
		"/etc/ssh/ssh_host_ed25519_key",
		"/etc/ssh/ssh_host_rsa_key",
		"/etc/ssh/ssh_host_dsa_key.pub",
		"/etc/ssh/ssh_host_ecdsa_key.pub",
		"/etc/ssh/ssh_host_ed25519_key.pub",
		"/etc/ssh/ssh_host_rsa_key.pub",
		// machine-id *should* be different
		"/etc/machine-id",
	}
	ignoreDiff := func(in string) bool {
		for _, ok := range ignore {
			if ok == in {
				return true
			}
		}
		return false
	}

	unordered := []string{
		"/etc/passwd",
		"/etc/passwd-",
		"/etc/shadow",
		"/etc/shadow-",
		"/etc/group",
		"/etc/group-",
		"/etc/gpasswd",
		"/etc/gpasswd-",
		"/etc/gshadow",
		"/etc/gshadow-",
	}

	mapLines := func(in []byte) map[string]struct{} {
		ret := map[string]struct{}{}
		for _, line := range strings.Split(string(in), "\n") {
			ret[line] = struct{}{}
		}
		return ret
	}

	// unorderedEqual returns true if the file k has the same lines on both m1 and m2 regardless of
	// order.
	unorderedEqual := func(c cluster.TestCluster, m1, m2 platform.Machine, k string) bool {
		valid := false
		for _, ok := range unordered {
			if k == ok {
				valid = true
				break;
			}
		}
		if !valid {
			return false
		}
		cmd := fmt.Sprintf("sudo cat %v", k)
		m1contents := mapLines(c.MustSSH(m1, cmd))
		m2contents := mapLines(c.MustSSH(m2, cmd))
		return reflect.DeepEqual(m1contents, m2contents)
	}



	notWiped := getAllFiles(c, c.Machines()[0])
	wiped := getAllFiles(c, wipedMachine)

	var diffs, notInWiped, notInUnwiped []string

	for k, v := range notWiped {
		if v2, ok := wiped[k]; ok {
			if v != v2 && !ignoreDiff(k) && !unorderedEqual(c, c.Machines()[0], wipedMachine, k) {
				diffs = append(diffs, k)
			}
		} else {
			notInWiped = append(notInWiped, k)
		}
	}
	for k, _ := range wiped {
		if _, ok := notWiped[k]; !ok {
			notInUnwiped = append(notInUnwiped, k)
		}
	}

	failed := false;

	if len(diffs) != 0 {
		failed = true
		c.Logf("Files that differ: ")
		for _, file := range diffs {
			c.Logf("%v", file)
		}
	}

	if len(notInWiped) != 0 {
		failed = true
		c.Logf("Files absent after wipe: ")
		for _, file := range notInWiped {
			c.Logf("%v", file)
		}
	}

	if len(notInUnwiped) != 0 {
		failed = true
		c.Logf("New files after wipe: ")
		for _, file := range notInUnwiped {
			c.Logf("%v", file)
		}
	}

	if failed {
		c.Fatalf("Root partition differed after wipe")
	}
}
