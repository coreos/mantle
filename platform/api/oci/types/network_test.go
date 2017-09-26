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

func TestVNICDetails(t *testing.T) {
	type in struct {
		input VNICDetails
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
			in: in{input: VNICDetails{
				AssignPublicIP:      BoolToPtr(false),
				DisplayName:         StrToPtr("example-vnic-details"),
				HostnameLabel:       StrToPtr("bminstance-1"),
				PrivateIP:           StrToPtr("10.0.3.3"),
				SkipSourceDestCheck: BoolToPtr(true),
				SubnetID:            "example-subnet-id",
			}},
			out:  out{jsonStr: `{"assignPublicIp":false,"displayName":"example-vnic-details","hostnameLabel":"bminstance-1","privateIp":"10.0.3.3","skipSourceDestCheck":true,"subnetId":"example-subnet-id"}`},
			name: "All Fields",
		},
		{
			in: in{input: VNICDetails{
				SubnetID: "example-subnet-id",
			}},
			out:  out{jsonStr: `{"subnetId":"example-subnet-id"}`},
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

func TestCreatePrivateIPInput(t *testing.T) {
	type in struct {
		input CreatePrivateIPInput
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
			in: in{input: CreatePrivateIPInput{
				DisplayName:   StrToPtr("example-private-ip"),
				HostnameLabel: StrToPtr("bminstance-1"),
				IPAddress:     StrToPtr("10.0.3.3"),
				VNICID:        "ocid1.example.oc1..aslkdfjaskldfjalksfjdlkasjfdsalfjdsalfjldsjafl",
			}},
			out:  out{jsonStr: `{"displayName":"example-private-ip","hostnameLabel":"bminstance-1","ipAddress":"10.0.3.3","vnicId":"ocid1.example.oc1..aslkdfjaskldfjalksfjdlkasjfdsalfjdsalfjldsjafl"}`},
			name: "All Fields",
		},
		{
			in: in{input: CreatePrivateIPInput{
				VNICID: "ocid1.example.oc1..aslkdfjaskldfjalksfjdlkasjfdsalfjdsalfjldsjafl",
			}},
			out:  out{jsonStr: `{"vnicId":"ocid1.example.oc1..aslkdfjaskldfjalksfjdlkasjfdsalfjdsalfjldsjafl"}`},
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

func TestUpdatePrivateIPInput(t *testing.T) {
	type in struct {
		input UpdatePrivateIPInput
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
			in: in{input: UpdatePrivateIPInput{
				DisplayName:   StrToPtr("example-private-ip"),
				HostnameLabel: StrToPtr("bminstance-1"),
				VNICID:        StrToPtr("ocid1.example.oc1..aslkdfjaskldfjalksfjdlkasjfdsalfjdsalfjldsjafl"),
			}},
			out:  out{jsonStr: `{"displayName":"example-private-ip","hostnameLabel":"bminstance-1","vnicId":"ocid1.example.oc1..aslkdfjaskldfjalksfjdlkasjfdsalfjdsalfjldsjafl"}`},
			name: "All Fields",
		},
		{
			in:   in{input: UpdatePrivateIPInput{}},
			out:  out{jsonStr: `{}`},
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

func TestCreateInternetGatewayInput(t *testing.T) {
	type in struct {
		input CreateInternetGatewayInput
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
			in: in{input: CreateInternetGatewayInput{
				CompartmentID: "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				DisplayName:   StrToPtr("example-internet-gateway"),
				IsEnabled:     true,
				VNICID:        "ocid1.example.oc1..aslkdfjaskldfjalksfjdlkasjfdsalfjdsalfjldsjafl",
			}},
			out:  out{jsonStr: `{"compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","displayName":"example-internet-gateway","isEnabled":true,"vnicId":"ocid1.example.oc1..aslkdfjaskldfjalksfjdlkasjfdsalfjdsalfjldsjafl"}`},
			name: "All Fields",
		},
		{
			in: in{input: CreateInternetGatewayInput{
				CompartmentID: "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				IsEnabled:     true,
				VNICID:        "ocid1.example.oc1..aslkdfjaskldfjalksfjdlkasjfdsalfjdsalfjldsjafl",
			}},
			out:  out{jsonStr: `{"compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","isEnabled":true,"vnicId":"ocid1.example.oc1..aslkdfjaskldfjalksfjdlkasjfdsalfjdsalfjldsjafl"}`},
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

func TestUpdateInternetGatewayInput(t *testing.T) {
	type in struct {
		input UpdateInternetGatewayInput
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
			in: in{input: UpdateInternetGatewayInput{
				DisplayName: StrToPtr("example-internet-gateway"),
				IsEnabled:   BoolToPtr(true),
			}},
			out:  out{jsonStr: `{"displayName":"example-internet-gateway","isEnabled":true}`},
			name: "All Fields",
		},
		{
			in:   in{input: UpdateInternetGatewayInput{}},
			out:  out{jsonStr: `{}`},
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

func TestAttachVNICInput(t *testing.T) {
	type in struct {
		input AttachVNICInput
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
			in: in{input: AttachVNICInput{
				CreateVNICDetails: VNICDetails{
					AssignPublicIP:      BoolToPtr(false),
					DisplayName:         StrToPtr("example-vnic-details"),
					HostnameLabel:       StrToPtr("bminstance-1"),
					PrivateIP:           StrToPtr("10.0.3.3"),
					SkipSourceDestCheck: BoolToPtr(true),
					SubnetID:            "example-subnet-id",
				},
				DisplayName: StrToPtr("example-attach-vnic"),
				InstanceID:  "ocid1.instance.oc1..aaaaaaaayzfqeibduyox6iib3olcmjsdlfjasldfjasldfjasdlfjaasdfds",
			}},
			out:  out{jsonStr: `{"createVnicDetails":{"assignPublicIp":false,"displayName":"example-vnic-details","hostnameLabel":"bminstance-1","privateIp":"10.0.3.3","skipSourceDestCheck":true,"subnetId":"example-subnet-id"},"displayName":"example-attach-vnic","instanceId":"ocid1.instance.oc1..aaaaaaaayzfqeibduyox6iib3olcmjsdlfjasldfjasldfjasdlfjaasdfds"}`},
			name: "All Fields",
		},
		{
			in: in{input: AttachVNICInput{
				CreateVNICDetails: VNICDetails{
					SubnetID: "example-subnet-id",
				},
				InstanceID: "ocid1.instance.oc1..aaaaaaaayzfqeibduyox6iib3olcmjsdlfjasldfjasldfjasdlfjaasdfds",
			}},
			out:  out{jsonStr: `{"createVnicDetails":{"subnetId":"example-subnet-id"},"instanceId":"ocid1.instance.oc1..aaaaaaaayzfqeibduyox6iib3olcmjsdlfjasldfjasldfjasdlfjaasdfds"}`},
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

func TestUpdateVNICInput(t *testing.T) {
	type in struct {
		input UpdateVNICInput
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
			in: in{input: UpdateVNICInput{
				DisplayName:         StrToPtr("example-update-vnic"),
				HostnameLabel:       StrToPtr("bminstance-1"),
				SkipSourceDestCheck: BoolToPtr(true),
			}},
			out:  out{jsonStr: `{"displayName":"example-update-vnic","hostnameLabel":"bminstance-1","skipSourceDestCheck":true}`},
			name: "All Fields",
		},
		{
			in:   in{input: UpdateVNICInput{}},
			out:  out{jsonStr: `{}`},
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

func TestVNIC(t *testing.T) {
	type in struct {
		jsonStr string
	}
	type out struct {
		output VNIC
	}

	tests := []struct {
		in   in
		out  out
		name string
	}{
		{
			in: in{jsonStr: `{"availabilityDomain":"Uocm:PHX-AD-1","compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","displayName":"example-vnic","hostnameLabel":"bminstance-1","id":"ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds","isPrimary":true,"lifecycleState":"AVAILABLE","macAddress":"00:00:17:B6:4D:DD","privateIp":"10.0.3.3","publicIp":"8.8.8.8","skipSourceDestCheck":true,"subnetId":"example-subnet-id","timeCreated":"2016-08-25T21:10:29.600Z"}`},
			out: out{output: VNIC{
				AvailabilityDomain:  "Uocm:PHX-AD-1",
				CompartmentID:       "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				DisplayName:         StrToPtr("example-vnic"),
				HostnameLabel:       StrToPtr("bminstance-1"),
				ID:                  "ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds",
				IsPrimary:           BoolToPtr(true),
				LifecycleState:      "AVAILABLE",
				MACAddress:          StrToPtr("00:00:17:B6:4D:DD"),
				PrivateIP:           "10.0.3.3",
				PublicIP:            StrToPtr("8.8.8.8"),
				SkipSourceDestCheck: BoolToPtr(true),
				SubnetID:            "example-subnet-id",
				TimeCreated:         "2016-08-25T21:10:29.600Z",
			}},
			name: "All Fields",
		},
		{
			in: in{jsonStr: `{"availabilityDomain":"Uocm:PHX-AD-1","compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","id":"ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds","lifecycleState":"AVAILABLE","privateIp":"10.0.3.3","subnetId":"example-subnet-id","timeCreated":"2016-08-25T21:10:29.600Z"}`},
			out: out{output: VNIC{
				AvailabilityDomain: "Uocm:PHX-AD-1",
				CompartmentID:      "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				ID:                 "ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds",
				LifecycleState:     "AVAILABLE",
				PrivateIP:          "10.0.3.3",
				SubnetID:           "example-subnet-id",
				TimeCreated:        "2016-08-25T21:10:29.600Z",
			}},
			name: "Required Only",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			output := VNIC{}
			err := json.Unmarshal([]byte(test.in.jsonStr), &output)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.out.output, output, "%s failure", test.name)
		})
	}
}

func TestVNICAttachment(t *testing.T) {
	type in struct {
		jsonStr string
	}
	type out struct {
		output VNICAttachment
	}

	tests := []struct {
		in   in
		out  out
		name string
	}{
		{
			in: in{jsonStr: `{"availabilityDomain":"Uocm:PHX-AD-1","compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","displayName":"example-vnic","id":"ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds","instanceId":"ocid1.instance.oc1..aaaaaaaayzfqeibduyox6iib3olcmjsdlfjasldfjasldfjasdlfjaasdfds","lifecycleState":"AVAILABLE","subnetId":"example-subnet-id","timeCreated":"2016-08-25T21:10:29.600Z","vlanTag":0,"vnicId":"ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds"}`},
			out: out{output: VNICAttachment{
				AvailabilityDomain: "Uocm:PHX-AD-1",
				CompartmentID:      "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				DisplayName:        StrToPtr("example-vnic"),
				ID:                 "ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds",
				InstanceID:         "ocid1.instance.oc1..aaaaaaaayzfqeibduyox6iib3olcmjsdlfjasldfjasldfjasdlfjaasdfds",
				LifecycleState:     "AVAILABLE",
				SubnetID:           "example-subnet-id",
				TimeCreated:        "2016-08-25T21:10:29.600Z",
				VLANTag:            IntToPtr(0),
				VNICID:             StrToPtr("ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds"),
			}},
			name: "All Fields",
		},
		{
			in: in{jsonStr: `{"availabilityDomain":"Uocm:PHX-AD-1","compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","id":"ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds","instanceId":"ocid1.instance.oc1..aaaaaaaayzfqeibduyox6iib3olcmjsdlfjasldfjasldfjasdlfjaasdfds","lifecycleState":"AVAILABLE","subnetId":"example-subnet-id","timeCreated":"2016-08-25T21:10:29.600Z"}`},
			out: out{output: VNICAttachment{
				AvailabilityDomain: "Uocm:PHX-AD-1",
				CompartmentID:      "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				ID:                 "ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds",
				InstanceID:         "ocid1.instance.oc1..aaaaaaaayzfqeibduyox6iib3olcmjsdlfjasldfjasldfjasdlfjaasdfds",
				LifecycleState:     "AVAILABLE",
				SubnetID:           "example-subnet-id",
				TimeCreated:        "2016-08-25T21:10:29.600Z",
			}},
			name: "Required Only",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			output := VNICAttachment{}
			err := json.Unmarshal([]byte(test.in.jsonStr), &output)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.out.output, output, "%s failure", test.name)
		})
	}
}

func TestInternetGateway(t *testing.T) {
	type in struct {
		jsonStr string
	}
	type out struct {
		output InternetGateway
	}

	tests := []struct {
		in   in
		out  out
		name string
	}{
		{
			in: in{jsonStr: `{"compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","displayName":"example-vnic","id":"ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds","isEnabled":true,"lifecycleState":"AVAILABLE","timeCreated":"2016-08-25T21:10:29.600Z","vnicId":"ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds"}`},
			out: out{output: InternetGateway{
				CompartmentID:  "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				DisplayName:    StrToPtr("example-vnic"),
				ID:             "ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds",
				IsEnabled:      BoolToPtr(true),
				LifecycleState: "AVAILABLE",
				TimeCreated:    "2016-08-25T21:10:29.600Z",
				VNICID:         "ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds",
			}},
			name: "All Fields",
		},
		{
			in: in{jsonStr: `{"compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","id":"ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds","lifecycleState":"AVAILABLE","timeCreated":"2016-08-25T21:10:29.600Z","vnicId":"ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds"}`},
			out: out{output: InternetGateway{
				CompartmentID:  "ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq",
				ID:             "ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds",
				LifecycleState: "AVAILABLE",
				TimeCreated:    "2016-08-25T21:10:29.600Z",
				VNICID:         "ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds",
			}},
			name: "Required Only",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			output := InternetGateway{}
			err := json.Unmarshal([]byte(test.in.jsonStr), &output)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.out.output, output, "%s failure", test.name)
		})
	}
}

func TestPrivateIP(t *testing.T) {
	type in struct {
		jsonStr string
	}
	type out struct {
		output PrivateIP
	}

	tests := []struct {
		in   in
		out  out
		name string
	}{
		{
			in: in{jsonStr: `{"availabilityDomain":"Uocm:PHX-AD-1","compartmentId":"ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq","displayName":"example-private-ip","hostnameLabel":"bminstance-1","id":"ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds","ipAddress":"10.0.3.3","isPrimary":true,"subnetId":"example-subnet-id","timeCreated":"2016-08-25T21:10:29.600Z","vnicId":"ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds"}`},
			out: out{output: PrivateIP{
				AvailabilityDomain: StrToPtr("Uocm:PHX-AD-1"),
				CompartmentID:      StrToPtr("ocid1.compartment.oc1..aaaaaaaayzfqeibduyox6iib3olcmdar3ugly4fmameq4h7lcdlihrvur7xq"),
				DisplayName:        StrToPtr("example-private-ip"),
				HostnameLabel:      StrToPtr("bminstance-1"),
				ID:                 StrToPtr("ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds"),
				IPAddress:          StrToPtr("10.0.3.3"),
				IsPrimary:          BoolToPtr(true),
				SubnetID:           StrToPtr("example-subnet-id"),
				TimeCreated:        StrToPtr("2016-08-25T21:10:29.600Z"),
				VNICID:             StrToPtr("ocid1.vnic..oc1..fkldasjflkasjdflkasjfldsajflajsdlfjasldfjasldfjasdlfjaasdfds"),
			}},
			name: "All Fields",
		},
		{
			in:   in{jsonStr: `{}`},
			out:  out{output: PrivateIP{}},
			name: "Required Only",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			output := PrivateIP{}
			err := json.Unmarshal([]byte(test.in.jsonStr), &output)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.out.output, output, "%s failure", test.name)
		})
	}
}
