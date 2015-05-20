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

package main

import (
	"fmt"
	"os"

	"github.com/coreos/mantle/cli"
	"github.com/coreos/mantle/kola"
)

const (
	cliName        = "kolalet"
	cliDescription = "Native code runner for kola"
	// http://en.wikipedia.org/wiki/Kola_Superdeep_Borehole
)

// main test harness
var cmdRun = &cli.Command{
	Name:    "run",
	Summary: "Run native tests a group at a time",
	Run:     Run,
}

func init() {
	cli.Register(cmdRun)
}

func main() {
	cli.Run(cliName, cliDescription)
}

// test runner
func Run(args []string) int {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "FAIL: Extra arguements specified. Usage: 'kolalet run <test group name>'\n")
		return 2
	}

	// find test with matching name
	group, ok := kola.Groups[args[0]]
	if !ok {
		fmt.Fprintf(os.Stderr, "FAIL: test group not found\n")
		return 1
	}

	var ranTests int //count successful tests
	for _, f := range group.NativeTests {
		ranTests++

		err := f()
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: on native test %v: %v", ranTests, err)
			return 1
		}
	}
	return 0
}
