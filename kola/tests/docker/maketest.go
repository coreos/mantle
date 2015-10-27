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

package docker

import (
	"fmt"
	"time"

	"github.com/coreos/mantle/platform"
	"github.com/coreos/mantle/util"

	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/pkg/capnslog"
)

var (
	plog = capnslog.NewPackageLogger("github.com/coreos/mantle", "kola/tests/docker")
)

func makeTestBranch(c platform.TestCluster, branch string) error {
	mach := c.Machines()[0]

	commands := []struct {
		tries int
		cmd   string
	}{
		{3, "git clone -b " + branch + " https://github.com/docker/docker"},
		{1, "git -C docker checkout -b dry-run-test"},
		{1, "docker build -t dry-run-test docker"},
		{1, `docker run --privileged --rm -i -v ` + "`pwd`" + `/docker/:/go/src/github.com/docker/docker -w /go/src/github.com/docker/docker dry-run-test /bin/bash -c "DOCKER_GRAPHDRIVER=overlay DOCKER_TEST_HOST=tcp://172.17.42.1:2376 ./hack/make.sh dynbinary binary test-unit test-integration-cli"`},
	}

	for i, c := range commands {
		cmd := c.cmd
		cmdfunc := func() error {
			out, err := mach.SSH(cmd)
			if err != nil {
				return fmt.Errorf("command %q failed: %s", cmd, out)
			}

			plog.Debugf("%q output:\n%s", cmd, out)
			return nil
		}

		if err := util.Retry(c.tries, 5*time.Second, cmdfunc); err != nil {
			platform.Manhole(mach)
			return fmt.Errorf("docker test step %d failed: %v", i, err)
		}
	}

	return nil
}

// MakeTest runs 'make test' in a checkout of docker source.
func MakeTest(c platform.TestCluster) error {
	return makeTestBranch(c, "v1.8.3")
}
