package platform

import (
	"fmt"
	"time"

	"github.com/coreos/mantle/util"

	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/packethost/packngo"
	"github.com/coreos/mantle/Godeps/_workspace/src/golang.org/x/crypto/ssh"
)

type packetMachine struct {
	cluster *packetCluster
	mach    *packngo.Device
}

func (pm *packetMachine) ID() string {
	return pm.mach.ID
}

func (pm *packetMachine) IP() string {
	for _, ip := range pm.mach.Network {
		if ip.Family == 4 && ip.Public {
			return ip.Address
		}
	}

	// really should error
	return ""
}

func (pm *packetMachine) PrivateIP() string {
	for _, ip := range pm.mach.Network {
		if ip.Family == 4 && !ip.Public {
			return ip.Address
		}
	}

	// really should error
	return ""
}

func (pm *packetMachine) SSHClient() (*ssh.Client, error) {
	return pm.cluster.SSHClient(pm.IP())
}

func (pm *packetMachine) SSH(cmd string) ([]byte, error) {
	return pm.cluster.SSH(pm, cmd)
}

func (pm *packetMachine) Destroy() error {
	_, err := pm.cluster.api.Devices.Delete(pm.ID())
	return err
}

type PacketOptions struct {
	APIKey  string
	Project string
	Plan    string
	OS      string
}

type packetCluster struct {
	*baseCluster
	api  *packngo.Client
	conf PacketOptions
}

func NewPacketCluster(conf PacketOptions) (Cluster, error) {
	api := packngo.NewClient("mantle", conf.APIKey, nil)

	bc, err := newBaseCluster()
	if err != nil {
		return nil, err
	}

	pc := &packetCluster{
		baseCluster: bc,
		api:         api,
		conf:        conf,
	}

	return pc, nil
}

func (pc *packetCluster) NewMachine(userdata string) (Machine, error) {
	conf, err := NewConf(userdata)
	if err != nil {
		return nil, err
	}

	keys, err := pc.agent.List()
	if err != nil {
		return nil, err
	}

	conf.CopyKeys(keys)

	// get project, if it doesn't exist make it
	projects, _, err := pc.api.Projects.List()
	if err != nil {
		return nil, fmt.Errorf("failed listing projects: %v", err)
	}

	var projectID string
	for _, p := range projects {
		if p.Name == pc.conf.Project {
			projectID = p.ID
		}
	}

	var proj *packngo.Project

	if projectID != "" {
		proj, _, err = pc.api.Projects.Get(projectID)
	} else {
		projCreate := &packngo.ProjectCreateRequest{
			Name: pc.conf.Project,
		}

		proj, _, err = pc.api.Projects.Create(projCreate)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get packet project: %v", err)
	}

	devCreate := &packngo.DeviceCreateRequest{
		HostName:     "coreos",
		Plan:         pc.conf.Plan,
		Facility:     "ewr1",
		OS:           pc.conf.OS,
		BillingCycle: "hourly",
		ProjectID:    proj.ID,
		UserData:     conf.String(),
	}

	dev, _, err := pc.api.Devices.Create(devCreate)
	if err != nil {
		return nil, fmt.Errorf("failed to create device: %v", err)
	}

	devID := dev.ID

	deviceChecker := func() error {
		dev, _, err = pc.api.Devices.Get(devID)
		if err != nil {
			return err
		}

		if dev.State != "active" {
			return fmt.Errorf("device state not active")
		}
		return nil
	}

	// i was told provisioning can take ~5 minutes, so be safe and shoot for 10..
	err = util.Retry(15, 20*time.Second, deviceChecker)
	if err != nil {
		// terminate just in case
		pc.api.Devices.Delete(devID)
		return nil, fmt.Errorf("failed checking packet device: %v", err)
	}

	mach := &packetMachine{
		cluster: pc,
		mach:    dev,
	}

	if err := commonMachineChecks(mach); err != nil {
		return nil, fmt.Errorf("machine %q failed basic checks: %v", mach.ID(), err)
	}

	pc.addMach(mach)

	return mach, nil
}

func (pc *packetCluster) Destroy() error {
	machs := pc.Machines()
	for _, pm := range machs {
		pm.Destroy()
	}
	pc.agent.Close()

	// XXX: destroying the project *should* destroy all the devices too, so maybe do that here.
	return nil
}
