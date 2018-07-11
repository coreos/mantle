// Copyright 2018 Red Hat, Inc.
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

package crio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/context"

	"github.com/coreos/mantle/kola/cluster"
	"github.com/coreos/mantle/kola/register"
	"github.com/coreos/mantle/lang/worker"
	"github.com/coreos/mantle/platform"
)

// crioArguments abstracts arguments used within a crio json config
type crioArguments []string

// simplifiedCrioInfo represents the results from crio info
type simplifiedCrioInfo struct {
	StorageDriver string `json:"storage_driver"`
	StorageRoot   string `json:"storage_root"`
	CgroupDriver  string `json:"cgroup_driver"`
}

// crioPodTemplate is a simple string template required for running crio pods/containers
// It takes two strings: the name (which will be expanded) and the argument to run
var crioPodTemplate = `{
	"metadata": {
		"name": "rhcos-crio-%s",
		"namespace": "redhat.test.crio"
	},
	"image": {
			"image": "localhost/%s:latest"
	},
	"args": ["sh", "init.sh"],
	"readonly_rootfs": false,
	"log_path": "",
	"stdin": false,
	"stdin_once": false,
	"tty": true,
	"linux": {
			"resources": {
					"memory_limit_in_bytes": 209715200,
					"cpu_period": 10000,
					"cpu_quota": 20000,
					"cpu_shares": 512,
					"oom_score_adj": 30,
					"cpuset_cpus": "0",
					"cpuset_mems": "0"
			},
			"security_context": {
					"namespace_options": {
							"pid": 1
					},
					"capabilities": {
							"add_capabilities": [
								"sys_admin"
							]
					}
			}
	}
}`

// init runs when the package is imported and takes care of registering tests
func init() {
	register.Register(&register.Test{
		Run:         crioBaseTests,
		ClusterSize: 1,
		Name:        `crio.base`,
		Distros:     []string{"rhcos"},
	})
	// TODO: Enable these once crio.base works fully
	// register.Register(&register.Test{
	// 	Run:         crioNetwork,
	// 	ClusterSize: 2,
	// 	Name:        "crio.network",
	// 	Distros:     []string{"rhcos"},
	// })
}

// crioBaseTests executes multiple tests under the "base" name
func crioBaseTests(c cluster.TestCluster) {
	c.Run("crio-info", testCrioInfo)
	c.Run("networks-reliably", crioNetworksReliably)
}

// TODO: REMOVE THIS WHEN POSSIBLE
// hackStartCrio is needed while crio isn't auto started in the compose.
// TODO: REMOVE THIS WHEN POSSIBLE
func hackStartCrio(c cluster.TestCluster) {
	for _, m := range c.Machines() {
		if _, err := c.SSH(m, `sudo systemctl start crio`); err != nil {
			c.Fatal(err)
		}
	}
}

// generateCrioContainerConfig generates a crio container configuration
// based on the input name and arguments returning the path to the generated config.
func generateCrioConfig(name string) string {
	fileContents := fmt.Sprintf(crioPodTemplate, name, name)
	// TODO: Remove before final commit
	fmt.Println(fileContents)

	tmpFile, err := ioutil.TempFile("", name)
	if err != nil {
		panic(err.Error())
	}
	if _, err = tmpFile.Write([]byte(fileContents)); err != nil {
		panic(err.Error())
	}
	return tmpFile.Name()
}

// genContainer makes a container out of binaries on the host. This function uses podman to build.
// The string returned by this function is the config to used with crictl runp. It will be dropped
// on to all machines in the cluster as ~/$STRING_RETURNED_FROM_FUNCTION. Note that the string returned
// here is just the name, not the full path on the cluster machine(s).
func genContainer(c cluster.TestCluster, m platform.Machine, name string, binnames []string, shellCommands []string) (string, error) {
	configPath := generateCrioConfig(name)
	if err := c.DropFile(configPath); err != nil {
		return "", err
	}
	// Generate the shell script
	file, err := ioutil.TempFile("", "init")
	if err != nil {
		c.Fatal(err)
	}
	file.WriteString("#!/bin/sh\n")
	for _, shellCmd := range(shellCommands) {
		file.WriteString(fmt.Sprintf("%s\n", shellCmd))
	}
	if err := c.DropFile(file.Name()); err != nil {
		return "", err
	}
	initName := path.Base(file.Name())

	// This shell script creates both the image for testing and the fake pause image
	// required by crio
	cmd := `tmpdir=$(mktemp -d); cd $tmpdir; echo -e "FROM scratch\nCOPY . /" > Dockerfile;
	        b=$(which %s); libs=$(sudo ldd $b | grep -o /lib'[^ ]*' | sort -u);
			sudo rsync -av --relative --copy-links $b $libs ./;
			c=$(which sleep); libs=$(sudo ldd $c | grep -o /lib'[^ ]*' | sort -u);
			sudo rsync -av --relative --copy-links $c $libs ./;
			echo "#!/bin/bash\n$c 30" > ./pause;
			chmod a+x ./pause;
			sudo cp ~/%s ./;
			sudo podman build -t %s -t k8s.gcr.io/pause:3.1 .`
	// TODO: Remove before final commit
	fmt.Println(cmd)
	c.MustSSH(m, fmt.Sprintf(cmd, strings.Join(binnames, " "), initName, name))
	return path.Base(configPath), nil
}

// TODO: MOVE TO genContainer
// genCrioPodConfig generates a pod config so crictl runp will run the container
func genCrioPodConfig(c cluster.TestCluster, imageName, fileName, arg string) (string, error) {
	crioPodConfig := fmt.Sprintf(crioPodTemplate, imageName, arg)
	if err := ioutil.WriteFile(fileName, []byte(crioPodConfig), 0644); err != nil {
		return "", err
	}
	if err := c.DropFile(fileName); err != nil {
		return "", err
	}
	return fileName, nil
}

// crioNetwork ensures that crio containers can make network connections outside of the host
func crioNetwork(c cluster.TestCluster) {
	machines := c.Machines()
	src, dest := machines[0], machines[1]

	c.Log("creating ncat containers")

	genContainer(c, src, "ncat", []string{"ncat"}, []string{"ncat"})
	genContainer(c, dest, "ncat", []string{"ncat"}, []string{"ncat"})

	listener := func(ctx context.Context) error {
		// Will block until a message is recieved
		// TODO: use genContainer instead
		genCrioPodConfig(c, "ncat", "ncat-server.json",
			`echo "HELLO FROM SERVER" | ncat --idle-timeout 20 --listen 0.0.0.0 9988`)
		out, err := c.SSH(dest, "crictl runp ncat-server.json")
		if err != nil {
			return err
		}

		if !bytes.Equal(out, []byte("HELLO FROM CLIENT")) {
			return fmt.Errorf("unexpected result from listener: %q", out)
		}

		return nil
	}

	talker := func(ctx context.Context) error {
		// Wait until listener is ready before trying anything
		for {
			_, err := c.SSH(dest, "sudo lsof -i TCP:9988 -s TCP:LISTEN | grep 9988 -q")
			if err == nil {
				break // socket is ready
			}

			exit, ok := err.(*ssh.ExitError)
			if !ok || exit.Waitmsg.ExitStatus() != 1 { // 1 is the expected exit of grep -q
				return err
			}

			select {
			case <-ctx.Done():
				return fmt.Errorf("timeout waiting for server")
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}

		genCrioPodConfig(c, "ncat", "ncat-client.json",
			fmt.Sprintf(`echo "HELLO FROM CLIENT" | ncat %s 9988`, dest.PrivateIP()))
		out, err := c.SSH(src, "crictl runp ncat-client.json")

		if err != nil {
			return err
		}

		if !bytes.Equal(out, []byte("HELLO FROM SERVER")) {
			return fmt.Errorf(`unexpected result from listener: "%v"`, out)
		}

		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := worker.Parallel(ctx, listener, talker); err != nil {
		c.Fatal(err)
	}
}

// crioNetworksReliably verifies that crio containers have a reliable network
func crioNetworksReliably(c cluster.TestCluster) {
	m := c.Machines()[0]

	crioConfig, err := genContainer(
		c, m, "ping", []string{"ping"},
		[]string{"sh", "-c", "ping -i 0.2 172.17.0.1 -w 1 > /dev/null && echo PASS || echo FAIL"})
	if err != nil {
		c.Fatal(err)
	}
	cmd := fmt.Sprintf("sudo crictl runp %s", crioConfig)
	output := ""
	for x := 1; x <= 10; x++ {
		fmt.Printf("\n%d: ", x)
		output = output + string(c.MustSSH(m, cmd))
	}

	numPass := strings.Count(string(output), "PASS")

	if numPass != 10 {
		c.Fatalf("Expected 10 passes, but output was: %s", output)
	}

}

// getCrioInfo parses and returns the information crio provides via socket
func getCrioInfo(c cluster.TestCluster, m platform.Machine) (simplifiedCrioInfo, error) {
	target := simplifiedCrioInfo{}
	crioInfoJSON, err := c.SSH(m, `sudo curl -s --unix-socket /var/run/crio/crio.sock http://crio/info`)

	if err != nil {
		return target, fmt.Errorf("could not get info: %v", err)
	}

	err = json.Unmarshal(crioInfoJSON, &target)
	if err != nil {
		return target, fmt.Errorf("could not unmarshal info %q into known json: %v", string(crioInfoJSON), err)
	}
	return target, nil
}

// testCrioInfo test that crio info's output is as expected.
func testCrioInfo(c cluster.TestCluster) {
	m := c.Machines()[0]
	// TODO: Remove when possible
	hackStartCrio(c)

	if _, err := c.SSH(m, `sudo systemctl start crio`); err != nil {
		c.Fatal(err)
	}

	info, err := getCrioInfo(c, m)
	if err != nil {
		c.Fatal(err)
	}
	expectedStorageDriver := "overlay"
	if info.StorageDriver != expectedStorageDriver {
		c.Errorf("unexpected storage driver: %v != %v", expectedStorageDriver, info.StorageDriver)
	}
	expectedStorageRoot := "/var/lib/containers/storage"
	if info.StorageRoot != expectedStorageRoot {
		c.Errorf("unexpected storage root: %v != %v", expectedStorageRoot, info.StorageRoot)
	}
	expectedCgroupDriver := "systemd"
	if info.CgroupDriver != expectedCgroupDriver {
		c.Errorf("unexpected cgroup driver: %v != %v", expectedCgroupDriver, info.CgroupDriver)
	}

}
