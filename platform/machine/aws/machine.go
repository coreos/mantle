// Copyright 2016 CoreOS, Inc.
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

package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"golang.org/x/crypto/ssh"

	"github.com/coreos/mantle/platform"
	"github.com/coreos/mantle/platform/api/aws"
	"github.com/coreos/mantle/platform/conf"
	"github.com/coreos/mantle/platform/util"
)

type awsMachine struct {
	cluster *awsCluster
	mach    *ec2.Instance
}

func (am *awsMachine) ID() string {
	return *am.mach.InstanceId
}

func (am *awsMachine) IP() string {
	return *am.mach.PublicIpAddress
}

func (am *awsMachine) PrivateIP() string {
	return *am.mach.PrivateIpAddress
}

func (am *awsMachine) SSHClient() (*ssh.Client, error) {
	return am.cluster.SSHClient(am.IP())
}

func (am *awsMachine) PasswordSSHClient(user string, password string) (*ssh.Client, error) {
	return am.cluster.PasswordSSHClient(am.IP(), user, password)
}

func (am *awsMachine) SSH(cmd string) ([]byte, error) {
	return am.cluster.SSH(am, cmd)
}

func (am *awsMachine) Destroy() error {
	// XXX: defer DelMach here?
	if err := am.cluster.api.TerminateInstance(am.ID()); err != nil {
		return err
	}

	am.cluster.DelMach(am)
	return nil
}

type awsCluster struct {
	*platform.BaseCluster
	api     *aws.API
	options *aws.Options
}

// NewCluster creates an instance of a Cluster suitable for spawning
// instances on Amazon Web Services' Elastic Compute platform.
//
// NewCluster will consume the environment variables $AWS_REGION,
// $AWS_ACCESS_KEY_ID, and $AWS_SECRET_ACCESS_KEY to determine the region to
// spawn instances in and the credentials to use to authenticate.
func NewCluster(opts *aws.Options) (platform.Cluster, error) {
	api, err := aws.New(opts)
	if err != nil {
		return nil, err
	}

	bc, err := platform.NewBaseCluster(opts.BaseName)
	if err != nil {
		return nil, err
	}

	var ac platform.Cluster

	ac = &awsCluster{
		BaseCluster: bc,
		api:         api,
		options:     opts,
	}

	keys, err := ac.Keys()
	if err != nil {
		return nil, err
	}

	if err := api.AddKey(bc.Name(), keys[0].String()); err != nil {
		return nil, err
	}

	return ac, nil
}

func (ac *awsCluster) NewMachine(userdata string) (platform.Machine, error) {
	conf, err := conf.New(userdata)
	if err != nil {
		return nil, err
	}

	keys, err := ac.Keys()
	if err != nil {
		return nil, err
	}

	conf.CopyKeys(keys)

	instances, err := ac.api.CreateInstances(ac.options.AMI, ac.Name(), conf.String(), ac.options.InstanceType, ac.options.SecurityGroup, 1, true)

	mach := &awsMachine{
		cluster: ac,
		mach:    instances[0],
	}

	if err := util.CommonMachineChecks(mach); err != nil {
		return nil, fmt.Errorf("machine %q failed basic checks: %v", mach.ID(), err)
	}

	ac.AddMach(mach)

	return mach, nil
}

func (ac *awsCluster) Destroy() error {
	if err := ac.api.DeleteKey(ac.Name()); err != nil {
		return err
	}

	return ac.BaseCluster.Destroy()
}
