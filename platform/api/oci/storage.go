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
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/objectstorage"
)

func (a *API) DeleteObject(bucketName, name string) error {
	namespace, err := a.GetNamespace()
	if err != nil {
		return err
	}

	_, err = a.os.DeleteObject(context.Background(), objectstorage.DeleteObjectRequest{
		NamespaceName: &namespace,
		BucketName:    &bucketName,
		ObjectName:    &name,
	})

	return err
}

func (a *API) GetNamespace() (string, error) {
	namespace, err := a.os.GetNamespace(context.Background(), objectstorage.GetNamespaceRequest{})
	if err != nil {
		return "", err
	}
	if namespace.Value == nil {
		return "", fmt.Errorf("received namespace nil")
	}
	return *namespace.Value, nil
}

func (a *API) HeadObject(bucketName, name string) (objectstorage.HeadObjectResponse, error) {
	namespace, err := a.GetNamespace()
	if err != nil {
		return objectstorage.HeadObjectResponse{}, err
	}

	return a.os.HeadObject(context.Background(), objectstorage.HeadObjectRequest{
		NamespaceName: &namespace,
		BucketName:    &bucketName,
		ObjectName:    &name,
	})
}
