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
	"io"
	"net/url"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func (a *API) Get(w io.WriterAt, key string) error {
	dl := s3manager.NewDownloaderWithClient(a.s3api)

	download := &s3.GetObjectInput{
		Bucket: &a.options.Bucket,
		Key:    &key,
	}

	_, err := dl.Download(w, download)

	return err
}

func (a *API) Put(r io.Reader, key, acl string, overwrite bool) (*url.URL, error) {
	if !overwrite {
		// check if object exists first
		_, err := a.s3api.HeadObject(&s3.HeadObjectInput{
			Bucket: &a.options.Bucket,
			Key:    &key,
		})

		if err == nil {
			// object exists - return the old data.
			u := &url.URL{
				Scheme: "https",
				Host:   fmt.Sprintf("%s.s3.amazonaws.com", a.options.Bucket),
				Path:   key,
			}

			return u, nil
		}
	}

	s3uploader := &s3manager.Uploader{
		PartSize:    8 * 1024 * 1024,
		Concurrency: 2,
		S3:          a.s3api,
	}

	upload := &s3manager.UploadInput{
		ACL:    &acl,
		Bucket: &a.options.Bucket,
		Key:    &key,
		Body:   r,
	}

	_, err := s3uploader.Upload(upload)
	if err != nil {
		return nil, err
	}

	u := &url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s.s3.amazonaws.com", a.options.Bucket),
		Path:   key,
	}

	return u, nil
}

func (a *API) Delete(key string) error {
	di := &s3.DeleteObjectInput{
		Bucket: &a.options.Bucket,
		Key:    &key,
	}

	_, err := a.s3api.DeleteObject(di)
	return err
}
