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

func TestLaunchInstanceInput(t *testing.T) {
	type in struct {
		input LaunchInstanceInput
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
			in: in{input: LaunchInstanceInput{
				AvailabilityDomain: "Uocm:PHX-AD-1",
				CompartmentID:      "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				CreateVNICDetails: &VNICDetails{
					AssignPublicIP:      BoolToPtr(false),
					DisplayName:         StrToPtr("example-vnic-details"),
					HostnameLabel:       StrToPtr("bminstance-1"),
					PrivateIP:           StrToPtr("10.0.3.3"),
					SkipSourceDestCheck: BoolToPtr(true),
					SubnetID:            "example-subnet-id",
				},
				DisplayName:      StrToPtr("example-instance"),
				HostnameLabel:    StrToPtr("bminstance-1"),
				ImageID:          "ocid1.image...",
				IPXEScript:       StrToPtr("example-ipxe-script"),
				Metadata:         &map[string]string{"foo": "bar"},
				Shape:            "VM.Standard1.1",
				SubnetID:         StrToPtr("example-subnet-id"),
				ExtendedMetadata: RawMessageToPtr(json.RawMessage(`{"extended":{"metadata":"example"}}`)),
			}},
			out:  out{jsonStr: `{"availabilityDomain":"Uocm:PHX-AD-1","compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","createVnicDetails":{"assignPublicIp":false,"displayName":"example-vnic-details","hostnameLabel":"bminstance-1","privateIp":"10.0.3.3","skipSourceDestCheck":true,"subnetId":"example-subnet-id"},"displayName":"example-instance","hostnameLabel":"bminstance-1","imageId":"ocid1.image...","ipxeScript":"example-ipxe-script","metadata":{"foo":"bar"},"shape":"VM.Standard1.1","subnetId":"example-subnet-id","extendedMetadata":{"extended":{"metadata":"example"}}}`},
			name: "All Fields",
		},
		{
			in: in{input: LaunchInstanceInput{
				AvailabilityDomain: "Uocm:PHX-AD-1",
				CompartmentID:      "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				ImageID:            "ocid1.image...",
				Shape:              "VM.Standard1.1",
			}},
			out:  out{jsonStr: `{"availabilityDomain":"Uocm:PHX-AD-1","compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","imageId":"ocid1.image...","shape":"VM.Standard1.1"}`},
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

func TestInstance(t *testing.T) {
	type in struct {
		jsonStr string
	}
	type out struct {
		output Instance
	}

	tests := []struct {
		in   in
		out  out
		name string
	}{
		{
			in: in{jsonStr: `{"availabilityDomain":"Uocm:PHX-AD-1","compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","displayName":"example-instance","id":"ocid1.instance...","imageId":"ocid1.image...","ipxeScript":"example-ipxe-script","lifecycleState":"RUNNING","metadata":{"foo":"bar"},"region":"phx","shape":"VM.Standard1.1","timeCreated":"2016-08-25T21:10:29.600Z","extendedMetadata":{"extended":{"metadata":"example"}}}`},
			out: out{output: Instance{
				AvailabilityDomain: "Uocm:PHX-AD-1",
				CompartmentID:      "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				DisplayName:        StrToPtr("example-instance"),
				ID:                 "ocid1.instance...",
				ImageID:            StrToPtr("ocid1.image..."),
				IPXEScript:         StrToPtr("example-ipxe-script"),
				LifecycleState:     "RUNNING",
				Metadata:           &map[string]string{"foo": "bar"},
				Region:             "phx",
				Shape:              "VM.Standard1.1",
				TimeCreated:        "2016-08-25T21:10:29.600Z",
				ExtendedMetadata:   RawMessageToPtr(json.RawMessage(`{"extended":{"metadata":"example"}}`)),
			}},
			name: "All Fields",
		},
		{
			in: in{jsonStr: `{"availabilityDomain":"Uocm:PHX-AD-1","compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","id":"ocid1.instance...","lifecycleState":"RUNNING","region":"phx","shape":"VM.Standard1.1","timeCreated":"2016-08-25T21:10:29.600Z"}`},
			out: out{output: Instance{
				AvailabilityDomain: "Uocm:PHX-AD-1",
				CompartmentID:      "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				ID:                 "ocid1.instance...",
				LifecycleState:     "RUNNING",
				Region:             "phx",
				Shape:              "VM.Standard1.1",
				TimeCreated:        "2016-08-25T21:10:29.600Z",
			}},
			name: "Required Only",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			output := Instance{}
			err := json.Unmarshal([]byte(test.in.jsonStr), &output)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.out.output, output, "%s failure", test.name)
		})
	}
}

func TestListInstancesInput(t *testing.T) {
	type in struct {
		input ListInstancesInput
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
			in: in{input: ListInstancesInput{
				AvailabilityDomain: StrToPtr("Uocm:PHX-AD-1"),
				CompartmentID:      "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				DisplayName:        StrToPtr("example-instance"),
			}},
			out:  out{jsonStr: `{"availabilityDomain":"Uocm:PHX-AD-1","compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","displayName":"example-instance"}`},
			name: "All Fields",
		},
		{
			in: in{input: ListInstancesInput{
				CompartmentID: "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
			}},
			out:  out{jsonStr: `{"compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq"}`},
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

func TestInstanceConsoleConnectionInput(t *testing.T) {
	type in struct {
		input InstanceConsoleConnectionInput
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
			in: in{input: InstanceConsoleConnectionInput{
				InstanceID: "ocid1.instance.oc1.phx.abyhqljrqyriphyccj75yut36ybxmlfgawtl7m77vqanhg6w4bdszaitd3da",
				PublicKey:  "example-public-key",
			}},
			out:  out{jsonStr: `{"instanceId":"ocid1.instance.oc1.phx.abyhqljrqyriphyccj75yut36ybxmlfgawtl7m77vqanhg6w4bdszaitd3da","publicKey":"example-public-key"}`},
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

func TestCaptureConsoleHistoryInput(t *testing.T) {
	type in struct {
		input CaptureConsoleHistoryInput
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
			in: in{input: CaptureConsoleHistoryInput{
				InstanceID: "ocid1.instance.oc1.phx.abyhqljrqyriphyccj75yut36ybxmlfgawtl7m77vqanhg6w4bdszaitd3da",
			}},
			out:  out{jsonStr: `{"instanceId":"ocid1.instance.oc1.phx.abyhqljrqyriphyccj75yut36ybxmlfgawtl7m77vqanhg6w4bdszaitd3da"}`},
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

func TestShape(t *testing.T) {
	type in struct {
		jsonStr string
	}
	type out struct {
		output Shape
	}

	tests := []struct {
		in   in
		out  out
		name string
	}{
		{
			in: in{jsonStr: `{"shape":"VM-Standard1.1"}`},
			out: out{output: Shape{
				Shape: "VM-Standard1.1",
			}},
			name: "All Fields",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			output := Shape{}
			err := json.Unmarshal([]byte(test.in.jsonStr), &output)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.out.output, output, "%s failure", test.name)
		})
	}
}

func TestInstanceConsoleConnection(t *testing.T) {
	type in struct {
		jsonStr string
	}
	type out struct {
		output InstanceConsoleConnection
	}

	tests := []struct {
		in   in
		out  out
		name string
	}{
		{
			in: in{jsonStr: `{"compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","connectionString":"example-connection-string","fingerprint":"aa:bb:cc:dd:00:11:22:33:44","id":"ocid1.console.connection...","instanceId":"ocid1.instance.oc1..aaaaaaaayzfqeibduyox6iib3olcmjsdlfjasldfjasldfjasdlfjaasdfds","lifecycleState":"ACTIVE"}`},
			out: out{output: InstanceConsoleConnection{
				CompartmentID:    StrToPtr("ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq"),
				ConnectionString: StrToPtr("example-connection-string"),
				Fingerprint:      StrToPtr("aa:bb:cc:dd:00:11:22:33:44"),
				ID:               StrToPtr("ocid1.console.connection..."),
				InstanceID:       StrToPtr("ocid1.instance.oc1..aaaaaaaayzfqeibduyox6iib3olcmjsdlfjasldfjasldfjasdlfjaasdfds"),
				LifecycleState:   StrToPtr("ACTIVE"),
			}},
			name: "All Fields",
		},
		{
			in:   in{jsonStr: `{}`},
			out:  out{output: InstanceConsoleConnection{}},
			name: "Required Only",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			output := InstanceConsoleConnection{}
			err := json.Unmarshal([]byte(test.in.jsonStr), &output)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.out.output, output, "%s failure", test.name)
		})
	}
}

func TestConsoleHistory(t *testing.T) {
	type in struct {
		jsonStr string
	}
	type out struct {
		output ConsoleHistory
	}

	tests := []struct {
		in   in
		out  out
		name string
	}{
		{
			in: in{jsonStr: `{"availabilityDomain":"Uocm:PHX-AD-1","compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","displayName":"example-instance","id":"ocid1.console...","instanceId":"ocid1.instance...","lifecycleState":"RUNNING","timeCreated":"2016-08-25T21:10:29.600Z"}`},
			out: out{output: ConsoleHistory{
				AvailabilityDomain: "Uocm:PHX-AD-1",
				CompartmentID:      "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				DisplayName:        StrToPtr("example-instance"),
				ID:                 "ocid1.console...",
				InstanceID:         "ocid1.instance...",
				LifecycleState:     "RUNNING",
				TimeCreated:        "2016-08-25T21:10:29.600Z",
			}},
			name: "All Fields",
		},
		{
			in: in{jsonStr: `{"availabilityDomain":"Uocm:PHX-AD-1","compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","id":"ocid1.console...","instanceId":"ocid1.instance...","lifecycleState":"RUNNING","timeCreated":"2016-08-25T21:10:29.600Z"}`},
			out: out{output: ConsoleHistory{
				AvailabilityDomain: "Uocm:PHX-AD-1",
				CompartmentID:      "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				ID:                 "ocid1.console...",
				InstanceID:         "ocid1.instance...",
				LifecycleState:     "RUNNING",
				TimeCreated:        "2016-08-25T21:10:29.600Z",
			}},
			name: "Required Only",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			output := ConsoleHistory{}
			err := json.Unmarshal([]byte(test.in.jsonStr), &output)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.out.output, output, "%s failure", test.name)
		})
	}
}
