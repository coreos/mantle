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

// crioPodTemplate is a simple string template required for creating a pod in crio
// It takes two strings: the name (which will be expanded) and the generated image name
var crioPodTemplate = `{
	"metadata": {
		"name": "rhcos-crio-pod-%s",
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
			"cgroup_parent": "Burstable-pod-123.slice",
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

// crioContainerTemplate is a simple string template required for running a container
// It takes three strings: the name (which will be expanded), the image, and the argument to run
var crioContainerTemplate = `{
	"metadata": {
		"name": "rhcos-crio-container-%s",
		"attempt": 1
	},
	"image": {
		"image": "docker.io/library/%s"
	},
	"command": [
		"%s"
	],
	"args": [],
	"working_dir": "/",
	"envs": [
		{
			"key": "PATH",
			"value": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
		},
		{
			"key": "TERM",
			"value": "xterm"
		}
	],
	"labels": {
		"type": "small",
		"batch": "no"
	},
	"annotations": {
		"daemon": "crio"
	},
	"privileged": true,
	"log_path": "",
	"stdin": false,
	"stdin_once": false,
	"tty": false,
	"linux": {
		"resources": {
			"cpu_period": 10000,
			"cpu_quota": 20000,
			"cpu_shares": 512,
			"oom_score_adj": 30,
			"memory_limit_in_bytes": 268435456
		},
		"security_context": {
			"readonly_rootfs": false,
			"selinux_options": {
				"user": "system_u",
				"role": "system_r",
				"type": "svirt_lxc_net_t",
				"level": "s0:c4,c5"
			},
			"capabilities": {
				"add_capabilities": [
					"setuid",
					"setgid"
				],
				"drop_capabilities": [
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

// generateCrioConfig generates a crio pod/container configuration
// based on the input name and arguments returning the path to the generated configs.
func generateCrioConfig(name string, command []string) (string, string) {
	fileContentsPod := fmt.Sprintf(crioPodTemplate, name, name)

	tmpFilePod, err := ioutil.TempFile("", name + "Pod")
	if err != nil {
		panic(err.Error())
	}
	if _, err = tmpFilePod.Write([]byte(fileContentsPod)); err != nil {
		panic(err.Error())
	}

	cmd := strings.Join(command, " ")
	fileContentsContainer := fmt.Sprintf(crioContainerTemplate, name, name, cmd)

	tmpFileContainer, err := ioutil.TempFile("", name + "Container")
	if err != nil {
		panic(err.Error())
	}
	if _, err = tmpFileContainer.Write([]byte(fileContentsContainer)); err != nil {
		panic(err.Error())
	}

	return tmpFilePod.Name(), tmpFileContainer.Name()
}

// genContainer makes a container out of binaries on the host. This function uses podman to build.
// The first string returned by this function is the pod config to be used with crictl runp. The second
// string returned is the container config to be used with crictl create/exec. They will be dropped
// on to all machines in the cluster as ~/$STRING_RETURNED_FROM_FUNCTION. Note that the string returned
// here is just the name, not the full path on the cluster machine(s).
func genContainer(c cluster.TestCluster, m platform.Machine, name string, binnames []string, shellCommands []string) (string, string, error) {
	configPathPod, configPathContainer := generateCrioConfig(name, shellCommands)
	if err := c.DropFile(configPathPod); err != nil {
		return "", "", err
	}
	if err := c.DropFile(configPathContainer); err != nil {
		return "", "", err
	}

	// This shell script creates both the image for testing and the fake pause image
	// required by crio
	cmd := `tmpdir=$(mktemp -d); cd $tmpdir; echo -e "FROM scratch\nCOPY . /" > Dockerfile;
	        b=$(which %s); libs=$(sudo ldd $b | grep -o /lib'[^ ]*' | sort -u);
			sudo rsync -av --relative --copy-links $b $libs ./;
			c=$(which sleep); libs=$(sudo ldd $c | grep -o /lib'[^ ]*' | sort -u);
			sudo rsync -av --relative --copy-links $c $libs ./;
			echo "#!/bin/bash\n$c 30" > ./pause;
			chmod a+x ./pause;
			sudo podman build -t %s -t kubernetes/pause .`
	c.MustSSH(m, fmt.Sprintf(cmd, strings.Join(binnames, " "), name))

	return path.Base(configPathPod), path.Base(configPathContainer), nil
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

	crioConfigPod, crioConfigContainer, err := genContainer(
		c, m, "ping", []string{"ping"},
		[]string{"ping"})
	if err != nil {
		c.Fatal(err)
	}

	// Here we generate 10 pods, each will run a container responsible for
	// pinging to host
	cmdCreatePod := fmt.Sprintf("sudo crictl runp %s", crioConfigPod)
	output := ""
	for x := 1; x <= 10; x++ {
		podID := c.MustSSH(m, cmdCreatePod)
		cmdCreateContainer := fmt.Sprintf("sudo crictl create %s %s %s", podID, crioConfigContainer, crioConfigPod)
		containerID := c.MustSSH(m, cmdCreateContainer)
		cmdExecContainer := fmt.Sprintf("sudo crictl exec -t %s ping -i 0.2 172.17.0.1 -w 1 >/dev/null && echo PASS || echo FAIL", containerID)
		output = output + string(c.MustSSH(m, cmdExecContainer))
	}

	numPass := strings.Count(string(output), "PASS")

	if numPass != 10 {
		c.Fatalf("Expected 10 passes, but received %d passes with output: %s", numPass, output)
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
