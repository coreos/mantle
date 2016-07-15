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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/coreos/pkg/capnslog"
)

var (
	plog = capnslog.NewPackageLogger("github.com/coreos/mantle", "platform/amazon")
)

// AWSAPI is a convenience wrapper for Amazon S3 and EC2 APIs.
type AWSAPI struct {
	s3api  *s3.S3
	ec2api *ec2.EC2

	// Region to do AWS operations in.
	region Region

	// A description prefix attached to each request
	shortDescription string
}

func NewAWSAPI(region Region) (*AWSAPI, error) {
	creds := credentials.NewEnvCredentials()
	if _, err := creds.Get(); err != nil {
		return nil, fmt.Errorf("no AWS credentials provided: %v", err)
	}

	awsconfig := aws.NewConfig().WithCredentials(creds).WithRegion(region.String())

	api := &AWSAPI{
		s3api:  s3.New(awsconfig),
		ec2api: ec2.New(awsconfig),
		region: region,
	}

	return api, nil
}
