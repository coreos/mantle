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

func TestCreateVolumeInput(t *testing.T) {
	type in struct {
		input CreateVolumeInput
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
			in: in{input: CreateVolumeInput{
				AvailabilityDomain: "Uocm:PHX-AD-1",
				CompartmentID:      "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				DisplayName:        StrToPtr("MyCustomVolume"),
				SizeInMBs:          IntToPtr(2048),
				VolumeBackupID:     StrToPtr("ocid1.volume.oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds"),
			}},
			out:  out{jsonStr: `{"availabilityDomain":"Uocm:PHX-AD-1","compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","displayName":"MyCustomVolume","sizeInMBs":2048,"volumeBackupId":"ocid1.volume.oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds"}`},
			name: "All Fields",
		},
		{
			in: in{input: CreateVolumeInput{
				AvailabilityDomain: "Uocm:PHX-AD-1",
				CompartmentID:      "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
			}},
			out:  out{jsonStr: `{"availabilityDomain":"Uocm:PHX-AD-1","compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq"}`},
			name: "Required Only",
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

func TestAttachVolumeInput(t *testing.T) {
	type in struct {
		input AttachVolumeInput
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
			in: in{input: AttachVolumeInput{
				DisplayName: StrToPtr("MyCustomVolume"),
				InstanceID:  "ocid1.instance.oc1..aaaaaaaayzfqeibduyox6iib3olcmjsdlfjasldfjasldfjasdlfjaasdfds",
				Type:        "iscsi",
				VolumeID:    "ocid1.volume.oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds",
			}},
			out:  out{jsonStr: `{"displayName":"MyCustomVolume","instanceId":"ocid1.instance.oc1..aaaaaaaayzfqeibduyox6iib3olcmjsdlfjasldfjasldfjasdlfjaasdfds","type":"iscsi","volumeId":"ocid1.volume.oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds"}`},
			name: "All Fields",
		},
		{
			in: in{input: AttachVolumeInput{
				InstanceID: "ocid1.instance.oc1..aaaaaaaayzfqeibduyox6iib3olcmjsdlfjasldfjasldfjasdlfjaasdfds",
				Type:       "iscsi",
				VolumeID:   "ocid1.volume.oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds",
			}},
			out:  out{jsonStr: `{"instanceId":"ocid1.instance.oc1..aaaaaaaayzfqeibduyox6iib3olcmjsdlfjasldfjasldfjasdlfjaasdfds","type":"iscsi","volumeId":"ocid1.volume.oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds"}`},
			name: "Required Only",
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

func TestVolume(t *testing.T) {
	type in struct {
		jsonStr string
	}
	type out struct {
		output Volume
	}

	tests := []struct {
		in   in
		out  out
		name string
	}{
		{
			in: in{jsonStr: `{"availabilityDomain":"Uocm:PHX-AD-1","compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","displayName":"MyCustomVolume","id":"ocid1.volume..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds","lifecycleState":"PROVISIONING","sizeInMBs":2048,"timeCreated":"2017-09-22T21:29:30.600Z"}`},
			out: out{output: Volume{
				AvailabilityDomain: "Uocm:PHX-AD-1",
				CompartmentID:      "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				DisplayName:        "MyCustomVolume",
				ID:                 "ocid1.volume..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds",
				LifecycleState:     "PROVISIONING",
				SizeInMBs:          2048,
				TimeCreated:        "2017-09-22T21:29:30.600Z",
			}},
			name: "All Fields",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			output := Volume{}
			err := json.Unmarshal([]byte(test.in.jsonStr), &output)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.out.output, output, "%s failure", test.name)
		})
	}
}

func TestVolumeAttachment(t *testing.T) {
	type in struct {
		jsonStr string
	}
	type out struct {
		output VolumeAttachment
	}

	tests := []struct {
		in   in
		out  out
		name string
	}{
		{
			in: in{jsonStr: `{"attachmentType":"example-type","availabilityDomain":"Uocm:PHX-AD-1","compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","displayName":"MyCustomVolumeAttachment","id":"ocid1.volume..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds","instanceId":"ocid1.instance.oc1..aaaaaaaayzfqeibduyox6iib3olcmjsdlfjasldfjasldfjasdlfjaasdfds","lifecycleState":"PROVISIONING","timeCreated":"2017-09-22T21:29:30.600Z","volumeId":"ocid1.volume..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds"}`},
			out: out{output: VolumeAttachment{
				AttachmentType:     "example-type",
				AvailabilityDomain: "Uocm:PHX-AD-1",
				CompartmentID:      "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				DisplayName:        StrToPtr("MyCustomVolumeAttachment"),
				ID:                 "ocid1.volume..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds",
				InstanceID:         "ocid1.instance.oc1..aaaaaaaayzfqeibduyox6iib3olcmjsdlfjasldfjasldfjasdlfjaasdfds",
				LifecycleState:     "PROVISIONING",
				TimeCreated:        "2017-09-22T21:29:30.600Z",
				VolumeID:           "ocid1.volume..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds",
			}},
			name: "All Fields",
		},
		{
			in: in{jsonStr: `{"attachmentType":"example-type","availabilityDomain":"Uocm:PHX-AD-1","compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","id":"ocid1.volume..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds","instanceId":"ocid1.instance.oc1..aaaaaaaayzfqeibduyox6iib3olcmjsdlfjasldfjasldfjasdlfjaasdfds","lifecycleState":"PROVISIONING","timeCreated":"2017-09-22T21:29:30.600Z","volumeId":"ocid1.volume..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds"}`},
			out: out{output: VolumeAttachment{
				AttachmentType:     "example-type",
				AvailabilityDomain: "Uocm:PHX-AD-1",
				CompartmentID:      "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				ID:                 "ocid1.volume..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds",
				InstanceID:         "ocid1.instance.oc1..aaaaaaaayzfqeibduyox6iib3olcmjsdlfjasldfjasldfjasdlfjaasdfds",
				LifecycleState:     "PROVISIONING",
				TimeCreated:        "2017-09-22T21:29:30.600Z",
				VolumeID:           "ocid1.volume..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds",
			}},
			name: "Required Only",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			output := VolumeAttachment{}
			err := json.Unmarshal([]byte(test.in.jsonStr), &output)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.out.output, output, "%s failure", test.name)
		})
	}
}
