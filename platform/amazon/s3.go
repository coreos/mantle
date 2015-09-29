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
	"fmt"
	"io"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var ErrS3ObjectExists = errors.New("s3 object already exists")

// S3Put reads data and uploads it to the S3 bucket, and stores it in key. If
// overwrite is true, existing data will be overwritten, otherwise an error
// will be returned if it exists. S3Put returns a URL that can be used to
// access the uploaded data, or an error on failure.
func (a *AWSAPI) S3Put(data io.Reader, bucket Bucket, key string, overwrite bool) (*url.URL, error) {
	if !overwrite {
		// check if object exists first
		_, err := a.s3api.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(bucket.String()),
			Key:    &key,
		})

		if err == nil {
			return nil, ErrS3ObjectExists
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
		Bucket: aws.String(bucket.String()),
		Key:    &key,
		Body:   data,
	}

	_, err := s3uploader.Upload(upload)
	if err != nil {
		return nil, err
	}

	return S3ObjectUrl(bucket, key), nil
}

// S3Delete deletes the object key in the S3 bucket.
func (a *AWSAPI) S3Delete(bucket Bucket, key string) error {
	di := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket.String()),
		Key:    &key,
	}

	_, err := a.s3api.DeleteObject(di)
	return err
}

// S3ObjectUrl constructs a URL that can access the key in the S3 bucket.
func S3ObjectUrl(bucket Bucket, key string) *url.URL {
	u := &url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s.s3.amazonaws.com", bucket),
		Path:   key,
	}

	return u
}
