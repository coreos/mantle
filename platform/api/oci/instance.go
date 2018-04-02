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
	"encoding/base64"
	"fmt"
	"time"

	"github.com/oracle/oci-go-sdk/core"

	"github.com/coreos/mantle/util"
)

const (
	terminated  = "TERMINATED"
	terminating = "TERMINATING"
	succeeded   = "SUCCEEDED"
	running     = "RUNNING"
)

func (a *API) CreateInstance(name, userdata, sshKey string) (*Machine, error) {
	vcn, err := a.GetVCN("kola")
	if err != nil {
		return nil, err
	}

	if vcn.Id == nil {
		return nil, fmt.Errorf("received VCN id nil")
	}

	subnet, err := a.getSubnetOnVCN(*vcn.Id)
	if err != nil {
		return nil, err
	}

	metadata := map[string]string{
		"created_by": "mantle",
	}
	if userdata != "" {
		metadata["user_data"] = base64.StdEncoding.EncodeToString([]byte(userdata))
	}
	if sshKey != "" {
		metadata["ssh_authorized_keys"] = sshKey
	}

	resp, err := a.compute.LaunchInstance(context.Background(), core.LaunchInstanceRequest{
		LaunchInstanceDetails: core.LaunchInstanceDetails{
			AvailabilityDomain: subnet.AvailabilityDomain,
			CompartmentId:      &a.opts.CompartmentID,
			Shape:              &a.opts.Shape,
			CreateVnicDetails: &core.CreateVnicDetails{
				SubnetId:       subnet.Id,
				AssignPublicIp: boolToPtr(true),
				HostnameLabel:  &name,
			},
			DisplayName:   &name,
			HostnameLabel: &name,
			ImageId:       &a.opts.Image,
			Metadata:      metadata,
		},
	})
	if err != nil {
		return nil, err
	}

	inst := resp.Instance
	id := inst.Id

	if id == nil {
		return nil, fmt.Errorf("received instance id nil")
	}

	err = util.WaitUntilReady(10*time.Minute, 10*time.Second, func() (bool, error) {
		getInst, err := a.compute.GetInstance(context.Background(), core.GetInstanceRequest{
			InstanceId: id,
		})
		if err != nil {
			return false, err
		}
		inst = getInst.Instance

		if inst.LifecycleState != running {
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		a.TerminateInstance(*id)
		return nil, fmt.Errorf("waiting for machine to become active: %v", err)
	}

	if inst.Id == nil {
		return nil, fmt.Errorf("received instance id nil")
	}

	vnicAttachments, err := a.compute.ListVnicAttachments(context.Background(), core.ListVnicAttachmentsRequest{
		CompartmentId: &a.opts.CompartmentID,
		InstanceId:    inst.Id,
	})
	if err != nil {
		a.TerminateInstance(*inst.Id)
		return nil, fmt.Errorf("listing vnic attachments: %v", err)
	}

	if len(vnicAttachments.Items) < 1 {
		a.TerminateInstance(*inst.Id)
		return nil, fmt.Errorf("couldn't get VM information")
	}
	vnic, err := a.vn.GetVnic(context.Background(), core.GetVnicRequest{
		VnicId: vnicAttachments.Items[0].VnicId,
	})
	if err != nil {
		return nil, fmt.Errorf("getting vnic: %v", err)
	}

	if inst.DisplayName == nil {
		return nil, fmt.Errorf("received instance display name nil")
	}

	if vnic.Vnic.PublicIp == nil {
		return nil, fmt.Errorf("received vnic public ip nil")
	}

	if vnic.Vnic.PrivateIp == nil {
		return nil, fmt.Errorf("received vnic private ip nil")
	}

	return &Machine{
		Name:             *inst.DisplayName,
		ID:               *inst.Id,
		PublicIPAddress:  *vnic.Vnic.PublicIp,
		PrivateIPAddress: *vnic.Vnic.PrivateIp,
	}, nil
}

func (a *API) TerminateInstance(instanceID string) error {
	_, err := a.compute.TerminateInstance(context.Background(), core.TerminateInstanceRequest{
		InstanceId: &instanceID,
	})
	return err
}

// ConsoleHistory is deleted when an instance is terminated, as such
// we just return errors and let the history be deleted when the instance
// is terminated.
func (a *API) GetConsoleOutput(instanceID string) (string, error) {
	metadata, err := a.compute.CaptureConsoleHistory(context.Background(), core.CaptureConsoleHistoryRequest{
		CaptureConsoleHistoryDetails: core.CaptureConsoleHistoryDetails{
			InstanceId: &instanceID,
		},
	})
	if err != nil {
		return "", fmt.Errorf("capturing console history: %v", err)
	}

	consoleHistoryStatus, err := a.compute.GetConsoleHistory(context.Background(), core.GetConsoleHistoryRequest{
		InstanceConsoleHistoryId: metadata.ConsoleHistory.Id,
	})
	if err != nil {
		return "", fmt.Errorf("getting console history status: %v", err)
	}

	err = util.WaitUntilReady(5*time.Minute, 10*time.Second, func() (bool, error) {
		consoleHistoryStatus, err := a.compute.GetConsoleHistory(context.Background(), core.GetConsoleHistoryRequest{
			InstanceConsoleHistoryId: metadata.ConsoleHistory.Id,
		})
		if err != nil {
			return false, fmt.Errorf("getting console history status: %v", err)
		}

		if consoleHistoryStatus.ConsoleHistory.LifecycleState != succeeded {
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return "", fmt.Errorf("waiting for console history data to be ready: %v", err)
	}

	// OCI limits console history to 1 MB; request 2 to be safe
	content, err := a.compute.GetConsoleHistoryContent(context.Background(), core.GetConsoleHistoryContentRequest{
		InstanceConsoleHistoryId: consoleHistoryStatus.ConsoleHistory.Id,
		Length: intToPtr(2 * 1024 * 1024),
		Offset: intToPtr(0),
	})
	if err != nil {
		return "", fmt.Errorf("getting console history data: %v", err)
	}

	if content.Value == nil {
		return "", fmt.Errorf("received console history data nil")
	}

	return *content.Value, nil
}

func (a *API) gcInstances(gracePeriod time.Duration) error {
	threshold := time.Now().Add(-gracePeriod)

	result, err := a.compute.ListInstances(context.Background(), core.ListInstancesRequest{
		CompartmentId: &a.opts.CompartmentID,
	})
	if err != nil {
		return err
	}
	for _, instance := range result.Items {
		if instance.Metadata["created_by"] != "mantle" {
			continue
		}

		if instance.TimeCreated.After(threshold) {
			continue
		}

		switch instance.LifecycleState {
		case terminating, terminated:
			continue
		}

		if instance.Id == nil {
			return fmt.Errorf("received instance id nil")
		}

		if err := a.TerminateInstance(*instance.Id); err != nil {
			return fmt.Errorf("couldn't terminate instance %v: %v", instance.Id, err)
		}
	}
	return nil
}
