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
	"encoding/json"
	"fmt"

	"github.com/coreos/go-semver/semver"
	"github.com/coreos/ignition/config"
	"github.com/coreos/ignition/config/types"
	"github.com/coreos/mantle/kola/register"
)

type dockerTestTarget string

const (
	dockerTestTargetHost     dockerTestTarget = "docker-host"
	dockerTestTargetSkim1126                  = "docker-skim-1.12.6"
)

type dockerTestRegistrationFn func(t dockerTest)

func registerDockerTests(t dockerTest) {
	targets := t.targets
	if len(targets) == 0 {
		targets = []dockerTestTarget{dockerTestTargetSkim1126, dockerTestTargetHost}
	}
	for _, target := range targets {
		targetMap[target](t)
	}
}

// targetMap specifies functions that will take a generic dockerTest and
// transform it into a kola test for the given target.
var targetMap = map[dockerTestTarget]dockerTestRegistrationFn{
	dockerTestTargetHost:     targetHostDocker,
	dockerTestTargetSkim1126: targetDockerSkim1126,
}

// dockerTest specifies the reusable portions of a test, and which docker
// targets it may be run against.
type dockerTest struct {
	// name is the name of the test. It maps to a kola name, but will
	// automatically be prefixed with a target-specific string (e.g. `docker.` +
	// name).
	name string
	// test specifies the test to register.
	// Note that the 'name', 'userdata', and 'platform' variables may be modified
	// by a given registration function
	test register.Test
	// targets, if set, will choose the specific docker targets for this test
	// from the above list.  If unset or empty, all targets will be run
	// Only ignition config is supported by the docker skim target.
	targets []dockerTestTarget
}

// targetDockerSkim1126 transforms a test to run on docker skim
func targetDockerSkim1126(t dockerTest) {
	name := t.name
	test := &t.test
	currentPlatforms := test.Platforms
	dockerSkimPlatforms := []string{"aws", "gce"}

	// Filter for the set of supported platforms from those the caller chose
	supportedSpecifiedPlatforms := currentPlatforms[:0]
	for _, platform := range currentPlatforms {
		for _, supported := range dockerSkimPlatforms {
			if supported == platform {
				supportedSpecifiedPlatforms = append(supportedSpecifiedPlatforms, platform)
				break
			}
		}
	}

	if len(currentPlatforms) == 0 {
		// Default to the full supported set
		test.Platforms = dockerSkimPlatforms
	} else {
		test.Platforms = supportedSpecifiedPlatforms
	}

	test.Name = fmt.Sprintf("%s.%s", dockerTestTargetSkim1126, name)

	ign := types.Config{
		Ignition: types.Ignition{
			Version: types.IgnitionVersion(semver.Version{
				Major: 2,
				Minor: 0,
				Patch: 0,
			}),
		},
	}
	if test.UserData != "" {
		var err error
		ign, err = config.Parse([]byte(test.UserData))
		if err != nil {
			plog.Fatalf("invalid ignition config for %v: %v", test.Name, err)
		}
	}

	skimUnit := types.SystemdUnit{
		Name: "docker.service",
		// https://github.com/coreos/docker-skim/blob/master/Documentation/using-docker-skim.md
		Contents: `
[Unit]
Description=Docker Application Container Engine
Documentation=http://docs.docker.com
After=docker.socket
Requires=docker.socket

[Service]
Type=simple

ExecStart=/usr/bin/rkt run --dns=host --interactive \
  --trust-keys-from-https \
  --stage1-name=users.developer.core-os.net/skim/stage1-skim:0.0.1 \
  users.developer.core-os.net/skim/docker:1.12.6_coreos.0 \
  --exec=/usr/lib/coreos/dockerd -- \
  --host=fd:// $DOCKER_OPTS $DOCKER_CGROUPS $DOCKER_OPT_BIP $DOCKER_OPT_MTU $DOCKER_OPT_IPMASQ

ExecReload=/bin/kill -s HUP $MAINPID
LimitNOFILE=1048576
LimitNPROC=infinity
LimitCORE=infinity
TasksMax=infinity
TimeoutStartSec=0
# set delegate yes so that systemd does not reset the cgroups of docker containers
Delegate=yes

[Install]
WantedBy=multi-user.target`,
	}

	// Modify ignition to force skim to be used
	ign.Systemd.Units = append(ign.Systemd.Units, skimUnit)
	ignConfigData, err := json.Marshal(ign)
	if err != nil {
		plog.Fatalf("unable to marshal ignition config: %v", err)
	}
	test.UserData = string(ignConfigData)

	register.Register(test)
}

// targetDockerHost runs a test against the docker version shipped in container linux
func targetHostDocker(t dockerTest) {
	name := t.name
	test := &t.test
	test.Name = fmt.Sprintf("%s.%s", dockerTestTargetHost, name)
	if test.UserData == "" {
		test.UserData = "#cloud-config"
	}
	register.Register(test)
}
