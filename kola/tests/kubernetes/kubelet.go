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

package kubernetes

import (
	"bytes"
	"fmt"
	"time"

	"github.com/vincent-petithory/dataurl"

	"github.com/coreos/mantle/kola/cluster"
	"github.com/coreos/mantle/kola/register"
	"github.com/coreos/mantle/platform/conf"
	"github.com/coreos/mantle/util"
)

var (
	kubeletConfigYAML = dataurl.Escape([]byte(`kind: KubeletConfiguration
apiVersion: kubelet.config.k8s.io/v1beta1
clusterDomain: kube.local
clusterDNS: [10.254.0.10]
cgroupDriver: systemd
staticPodPath: /etc/kubernetes/manifests/
authorization:
  mode: AlwaysAllow
authentication:
  anonymous:
    enabled: true
  webhook:
    enabled: false`))

	staticPodYAML = dataurl.Escape([]byte(`apiVersion: v1
kind: Pod
metadata:
  name: static-web
  labels:
    role: myrole
spec:
  containers:
    - name: web
      image: docker.io/nginx
      ports:
        - name: web
          containerPort: 80
          hostPort: 80`))

	kubeletUnit = `[Unit]
Description=Kubernetes Kubelet
Requires=crio.service
After=crio.service

[Service]
ExecStart=/usr/bin/hyperkube kubelet --config /etc/kubernetes/kubeletconfig --container-runtime=remote --container-runtime-endpoint=unix:///var/run/crio/crio.sock --runtime-request-timeout=10m

Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target`

	staticPodIgnitionConfig = conf.Ignition(fmt.Sprintf(`{
    "ignition": {
        "config": {},
        "security": {
            "tls": {}
        },
        "timeouts": {},
        "version": "2.2.0"
    },
    "networkd": {},
    "passwd": {},
    "storage": {
        "files": [
            {
                "contents": {
                    "source": "data:,%s",
                    "verification": {}
                },
                "filesystem": "root",
                "mode": 420,
                "path": "/etc/kubernetes/kubeletconfig"
            },
            {
                "contents": {
                    "source": "data:,%s",
                    "verification": {}
                },
                "filesystem": "root",
                "mode": 420,
                "path": "/etc/kubernetes/manifests/static-pod.yaml"
            }
        ]
    },
    "systemd": {
        "units": [
            {
                "contents": %q,
                "enabled": true,
                "name": "kubelet.service"
            }
        ]
    }
}`, kubeletConfigYAML, staticPodYAML, kubeletUnit))
)

func init() {
	// Test: verify kubelet can schedule a static pod
	register.Register(&register.Test{
		Name:             "kubernetes.kubelet.static-pod",
		Run:              kubeletStaticPodTest,
		ClusterSize:      1,
		UserData:         staticPodIgnitionConfig,
		ExcludePlatforms: []string{"qemu"}, // These tests require networking
		Distros:          []string{"rhcos"},
	})
}

func kubeletStaticPodTest(c cluster.TestCluster) {
	m := c.Machines()[0]

	podIsRunning := func() error {
		b, err := c.SSH(m, `curl -f http://localhost 2>/dev/null`)
		if err != nil {
			return err
		}
		if !bytes.Contains(b, []byte("nginx")) {
			return fmt.Errorf("nginx pod is not running %s", b)
		}
		return nil
	}

	if err := util.Retry(10, 10*time.Second, podIsRunning); err != nil {
		c.Fatal(err)
	}
}
