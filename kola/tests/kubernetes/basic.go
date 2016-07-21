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

package kubernetes

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/mantle/kola/register"
	"github.com/coreos/mantle/platform"
	putil "github.com/coreos/mantle/platform/util"
	"github.com/coreos/mantle/util"
)

// register a separate test for each version tag
var basicTags = []string{
	"v1.2.0_coreos.1",
}

func init() {
	for i := range basicTags {
		// use closure to store a version tag in a Test
		t := basicTags[i]
		f := func(c platform.TestCluster) error {
			return CoreOSBasic(c, t)
		}

		register.Register(&register.Test{
			Name:        "google.kubernetes.basic." + t,
			Run:         f,
			ClusterSize: 0,
			Platforms:   []string{"gce"},
		})
	}
}

// Run basic smoke tests on cluster. Assumes master is machine index 1,
// workers make up the rest.
func CoreOSBasic(c platform.TestCluster, version string) error {
	k, err := setupCluster(c, 2, version)
	if err != nil {
		return err
	}

	// start nginx pod and curl endpoint
	if err := nginxCheck(k.master, k.workers); err != nil {
		return err
	}

	// http://kubernetes.io/v1.0/docs/user-guide/secrets/ Also, ensures
	// https://github.com/coreos/bugs/issues/447 does not re-occur.
	if err := secretCheck(k.master, k.workers); err != nil {
		return err
	}

	return nil
}

func nodeCheck(master platform.Machine, nodes []platform.Machine) error {
	b, err := master.SSH("./kubectl get nodes")
	if err != nil {
		return err
	}

	// parse kubectl output for IPs
	addrMap := map[string]struct{}{}
	for _, line := range strings.Split(string(b), "\n")[1:] {
		addrMap[strings.SplitN(line, " ", 2)[0]] = struct{}{}
	}

	// add master to node list because it is now normal to register
	// master nodes but have it set as unschedulable in kubernetes v1.2+
	nodes = append(nodes, master)

	if len(addrMap) != len(nodes) {
		return fmt.Errorf("cannot detect all nodes in kubectl output \n%v\n%v", addrMap, nodes)
	}
	for _, node := range nodes {
		if _, ok := addrMap[node.PrivateIP()]; !ok {
			return fmt.Errorf("node IP missing from kubectl get nodes")
		}
	}
	return nil
}

func nginxCheck(master platform.Machine, nodes []platform.Machine) error {
	pod := strings.NewReader(nginxPodYAML)
	if err := putil.InstallFile(pod, master, "./nginx-pod.yaml"); err != nil {
		return err
	}
	if _, err := master.SSH("./kubectl create -f nginx-pod.yaml"); err != nil {
		return err
	}
	// wait for pod status to be 'Running'
	podIsRunning := func() error {
		b, err := master.SSH("./kubectl get pod nginx --template={{.status.phase}}")
		if err != nil {
			return err
		}
		if !bytes.Equal(b, []byte("Running")) {
			return fmt.Errorf("nginx pod not running: %s", b)
		}
		return nil
	}
	if err := util.Retry(10, 10*time.Second, podIsRunning); err != nil {
		return err
	}

	// delete pod
	_, err := master.SSH("./kubectl delete pods nginx")
	if err != nil {
		return err
	}

	return nil
}

func secretCheck(master platform.Machine, nodes []platform.Machine) error {
	// create yaml files
	secret := strings.NewReader(secretYAML)
	pod := strings.NewReader(secretPodYAML)
	if err := putil.InstallFile(secret, master, "./secret.yaml"); err != nil {
		return err
	}
	if err := putil.InstallFile(pod, master, "./secret-pod.yaml"); err != nil {
		return err
	}

	if _, err := master.SSH("./kubectl create -f secret.yaml"); err != nil {
		return err
	}
	_, err := master.SSH("./kubectl describe secret test-secret")
	if err != nil {
		return err
	}

	b, err := master.SSH("./kubectl create -f secret-pod.yaml")
	if err != nil {
		return err
	}
	expectedOutput := "value-1"
	if strings.Contains(strings.TrimSpace(string(b)), expectedOutput) {
		return fmt.Errorf("error detecting secret pod")
	}

	return nil
}

const (
	secretPodYAML = `apiVersion: v1
kind: Pod
metadata:
  name: secret-test-pod
spec:
  containers:
    - name: test-container
      image: kubernetes/mounttest:0.1
      command: [ "/mt", "--file_content=/etc/secret-volume/data-1" ]
      volumeMounts:
          # name must match the volume name below
          - name: secret-volume
            mountPath: /etc/secret-volume
  volumes:
    - name: secret-volume
      secret:
        secretName: test-secret
  restartPolicy: Never`

	secretYAML = `apiVersion: v1
kind: Secret
metadata:
  name: test-secret
data:
  data-1: dmFsdWUtMQ0K
  data-2: dmFsdWUtMg0KDQo=`

	nginxPodYAML = `apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  containers:
  - name: nginx
    image: nginx
    ports:
    - containerPort: 80`
)
