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

	"github.com/coreos/mantle/kola/register"
	"github.com/coreos/mantle/platform"

	"github.com/coreos/go-semver/semver"
)

func init() {
	// regression test for https://github.com/coreos/bugs/issues/1092
	register.Register(&register.Test{
		Run:         cephDemo,
		ClusterSize: 1,
		Name:        "linux.cephdemo",
		Platforms:   []string{"aws", "gce"},
		UserData:    `#cloud-config`,
		// ceph xattrs were fixed in kernel 4.6.0
		MinVersion: semver.Version{Major: 1053},
	})
}

func cephDemo(c platform.TestCluster) error {
	m := c.Machines()[0]

	for _, cmd := range []struct {
		cmdline     string
		checkoutput bool
		output      string
	}{
		{"docker run -d --net=host -v /etc/ceph:/etc/ceph -e MON_IP=127.0.0.1 -e CEPH_NETWORK=127.0.0.1/8 ceph/demo", false, ""},
		{`SECRET=$(docker run --rm -v /etc/ceph:/etc/ceph --entrypoint=/usr/bin/ceph-authtool ceph/demo -p /etc/ceph/ceph.client.admin.keyring)
sudo mount -t ceph -o rw,relatime,name=admin,secret=$SECRET,fsc 127.0.0.1:/ /mnt`, false, ""},
		{"sudo chown :core /mnt", false, ""},
		{"sudo chmod g+w /mnt", false, ""},
		{"git -C /mnt clone https://anongit.gentoo.org/git/repo/gentoo.git", false, ""},
	} {
		output, err := m.SSH(cmd.cmdline)
		if err != nil {
			return fmt.Errorf("failed to run %q: output: %q status: %q", cmd.cmdline, output, err)
		}

		if cmd.checkoutput && string(output) != cmd.output {
			return fmt.Errorf("command %q has unexpected output: want %q got %q", cmd.cmdline, cmd.output, string(output))
		}
	}

	return nil
}
