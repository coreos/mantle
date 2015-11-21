// Copyright 2014 CoreOS, Inc.
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
	"io"
	"os"
	"time"

	"path/filepath"
	"sync"

	"github.com/coreos/mantle/Godeps/_workspace/src/golang.org/x/crypto/ssh"
	"github.com/coreos/mantle/util"
)

const (
	sshRetries = 10
	sshTimeout = 2 * time.Second
)

// Machine represents a CoreOS instance.
type Machine interface {
	// ID returns the plaform-specific machine identifier.
	ID() string

	// IP returns the machine's public IP.
	IP() string

	// PrivateIP returns the machine's private IP.
	PrivateIP() string

	// SSHClient establishes a new SSH connection to the machine.
	SSHClient() (*ssh.Client, error)

	// SSH runs a single command over a new SSH connection.
	SSH(cmd string) ([]byte, error)

	// Destroy terminates the machine and frees associated resources.
	Destroy() error
}

// Cluster represents a cluster of CoreOS machines within a single platform.
type Cluster interface {
	// NewMachine creates a new CoreOS machine.
	NewMachine(config string) (Machine, error)

	// Machines returns a slice of the active machines in the Cluster.
	Machines() []Machine

	// GetDiscoveryURL returns a new etcd discovery URL.
	GetDiscoveryURL(size int) (string, error)

	// Destroy terminates each machine in the cluster and frees any other
	// associated resources.
	Destroy() error
}

// TestCluster embedds a Cluster to provide platform independant helper
// methods.
type TestCluster struct {
	Name        string
	NativeFuncs []string
	Options     map[string]string
	Cluster
}

// RunNative runs a registered NativeFunc on a remote machine
func (t *TestCluster) RunNative(funcName string, m Machine) error {
	// scp and execute kolet on remote machine
	client, err := m.SSHClient()
	if err != nil {
		return fmt.Errorf("kolet SSH client: %v", err)
	}

	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("kolet SSH session: %v", err)
	}

	defer session.Close()

	b, err := session.CombinedOutput(fmt.Sprintf("./kolet run %q %q", t.Name, funcName))
	if err != nil {
		return fmt.Errorf("%s", b) // return function std output, not the exit status
	}
	return nil
}

// ListNativeFunctions returns a slice of function names that can be executed
// directly on machines in the cluster.
func (t *TestCluster) ListNativeFunctions() []string {
	return t.NativeFuncs
}

// DropFile places file from localPath to ~/ on every machine in cluster
func (t *TestCluster) DropFile(localPath string) error {
	in, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer in.Close()

	for _, m := range t.Machines() {
		if _, err := in.Seek(0, 0); err != nil {
			return err
		}
		if err := InstallFile(in, m, filepath.Base(localPath)); err != nil {
			return err
		}
	}
	return nil
}

// InstallFile copies data from in to the path to on m.
func InstallFile(in io.Reader, m Machine, to string) error {
	dir := filepath.Dir(to)
	out, err := m.SSH(fmt.Sprintf("sudo mkdir -p %s", dir))
	if err != nil {
		return fmt.Errorf("failed creating directory %s: %s", dir, out)
	}

	client, err := m.SSHClient()
	if err != nil {
		return fmt.Errorf("failed creating SSH client: %v", err)
	}

	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed creating SSH session: %v", err)
	}

	defer session.Close()

	// write file to fs from stdin
	session.Stdin = in
	err = session.Run(fmt.Sprintf("install -m 0755 /dev/stdin %s", to))
	if err != nil {
		return fmt.Errorf("failed executing install: %v", err)
	}

	return nil
}

// NewMachines spawns len(userdatas) instances in cluster c, with
// each instance passed the respective userdata.
func NewMachines(c Cluster, userdatas []string) ([]Machine, error) {
	var wg sync.WaitGroup

	n := len(userdatas)

	if n <= 0 {
		return nil, fmt.Errorf("must provide one or more userdatas")
	}

	mchan := make(chan Machine, n)
	errchan := make(chan error, n)

	for i := 0; i < n; i++ {
		ud := userdatas[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			m, err := c.NewMachine(ud)
			if err != nil {
				errchan <- err
			}
			if m != nil {
				mchan <- m
			}
		}()
	}

	wg.Wait()
	close(mchan)
	close(errchan)

	machs := []Machine{}

	for m := range mchan {
		machs = append(machs, m)
	}

	if firsterr, ok := <-errchan; ok {
		for _, m := range machs {
			m.Destroy()
		}
		return nil, firsterr
	}

	return machs, nil
}

// commonMachineChecks tests a machine for various error conditions such as ssh
// being available and no systemd units failing at the time ssh is reachable.
// It also ensures the remote system is running CoreOS.
//
// TODO(mischief): better error messages.
func commonMachineChecks(m Machine) error {
	// ensure ssh works
	sshChecker := func() error {
		_, err := m.SSH("true")
		if err != nil {
			return err
		}
		return nil
	}

	if err := util.Retry(sshRetries, sshTimeout, sshChecker); err != nil {
		return fmt.Errorf("ssh unreachable: %v", err)
	}

	// ensure we're talking to a CoreOS system
	out, err := m.SSH("grep ^ID= /etc/os-release")
	if err != nil {
		return fmt.Errorf("no /etc/os-release file")
	}

	if !bytes.Equal(out, []byte("ID=coreos")) {
		return fmt.Errorf("not a CoreOS instance")
	}

	// ensure no systemd units failed during boot
	out, err = m.SSH("systemctl --no-legend --state failed list-units")
	if err != nil {
		return fmt.Errorf("systemctl: %v: %v", out, err)
	}

	if len(out) > 0 {
		return fmt.Errorf("some systemd units failed:\n%s", out)
	}

	return nil
}
