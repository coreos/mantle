package etcdtests

import (
	"fmt"

	"github.com/coreos/mantle/platform"
)

func Discovery(cluster platform.Cluster) error {
	// get journalctl -f from all machines before starting
	for _, m := range cluster.Machines() {
		if err := m.StartJournal(); err != nil {
			return fmt.Errorf("Failed to start journal: %v", err)
		}
	}

	// start etcd cluster
	err := startEtcd(cluster)
	if err != nil {
		return err
	}

	// check health on first machine
	err = getClusterHealth(cluster)
	if err != nil {
		return fmt.Errorf("discovery failed health check: %v", err)
	}

	return nil
}
