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
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/coreos/pkg/capnslog"

	"github.com/coreos/mantle/platform"
)

var (
	plog = capnslog.NewPackageLogger("github.com/coreos/mantle", "platform/api/aws")

	DefaultRegion = "us-west-1"
)

// Options contains AWS-specific options.
type Options struct {
	*platform.Options
	Region        string
	Bucket        string
	AMI           string
	InstanceType  string
	SecurityGroup string
}

// API is a convenience wrapper for Amazon S3 and EC2 APIs.
type API struct {
	session *session.Session
	s3api   *s3.S3
	ec2api  *ec2.EC2

	options *Options
}

// New creates a new AWS API wrapper. It uses credentials from AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables.
func New(opts *Options) (*API, error) {
	creds := credentials.NewEnvCredentials()
	if _, err := creds.Get(); err != nil {
		return nil, fmt.Errorf("no AWS credentials provided: %v", err)
	}

	cfg := aws.NewConfig().WithRegion(opts.Region).WithCredentials(creds)

	sess := session.New(cfg)

	api := &API{
		s3api:   s3.New(sess),
		ec2api:  ec2.New(sess),
		options: opts,
	}

	return api, nil
}
