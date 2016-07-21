// Copyright 2016 CoreOS, Inc.
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

package aws

import (
	"encoding/base64"
	"fmt"
	"net"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var (
	// Kernel images used for CoreOS PV AMIs.
	akiids = map[string]string{
		"us-east-1":      "aki-919dcaf8",
		"us-west-1":      "aki-880531cd",
		"us-west-2":      "aki-fc8f11cc",
		"eu-west-1":      "aki-52a34525",
		"eu-central-1":   "aki-184c7a05",
		"ap-southeast-1": "aki-503e7402",
		"ap-southeast-2": "aki-c362fff9",
		"ap-northeast-1": "aki-176bf516",
		"sa-east-1":      "aki-5553f448",
	}
)

func (a *API) ImportVolume(manifestURL, format string, volumeSize int64) (string, error) {
	// TODO(mischief): ensure this lands on gp2 (ssd), not standard storage
	ami := &ec2.ImportVolumeInput{
		AvailabilityZone: aws.String(a.options.Region + "a"),
		// TODO(mischief): better description
		Description: aws.String("mantle"),
		Image: &ec2.DiskImageDetail{
			Bytes:             &volumeSize,
			Format:            &format,
			ImportManifestUrl: &manifestURL,
		},
		Volume: &ec2.VolumeDetail{
			Size: &volumeSize,
		},
	}

	res, err := a.ec2api.ImportVolume(ami)
	if err != nil {
		return "", err
	}

	for {
		convs := &ec2.DescribeConversionTasksInput{
			ConversionTaskIds: []*string{
				res.ConversionTask.ConversionTaskId,
			},
		}

		convres, err := a.ec2api.DescribeConversionTasks(convs)
		if err != nil {
			return "", err
		}

		if len(convres.ConversionTasks) > 0 {
			st := convres.ConversionTasks[0]
			if st.State != nil {
				// Status == pending || active || completed || error
				switch *st.State {
				case "cancelled":
					fallthrough
				case "error":
					fallthrough
				case "completed":
					// get volume id
					v := st.ImportVolume
					if v != nil && v.Volume.Id != nil {
						return *v.Volume.Id, nil
					}

					return "", fmt.Errorf("no volume id")
				}
			}
		}

		time.Sleep(10 * time.Second)
	}
}

// CreateSnapshot takes a volume id and returns a snapshot id.
func (a *API) CreateSnapshot(description, volume string) (string, error) {
	vi := &ec2.CreateSnapshotInput{
		Description: &description,
		VolumeId:    &volume,
	}

	snap, err := a.ec2api.CreateSnapshot(vi)
	if err != nil {
		return "", err
	}

	snapid := snap.SnapshotId

	for {
		ds := &ec2.DescribeSnapshotsInput{
			SnapshotIds: []*string{
				snapid,
			},
		}

		snaps, err := a.ec2api.DescribeSnapshots(ds)
		if err != nil {
			return "", err
		}

		// XXX: check pointers
		snap = snaps.Snapshots[0]

		switch *snap.State {
		case "pending":
		case "completed":
			return *snap.SnapshotId, nil
		case "error":
			return "", fmt.Errorf("failed creating snapshot: %s", *snap.StateMessage)
		}

		time.Sleep(10 * time.Second)
	}
}

func (a *API) RegisterImage(name, description, snap string, hvm bool) (string, error) {
	ri := &ec2.RegisterImageInput{
		Architecture: aws.String("x86_64"),
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			// CoreOS root
			&ec2.BlockDeviceMapping{
				DeviceName: aws.String("/dev/xvda"),
				Ebs: &ec2.EbsBlockDevice{
					DeleteOnTermination: aws.Bool(true),
					SnapshotId:          &snap,
				},
			},
			// ephemeral volume
			&ec2.BlockDeviceMapping{
				DeviceName:  aws.String("/dev/xvdb"),
				VirtualName: aws.String("ephemeral0"),
			},
		},
		Description: &description,
		Name:        &name,
	}

	if hvm {
		ri.BlockDeviceMappings[0].DeviceName = aws.String("/dev/xvda")
		ri.BlockDeviceMappings[1].DeviceName = aws.String("/dev/xvdb")
		ri.RootDeviceName = aws.String("/dev/xvda")
		ri.SriovNetSupport = aws.String("simple")
		ri.VirtualizationType = aws.String("hvm")
	} else {
		ri.BlockDeviceMappings[0].DeviceName = aws.String("/dev/sda")
		ri.BlockDeviceMappings[1].DeviceName = aws.String("/dev/sdb")
		// XXX: check that kernel exists
		ri.KernelId = aws.String(akiids[a.options.Region])
		ri.RootDeviceName = aws.String("/dev/sda")
		ri.VirtualizationType = aws.String("paravirtual")
	}

	ami, err := a.ec2api.RegisterImage(ri)
	if err != nil {
		return "", err
	}

	return *ami.ImageId, nil
}

func (a *API) AddKey(name, key string) error {
	_, err := a.ec2api.ImportKeyPair(&ec2.ImportKeyPairInput{
		KeyName:           &name,
		PublicKeyMaterial: []byte(key),
	})

	return err
}

func (a *API) DeleteKey(name string) error {
	_, err := a.ec2api.DeleteKeyPair(&ec2.DeleteKeyPairInput{
		KeyName: &name,
	})

	return err
}

// waitForAWSInstance waits until a set of aws ec2 instance is accessible by ssh.
func waitForAWSInstances(api *ec2.EC2, ids []*string, d time.Duration) error {
	after := time.After(d)

	online := make(map[string]bool)

	for len(ids) != len(online) {
		select {
		case <-after:
			return fmt.Errorf("timed out waiting for instances to run")
		default:
		}

		// don't make api calls too quickly, or we will hit the rate limit

		time.Sleep(10 * time.Second)

		getinst := &ec2.DescribeInstancesInput{
			InstanceIds: ids,
		}

		insts, err := api.DescribeInstances(getinst)
		if err != nil {
			return err
		}

		for _, r := range insts.Reservations {
			for _, i := range r.Instances {
				// skip instances known to be up
				if online[*i.InstanceId] {
					continue
				}

				// "running"
				if *i.State.Code == int64(16) {
					// XXX: ssh is a terrible way to check this, but it is all we have.
					c, err := net.DialTimeout("tcp", *i.PublicIpAddress+":22", 10*time.Second)
					if err != nil {
						continue
					}
					c.Close()

					online[*i.InstanceId] = true
				}
			}
		}
	}

	return nil
}

func (a *API) CreateInstances(imageid, keyname, userdata, instancetype, securitygroup string, count uint64, wait bool) ([]*ec2.Instance, error) {
	cnt := int64(count)

	var ud *string
	if len(userdata) > 0 {
		tud := base64.StdEncoding.EncodeToString([]byte(userdata))
		ud = &tud
	}

	inst := ec2.RunInstancesInput{
		ImageId:        &imageid,
		MinCount:       &cnt,
		MaxCount:       &cnt,
		KeyName:        &keyname,
		InstanceType:   &instancetype,
		SecurityGroups: []*string{&securitygroup},
		UserData:       ud,
	}

	reservations, err := a.ec2api.RunInstances(&inst)
	if err != nil {
		return nil, err
	}

	if !wait {
		return reservations.Instances, nil
	}

	ids := make([]*string, len(reservations.Instances))
	for i, inst := range reservations.Instances {
		ids[i] = inst.InstanceId
	}

	if err := waitForAWSInstances(a.ec2api, ids, 5*time.Minute); err != nil {
		return nil, err
	}

	// call DescribeInstances to get machine IP
	getinst := &ec2.DescribeInstancesInput{
		InstanceIds: ids,
	}

	insts, err := a.ec2api.DescribeInstances(getinst)
	if err != nil {
		return nil, err
	}

	return insts.Reservations[0].Instances, err
}

func (a *API) TerminateInstance(id string) error {
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{&id},
	}

	if _, err := a.ec2api.TerminateInstances(input); err != nil {
		return err
	}

	return nil
}

// ImageType returns a EC2 image format name based on the path p.
func ImageType(p string) (typ string, err error) {
	ext := path.Ext(p)
	switch ext {
	case ".bin":
		typ = "RAW"
	case ".vmdk":
		typ = "VMDK"
	case ".vhd":
		typ = "VHD"
	default:
		return "", fmt.Errorf("unrecognized image type %q", ext)
	}

	return typ, nil
}
