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
	"golang.org/x/crypto/ssh"
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

	// PasswordSSHClient establishes a new SSH connection using the provided credentials.
	PasswordSSHClient(user string, password string) (*ssh.Client, error)

	// SSH runs a single command over a new SSH connection.
	SSH(cmd string) ([]byte, error)

	// Destroy terminates the machine and frees associated resources.
	Destroy() error
}
