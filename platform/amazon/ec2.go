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
	// See http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/UserProvidedKernels.html#AmazonKernelImageIDs for details.
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

func (a *AWSAPI) ImportVolume(manifestURL string, format VolumeFormat, volumeSize int64) (*VolumeID, error) {
	// TODO(mischief): ensure this lands on gp2 (ssd), not standard storage
	ami := &ec2.ImportVolumeInput{
		AvailabilityZone: aws.String(a.region.String() + "a"),
		Description:      &a.shortDescription,
		Image: &ec2.DiskImageDetail{
			Bytes:             &volumeSize,
			Format:            aws.String(format.String()),
			ImportManifestUrl: &manifestURL,
		},
		Volume: &ec2.VolumeDetail{
			Size: &volumeSize,
		},
	}

	res, err := a.ec2api.ImportVolume(ami)
	if err != nil {
		return nil, err
	}

	for {
		convs := &ec2.DescribeConversionTasksInput{
			ConversionTaskIds: []*string{
				res.ConversionTask.ConversionTaskId,
			},
		}

		convres, err := a.ec2api.DescribeConversionTasks(convs)
		if err != nil {
			return nil, err
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
						vid := VolumeID(*v.Volume.Id)
						return &vid, nil
					}

					return nil, err
				}
			}
		}

		time.Sleep(10 * time.Second)
	}
}

// CreateSnapshot takes a volume id and returns a snapshot id.
func (a *AWSAPI) CreateSnapshot(vid *VolumeID) (*SnapshotID, error) {
	vi := &ec2.CreateSnapshotInput{
		Description: &a.shortDescription,
		VolumeId:    aws.String(vid.String()),
	}

	snap, err := a.ec2api.CreateSnapshot(vi)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		// XXX: check pointers
		snap = snaps.Snapshots[0]

		switch *snap.State {
		case "pending":
			// keep waiting
		case "completed":
			sid := SnapshotID(*snap.SnapshotId)
			return &sid, nil
		case "error":
			return nil, fmt.Errorf("failed creating snapshot: %s", *snap.StateMessage)
		}

		time.Sleep(10 * time.Second)
	}
}

func (a *AWSAPI) RegisterImage(name string, sid *SnapshotID, hvm bool) (*AMIID, error) {
	ri := &ec2.RegisterImageInput{
		Architecture: aws.String("x86_64"),
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			// CoreOS root
			&ec2.BlockDeviceMapping{
				DeviceName: aws.String("/dev/xvda"),
				Ebs: &ec2.EbsBlockDevice{
					DeleteOnTermination: aws.Bool(true),
					SnapshotId:          aws.String(sid.String()),
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
		ri.KernelId = aws.String(akiids[a.region.String()])
		ri.RootDeviceName = aws.String("/dev/sda")
		ri.VirtualizationType = aws.String("paravirtual")
	}

	ami, err := a.ec2api.RegisterImage(ri)
	if err != nil {
		return nil, err
	}

	amiid := AMIID(*ami.ImageId)

	return &amiid, nil
}

// ImageType returns a EC2 image format name derived from the path p.
func ImageType(p string) (VolumeFormat, error) {
	ext := path.Ext(p)
	switch ext {
	case ".bin":
		return VolumeFormatRAW, nil
	case ".vmdk":
		return VolumeFormatVMDK, nil
	case ".vhd":
		return VolumeFormatVHD, nil
	}

	return "", ErrInvalidVolumeFormat
}
