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

package systemd

import (
	"fmt"
	"io"
	"time"

	"github.com/coreos/mantle/kola/register"
	"github.com/coreos/mantle/platform"
	"github.com/coreos/mantle/system/exec"
	"github.com/coreos/mantle/system/targen"
	"github.com/coreos/mantle/util"

	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/coreos-cloudinit/config"
)

var (
	// makes some assumptions about the qemu platform's network.
	// see platform/local/dnsmasq.go
	networkdstaticconf = config.CloudConfig{
		CoreOS: config.CoreOS{
			Units: []config.Unit{
				config.Unit{
					Name:    "00-eth.network",
					Runtime: true,
					Content: `[Match]
Name=eth0

[Network]
DNS=10.0.0.1
Address=10.0.0.2/24
Gateway=10.0.0.1
`,
				},
			},
		},
		Hostname: "gateway",
	}
)

func init() {
	register.Register(&register.Test{
		Run:         NetworkdStaticDNS,
		ClusterSize: 1,
		NativeFuncs: map[string]func() error{
			"dnscontainer": func() error {
				return genContainer("dns", []string{"dig"})
			},
		},
		Name:      "systemd.networkd.static.dns",
		UserData:  networkdstaticconf.String(),
		Platforms: []string{"qemu"},
	})
}

func genContainer(name string, binnames []string) error {
	tg := targen.New()

	for _, bin := range binnames {
		binpath, err := exec.LookPath(bin)
		if err != nil {
			return fmt.Errorf("failed to find %q binary: %v", bin, err)
		}

		tg.AddBinary(binpath)
	}

	pr, pw := io.Pipe()
	dimport := exec.Command("docker", "import", "-", name)
	dimport.Stdin = pr

	if err := dimport.Start(); err != nil {
		return fmt.Errorf("starting docker import failed %v", err)
	}

	if err := tg.Generate(pw); err != nil {
		return fmt.Errorf("failed to generate tarball: %v", err)
	}

	// err is always nil.
	_ = pw.Close()

	if err := dimport.Wait(); err != nil {
		return fmt.Errorf("waiting for docker import failed %v", err)
	}

	return nil
}

// NetworkdStaticDNS tests that dns works inside a container, when networkd is configured to use static ips.
//
// Created in response to https://github.com/coreos/bugs/issues/1140.
func NetworkdStaticDNS(c platform.TestCluster) error {
	m := c.Machines()[0]

	plog.Debug("creating dns container")

	if err := c.RunNative("dnscontainer", m); err != nil {
		return fmt.Errorf("failed to create dig container: %v", err)
	}

	plog.Debug("checking for networkd idleness")

	// wait for networkd to exit. we'll error out if it's still running after 120s.
	networkdIdleCheck := func() error {
		_, err := m.SSH("pidof systemd-networkd")
		if err == nil {
			return fmt.Errorf("systemd-networkd is still running")
		}
		return nil
	}

	if err := util.Retry(10, 12*time.Second, networkdIdleCheck); err != nil {
		return err
	}

	plog.Debug("running dns container")

	// if we've hit the bug in #1140 this won't work because /etc/resolv.conf is gone.
	// makes some assumptions about the qemu platform's dns.
	// see platform/local/dnsmasq.go
	out, err := m.SSH("docker run --rm dns dig +short br0.local")
	if err != nil {
		return fmt.Errorf("failed to run dig container: %s: %v", out, err)
	}

	plog.Debugf("dig says: %s", string(out))

	return nil
}
