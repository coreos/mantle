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
	cmdDelete = &cobra.Command{
		Use:   "delete-kola-vcn",
		Short: "Delete OCI networking",
		Long:  "Remove kola virtual cloud network from an OCI account",
		Run:   runDeleteVCN,
	}
)

func runDeleteVCN(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		fmt.Fprintf(os.Stderr, "Unrecognized args in ore clear cmd: %v\n", args)
		os.Exit(2)
	}

	if err := deleteVCN(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func deleteVCN() error {
	vcn, err := API.GetVCN("kola")
	if err != nil {
		return fmt.Errorf("A Virtual Cloud Network named `kola` doesn't exist!")
	}

	if vcn.Id == nil {
		return fmt.Errorf("received virtual cloud network id nil")
	}
	if vcn.DisplayName == nil {
		return fmt.Errorf("received virtual cloud network display name nil")
	}

	secLists, err := API.ListSecurityLists(*vcn.Id)
	if err != nil {
		return fmt.Errorf("Getting Security Lists: %v", err)
	}

	for _, secList := range secLists {
		if secList.DisplayName == nil {
			return fmt.Errorf("received security list display name nil")
		}
		if secList.Id == nil {
			return fmt.Errorf("received security list id nil")
		}
		if *secList.DisplayName != fmt.Sprintf("Default Security List for %s", *vcn.DisplayName) {
			err = API.DeleteSecurityList(*secList.Id)
			if err != nil {
				return fmt.Errorf("Deleting Security List %s: %v", *secList.DisplayName, err)
			}
		}
	}

	subnets, err := API.ListSubnets(*vcn.Id)
	if err != nil {
		return fmt.Errorf("Getting Subnets: %v", err)
	}

	for _, subnet := range subnets {
		if subnet.Id == nil {
			return fmt.Errorf("received subnet id nil")
		}
		if subnet.DisplayName == nil {
			return fmt.Errorf("received subnet display name nil")
		}
		err = API.DeleteSubnet(*subnet.Id)
		if err != nil {
			return fmt.Errorf("Deleting Subnet %s: %v", *subnet.DisplayName, err)
		}
	}

	rts, err := API.ListRouteTables(*vcn.Id)
	if err != nil {
		return fmt.Errorf("Getting Route Tables: %v", err)
	}

	for _, rt := range rts {
		if rt.DisplayName == nil {
			return fmt.Errorf("received route table display name nil")
		}
		if rt.Id == nil {
			return fmt.Errorf("received route table id nil")
		}
		if *rt.DisplayName != fmt.Sprintf("Default Route Table for %s", *vcn.DisplayName) {
			err = API.DeleteRouteTable(*rt.Id)
			if err != nil {
				return fmt.Errorf("Deleting Route Table %s: %v", *rt.DisplayName, err)
			}
		}
	}

	igws, err := API.ListInternetGateways(*vcn.Id)
	if err != nil {
		return fmt.Errorf("Getting Internet Gateways: %v", err)
	}

	for _, igw := range igws {
		if igw.Id == nil {
			return fmt.Errorf("received internet gateway id nil")
		}
		if igw.DisplayName == nil {
			return fmt.Errorf("received internet gateway display name nil")
		}
		err = API.DeleteInternetGateway(*igw.Id)
		if err != nil {
			return fmt.Errorf("Deleting Internet Gateway %s: %v", *igw.DisplayName, err)
		}
	}

	err = API.DeleteVCN(*vcn.Id)
	if err != nil {
		return fmt.Errorf("Deleting Virtual Cloud Network %s: %v", *vcn.DisplayName, err)
	}

	return nil
}
