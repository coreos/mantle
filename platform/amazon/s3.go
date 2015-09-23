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
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func (a *AWSAPI) S3Put(data io.Reader, key string, overwrite bool) (string, error) {
	if !overwrite {
		// check if object exists first
		_, err := a.s3api.HeadObject(&s3.HeadObjectInput{
			Bucket: &a.s3bucket,
			Key:    &key,
		})

		if err == nil {
			// object exists - return the old data.
			return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", a.s3bucket, key), nil
		}
	}

	s3uploader := s3manager.NewUploader(&s3manager.UploadOptions{
		PartSize:    8 * 1024 * 1024,
		Concurrency: 2,
		S3:          a.s3api,
	})

	upload := &s3manager.UploadInput{
		// XXX: should be private?
		ACL:    aws.String("public-read"),
		Bucket: &a.s3bucket,
		Key:    &key,
		Body:   data,
	}

	_, err := s3uploader.Upload(upload)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", a.s3bucket, key), nil
}

func (a *AWSAPI) S3Delete(key string) error {
	di := &s3.DeleteObjectInput{
		Bucket: &a.s3bucket,
		Key:    &key,
	}

	_, err := a.s3api.DeleteObject(di)
	return err
}
