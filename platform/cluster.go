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

package platform

import (
	"golang.org/x/crypto/ssh/agent"
)

// Cluster represents a cluster of CoreOS machines within a single platform.
type Cluster interface {
	// NewMachine creates a new CoreOS machine.
	NewMachine(config string) (Machine, error)

	// Machines returns a slice of the active machines in the Cluster.
	Machines() []Machine

	// Keys returns the SSH public keys used by the cluster to authenciate with machines.
	Keys() ([]*agent.Key, error)

	// GetDiscoveryURL returns a new etcd discovery URL.
	GetDiscoveryURL(size int) (string, error)

	// Destroy terminates each machine in the cluster and frees any other
	// associated resources.
	Destroy() error

	// Name returns a identifying name for the cluster.
	Name() string
}
