package etcdtests

import (
	"fmt"
	"os"

	"github.com/coreos/mantle/platform"
)

// setup etcd 0.4.x cluster and add some values -- shutdown and start etcd
// 2.0.x binary -- does it fallback to etcd 0.4.x binary and retain
// data?
func Fallback(cluster platform.Cluster) error {
	const (
		testKey   = "foo"
		testValue = "etcdFallback"
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

	// stop etcd 0.4.x
	err = stopEtcd(cluster)
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

	fmt.Fprintf(os.Stderr, "bringing up etcd2.0...\n")
	// start etcd using new binary
	err = startEtcd(cluster)
	if err != nil {
		return err
	}
	// check cluster health
	err = getClusterHealth(cluster)
	if err != nil {
		return fmt.Errorf("discovery failed health check: %v", err)
	}

	// check that etcd 2 binary falls back to internal version 1
	version, err := getEtcdInternalVersion(cluster)
	if err != nil {
		return err
	}
	if version != 1 {
		return fmt.Errorf("etcd did not fallback to version 1, got version: %v", version)
	}
	fmt.Fprintf(os.Stderr, "etcd detected version 1 WAL and has fallen back to version 1\n")

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
