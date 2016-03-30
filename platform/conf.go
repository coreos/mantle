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
	"encoding/json"

	cci "github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/mantle/Godeps/_workspace/src/golang.org/x/crypto/ssh/agent"
	"github.com/coreos/mantle/config/ignition/v1/config"
	"github.com/coreos/mantle/config/ignition/v1/config/types"
)

// Conf is a configuration for a CoreOS machine. It may be either a
// coreos-cloudconfig or an ignition configuration.
type Conf struct {
	ignition    *types.Config
	cloudconfig *cci.CloudConfig
}

// NewConf parses userdata and returns a new Conf. It returns an error if the
// userdata can't be parsed as a coreos-cloudinit or ignition configuration.
func NewConf(userdata string) (*Conf, error) {
	c := &Conf{}

	ignc, err := config.Parse([]byte(userdata))
	switch err {
	case config.ErrEmpty:
		// empty, noop
	case config.ErrCloudConfig:
		// fall back to cloud-config
		c.cloudconfig, err = cci.NewCloudConfig(userdata)
		if err != nil {
			return nil, err
		}
	default:
		// some other error (invalid json, script)
		return nil, err
	case nil:
		c.ignition = &ignc
	}

	return c, nil
}

// String returns the string representation of the userdata in Conf.
func (c *Conf) String() string {
	if c.ignition != nil {
		var buf bytes.Buffer

		err := json.NewEncoder(&buf).Encode(c.ignition)
		if err != nil {
			return ""
		}

		return buf.String()
	} else if c.cloudconfig != nil {
		return c.cloudconfig.String()
	}

	return ""
}

func (c *Conf) copyKeysIgnition(keys []*agent.Key) {
	// lookup core user entry
	var usr *types.User

	users := c.ignition.Passwd.Users

	for i, u := range users {
		if u.Name == "core" {
			usr = &users[i]
		}
	}

	// doesn't exist yet - create it
	if usr == nil {
		u := types.User{Name: "core"}
		users = append(users, u)
		c.ignition.Passwd.Users = users
		usr = &users[len(users)-1]
	}

	for _, key := range keys {
		usr.SSHAuthorizedKeys = append(usr.SSHAuthorizedKeys, key.String())
	}
}

func (c *Conf) copyKeysCloudConfig(keys []*agent.Key) {
	for _, key := range keys {
		c.cloudconfig.SSHAuthorizedKeys = append(c.cloudconfig.SSHAuthorizedKeys, key.String())
	}
}

// CopyKeys copies public keys from agent ag into the configuration to the
// appropriate configuration section for the core user.
func (c *Conf) CopyKeys(keys []*agent.Key) {
	if c.ignition != nil {
		c.copyKeysIgnition(keys)
	} else if c.cloudconfig != nil {
		c.copyKeysCloudConfig(keys)
	}
}
