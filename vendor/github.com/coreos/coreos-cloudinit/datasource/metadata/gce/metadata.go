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

package gce

import (
	"fmt"
	"net"
	"net/http"

	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/datasource/metadata"
)

const (
	apiVersion   = "computeMetadata/v1/"
	metadataPath = apiVersion + "instance/"
	userdataPath = apiVersion + "instance/attributes/user-data"
)

type metadataService struct {
	metadata.MetadataService
}

func NewDatasource(root string) *metadataService {
	return &metadataService{metadata.NewDatasource(root, apiVersion, userdataPath, metadataPath, http.Header{"Metadata-Flavor": {"Google"}})}
}

func (ms metadataService) FetchMetadata() (datasource.Metadata, error) {
	public, err := ms.fetchIP("network-interfaces/0/access-configs/0/external-ip")
	if err != nil {
		return datasource.Metadata{}, err
	}
	local, err := ms.fetchIP("network-interfaces/0/ip")
	if err != nil {
		return datasource.Metadata{}, err
	}
	hostname, err := ms.fetchString("hostname")
	if err != nil {
		return datasource.Metadata{}, err
	}

	return datasource.Metadata{
		PublicIPv4:  public,
		PrivateIPv4: local,
		Hostname:    hostname,
	}, nil
}

func (ms metadataService) Type() string {
	return "gce-metadata-service"
}

func (ms metadataService) fetchString(key string) (string, error) {
	data, err := ms.FetchData(ms.MetadataUrl() + key)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (ms metadataService) fetchIP(key string) (net.IP, error) {
	str, err := ms.fetchString(key)
	if err != nil {
		return nil, err
	}

	if str == "" {
		return nil, nil
	}

	if ip := net.ParseIP(str); ip != nil {
		return ip, nil
	} else {
		return nil, fmt.Errorf("couldn't parse %q as IP address", str)
	}
}
