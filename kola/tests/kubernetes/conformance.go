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

package kubernetes

import (
	"fmt"

	"github.com/coreos/mantle/kola/register"
	"github.com/coreos/mantle/platform"
)

// register a separate test for each version tag
var confTags = []string{
	"v1.1.7_coreos.2",
	"v1.1.8_coreos.0",
}

func init() {
	for i := range confTags {
		// use closure to store a version tag in a Test
		t := confTags[i]
		f := func(c platform.TestCluster) error {
			return CoreOSConformance(c, t)
		}

		register.Register(&register.Test{
			Name:        "google.kubernetes.conformance." + t,
			Run:         f,
			ClusterSize: 0,
			Platforms:   []string{"gce", "aws"},
		})
	}
}

// Run kubernetes conformance tests. Assumes there is a container with
// all the upstream test binaries for each version tag.
func CoreOSConformance(c platform.TestCluster, version string) error {
	if err := setupCluster(c, 4, version); err != nil {
		return err
	}
	master := c.Machines()[1]

	version, err := stripSemverSuffix(version)
	if err != nil {
		return err
	}

	runTests := `sudo rkt  run \
		--volume kube,kind=host,source=/home/core \
		--mount volume=kube,target=/home/core \
		--net=host \
		--trust-keys-from-https \
		quay.io/peanutbutter/kubeconformance:%v \
		--exec=/bin/bash -- -c \
		"cd /kubernetes/ && KUBECONFIG=/home/core/.kube/config hack/conformance-test.sh 2>&1"`
	runTests = fmt.Sprintf(runTests, version)
	plog.Errorf("%v", runTests)

	b, err := master.SSH(runTests)

	// The exceptional case: all conformance tests pass
	if err == nil {
		plog.Noticef("All conformance tests passed!")
		plog.Noticef("%s", b)
		return nil
	}
	plog.Noticef("%s", b)

	// Check failures against our whitelist and report those.

	// If any failures remain report and fail test
	return nil
}
