// Copyright 2015 CoreOS, Inc.
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

package platform

import (
	"net"
)

type Metadata struct {
	Platform  string `json:"platform"`
	ID        string `json:"id"`
	PublicIP  net.IP `json:"public_ip"`
	PrivateIP net.IP `json:"private_ip"`
}

func MachineMetadata(m Machine) (*Metadata, error) {
	meta := &Metadata{
		ID:        m.ID(),
		PublicIP:  net.ParseIP(m.IP()),
		PrivateIP: net.ParseIP(m.PrivateIP()),
	}

	return meta, nil
}
