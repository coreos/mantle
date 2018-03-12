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

package oci

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cmdCreate = &cobra.Command{
		Use:   "create-kola-vcn",
		Short: "Create OCI networking",
		Long:  "Create kola virtual cloud network in an OCI account",
		Run:   runCreateVCN,
	}
)

func init() {
	OCI.AddCommand(cmdCreate)
}

func runCreateVCN(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		fmt.Fprintf(os.Stderr, "Unrecognized args in ore setup cmd: %v\n", args)
		os.Exit(2)
	}
	if err := createVCN(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func createVCN() error {
	_, err := API.GetVCN("kola")
	if err == nil {
		return fmt.Errorf("A Virtual Cloud Network named `kola` already exists!")
	}

	vcnResp, err := API.CreateVCN("kola", "10.0.0.0/16")
	if err != nil {
		return fmt.Errorf("Creating VCN: %v", err)
	}
	vcn := vcnResp.Vcn

	if vcn.Id == nil {
		return fmt.Errorf("received virtual cloud network id nil")
	}

	secList, err := API.CreateDefaultSecurityList(*vcn.Id)
	if err != nil {
		return fmt.Errorf("Creating default Security List: %v", err)
	}

	if secList.Id == nil {
		return fmt.Errorf("received security list id nil")
	}

	igw, err := API.CreateInternetGateway(*vcn.Id)
	if err != nil {
		return fmt.Errorf("Creating Internet Gateway: %v", err)
	}

	if igw.Id == nil {
		return fmt.Errorf("received internet gateway id nil")
	}

	rt, err := API.CreateDefaultRouteTable(*vcn.Id, *igw.Id)
	if err != nil {
		return fmt.Errorf("Creating default Route Table: %v", err)
	}

	if rt.Id == nil {
		return fmt.Errorf("received route table id nil")
	}

	ads, err := API.ListAvailabilityDomains()
	if err != nil {
		return fmt.Errorf("Listing Availability Domains: %v", err)
	}

	ad := ads[0]

	if ad.Name == nil {
		return fmt.Errorf("received availability domain name nil")
	}

	_, err = API.CreateSubnet("kola1", *ad.Name, "10.0.0.0/16", *vcn.Id, *secList.Id, *rt.Id)
	if err != nil {
		return fmt.Errorf("Creating Subnet: %v", err)
	}

	return nil
}
