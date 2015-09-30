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

package fleet

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/coreos/mantle/platform"
	"github.com/coreos/mantle/util"

	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/pkg/capnslog"
)

var (
	plog = capnslog.NewPackageLogger("github.com/coreos/mantle", "kola/tests/fleet")

	masterconf = config.CloudConfig{
		CoreOS: config.CoreOS{
			Etcd2: config.Etcd2{
				AdvertiseClientURLs:      "http://$private_ipv4:2379",
				InitialAdvertisePeerURLs: "http://$private_ipv4:2380",
				ListenClientURLs:         "http://0.0.0.0:2379,http://0.0.0.0:4001",
				ListenPeerURLs:           "http://$private_ipv4:2380,http://$private_ipv4:7001",
			},
			Units: []config.Unit{
				config.Unit{
					Name:    "etcd2.service",
					Command: "start",
				},
				config.Unit{
					Name:    "fleet.service",
					Command: "start",
				},
			},
		},
		Hostname: "master",
	}

	proxyconf = config.CloudConfig{
		CoreOS: config.CoreOS{
			Etcd2: config.Etcd2{
				Proxy:            "on",
				ListenClientURLs: "http://0.0.0.0:2379,http://0.0.0.0:4001",
			},
			Units: []config.Unit{
				config.Unit{
					Name:    "etcd2.service",
					Command: "start",
				},
				config.Unit{
					Name:    "fleet.service",
					Command: "start",
				},
			},
		},
		Hostname: "proxy",
	}

	fleetunit = `
[Unit]
Description=simple fleet test
[Service]
ExecStart=/bin/sh -c "while sleep 1; do echo hello world; done"
`
)

func copyFile(mach platform.Machine, r io.Reader, path string) error {
	sess, err := mach.SSHSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	wr, err := sess.StdinPipe()
	if err != nil {
		return err
	}

	if err := sess.Start(fmt.Sprintf("tee %s", path)); err != nil {
		return err
	}

	go func() {
		io.Copy(wr, r)
		wr.Close()
	}()

	return sess.Wait()
}

// Test fleet running through an etcd2 proxy.
func Proxy(c platform.TestCluster) error {
	masterconf.CoreOS.Etcd2.Discovery, _ = c.GetDiscoveryURL(1)
	master, err := c.NewMachine(masterconf.String())
	if err != nil {
		return fmt.Errorf("Cluster.NewMachine: %s", err)
	}
	defer master.Destroy()

	proxyconf.CoreOS.Etcd2.Discovery = masterconf.CoreOS.Etcd2.Discovery
	proxy, err := c.NewMachine(proxyconf.String())
	if err != nil {
		return fmt.Errorf("Cluster.NewMachine: %s", err)
	}
	defer proxy.Destroy()

	// settling...
	time.Sleep(5 * time.Second)

	err = copyFile(proxy, strings.NewReader(fleetunit), "/home/core/hello.service")
	if err != nil {
		return fmt.Errorf("copyFile: %s", err)
	}

	var retryFuncs []func() error

	retryFuncs = append(retryFuncs, func() error {
		_, err = proxy.SSH("fleetctl start /home/core/hello.service")
		if err != nil {
			return fmt.Errorf("fleetctl start: %s", err)
		}
		return nil
	})

	var status []byte
	retryFuncs = append(retryFuncs, func() error {
		status, err = proxy.SSH("fleetctl list-units -l -fields active -no-legend")
		if err != nil {
			return fmt.Errorf("fleetctl list-units: %s", err)
		}

		if !bytes.Equal(status, []byte("active")) {
			return fmt.Errorf("unit not active")
		}

		return nil
	})

	for _, retry := range retryFuncs {
		if err := util.Retry(5, 2*time.Second, retry); err != nil {
			return err
		}
	}

	return nil
}
