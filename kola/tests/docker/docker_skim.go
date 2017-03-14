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

package docker

import (
	"fmt"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/coreos/mantle/kola/cluster"
	"github.com/coreos/mantle/kola/register"
)

func init() {
	registerDockerTests(dockerTest{
		name: "skim1126",
		test: register.Test{
			Run:         dockerSkim1126Version,
			ClusterSize: 1,
			MinVersion:  semver.Version{Major: 1192},
		},
		targets: []dockerTestTarget{dockerTestTargetSkim1126},
	})
}

func dockerSkim1126Version(c cluster.TestCluster) error {
	m := c.Machines()[0]

	output, err := m.SSH(`docker version -f '{{.Server.Version}}'`)
	if err != nil {
		return fmt.Errorf("error determining version: %v, %v", err, string(output))
	}
	if strings.TrimSpace(string(output)) != "1.12.6" {
		return fmt.Errorf("expected 1.12.6; got %v", string(output))
	}

	return nil
}
