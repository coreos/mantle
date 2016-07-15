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
	"errors"
	"strings"
)

var (
	ErrInvalidRegion       = errors.New("invalid region name")
	ErrInvalidVolumeFormat = errors.New("invalid volume format")
	ErrInvalidVolumeID     = errors.New("invalid volume ID")
	ErrInvalidSnapshotID   = errors.New("invalid snapshot ID")
	ErrInvalidAMIID        = errors.New("invalid ami ID")
	ErrInvalidInstanceID   = errors.New("invalid instance ID")
)

// Region is an AWS region name.
// See http://docs.aws.amazon.com/general/latest/gr/rande.html for details on region names.
type Region string

var Regions = []Region{
	Region("us-east-1"),
	Region("us-west-2"),
	Region("us-west-1"),
	Region("eu-west-1"),
	Region("eu-central-1"),
	Region("ap-southeast-1"),
	Region("ap-southeast-2"),
	Region("ap-northeast-1"),
	Region("sa-east-1"),
}

func (r Region) String() string {
	return string(r)
}

func (r Region) Set(region string) error {
	valid := false
	for _, reg := range Regions {
		if reg == Region(region) {
			valid = true
			r = reg
			break
		}
	}

	if !valid {
		return ErrInvalidRegion
	}

	return nil
}

func (r Region) Type() string {
	return "amazonRegion"
}

// Bucket is an S3 bucket name.
type Bucket string

func (b Bucket) String() string {
	return string(b)
}

func (b Bucket) Set(bucket string) error {
	b = Bucket(bucket)
	return nil
}

func (b Bucket) Type() string {
	return "amazonBucket"
}

// VolumeFormat is a EC2 image type.
// See http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DiskImageDetail.html for valid values.
type VolumeFormat string

var (
	VolumeFormatVMDK = VolumeFormat("VMDK")
	VolumeFormatRAW  = VolumeFormat("RAW")
	VolumeFormatVHD  = VolumeFormat("VHD")

	VolumeFormats = []VolumeFormat{
		VolumeFormatVMDK,
		VolumeFormatRAW,
		VolumeFormatVHD,
	}
)

func (vf VolumeFormat) String() string {
	return string(vf)
}

func (vf VolumeFormat) Set(volumetype string) error {
	valid := false
	for _, voltype := range VolumeFormats {
		if voltype == VolumeFormat(volumetype) {
			valid = true
			vf = voltype
			break
		}
	}

	if !valid {
		return ErrInvalidVolumeFormat
	}

	return nil
}

func (vf VolumeFormat) Type() string {
	return "amazonVolumeFormat"
}

// VolumeID is an EC2 volume identifier.
// It must begin with "vol-".
type VolumeID string

func (vi VolumeID) String() string {
	return string(vi)
}

func (vi VolumeID) Set(volumeid string) error {
	if !strings.HasPrefix(volumeid, "vol-") {
		return ErrInvalidVolumeID
	}

	vi = VolumeID(volumeid)
	return nil
}

func (vi VolumeID) Type() string {
	return "amazonVolumeID"
}

// SnapshotID is an EC2 snapshot identifier.
// It must begin with "snap-".
type SnapshotID string

func (si SnapshotID) String() string {
	return string(si)
}

func (si SnapshotID) Set(snapshotid string) error {
	if !strings.HasPrefix(snapshotid, "snap-") {
		return ErrInvalidSnapshotID
	}

	si = SnapshotID(snapshotid)
	return nil
}

func (si SnapshotID) Type() string {
	return "amazonSnapshotID"
}

// AMIID is an EC2 AMI identifier.
// It must begin with "ami-".
type AMIID string

func (ai AMIID) String() string {
	return string(ai)
}

func (ai AMIID) Set(amiid string) error {
	if !strings.HasPrefix(amiid, "ami-") {
		return ErrInvalidAMIID
	}

	ai = AMIID(amiid)
	return nil
}

func (ai AMIID) Type() string {
	return "amazonAMIID"
}

// InstanceID is an EC2 instance identifier.
// It must begin with "i-".
type InstanceID string

func (ii InstanceID) String() string {
	return string(ii)
}

func (ii InstanceID) Set(instanceid string) error {
	if !strings.HasPrefix(instanceid, "i-") {
		return ErrInvalidInstanceID
	}

	ii = InstanceID(instanceid)
	return nil
}

func (ii InstanceID) Type() string {
	return "amazonInstanceID"
}
