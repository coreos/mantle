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

package oci

import (
	"context"
	"fmt"
	"os"

	"github.com/oracle/oci-go-sdk/objectstorage"
)

func (a *API) UploadImage(bucketName, name, filePath string) (objectstorage.PutObjectResponse, error) {
	namespace, err := a.os.GetNamespace(context.Background(), objectstorage.GetNamespaceRequest{})
	if err != nil {
		return objectstorage.PutObjectResponse{}, err
	}
	if namespace.Value == nil {
		return objectstorage.PutObjectResponse{}, fmt.Errorf("received namespace nil")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return objectstorage.PutObjectResponse{}, err
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return objectstorage.PutObjectResponse{}, err
	}
	size := int(fi.Size())

	return a.os.PutObject(context.Background(), objectstorage.PutObjectRequest{
		NamespaceName: namespace.Value,
		BucketName:    &bucketName,
		ObjectName:    &name,
		ContentLength: &size,
		PutObjectBody: file,
		ContentType:   strToPtr("application/octet-stream"),
	})
}
