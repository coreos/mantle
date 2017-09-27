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

package types

import (
	"encoding/json"
)

type LaunchInstanceInput struct {
	AvailabilityDomain string             `json:"availabilityDomain"`
	CompartmentID      string             `json:"compartmentId"`
	CreateVNICDetails  *VNICDetails       `json:"createVnicDetails,omitempty"`
	DisplayName        *string            `json:"displayName,omitempty"`
	HostnameLabel      *string            `json:"hostnameLabel,omitempty"`
	ImageID            string             `json:"imageId"`
	IPXEScript         *string            `json:"ipxeScript,omitempty"`
	Metadata           *map[string]string `json:"metadata,omitempty"`
	Shape              string             `json:"shape"`
	SubnetID           *string            `json:"subnetId,omitempty"`

	// The ExtendedMetadata property allows nested JSON objects.
	// It is strongly advised that the Metadata attribute is used
	// instead.
	ExtendedMetadata *json.RawMessage `json:"extendedMetadata,omitempty"`
}

type ListInstancesInput struct {
	AvailabilityDomain *string `json:"availabilityDomain,omitempty"`
	CompartmentID      string  `json:"compartmentId"`
	DisplayName        *string `json:"displayName,omitempty"`
}

type InstanceConsoleConnectionInput struct {
	InstanceID string `json:"instanceId"`
	PublicKey  string `json:"publicKey"`
}

type CaptureConsoleHistoryInput struct {
	InstanceID string `json:"instanceId"`
}

type Shape struct {
	Shape string `json:"shape"`
}

type InstanceConsoleConnection struct {
	CompartmentID    *string `json:"compartmentId"`
	ConnectionString *string `json:"connectionString"`
	Fingerprint      *string `json:"fingerprint"`
	ID               *string `json:"id"`
	InstanceID       *string `json:"instanceId"`
	LifecycleState   *string `json:"lifecycleState"`
}

type Instance struct {
	AvailabilityDomain string             `json:"availabilityDomain"`
	CompartmentID      string             `json:"compartmentId"`
	DisplayName        *string            `json:"displayName"`
	ID                 string             `json:"id"`
	ImageID            *string            `json:"imageId"`
	IPXEScript         *string            `json:"ipxeScript"`
	LifecycleState     string             `json:"lifecycleState"`
	Metadata           *map[string]string `json:"metadata"`
	Region             string             `json:"region"`
	Shape              string             `json:"shape"`
	TimeCreated        string             `json:"timeCreated"`

	// The ExtendedMetadata property allows nested JSON objects.
	// It is strongly advised that the Metadata attribute is used
	// instead.
	ExtendedMetadata *json.RawMessage `json:"extendedMetadata"`
}

type ConsoleHistory struct {
	AvailabilityDomain string  `json:"availabilityDomain"`
	CompartmentID      string  `json:"compartmentId"`
	DisplayName        *string `json:"displayName"`
	ID                 string  `json:"id"`
	InstanceID         string  `json:"instanceId"`
	LifecycleState     string  `json:"lifecycleState"`
	TimeCreated        string  `json:"timeCreated"`
}
