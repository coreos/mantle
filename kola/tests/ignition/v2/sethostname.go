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

package ignition

import (
	"fmt"
	"strings"

	"github.com/coreos/mantle/kola/register"
	"github.com/coreos/mantle/platform"

	"github.com/coreos/go-semver/semver"
	"github.com/coreos/pkg/capnslog"
)

var (
	plog = capnslog.NewPackageLogger("github.com/coreos/mantle", "kola/tests/ignition")
)

func init() {
	// Set the hostname
	config := `{
		          "ignition": {
		              "version": "2.0.0"
		          },
		          "storage": {
		              "files": [
		                  {
		                      "filesystem": "root",
		                      "path": "/etc/hostname",
		                      "mode": 420,
		                      "contents": {
		                          "source": "data:,core1"
		                      }
		                  }
		              ]
		          }
		      }`
	register.Register(&register.Test{
		Name:        "coreos.ignition.v2.sethostname.aws",
		Run:         setHostname,
		ClusterSize: 1,
		Platforms:   []string{"aws"},
		MinVersion:  semver.Version{Major: 1010},
		UserData:    config,
	})
	register.Register(&register.Test{
		Name:        "coreos.ignition.v2.sethostname.gce",
		Run:         setHostname,
		ClusterSize: 1,
		Platforms:   []string{"gce"},
		MinVersion:  semver.Version{Major: 1045},
		UserData:    config,
	})
}

func setHostname(c platform.TestCluster) error {
	m := c.Machines()[0]

	out, err := m.SSH("hostnamectl")
	if err != nil {
		return fmt.Errorf("failed to run hostnamectl: %s: %v", out, err)
	}

	if !strings.Contains(string(out), "Static hostname: core1") {
		return fmt.Errorf("hostname wasn't set correctly:\n%s", out)
	}

	return nil
}
