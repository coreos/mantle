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
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/coreos/mantle/network"
)

// BaseCluster forms the basis of a cluster of CoreOS machines.
//
// It holds references to each Machine, and has ssh credentials for
// each machine. It partially implements Cluster.
type BaseCluster struct {
	agent *network.SSHAgent

	machlock sync.Mutex
	machmap  map[string]Machine

	name string
}

// NewBaseCluster creates a BaseCluster with the given name.
func NewBaseCluster(basename string) (*BaseCluster, error) {
	// set reasonable timeout and keepalive interval
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	agent, err := network.NewSSHAgent(dialer)
	if err != nil {
		return nil, err
	}

	bc := &BaseCluster{
		agent:   agent,
		machmap: make(map[string]Machine),
		name:    fmt.Sprintf("%s-%s", basename, uuid.NewV4()),
	}

	return bc, nil
}

// SSHClient creates an ssh session to a machine using the keys in the agent in the BaseCluster.
func (bc *BaseCluster) SSHClient(ip string) (*ssh.Client, error) {
	sshClient, err := bc.agent.NewClient(ip)
	if err != nil {
		return nil, err
	}

	return sshClient, nil
}

// PasswordSSHClient creates an ssh session to a machine using a password with the agent in the BaseCluster.
func (bc *BaseCluster) PasswordSSHClient(ip string, user string, password string) (*ssh.Client, error) {
	sshClient, err := bc.agent.NewPasswordClient(ip, user, password)
	if err != nil {
		return nil, err
	}

	return sshClient, nil
}

// SSH runs a command via ssh on a machine using the keys of agent in the BaseCluster.
func (bc *BaseCluster) SSH(m Machine, cmd string) ([]byte, error) {
	client, err := bc.SSHClient(m.IP())
	if err != nil {
		return nil, err
	}

	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}

	defer session.Close()

	session.Stderr = os.Stderr
	out, err := session.Output(cmd)
	out = bytes.TrimSpace(out)
	return out, err
}

// Machines returns a slice of machines in the cluster.
func (bc *BaseCluster) Machines() []Machine {
	bc.machlock.Lock()
	defer bc.machlock.Unlock()
	machs := make([]Machine, 0, len(bc.machmap))
	for _, m := range bc.machmap {
		machs = append(machs, m)
	}
	return machs
}

func (bc *BaseCluster) Keys() ([]*agent.Key, error) {
	return bc.agent.List()
}

// AddMach adds a machine to the cluster.
func (bc *BaseCluster) AddMach(m Machine) {
	bc.machlock.Lock()
	defer bc.machlock.Unlock()
	bc.machmap[m.ID()] = m
}

// DelMach removes a machine from the cluster.
func (bc *BaseCluster) DelMach(m Machine) {
	bc.machlock.Lock()
	defer bc.machlock.Unlock()
	delete(bc.machmap, m.ID())
}

// GetDiscoveryURL returns an etcd discovery URL for a cluster of a given size.
//
// XXX(mischief): i don't really think this belongs here, but it completes the
// interface we've established.
func (bc *BaseCluster) GetDiscoveryURL(size int) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://discovery.etcd.io/new?size=%d", size))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// Destroy destroys each machine in the cluster and closes the SSH agent.
func (bc *BaseCluster) Destroy() error {
	for _, m := range bc.Machines() {
		// TODO(mischief): should log failures here.
		m.Destroy()
	}

	bc.agent.Close()

	return nil
}

func (bc *BaseCluster) Name() string {
	return bc.name
}
