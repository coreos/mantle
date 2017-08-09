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

package misc

import (
	"strings"

	"github.com/coreos/mantle/kola/cluster"
	"github.com/coreos/mantle/kola/register"
	"github.com/coreos/mantle/platform/conf"
)

func init() {
	register.Register(&register.Test{
		Run:              CheckUserShells,
		ClusterSize:      1,
		ExcludePlatforms: []string{"gce"},
		Name:             "coreos.users.shells",
	})
	register.Register(&register.Test{
		Run:         UserCreate,
		ClusterSize: 1,
		Name:        "coreos.misc.usercreate",
		UserData: conf.ContainerLinuxConfig(`passwd:
  users:
    - name: user1
      password_hash: "$1$xyz$hS2WeqUH/1Z2RyvZMLfvz/"
      ssh_authorized_keys:
        - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC4iOch8YzSh6v8yMOLRBwWftCQqdANaHK5oClC5oGVVT5C+1Ai7FF2HGXmoprbE70pmzdb8cyLEUwuew5w0YNxIhwYj1aPQeHlU23vyguUgCEDHjn2bvLDvvyEDfEw4/7JeHETC+GNmGOlReq3pO7iQsSFOrSCMCwhMN1KCHyhJ+Ve3UKt6ZIl7sJWsEeP+62hEf0bE2M+Jg8TSFVH8V0K9wG0DrKFf+NNZwdW5VAhkFN6qP+HdEsvn/QkcdQ5APDcnpS6OPVLZlRTW4sNrAs+2muCZCVoNWVUvSXul8+sEzAs3+ODFHezJ9lKUE/KqXDBZeQ+R9lFQxkQhW6iEcJj
      create:
        home_dir: /home/user1
        groups:
          - wheel
        shell: /bin/sh`),
	})
}

func CheckUserShells(c cluster.TestCluster) {
	m := c.Machines()[0]
	var badusers []string

	ValidUsers := map[string]string{
		"root":     "/bin/bash",
		"sync":     "/bin/sync",
		"shutdown": "/sbin/shutdown",
		"halt":     "/sbin/halt",
		"core":     "/bin/bash",
	}

	output, err := m.SSH("getent passwd")
	if err != nil {
		c.Fatalf("Failed to run grep: output %s, status: %v", output, err)
	}

	users := strings.Split(string(output), "\n")

	for _, user := range users {
		userdata := strings.Split(user, ":")
		if len(userdata) != 7 {
			badusers = append(badusers, user)
			continue
		}

		username := userdata[0]
		shell := userdata[6]
		if shell != ValidUsers[username] && shell != "/sbin/nologin" {
			badusers = append(badusers, user)
		}
	}

	if len(badusers) != 0 {
		c.Fatalf("Invalid users: %v", badusers)
	}
}

func UserCreate(c cluster.TestCluster) {
	m := c.Machines()[0]

	shadowContents, err := m.SSH("sudo cat /etc/shadow")
	if err != nil {
		c.Fatalf("could not check password: %v", err)
	}
	if !strings.Contains(string(shadowContents), "$1$xyz$hS2WeqUH/1Z2RyvZMLfvz/") {
		c.Fatalf("core user password not set correctly")
	}
	sshKeysContents, err := m.SSH("sudo cat /home/user1/.ssh/authorized_keys")
	if err != nil {
		c.Fatalf("could not check user1's ssh keys: %v", err)
	}
	if string(sshKeysContents) != "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC4iOch8YzSh6v8yMOLRBwWftCQqdANaHK5oClC5oGVVT5C+1Ai7FF2HGXmoprbE70pmzdb8cyLEUwuew5w0YNxIhwYj1aPQeHlU23vyguUgCEDHjn2bvLDvvyEDfEw4/7JeHETC+GNmGOlReq3pO7iQsSFOrSCMCwhMN1KCHyhJ+Ve3UKt6ZIl7sJWsEeP+62hEf0bE2M+Jg8TSFVH8V0K9wG0DrKFf+NNZwdW5VAhkFN6qP+HdEsvn/QkcdQ5APDcnpS6OPVLZlRTW4sNrAs+2muCZCVoNWVUvSXul8+sEzAs3+ODFHezJ9lKUE/KqXDBZeQ+R9lFQxkQhW6iEcJj" {
		c.Fatalf("user1 ssh key not set correctly: %q", sshKeysContents)
	}
}
