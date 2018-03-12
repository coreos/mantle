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
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/identity"
)

const (
	tcp  = "6"
	udp  = "17"
	icmp = "1"
)

func (a *API) GetVCN(name string) (core.Vcn, error) {
	vcns, err := a.vn.ListVcns(context.Background(), core.ListVcnsRequest{
		CompartmentId: &a.opts.CompartmentID,
		DisplayName:   &name,
	})
	if err != nil {
		return core.Vcn{}, err
	}

	for _, v := range vcns.Items {
		if v.DisplayName != nil && name == *v.DisplayName {
			return v, nil
		}
	}

	return core.Vcn{}, fmt.Errorf("couldn't find Virtual Network %s", name)
}

func (a *API) ListAvailabilityDomains() ([]identity.AvailabilityDomain, error) {
	ads, err := a.identity.ListAvailabilityDomains(context.Background(), identity.ListAvailabilityDomainsRequest{
		CompartmentId: &a.opts.CompartmentID,
	})
	if err != nil {
		return nil, err
	}

	return ads.Items, err
}

func (a *API) CreateVCN(name, cidrBlock string) (core.CreateVcnResponse, error) {
	return a.vn.CreateVcn(context.Background(), core.CreateVcnRequest{
		CreateVcnDetails: core.CreateVcnDetails{
			CidrBlock:     &cidrBlock,
			CompartmentId: &a.opts.CompartmentID,
			DisplayName:   &name,
			DnsLabel:      &name,
		},
	})
}

func (a *API) DeleteVCN(ID string) error {
	_, err := a.vn.DeleteVcn(context.Background(), core.DeleteVcnRequest{
		VcnId: &ID,
	})
	return err
}

func (a *API) CreateSubnet(subdomain, availabilityDomain, cidrBlock, vcnID, securityListID, routeTableID string) (core.CreateSubnetResponse, error) {
	return a.vn.CreateSubnet(context.Background(), core.CreateSubnetRequest{
		CreateSubnetDetails: core.CreateSubnetDetails{
			AvailabilityDomain: &availabilityDomain,
			CidrBlock:          &cidrBlock,
			CompartmentId:      &a.opts.CompartmentID,
			VcnId:              &vcnID,
			DnsLabel:           &subdomain,
			RouteTableId:       &routeTableID,
			SecurityListIds:    []string{securityListID},
		},
	})
}

func (a *API) CreateInternetGateway(vcnID string) (core.CreateInternetGatewayResponse, error) {
	return a.vn.CreateInternetGateway(context.Background(), core.CreateInternetGatewayRequest{
		CreateInternetGatewayDetails: core.CreateInternetGatewayDetails{
			CompartmentId: &a.opts.CompartmentID,
			IsEnabled:     boolToPtr(true),
			VcnId:         &vcnID,
		},
	})
}

func (a *API) ListSecurityLists(vcnID string) ([]core.SecurityList, error) {
	secLists, err := a.vn.ListSecurityLists(context.Background(), core.ListSecurityListsRequest{
		CompartmentId: &a.opts.CompartmentID,
		VcnId:         &vcnID,
	})
	if err != nil {
		return nil, err
	}
	return secLists.Items, nil
}

func (a *API) DeleteSecurityList(ID string) error {
	_, err := a.vn.DeleteSecurityList(context.Background(), core.DeleteSecurityListRequest{
		SecurityListId: &ID,
	})
	return err
}

func (a *API) ListInternetGateways(vcnID string) ([]core.InternetGateway, error) {
	igws, err := a.vn.ListInternetGateways(context.Background(), core.ListInternetGatewaysRequest{
		CompartmentId: &a.opts.CompartmentID,
		VcnId:         &vcnID,
	})
	if err != nil {
		return nil, err
	}
	return igws.Items, nil
}

func (a *API) DeleteInternetGateway(ID string) error {
	_, err := a.vn.DeleteInternetGateway(context.Background(), core.DeleteInternetGatewayRequest{
		IgId: &ID,
	})
	return err
}

func (a *API) ListRouteTables(vcnID string) ([]core.RouteTable, error) {
	rts, err := a.vn.ListRouteTables(context.Background(), core.ListRouteTablesRequest{
		CompartmentId: &a.opts.CompartmentID,
		VcnId:         &vcnID,
	})
	if err != nil {
		return nil, err
	}
	return rts.Items, nil
}

func (a *API) DeleteRouteTable(ID string) error {
	_, err := a.vn.DeleteRouteTable(context.Background(), core.DeleteRouteTableRequest{
		RtId: &ID,
	})
	return err
}

func (a *API) CreateDefaultSecurityList(vcnID string) (core.CreateSecurityListResponse, error) {
	return a.vn.CreateSecurityList(context.Background(), core.CreateSecurityListRequest{
		CreateSecurityListDetails: core.CreateSecurityListDetails{
			CompartmentId: &a.opts.CompartmentID,
			VcnId:         &vcnID,
			EgressSecurityRules: []core.EgressSecurityRule{
				{
					Destination: strToPtr("0.0.0.0/0"),
					Protocol:    strToPtr("all"),
				},
			},
			IngressSecurityRules: []core.IngressSecurityRule{
				{
					// Allow all TCP on private network
					Protocol: strToPtr(tcp),
					Source:   strToPtr("10.0.0.0/16"),
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Min: intToPtr(1),
							Max: intToPtr(65535),
						},
					},
				},
				{
					// Allow all UDP on private network
					Protocol: strToPtr(udp),
					Source:   strToPtr("10.0.0.0/16"),
					UdpOptions: &core.UdpOptions{
						DestinationPortRange: &core.PortRange{
							Min: intToPtr(1),
							Max: intToPtr(65535),
						},
					},
				},
				{
					// Allow all ICMP on private network
					Protocol: strToPtr(icmp),
					Source:   strToPtr("10.0.0.0/16"),
				},
				{
					// Default setting:
					// open inbound TCP traffic to port 22
					// from any source to allow for SSH
					Protocol: strToPtr(tcp),
					Source:   strToPtr("0.0.0.0/0"),
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Min: intToPtr(22),
							Max: intToPtr(22),
						},
					},
				},
				{
					// Default setting:
					// allow type 3 ICMP to the machine
					// "Destination Unreachable" from any source
					// to allow for MTU negotiation
					Protocol: strToPtr(icmp),
					Source:   strToPtr("0.0.0.0/0"),
					IcmpOptions: &core.IcmpOptions{
						Code: intToPtr(4),
						Type: intToPtr(3),
					},
				},
			},
		},
	})
}

func (a *API) ListSubnets(vcnID string) ([]core.Subnet, error) {
	subnets, err := a.vn.ListSubnets(context.Background(), core.ListSubnetsRequest{
		CompartmentId: &a.opts.CompartmentID,
		VcnId:         &vcnID,
	})
	if err != nil {
		return nil, err
	}
	return subnets.Items, nil

}

func (a *API) DeleteSubnet(ID string) error {
	_, err := a.vn.DeleteSubnet(context.Background(), core.DeleteSubnetRequest{
		SubnetId: &ID,
	})
	return err
}

func (a *API) getSubnetOnVCN(vcnID string) (core.Subnet, error) {
	subnets, err := a.ListSubnets(vcnID)
	if err != nil {
		return core.Subnet{}, err
	}

	if len(subnets) < 1 {
		return core.Subnet{}, fmt.Errorf("could't find Subnet")
	}
	return subnets[0], nil
}

func (a *API) CreateDefaultRouteTable(vcnID, igwID string) (core.RouteTable, error) {
	rt, err := a.vn.CreateRouteTable(context.Background(), core.CreateRouteTableRequest{
		CreateRouteTableDetails: core.CreateRouteTableDetails{
			CompartmentId: &a.opts.CompartmentID,
			RouteRules: []core.RouteRule{
				{
					CidrBlock:       strToPtr("0.0.0.0/0"),
					NetworkEntityId: &igwID,
				},
			},
			VcnId: &vcnID,
		},
	})
	if err != nil {
		return core.RouteTable{}, err
	}
	return rt.RouteTable, nil
}
