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
)

func init() {
	register.Register(&register.Test{
		Run:         SelinuxEnforce,
		ClusterSize: 1,
		Name:        "coreos.selinux.enforce",
		UserData:    `#cloud-config`,
	})
}

// SelinuxEnforce checks that some basic things work after setting SELINUX=enforcing.
// updated to check for polkit failing in https://github.com/coreos/bugs/issues/1258.
func SelinuxEnforce(c platform.TestCluster) error {
	m := c.Machines()[0]

	// configure selinux enforcing mode
	for _, cmd := range []string{
		"sudo cp --remove-destination $(readlink -f /etc/selinux/config) /etc/selinux/config",
		"sudo sed --in-place --expression 's/SELINUX=permissive/SELINUX=enforcing/' /etc/selinux/config",
	} {
		output, err := m.SSH(cmd)
		if err != nil {
			return fmt.Errorf("failed to run %q: output: %q status: %q", cmd, output, err)
		}
	}

	// reboot so we've got selinux from the start
	err := platform.Reboot(m)
	if err != nil {
		return fmt.Errorf("failed to reboot machine: %v", err)
	}

	// check selinux state
	for _, cmd := range []struct {
		cmdline     string
		checkoutput bool
		output      string
	}{
		{"getenforce", true, "Enforcing"},
		{"systemctl --no-pager is-active system.slice", true, "active"},
		// check that polkit works, coreos/bugs#1258
		{"sudo systemctl restart polkit", false, ""},
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
