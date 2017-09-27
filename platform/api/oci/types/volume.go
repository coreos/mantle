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

type CreateVolumeInput struct {
	AvailabilityDomain string  `json:"availabilityDomain"`
	CompartmentID      string  `json:"compartmentId"`
	DisplayName        *string `json:"displayName,omitempty"`
	SizeInMBs          *int    `json:"sizeInMBs,omitempty"`
	VolumeBackupID     *string `json:"volumeBackupId,omitempty"`
}

type AttachVolumeInput struct {
	DisplayName *string `json:"displayName,omitempty"`
	InstanceID  string  `json:"instanceId"`
	Type        string  `json:"type"`
	VolumeID    string  `json:"volumeId"`
}

type Volume struct {
	AvailabilityDomain string `json:"availabilityDomain"`
	CompartmentID      string `json:"compartmentId"`
	DisplayName        string `json:"displayName"`
	ID                 string `json:"id"`
	LifecycleState     string `json:"lifecycleState"`
	SizeInMBs          int    `json:"sizeInMBs"`
	TimeCreated        string `json:"timeCreated"`
}

type VolumeAttachment struct {
	AttachmentType     string  `json:"attachmentType"`
	AvailabilityDomain string  `json:"availabilityDomain"`
	CompartmentID      string  `json:"compartmentId"`
	DisplayName        *string `json:"displayName"`
	ID                 string  `json:"id"`
	InstanceID         string  `json:"instanceId"`
	LifecycleState     string  `json:"lifecycleState"`
	TimeCreated        string  `json:"timeCreated"`
	VolumeID           string  `json:"volumeId"`
}
