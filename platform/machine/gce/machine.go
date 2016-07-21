// Copyright 2016 CoreOS, Inc.
// Copyright 2015 The Go Authors.
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

package gce

import (
	"golang.org/x/crypto/ssh"

	"github.com/coreos/mantle/platform"
	"github.com/coreos/mantle/platform/api/gce"
	"github.com/coreos/mantle/platform/conf"
	"github.com/coreos/mantle/platform/util"
)

type gceCluster struct {
	*platform.BaseCluster
	api     *gce.API
	options *gce.Options
}

type gceMachine struct {
	gc    *gceCluster
	name  string
	intIP string
	extIP string
}

func NewCluster(opts *gce.Options) (platform.Cluster, error) {
	api, err := gce.New(opts)
	if err != nil {
		return nil, err
	}

	bc, err := platform.NewBaseCluster(opts.BaseName)
	if err != nil {
		return nil, err
	}

	gc := &gceCluster{
		BaseCluster: bc,
		api:         api,
		options:     opts,
	}

	return gc, nil
}

// Calling in parallel is ok
func (gc *gceCluster) NewMachine(userdata string) (platform.Machine, error) {
	conf, err := conf.New(userdata)
	if err != nil {
		return nil, err
	}

	keys, err := gc.Keys()
	if err != nil {
		return nil, err
	}

	conf.CopyKeys(keys)

	// Create gce VM and wait for creation to succeed.
	instance, err := gc.api.CreateInstance(conf.String(), keys)
	if err != nil {
		return nil, err
	}

	intIP, extIP := gce.InstanceIPs(instance)

	gm := &gceMachine{
		gc:    gc,
		name:  instance.Name,
		extIP: extIP,
		intIP: intIP,
	}

	if err := util.CommonMachineChecks(gm); err != nil {
		gm.Destroy()
		return nil, err
	}

	gc.AddMach(gm)

	return gm, nil
}

func (gm *gceMachine) ID() string {
	return gm.name
}

func (gm *gceMachine) IP() string {
	return gm.extIP
}

func (gm *gceMachine) PrivateIP() string {
	return gm.intIP
}

func (gm *gceMachine) SSHClient() (*ssh.Client, error) {
	return gm.gc.SSHClient(gm.IP())
}

func (gm *gceMachine) PasswordSSHClient(user string, password string) (*ssh.Client, error) {
	return gm.gc.PasswordSSHClient(gm.IP(), user, password)
}

func (gm *gceMachine) SSH(cmd string) ([]byte, error) {
	return gm.gc.SSH(gm, cmd)
}

func (gm *gceMachine) Destroy() error {
	if err := gm.gc.api.TerminateInstance(gm.name); err != nil {
		return err
	}

	gm.gc.DelMach(gm)
	return nil
}
