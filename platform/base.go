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

	"github.com/coreos/mantle/network"

	"github.com/coreos/mantle/Godeps/_workspace/src/golang.org/x/crypto/ssh"
)

type baseCluster struct {
	agent *network.SSHAgent

	machlock sync.Mutex
	machs    []Machine
}

func newBaseCluster() (*baseCluster, error) {
	// set reasonable timeout and keepalive interval
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	agent, err := network.NewSSHAgent(dialer)
	if err != nil {
		return nil, err
	}

	bc := &baseCluster{
		agent: agent,
		machs: make([]Machine, 0),
	}

	return bc, nil
}

func (bc *baseCluster) SSHClient(ip string) (*ssh.Client, error) {
	sshClient, err := bc.agent.NewClient(ip)
	if err != nil {
		return nil, err
	}

	return sshClient, nil
}

func (bc *baseCluster) SSH(m Machine, cmd string) ([]byte, error) {
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

func (bc *baseCluster) Machines() []Machine {
	bc.machlock.Lock()
	defer bc.machlock.Unlock()

	// copy to new underlying array for safety
	dup := make([]Machine, len(bc.machs))
	copy(dup, bc.machs)

	return dup
}

func (bc *baseCluster) addMach(m Machine) {
	bc.machlock.Lock()
	defer bc.machlock.Unlock()
	bc.machs = append(bc.machs, m)
}

func (bc *baseCluster) delMach(m Machine) {
	bc.machlock.Lock()
	defer bc.machlock.Unlock()

	// find machine in slice
	idx := -1
	for i := range bc.machs {
		if bc.machs[i].ID() == m.ID() {
			idx = i
			break
		}
	}
	if idx == -1 {
		panic("baseCluster does not contain machine to delete")
	}

	// delete from slice
	bc.machs = append(bc.machs[:idx], bc.machs[idx+1:]...)
}

// XXX(mischief): i don't really think this belongs here, but it completes the
// interface we've established.
func (bc *baseCluster) GetDiscoveryURL(size int) (string, error) {
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
