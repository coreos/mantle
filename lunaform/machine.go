// Copyright 2017 CoreOS, Inc.
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

package lunaform

import (
	"bytes"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"

	"github.com/coreos/mantle/platform"
)

type machine struct {
	cluster  *Cluster
	name     string
	localIP  string
	remoteIP string
	journal  *platform.Journal
}

var _ platform.Machine = &machine{}

func (m *machine) setup() error {
	dir := filepath.Join(m.cluster.OutputDir(), m.ID())
	if err := os.Mkdir(dir, 0777); err != nil {
		return err
	}
	var err error
	if m.journal, err = platform.NewJournal(dir); err != nil {
		return err
	}
	if err := m.journal.Start(m.cluster.Context(), m); err != nil {
		return err
	}
	if err := platform.CheckMachine(m); err != nil {
		return err
	}
	if err := platform.EnableSelinux(m); err != nil {
		return err
	}
	return nil
}

func (m *machine) destroy() error {
	if m == nil || m.journal != nil {
		return m.journal.Destroy()
	}
	return nil
}

func (m *machine) ID() string {
	return m.name
}

func (m *machine) IP() string {
	return m.remoteIP
}

func (m *machine) PrivateIP() string {
	return m.localIP
}

func (m *machine) SSHClient() (*ssh.Client, error) {
	return m.cluster.sshAgent.NewClient(m.IP())
}

func (m *machine) PasswordSSHClient(user string, password string) (*ssh.Client, error) {
	return m.cluster.sshAgent.NewPasswordClient(m.IP(), user, password)
}

func (m *machine) SSH(cmd string) ([]byte, error) {
	client, err := m.SSHClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	out, err := session.Output(cmd)
	out = bytes.TrimSpace(out)
	return out, err
}

func (m *machine) Reboot() error {
	if err := platform.StartReboot(m); err != nil {
		return err
	}
	if err := m.journal.Start(m.cluster.Context(), m); err != nil {
		return err
	}
	if err := platform.CheckMachine(m); err != nil {
		return err
	}
	if err := platform.EnableSelinux(m); err != nil {
		return err
	}
	return nil
}

// Machines cannot be individually destroyed by tests.
func (m *machine) Destroy() error {
	panic("Destroy not supported")
}
