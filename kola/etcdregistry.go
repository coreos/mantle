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

package kola

import (
	"github.com/coreos/mantle/kola/tests/etcd"
	"github.com/coreos/mantle/platform"
)

func init() {
	// test etcd discovery with 0.4.7
	Register(&TestGroup{
		ClusterTests: []func(platform.Cluster) error{etcd.DiscoveryV1},
		ClusterSize:  3,
		Name:         "Etcd1Discovery",
		CloudConfig: `#cloud-config
coreos:
  etcd:
    name: $name
    discovery: $discovery
    addr: $public_ipv4:4001
    peer-addr: $private_ipv4:7001`,
	})

	// test etcd discovery with 2.0 with new cloud config
	Register(&TestGroup{
		ClusterTests: []func(platform.Cluster) error{etcd.DiscoveryV2},
		ClusterSize:  3,
		Name:         "Etcd2Discovery",
		CloudConfig: `#cloud-config

coreos:
  etcd2:
    name: $name
    discovery: $discovery
    advertise-client-urls: http://$public_ipv4:2379
    initial-advertise-peer-urls: http://$private_ipv4:2380
    listen-client-urls: http://0.0.0.0:2379,http://0.0.0.0:4001
    listen-peer-urls: http://$private_ipv4:2380,http://$private_ipv4:7001`,
	})
}
