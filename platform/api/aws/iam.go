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

package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/coreos/mantle/util"
)

const (
	ec2AssumeRolePolicy = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`
	s3ReadOnlyAccess = `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:Get*",
                "s3:List*"
            ],
            "Resource": "*"
        }
    ]
}`
)

// ensureInstanceProfile checks that the specified instance profile exists,
// and creates it and its backing role if not. The role will have the
// AmazonS3RReadOnlyAccess permissions policy applied to allow fetches
// of S3 objects that are not owned by the root account.
func (a *API) ensureInstanceProfile(name string) error {
	profileExists := false
	policy := "AmazonS3ReadOnlyAccess"

	profile, err := a.iam.GetInstanceProfile(&iam.GetInstanceProfileInput{
		InstanceProfileName: &name,
	})
	if awsErr, ok := err.(awserr.Error); !ok || awsErr.Code() != "NoSuchEntity" {
		return fmt.Errorf("getting instance profile %q: %v", name, err)
	} else if err == nil {
		profileExists = true
		// check if a role of the same name already exists and is attached
		// to the instance profile
		for _, role := range profile.InstanceProfile.Roles {
			if role.RoleName != nil && *role.RoleName == name {
				if role.AssumeRolePolicyDocument != nil && *role.AssumeRolePolicyDocument != ec2AssumeRolePolicy {
					// the role exists but is missing the assume role policy
					// present this as an error to the user requiring manual intervention
					return fmt.Errorf("role %q exists but is not configured properly, manual intervention required", name)
				}
				// validate that the role has the correct policy
				_, err := a.iam.GetRolePolicy(&iam.GetRolePolicyInput{
					PolicyName: aws.String(policy),
					RoleName:   aws.String(name),
				})
				if awserr, ok := err.(awserr.Error); !ok || awserr.Code() != "NoSuchEntity" {
					return fmt.Errorf("getting role policy: %v", err)
				} else if awserr.Code() == "NoSuchEntity" {
					// the role does not contain the AmazonS3ReadOnlyAccess policy
					// attempt to add it to allow the existing instance prrofile
					// to be re-used
					_, err = a.iam.PutRolePolicy(&iam.PutRolePolicyInput{
						PolicyName:     &policy,
						PolicyDocument: aws.String(s3ReadOnlyAccess),
						RoleName:       &name,
					})
					if err != nil {
						return fmt.Errorf("adding %q policy to role %q: %v", policy, name, err)
					}
				}
				// the role exists with the correct policy and is attached to the instance profile
				// re-use the existing instance profile
				return nil
			}
		}
	}

	_, err = a.iam.CreateRole(&iam.CreateRoleInput{
		RoleName:                 &name,
		Description:              aws.String("mantle role for testing"),
		AssumeRolePolicyDocument: aws.String(ec2AssumeRolePolicy),
	})
	if err != nil {
		return fmt.Errorf("creating role %q: %v", name, err)
	}
	_, err = a.iam.PutRolePolicy(&iam.PutRolePolicyInput{
		PolicyName:     &policy,
		PolicyDocument: aws.String(s3ReadOnlyAccess),
		RoleName:       &name,
	})
	if err != nil {
		return fmt.Errorf("adding %q policy to role %q: %v", policy, name, err)
	}

	if !profileExists {
		_, err = a.iam.CreateInstanceProfile(&iam.CreateInstanceProfileInput{
			InstanceProfileName: &name,
		})
		if err != nil {
			return fmt.Errorf("creating instance profile %q: %v", name, err)
		}
	}

	_, err = a.iam.AddRoleToInstanceProfile(&iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: &name,
		RoleName:            &name,
	})
	if err != nil {
		return fmt.Errorf("adding role %q to instance profile %q: %v", name, name, err)
	}

	// wait for instance profile to fully exist in IAM before returning.
	// note that this does not guarantee that it will exist within ec2.
	err = util.WaitUntilReady(30*time.Second, 5*time.Second, func() (bool, error) {
		_, err = a.iam.GetInstanceProfile(&iam.GetInstanceProfileInput{
			InstanceProfileName: &name,
		})
		if err != nil {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("waiting for instance profile to become ready: %v", err)
	}

	return nil
}
