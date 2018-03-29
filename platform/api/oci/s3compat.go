// Copyright 2018 CoreOS, Inc.
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
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/oracle/oci-go-sdk/common"
)

func isNotFound(err error) bool {
	if ocierr, ok := err.(common.ServiceError); ok {
		return ocierr.GetHTTPStatusCode() == 404
	}
	return false
}

// UploadObject uploads an object to S3
func (a *API) UploadObject(r io.Reader, bucket, path string, force bool) error {
	s3uploader := s3manager.NewUploaderWithClient(a.s3)

	if !force {
		_, err := a.HeadObject(bucket, path)
		if err != nil {
			if !isNotFound(err) {
				return fmt.Errorf("unable to head object %v/%v: %v", bucket, path, err)
			}
		} else {
			plog.Infof("skipping upload since force was not set: s3://%v/%v", bucket, path)
			return nil
		}
	}

	_, err := s3uploader.Upload(&s3manager.UploadInput{
		Body:            r,
		Bucket:          aws.String(bucket),
		Key:             aws.String(path),
		ContentEncoding: aws.String("application/x-raw-disk-image"),
	})
	if err != nil {
		return fmt.Errorf("error uploading s3://%v/%v: %v", bucket, path, err)
	}
	return err
}
