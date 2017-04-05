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
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/coreos/mantle/harness"
	"github.com/coreos/mantle/lunaform/tfdata"
	"github.com/coreos/mantle/network"
	"github.com/coreos/mantle/platform"
	"github.com/coreos/mantle/platform/conf"
)

const gce = `
variable "name_prefix" {
	type = "string"
}

variable "cluster_size" {
	type = "string"
}

output "names" {
	value = ["${google_compute_instance.machine.*.name}"]
}

output "local_ips" {
	value = ["${google_compute_instance.machine.*.network_interface.0.address}"]
}

output "remote_ips" {
	value = ["${google_compute_instance.machine.*.network_interface.0.access_config.0.assigned_nat_ip}"]
}


provider "google" {
	project = "coreos-gce-testing"
	region = "us-central"
}

resource "google_compute_instance" "machine" {
	count = "${var.cluster_size}"

	name = "${var.name_prefix}-${count.index}"
	zone = "us-central1-a"
	machine_type = "n1-standard-1"

	metadata {
		user-data = "${file("user-data")}"
		block-project-ssh-keys = "TRUE"
	}

	disk {
		image = "coreos-cloud/coreos-alpha"
	}

	network_interface {
		network = "default"
		access_config {
			// Ephemeral IP
		}
	}
}
`

type tfInput struct {
	NamePrefix  string `json:"name_prefix"`
	ClusterSize string `json:"cluster_size"`
}

type tfOutput struct {
	Names     tfdata.OutputList `json:"names"`
	LocalIPs  tfdata.OutputList `json:"local_ips"`
	RemoteIPs tfdata.OutputList `json:"remote_ips"`
}

type Cluster struct {
	*harness.H
	test     Test
	sshAgent *network.SSHAgent
	machines []platform.Machine
}

var _ platform.Cluster = &Cluster{}

func newCluster(h *harness.H, test Test) *Cluster {
	return &Cluster{H: h, test: test}
}

func (c *Cluster) writeInputs() {
	data, err := json.Marshal(tfInput{
		NamePrefix:  NamePrefix(),
		ClusterSize: strconv.Itoa(c.test.ClusterSize),
	})
	if err != nil {
		c.Fatalf("encoding tfvars failed: %v", err)
	}

	path := filepath.Join(c.OutputDir(), "terraform.tfvars")
	if err := ioutil.WriteFile(path, data, 0666); err != nil {
		c.Fatalf("writing tfvars failed: %v", err)
	}
}

func (c *Cluster) writeUserData() {
	data, err := conf.New(`{"ignition": { "version": "2.0.0" }}`)
	if err != nil {
		c.Fatalf("parsing user data failed: %v", err)
	}

	keys, err := c.sshAgent.List()
	if err != nil {
		c.Fatalf("ssh agent failed: %v", err)
	}

	data.CopyKeys(keys)

	path := filepath.Join(c.OutputDir(), "user-data")
	if err := data.WriteFile(path); err != nil {
		c.Fatalf("writing user-data failed: %v", err)
	}
}

func (c *Cluster) writeConfig() {
	path := filepath.Join(c.OutputDir(), "terraform.tf")
	if err := ioutil.WriteFile(path, []byte(gce), 0666); err != nil {
		c.Fatalf("writing config failed: %v", err)
	}
}

func (c *Cluster) terraform(arg ...string) *exec.Cmd {
	arg = append(arg, "-no-color")
	cmd := exec.CommandContext(c.Context(), "terraform", arg...)
	cmd.Dir = c.OutputDir()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		c.Fatalf("opening pipe failed: %v", err)
	}

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			c.Log(scanner.Text())
		}
		if scanner.Err() != nil {
			c.Errorf("reading pipe failed: %v", err)
		}
	}()

	return cmd
}

func (c *Cluster) readOutputs() {
	cmd := c.terraform("output", "-json")
	data, err := cmd.Output()
	if err != nil {
		c.Fatalf("terraform output failed: %v", err)
	}

	var parsed tfOutput
	if err := tfdata.ParseOutput(data, &parsed); err != nil {
		c.Fatalf("Cannot parse terraform output: %v\n%s", err, data)
	}

	n := c.test.ClusterSize
	if len(parsed.Names.Value) != n ||
		len(parsed.LocalIPs.Value) != n ||
		len(parsed.RemoteIPs.Value) != n {
		c.Fatalf("List lengths are not %d: %#v", n, parsed)
	}

	c.machines = make([]platform.Machine, n)
	for i := 0; i < n; i++ {
		c.machines[i] = &machine{
			cluster:  c,
			name:     parsed.Names.Value[i],
			localIP:  parsed.LocalIPs.Value[i],
			remoteIP: parsed.RemoteIPs.Value[i],
		}
	}
}

func (c *Cluster) setup() {
	agent, err := network.NewSSHAgent(network.NewRetryDialer())
	if err != nil {
		c.Fatalf("ssh agent failed: %v", err)
	}
	c.sshAgent = agent

	c.writeInputs()
	c.writeUserData()
	c.writeConfig()

	tf := c.terraform("apply", "-input=false")
	tf.Stdout = tf.Stderr // log both stdout and stderr
	if err := tf.Run(); err != nil {
		c.Fatalf("terraform apply failed: %v", err)
	}

	c.readOutputs()

	// TODO: parallel
	for _, m := range c.machines {
		if err := m.(*machine).setup(); err != nil {
			c.Fatalf("instance %s failed: %v", m.ID(), err)
		}
	}
}

func (c *Cluster) destroy() {
	tf := c.terraform("destroy", "-force")
	tf.Stdout = tf.Stderr // log both stdout and stderr
	if err := tf.Run(); err != nil {
		c.Errorf("terraform destroy failed: %v", err)
	}

	// free up local resources, no need to be parallel.
	for _, m := range c.machines {
		if err := m.(*machine).destroy(); err != nil {
			c.Errorf("instance %s failed: %v", m.ID(), err)
		}
	}
	c.machines = nil

	if c.sshAgent != nil {
		if err := c.sshAgent.Close(); err != nil {
			c.Errorf("ssh agent failed: %v", err)
		}
		c.sshAgent = nil
	}
}

func (c *Cluster) Machines() []platform.Machine {
	return c.machines
}

// currently adding machines on the fly isn't implemented
func (c *Cluster) NewMachine(config string) (platform.Machine, error) {
	panic("NewMachine not supported")
}

// TODO and/or possibly removed from the Cluster interface
func (c *Cluster) GetDiscoveryURL(size int) (string, error) {
	panic("GetDiscoveryURL not supported")
}

// cluster destruction is tied to test execution so no public method
func (c *Cluster) Destroy() error {
	panic("Destroy not supported")
}
