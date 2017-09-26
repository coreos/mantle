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

type CreateImageInput struct {
	CompartmentID string             `json:"compartmentId"`
	DisplayName   *string            `json:"displayName,omitempty"`
	ImageSource   ImageSourceDetails `json:"imageSourceDetails,omitempty"`
	InstanceID    *string            `json:"instanceId,omitempty"`
}

type ImageSourceDetails interface{}

type ImageSourceViaObjectStorageTuple struct {
	BucketName    string `json:"bucketName"`
	NamespaceName string `json:"namespaceName"`
	ObjectName    string `json:"objectName"`
	SourceType    string `json:"sourceType"`

	ImageSourceDetails `json:"-"`
}

type ImageSourceViaObjectStorageURI struct {
	SourceURI  string `json:"sourceUri"`
	SourceType string `json:"sourceType"`

	ImageSourceDetails `json:"-"`
}

type ExportImageInput struct {
	DestinationType string  `json:"destinationType"`
	BucketName      *string `json:"bucketName,omitempty"`
	NamespaceName   *string `json:"namespaceName,omitempty"`
	ObjectName      *string `json:"objectName,omitempty"`
	DestinationURI  *string `json:"destinationUri,omitempty"`
}

type UpdateImageInput struct {
	DisplayName *string `json:"displayName,omitempty"`
}

type Image struct {
	BaseImageID            *string `json:"baseImageId"`
	CompartmentID          string  `json:"compartmentId"`
	CreateImageAllowed     bool    `json:"createImageAllowed"`
	DisplayName            *string `json:"displayName"`
	ID                     string  `json:"id"`
	LifecycleState         string  `json:"lifecycleState"`
	OperatingSystem        string  `json:"operatingSystem"`
	OperatingSystemVersion string  `json:"operatingSystemVersion"`
	TimeCreated            string  `json:"timeCreated"`
}
