// Copyright 2015 CoreOS, Inc.
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

package amazon

import (
	"fmt"
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

func (a *AWSAPI) ImportVolume(manifestURL, format string, volumeSize int64) (string, error) {
	// TODO(mischief): ensure this lands on gp2 (ssd), not standard storage
	ami := &ec2.ImportVolumeInput{
		AvailabilityZone: aws.String(a.region + "a"),
		Description:      &a.shortDescription,
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
func (a *AWSAPI) CreateSnapshot(volume string) (string, error) {
	vi := &ec2.CreateSnapshotInput{
		Description: &a.shortDescription,
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

func (a *AWSAPI) RegisterImage(name, snap string, hvm bool) (string, error) {
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
		Description: &a.shortDescription,
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
		ri.KernelId = aws.String(akiids[a.region])
		ri.RootDeviceName = aws.String("/dev/sda")
		ri.VirtualizationType = aws.String("paravirtual")
	}

	ami, err := a.ec2api.RegisterImage(ri)
	if err != nil {
		return "", err
	}

	return *ami.ImageId, nil
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
