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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateImageInput(t *testing.T) {
	type in struct {
		input CreateImageInput
	}
	type out struct {
		jsonStr string
	}

	tests := []struct {
		in   in
		out  out
		name string
	}{
		{
			in: in{input: CreateImageInput{
				CompartmentID: "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				DisplayName:   StrToPtr("MyCustomImage"),
				InstanceID:    StrToPtr("ocid1.instance.oc1.phx.abyhqljrqyriphyccj75yut36ybxmlfgawtl7m77vqanhg6w4bdszaitd3da"),
			}},
			out:  out{jsonStr: `{"compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","displayName":"MyCustomImage","instanceId":"ocid1.instance.oc1.phx.abyhqljrqyriphyccj75yut36ybxmlfgawtl7m77vqanhg6w4bdszaitd3da"}`},
			name: "No ImageSource",
		},
		{
			in: in{input: CreateImageInput{
				CompartmentID: "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				ImageSource: &ImageSourceViaObjectStorageTuple{
					ObjectName:    "image-to-import.qcow2",
					BucketName:    "MyBucket",
					NamespaceName: "MyNamespace",
					SourceType:    "objectStorageTuple",
				},
			}},
			out:  out{jsonStr: `{"compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","imageSourceDetails":{"bucketName":"MyBucket","namespaceName":"MyNamespace","objectName":"image-to-import.qcow2","sourceType":"objectStorageTuple"}}`},
			name: "Object Tuple ImageSource",
		},
		{
			in: in{input: CreateImageInput{
				CompartmentID: "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				DisplayName:   StrToPtr("MyImportedImage"),
				ImageSource: &ImageSourceViaObjectStorageURI{
					SourceURI:  "https://objectstorage.us-phoenix-1.oraclecloud.com/n/MyNamespace/b/MyBucket/o/image-to-import.qcow2",
					SourceType: "objectStorageUri",
				},
			}},
			out:  out{jsonStr: `{"compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","displayName":"MyImportedImage","imageSourceDetails":{"sourceUri":"https://objectstorage.us-phoenix-1.oraclecloud.com/n/MyNamespace/b/MyBucket/o/image-to-import.qcow2","sourceType":"objectStorageUri"}}`},
			name: "Object Storage Service URL",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			marshal, err := json.Marshal(test.in.input)
			if err != nil {
				t.Fatal(err)
			}

			if string(marshal) != test.out.jsonStr {
				t.Fatalf("JSON doens't match:\n\tExpected: %s\n\tReceived: %s", test.out.jsonStr, marshal)
			}
		})
	}
}

func TestExportImageInput(t *testing.T) {
	type in struct {
		input ExportImageInput
	}
	type out struct {
		jsonStr string
	}

	tests := []struct {
		in   in
		out  out
		name string
	}{
		{
			in: in{input: ExportImageInput{
				DestinationType: "objectStorageTuple",
				BucketName:      StrToPtr("MyBucket"),
				NamespaceName:   StrToPtr("MyNamespace"),
				ObjectName:      StrToPtr("exported-image.qcow2"),
			}},
			out:  out{jsonStr: `{"destinationType":"objectStorageTuple","bucketName":"MyBucket","namespaceName":"MyNamespace","objectName":"exported-image.qcow2"}`},
			name: "Namespace, Bucket Name, and Object Name",
		},
		{
			in: in{input: ExportImageInput{
				DestinationType: "objectStorageUri",
				DestinationURI:  StrToPtr("https://objectstorage.us-phoenix-1.oraclecloud.com/n/MyNamespace/b/MyBucket/o/exported-image.qcow2"),
			}},
			out:  out{jsonStr: `{"destinationType":"objectStorageUri","destinationUri":"https://objectstorage.us-phoenix-1.oraclecloud.com/n/MyNamespace/b/MyBucket/o/exported-image.qcow2"}`},
			name: "Object Storage URL",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			marshal, err := json.Marshal(test.in.input)
			if err != nil {
				t.Fatal(err)
			}

			if string(marshal) != test.out.jsonStr {
				t.Fatalf("JSON doens't match:\n\tExpected: %s\n\tReceived: %s", test.out.jsonStr, marshal)
			}
		})
	}
}

func TestUpdateImageInput(t *testing.T) {
	type in struct {
		input UpdateImageInput
	}
	type out struct {
		jsonStr string
	}

	tests := []struct {
		in   in
		out  out
		name string
	}{
		{
			in: in{input: UpdateImageInput{
				DisplayName: StrToPtr("MyFavoriteImage"),
			}},
			out:  out{jsonStr: `{"displayName":"MyFavoriteImage"}`},
			name: "All Fields",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			marshal, err := json.Marshal(test.in.input)
			if err != nil {
				t.Fatal(err)
			}

			if string(marshal) != test.out.jsonStr {
				t.Fatalf("JSON doens't match:\n\tExpected: %s\n\tReceived: %s", test.out.jsonStr, marshal)
			}
		})
	}
}

func TestImage(t *testing.T) {
	type in struct {
		jsonStr string
	}
	type out struct {
		output Image
	}

	tests := []struct {
		in   in
		out  out
		name string
	}{
		{
			in: in{jsonStr: `{"baseImageId":"example-base-image-id","compartmentId":"example-compartment-id","createImageAllowed":true,"displayName":"example-display-name","id":"example-id","lifecycleState":"AVAILABLE","operatingSystem":"coreos","operatingSystemVersion":"1535.2.0","timeCreated":"2017-09-22T21:29:30.600Z"}`},
			out: out{output: Image{
				BaseImageID:            StrToPtr("example-base-image-id"),
				CompartmentID:          "example-compartment-id",
				CreateImageAllowed:     true,
				DisplayName:            StrToPtr("example-display-name"),
				ID:                     "example-id",
				LifecycleState:         "AVAILABLE",
				OperatingSystem:        "coreos",
				OperatingSystemVersion: "1535.2.0",
				TimeCreated:            "2017-09-22T21:29:30.600Z",
			}},
			name: "All Fields",
		},
		{
			in: in{jsonStr: `{"compartmentId":"example-compartment-id","createImageAllowed":true,"id":"example-id","lifecycleState":"AVAILABLE","operatingSystem":"coreos","operatingSystemVersion":"1535.2.0","timeCreated":"2017-09-22T21:29:30.600Z"}`},
			out: out{output: Image{
				CompartmentID:          "example-compartment-id",
				CreateImageAllowed:     true,
				ID:                     "example-id",
				LifecycleState:         "AVAILABLE",
				OperatingSystem:        "coreos",
				OperatingSystemVersion: "1535.2.0",
				TimeCreated:            "2017-09-22T21:29:30.600Z",
			}},
			name: "Required Only",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			output := Image{}
			err := json.Unmarshal([]byte(test.in.jsonStr), &output)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.out.output, output, "%s failure", test.name)
		})
	}
}
