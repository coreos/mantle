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
	"fmt"
	"os"

	"github.com/coreos/mantle/harness"
)

type Test struct {
	Name        string
	ClusterSize int
	Run         func(c *Cluster)
}

var (
	tests harness.Tests
)

func Register(test Test) {
	if test.Name == "" {
		panic(fmt.Errorf("Missing Name: %#v", test))
	}
	if test.ClusterSize < 1 {
		panic(fmt.Errorf("Invalid ClusterSize: %#v", test))
	}
	tests.Add(test.Name, test.run)
}

func Run(opts harness.Options) {
	suite := harness.NewSuite(opts, tests)
	if err := suite.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Println("FAIL")
		os.Exit(1)
	}
	fmt.Println("PASS")
	os.Exit(0)
}

func (test Test) run(h *harness.H) {
	h.Parallel()

	c := newCluster(h, test)

	// setup may fail with an incomplete state so schedule
	// the destroy to cleanup anything first.
	defer c.destroy()

	c.setup()

	test.Run(c)
}
