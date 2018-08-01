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
	"github.com/coreos/mantle/kola/cluster"
	"github.com/coreos/mantle/kola/register"
	"strings"
)

func init() {
	register.Register(&register.Test{
		Run:         SelinuxLogCheck,
		ClusterSize: 1,
		Name:        "coreos.selinux.logcheck",
		Flags:       []register.Flag{register.NoEnableSelinux},
	})
	register.Register(&register.Test{
		Run:         SelinuxEnforce,
		ClusterSize: 1,
		Name:        "coreos.selinux.enforce",
		Flags:       []register.Flag{register.NoEnableSelinux},
		Distros:     []string{"cl", "rhcos"},
	})
}

// SelinuxLogCheck checks that no audit AVC messages appear in boot logs.
func SelinuxLogCheck(c cluster.TestCluster) {
	m := c.Machines()[0]

	cmd := "sudo journalctl -b --no-pager | egrep 'AVC avc'"
	stdout, stderr, err := m.SSH(cmd)

	if err == nil {
		c.Fatalf("Found audit AVC messages in boot logs: \n%v", string(stdout))
	}

	if err.Error() == "Process exited with status 1" &&
		strings.TrimSpace(string(stderr)) == "" {
		return // OK, nothing found.
	}

	c.Fatalf("cmd '%v' failed: %v: %v.\n", string(cmd), err, string(stderr))
}

// SelinuxEnforce checks that some basic things work after `setenforce 1`
func SelinuxEnforce(c cluster.TestCluster) {
	m := c.Machines()[0]

	for _, cmd := range []struct {
		cmdline     string
		checkoutput bool
		output      string
	}{
		{"getenforce", true, "Permissive"},
		{"sudo setenforce 1", false, ""},
		{"getenforce", true, "Enforcing"},
		{"systemctl --no-pager is-active system.slice", true, "active"},
		{"sudo cp --remove-destination $(readlink -f /etc/selinux/config) /etc/selinux/config", false, ""},
		{"sudo sed -i 's/SELINUX=permissive/SELINUX=enforcing/' /etc/selinux/config", false, ""},
	} {
		output := c.MustSSH(m, cmd.cmdline)

		if cmd.checkoutput && string(output) != cmd.output {
			c.Fatalf("command %q has unexpected output: want %q got %q", cmd.cmdline, cmd.output, string(output))
		}
	}

	err := m.Reboot()
	if err != nil {
		c.Fatalf("failed to reboot machine: %v", err)
	}

	o := c.MustSSH(m, "which getenforce")
	c.Logf(string(o))

	output := c.MustSSH(m, "getenforce")

	if string(output) != "Enforcing" {
		c.Fatalf("command \"getenforce\" has unexpected output: want \"Enforcing\" got %q", string(output))
	}
}
