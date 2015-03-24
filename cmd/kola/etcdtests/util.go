package etcdtests

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/mantle/platform"
)

// run etcd on each cluster machine
func startEtcd(cluster platform.Cluster) error {
	for i, m := range cluster.Machines() {
		etcdStart := "sudo systemctl start etcd.service"
		_, err := m.SSH(etcdStart)
		if err != nil {
			return fmt.Errorf("start etcd.service on %v failed: %s", m.IP(), err)
		}
		fmt.Fprintf(os.Stderr, "etcd instance%d started\n", i)
	}
	return nil
}

// stop etcd on each cluster machine
func stopEtcd(cluster platform.Cluster) error {
	for i, m := range cluster.Machines() {
		// start etcd instance
		etcdStop := "sudo systemctl stop etcd.service"
		_, err := m.SSH(etcdStop)
		if err != nil {
			return fmt.Errorf("stop etcd.service on %v failed: %s", m.IP(), err)
		}
		fmt.Fprintf(os.Stderr, "etcd instance%d stopped\n", i)
	}
	return nil
}

type Key struct {
	Node struct {
		Value string `json:"value"`
	} `json:"node"`
}

func setKey(cluster platform.Cluster, key, value string) error {
	cmd := cluster.NewCommand("curl", "-L", "http://10.0.0.3:4001/v2/keys/"+key, "-XPUT", "-d", "value="+value)
	b, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Error setting value to cluster: %v", err)
	}
	fmt.Fprintf(os.Stderr, "%s\n", b)

	var k Key
	err = json.Unmarshal(b, &k)
	if err != nil {
		return err
	}
	if k.Node.Value != value {
		fmt.Errorf("etcd key not set correctly")
	}
	return nil
}

func getKey(cluster platform.Cluster, key string) (string, error) {
	cmd := cluster.NewCommand("curl", "-L", "http://10.0.0.3:4001/v2/keys/"+key)
	b, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error getting old value from cluster: %v", err)
	}
	fmt.Fprintf(os.Stderr, "%s\n", b)

	var k Key
	err = json.Unmarshal(b, &k)
	if err != nil {
		return "", err
	}

	return k.Node.Value, nil
}

// replace default binary for etcd.service with internal version v
func replaceEtcdBin(cluster platform.Cluster, oldv, newv string) error {
	// escape file path forward-slashes for sed command
	oldv = strings.Replace(oldv, "/", `\/`, -1)
	newv = strings.Replace(newv, "/", `\/`, -1)
	fmt.Println("%v \n%v\n", oldv, newv)

	for _, m := range cluster.Machines() {
		// first err collector
		var err error
		firstErr := func(b []byte, e error) {
			if e != nil && err == nil {
				err = e
			}
		}
		firstErr(m.SSH(fmt.Sprintf("sed s/\"%v\"/\"%v\"/ /run/systemd/system/etcd.service.d/30-exec.conf > ./etcd.service.conf", oldv, newv)))
		firstErr(m.SSH("sudo mkdir -p /etc/systemd/system/etcd.service.d"))
		firstErr(m.SSH("sudo cp ./etcd.service.conf /etc/systemd/system/etcd.service.d/etcd.conf"))
		firstErr(m.SSH("sudo systemctl daemon-reload"))
		firstErr(m.SSH("sudo cat /etc/systemd/system/etcd.service.d/etcd.conf"))
		if err != nil {
			return fmt.Errorf("failed changing etcd version: %v", err)
		}
	}
	return nil
}

func getEtcdInternalVersion(cluster platform.Cluster) (int, error) {
	cmd := cluster.NewCommand("curl", "-L", "http://10.0.0.2:4001/version")
	b, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("error curling version: %v", err)
	}

	type Version struct {
		Internal string `json:"internalVersion"`
	}
	var v Version

	err = json.Unmarshal(b, &v)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(v.Internal)
}

// poll etcdctl cluster-health from first machine in cluster
func getClusterHealth(cluster platform.Cluster) error {
	const (
		retries   = 5
		retryWait = 3 * time.Second
	)
	var err error
	var b []byte
	machine := cluster.Machines()[0]

	for i := 0; i < retries; i++ {
		fmt.Fprintf(os.Stderr, "polling cluster health...\n")
		b, err = machine.SSH("etcdctl cluster-health")
		if err == nil {
			break
		}
		time.Sleep(retryWait)
	}
	if err != nil {
		return fmt.Errorf("health polling failed: %s", b)
	}

	// repsonse should include "healthy" for each machine and for cluster
	if strings.Count(string(b), "healthy") == len(cluster.Machines())+1 {
		fmt.Fprintf(os.Stderr, "%s\n", b)
		return nil
	} else {
		return fmt.Errorf("status unhealthy or incomplete: %s", b)
	}
}
