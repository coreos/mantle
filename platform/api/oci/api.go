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

package oci

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/identity"
	"github.com/oracle/oci-go-sdk/objectstorage"

	"github.com/coreos/mantle/auth"
	"github.com/coreos/mantle/platform"
)

type Options struct {
	*platform.Options

	ConfigPath string
	Profile    string

	TenancyID          string
	UserID             string
	Fingerprint        string
	KeyFile            string
	PrivateKeyPassword *string
	Region             string

	CompartmentID string
	Image         string
	Shape         string
}

type Machine struct {
	Name             string
	ID               string
	PublicIPAddress  string
	PrivateIPAddress string
}

type API struct {
	config   common.ConfigurationProvider
	compute  core.ComputeClient
	identity identity.IdentityClient
	os       objectstorage.ObjectStorageClient
	vn       core.VirtualNetworkClient
	opts     *Options
}

func New(opts *Options) (*API, error) {
	profiles, err := auth.ReadOCIConfig(opts.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't read OCI profile: %v", err)
	}

	if opts.Profile == "" {
		opts.Profile = "default"
	}

	profile, ok := profiles[opts.Profile]
	if !ok {
		return nil, fmt.Errorf("no such profile %q", opts.Profile)
	}

	if opts.TenancyID == "" {
		opts.TenancyID = profile.TenancyID
	}
	if opts.UserID == "" {
		opts.UserID = profile.UserID
	}
	if opts.Fingerprint == "" {
		opts.Fingerprint = profile.Fingerprint
	}
	if opts.KeyFile == "" {
		opts.KeyFile = profile.KeyFile
	}
	if opts.PrivateKeyPassword == nil {
		opts.PrivateKeyPassword = profile.PrivateKeyPassword
	}
	if opts.Region == "" {
		opts.Region = profile.Region
	}
	if opts.CompartmentID == "" {
		opts.CompartmentID = profile.CompartmentID
	}

	privateKey, err := ioutil.ReadFile(opts.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("reading RSA private key: %v", err)
	}

	config := common.NewRawConfigurationProvider(opts.TenancyID, opts.UserID, opts.Region, opts.Fingerprint, string(privateKey), opts.PrivateKeyPassword)

	computeClient, err := core.NewComputeClientWithConfigurationProvider(config)
	if err != nil {
		return nil, fmt.Errorf("creating compute client: %v", err)
	}

	objectClient, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(config)
	if err != nil {
		return nil, fmt.Errorf("creating objectstorage client: %v", err)
	}

	vnClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(config)
	if err != nil {
		return nil, fmt.Errorf("creating virtual network client: %v", err)
	}

	idClient, err := identity.NewIdentityClientWithConfigurationProvider(config)
	if err != nil {
		return nil, fmt.Errorf("creating identity client: %v", err)
	}

	return &API{
		config:   config,
		compute:  computeClient,
		identity: idClient,
		os:       objectClient,
		vn:       vnClient,
		opts:     opts,
	}, nil
}

func (a *API) GC(gracePeriod time.Duration) error {
	return a.gcInstances(gracePeriod)
}

func boolToPtr(b bool) *bool {
	return &b
}

func strToPtr(s string) *string {
	return &s
}

func intToPtr(i int) *int {
	return &i
}
