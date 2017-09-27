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

type VNICDetails struct {
	AssignPublicIP      *bool   `json:"assignPublicIp,omitempty"`
	DisplayName         *string `json:"displayName,omitempty"`
	HostnameLabel       *string `json:"hostnameLabel,omitempty"`
	PrivateIP           *string `json:"privateIp,omitempty"`
	SkipSourceDestCheck *bool   `json:"skipSourceDestCheck,omitempty"`
	SubnetID            string  `json:"subnetId,omitempty"`
}

type CreatePrivateIPInput struct {
	DisplayName   *string `json:"displayName,omitempty"`
	HostnameLabel *string `json:"hostnameLabel,omitempty"`
	IPAddress     *string `json:"ipAddress,omitempty"`
	VNICID        string  `json:"vnicId"`
}

type UpdatePrivateIPInput struct {
	DisplayName   *string `json:"displayName,omitempty"`
	HostnameLabel *string `json:"hostnameLabel,omitempty"`
	VNICID        *string `json:"vnicId,omitempty"`
}

type CreateInternetGatewayInput struct {
	CompartmentID string  `json:"compartmentId"`
	DisplayName   *string `json:"displayName,omitempty"`
	IsEnabled     bool    `json:"isEnabled"`
	VNICID        string  `json:"vnicId"`
}

type UpdateInternetGatewayInput struct {
	DisplayName *string `json:"displayName,omitempty"`
	IsEnabled   *bool   `json:"isEnabled,omitempty"`
}

type AttachVNICInput struct {
	CreateVNICDetails VNICDetails `json:"createVnicDetails"`
	DisplayName       *string     `json:"displayName,omitempty"`
	InstanceID        string      `json:"instanceId"`
}

type UpdateVNICInput struct {
	DisplayName         *string `json:"displayName,omitempty"`
	HostnameLabel       *string `json:"hostnameLabel,omitempty"`
	SkipSourceDestCheck *bool   `json:"skipSourceDestCheck,omitempty"`
}

type VNIC struct {
	AvailabilityDomain  string  `json:"availabilityDomain"`
	CompartmentID       string  `json:"compartmentId"`
	DisplayName         *string `json:"displayName"`
	HostnameLabel       *string `json:"hostnameLabel"`
	ID                  string  `json:"id"`
	IsPrimary           *bool   `json:"isPrimary"`
	LifecycleState      string  `json:"lifecycleState"`
	MACAddress          *string `json:"macAddress"`
	PrivateIP           string  `json:"privateIp"`
	PublicIP            *string `json:"publicIp"`
	SkipSourceDestCheck *bool   `json:"skipSourceDestCheck"`
	SubnetID            string  `json:"subnetId"`
	TimeCreated         string  `json:"timeCreated"`
}

type VNICAttachment struct {
	AvailabilityDomain string  `json:"availabilityDomain"`
	CompartmentID      string  `json:"compartmentId"`
	DisplayName        *string `json:"displayName"`
	ID                 string  `json:"id"`
	InstanceID         string  `json:"instanceId"`
	LifecycleState     string  `json:"lifecycleState"`
	SubnetID           string  `json:"subnetId"`
	TimeCreated        string  `json:"timeCreated"`
	VLANTag            *int    `json:"vlanTag"`
	VNICID             *string `json:"vnicId"`
}

type InternetGateway struct {
	CompartmentID  string  `json:"compartmentId"`
	DisplayName    *string `json:"displayName"`
	ID             string  `json:"id"`
	IsEnabled      *bool   `json:"isEnabled"`
	LifecycleState string  `json:"lifecycleState"`
	TimeCreated    string  `json:"timeCreated"`
	VNICID         string  `json:"vnicId"`
}

type PrivateIP struct {
	AvailabilityDomain *string `json:"availabilityDomain"`
	CompartmentID      *string `json:"compartmentId"`
	DisplayName        *string `json:"displayName"`
	HostnameLabel      *string `json:"hostnameLabel"`
	ID                 *string `json:"id"`
	IPAddress          *string `json:"ipAddress"`
	IsPrimary          *bool   `json:"isPrimary"`
	SubnetID           *string `json:"subnetId"`
	TimeCreated        *string `json:"timeCreated"`
	VNICID             *string `json:"vnicId"`
}
