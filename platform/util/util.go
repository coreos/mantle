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

// Package util contains utility functions for Clusters and Machines.
package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/coreos/mantle/platform"
	"github.com/coreos/mantle/util"
)

// Manhole connects os.Stdin, os.Stdout, and os.Stderr to an interactive shell
// session on the Machine m. Manhole blocks until the shell session has ended.
// If os.Stdin does not refer to a TTY, Manhole returns immediately with a nil
// error.
func Manhole(m platform.Machine) error {
	fd := int(os.Stdin.Fd())
	if !terminal.IsTerminal(fd) {
		return nil
	}

	tstate, _ := terminal.MakeRaw(fd)
	defer terminal.Restore(fd, tstate)

	client, err := m.SSHClient()
	if err != nil {
		return fmt.Errorf("SSH client failed: %v", err)
	}

	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("SSH session failed: %v", err)
	}

	defer session.Close()

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	modes := ssh.TerminalModes{
		ssh.TTY_OP_ISPEED: 115200,
		ssh.TTY_OP_OSPEED: 115200,
	}

	cols, lines, err := terminal.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}

	if err = session.RequestPty(os.Getenv("TERM"), lines, cols, modes); err != nil {
		return fmt.Errorf("failed to request pseudo terminal: %s", err)
	}

	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %s", err)
	}

	if err := session.Wait(); err != nil {
		return fmt.Errorf("failed to wait for session: %s", err)
	}

	return nil
}

// StreamJournal streams the remote system's journal to stdout.
func StreamJournal(m platform.Machine) error {
	c, err := m.SSHClient()
	if err != nil {
		return fmt.Errorf("SSH client failed: %v", err)
	}

	s, err := c.NewSession()
	if err != nil {
		return fmt.Errorf("SSH session failed: %v", err)
	}

	s.Stdout = os.Stdout
	s.Stderr = os.Stderr
	go func() {
		defer c.Close()
		defer s.Close()
		s.Run("journalctl -f")
	}()

	return nil
}

// Reboots a machine and blocks until the system to be accessible by SSH again.
// It will return an error if the machine is not accessible after a timeout.
func Reboot(m platform.Machine) error {
	// stop sshd so that commonMachineChecks will only work if the machine
	// actually rebooted
	out, err := m.SSH("sudo systemctl stop sshd.socket; sudo systemd-run --no-block systemctl reboot")
	if err != nil {
		return fmt.Errorf("issuing reboot command failed: %v", out)
	}

	return CommonMachineChecks(m)
}

// Wrap a StdoutPipe as a io.ReadCloser
type sshPipe struct {
	s   *ssh.Session
	c   *ssh.Client
	err *bytes.Buffer
	io.Reader
}

func (p *sshPipe) Close() error {
	if err := p.s.Wait(); err != nil {
		return fmt.Errorf("%s: %s", err, p.err)
	}
	if err := p.s.Close(); err != nil {
		return err
	}
	return p.c.Close()
}

// Copy a file between two machines in a cluster.
func TransferFile(src platform.Machine, srcPath string, dst platform.Machine, dstPath string) error {
	srcPipe, err := ReadFile(src, srcPath)
	if err != nil {
		return err
	}
	defer srcPipe.Close()

	if err := InstallFile(srcPipe, dst, dstPath); err != nil {
		return err
	}
	return nil
}

// ReadFile returns a io.ReadCloser that streams the requested file. The
// caller should close the reader when finished.
func ReadFile(m platform.Machine, path string) (io.ReadCloser, error) {
	client, err := m.SSHClient()
	if err != nil {
		return nil, fmt.Errorf("failed creating SSH client: %v", err)
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed creating SSH session: %v", err)
	}

	// connect session stdout
	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, err
	}

	// collect stderr
	errBuf := bytes.NewBuffer(nil)
	session.Stderr = errBuf

	// stream file to stdout
	err = session.Start(fmt.Sprintf("sudo cat %s", path))
	if err != nil {
		session.Close()
		client.Close()
		return nil, err
	}

	// pass stdoutPipe as a io.ReadCloser that cleans up the ssh session
	// on when closed.
	return &sshPipe{session, client, errBuf, stdoutPipe}, nil
}

// InstallFile copies data from in to the path to on m.
func InstallFile(in io.Reader, m platform.Machine, to string) error {
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

// DropFile places file from localPath to ~/ on every machine in cluster
func DropFile(c platform.Cluster, localPath string) error {
	in, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer in.Close()

	for _, m := range c.Machines() {
		if _, err := in.Seek(0, 0); err != nil {
			return err
		}
		if err := InstallFile(in, m, filepath.Base(localPath)); err != nil {
			return err
		}
	}
	return nil
}

// NewMachines spawns len(userdatas) instances in cluster c, with
// each instance passed the respective userdata.
func NewMachines(c platform.Cluster, userdatas []string) ([]platform.Machine, error) {
	var wg sync.WaitGroup

	n := len(userdatas)

	if n <= 0 {
		return nil, fmt.Errorf("must provide one or more userdatas")
	}

	mchan := make(chan platform.Machine, n)
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

	machs := []platform.Machine{}

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
func CommonMachineChecks(m platform.Machine) error {
	// ensure ssh works
	sshChecker := func() error {
		_, err := m.SSH("true")
		if err != nil {
			return err
		}
		return nil
	}

	if err := util.Retry(10, 2*time.Second, sshChecker); err != nil {
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
