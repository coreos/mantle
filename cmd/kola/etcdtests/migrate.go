package etcdtests

import (
	"fmt"
	"os"
	"time"

	"github.com/coreos/mantle/platform"
)

// setup etcd 0.4.x cluster and add some values -- shutdown and migrate
// data to version 2 format --start etcd version 2 -- does it stay on
// version 2 and have access to the same data?
func Migrate(cluster platform.Cluster) error {
	const (
		testKey   = "foo"
		testValue = "etcdMigrate"
	)

	// output journalctl -f from only one machine
	err := cluster.Machines()[0].StartJournal()
	if err != nil {
		return fmt.Errorf("Failed to start journal: %v", err)
	}

	// start etcd 0.4.x
	err = startEtcd(cluster)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Waiting for discovery to finish...\n")

	// get status of etcd instances -- retry until satisfied
	err = getClusterHealth(cluster)
	if err != nil {
		return fmt.Errorf("discovery failed health check: %v", err)
	}

	// set key
	err = setKey(cluster, testKey, testValue)
	if err != nil {
		return err
	}

	//replace etcd 0.4.x binary with default etcd bin to use starter
	const (
		oldp = "/usr/libexec/etcd/internal_versions/1/etcd"
		newp = "/usr/bin/etcd"
	)
	err = replaceEtcdBin(cluster, oldp, newp)
	if err != nil {
		return err
	}

	//migrate data from version 1 to version 2
	for _, m := range cluster.Machines() {
		b, err := m.SSH("etcdctl upgrade --peer-url http://127.0.0.1:7001")
		if err != nil {
			return fmt.Errorf("failure updating cluster: %v %s", err, b)
		}
		fmt.Fprintf(os.Stderr, "%s\n", b)
	}

	// etcdctl upgrade waits 10 seconds before restarting so sleep 20
	// before polling
	time.Sleep(20 * time.Second)

	// check cluster health
	err = getClusterHealth(cluster)
	if err != nil {
		return fmt.Errorf("discovery failed health check: %v", err)
	}

	// check running version
	version, err := getEtcdInternalVersion(cluster)
	if err != nil {
		return err
	}
	if version != 2 {
		return fmt.Errorf("etcd did not upgrade to version 2, got version: %v", version)
	}
	fmt.Fprintf(os.Stderr, "etcd upgraded to version 2\n")

	// check previously set key
	value, err := getKey(cluster, testKey)
	if err != nil {
		return err
	}
	if value != testValue {
		return fmt.Errorf("error getting previously set key: got %v instead of %v", value, testValue)
	}

	return nil
}
