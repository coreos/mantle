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

package azure

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/management"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/coreos/pkg/capnslog"
)

var (
	plog = capnslog.NewPackageLogger("github.com/coreos/mantle", "platform/api/azure")
)

type API struct {
	client management.Client
	opts   *Options
}

// New creates a new Azure client. If no publish settings file is provided or
// can't be parsed, an anonymous client is created.
func New(opts *Options) (*API, error) {
	conf := management.DefaultConfig()
	conf.APIVersion = "2015-04-01"

	if opts.ManagementURL != "" {
		conf.ManagementURL = opts.ManagementURL
	}

	if opts.StorageEndpointSuffix == "" {
		opts.StorageEndpointSuffix = storage.DefaultBaseURL
	}

	client, err := management.NewClientFromConfig(opts.SubscriptionID, opts.ManagementCertificate, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure client: %v", err)
	}

	api := &API{
		client: client,
		opts:   opts,
	}

	return api, nil
}
