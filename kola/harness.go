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

package kola

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/coreos/mantle/platform"
)

// Defines a group of tests to run on a cluster instance. These tests are
// repeated on a new cluster for each platform. Cluster test functions are run
// first using the platform.Cluster interface. Then the series of native code
// tests are run on each member of the cluster.
type TestGroup struct {
	ClusterTests []func(platform.Cluster) error // run sequentially accross cluster
	NativeTests  []func() error                 // run sequentially per machine
	Name         string                         // should be uppercase and unique
	CloudConfig  string
	ClusterSize  int
	Platforms    []string // whitelist of platforms to run tests against -- defaults to all
}

// map names to test groups
var Groups = map[string]*TestGroup{}

// error if we register existing name
func Register(t *TestGroup) {
	_, ok := Groups[t.Name]
	if ok {
		panic("testgroup already registered with same name")
	}
	Groups[t.Name] = t
}

// Test runner
func RunTests(args []string) int {
	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "Extra arguements specified. Usage: 'kola run [glob pattern]'\n")
		return 2
	}
	var pattern string
	if len(args) == 1 {
		pattern = args[0]
	} else {
		pattern = "*" // run all tests by default
	}

	var ranTests int //count successful tests
	for _, t := range Groups {
		match, err := filepath.Match(pattern, t.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		if !match {
			continue
		}

		// run all platforms if whitelist is nil
		if t.Platforms == nil {
			t.Platforms = []string{"qemu", "gce"}
		}

		for _, pltfrm := range t.Platforms {
			err := runTestGroup(t, pltfrm)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v failed on %v: %v\n", t.Name, pltfrm, err)
				return 1
			}
			fmt.Printf("test group %v ran successfully on %v\n", t.Name, pltfrm)
			ranTests++
		}
	}
	fmt.Fprintf(os.Stderr, "All %v test groups ran successfully!\n", ranTests)
	return 0
}

// starts a cluster and runs all tests until first error
func runTestGroup(t *TestGroup, pltfrm string) error {
	var err error
	var cluster platform.Cluster
	if pltfrm == "qemu" {
		cluster, err = platform.NewQemuCluster(*QemuImage)
	} else if pltfrm == "gce" {
		cluster, err = platform.NewGCECluster(GCEOpts())
	} else {
		fmt.Fprintf(os.Stderr, "Invalid platform: %v", pltfrm)
	}

	if err != nil {
		return fmt.Errorf("Cluster failed: %v", err)
	}
	defer func() {
		if err := cluster.Destroy(); err != nil {
			fmt.Fprintf(os.Stderr, "cluster.Destroy(): %v\n", err)
		}
	}()

	url, err := cluster.GetDiscoveryURL(t.ClusterSize)
	if err != nil {
		return fmt.Errorf("Failed to create discovery endpoint: %v", err)
	}

	cfgs := makeConfigs(url, t.CloudConfig, t.ClusterSize)

	for i := 0; i < t.ClusterSize; i++ {
		_, err := cluster.NewMachine(cfgs[i])
		if err != nil {
			return fmt.Errorf("Cluster failed starting machine: %v", err)
		}
		fmt.Fprintf(os.Stderr, "%v instance up\n", pltfrm)
	}

	// run tests
	if t.ClusterTests != nil {
		for _, f := range t.ClusterTests {
			err = f(cluster)
			if err != nil {
				return fmt.Errorf("FAIL in group %v: %v", t.Name, err)
			}
		}
	}

	if t.NativeTests != nil {
		// drop binary
		for _, m := range cluster.Machines() {
			err = scpFile(m, "./kolalet")
			if err != nil {
				return fmt.Errorf("FAIL dropping kolalet binaries: %v", err)
			}

			_, err := m.SSH("chmod +x ./kolalet")
			if err != nil {
				return fmt.Errorf("FAIL dropping kolalet binaries: %v", err)
			}
		}

		// run native tests
		for _, m := range cluster.Machines() {
			b, err := m.SSH("./kolalet run " + t.Name)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "%s\n", b)
		}
	}

	return nil
}

// scpFile copies file from src path to ~/ on machine
func scpFile(m platform.Machine, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	session, err := m.SSHSession()
	if err != nil {
		return fmt.Errorf("Error establishing ssh session: %v", err)
	}
	defer session.Close()

	// machine reads file from stdin
	session.Stdin = in

	// cat file to fs
	_, filename := filepath.Split(src)
	_, err = session.CombinedOutput(fmt.Sprintf("cat > ./%s", filename))
	if err != nil {
		return err
	}
	return nil
}

// replaces $discovery with discover url in etcd cloud config and
// replaces $name with a unique name
func makeConfigs(url, cfg string, csize int) []string {
	cfg = strings.Replace(cfg, "$discovery", url, -1)

	var cfgs []string
	for i := 0; i < csize; i++ {
		cfgs = append(cfgs, strings.Replace(cfg, "$name", "instance"+strconv.Itoa(i), -1))
	}
	return cfgs
}
