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
	"time"

	"github.com/oracle/oci-go-sdk/core"

	"github.com/coreos/mantle/util"
)

func (a *API) createImage(bucketName, name string) (core.CreateImageResponse, error) {
	namespace, err := a.GetNamespace()
	if err != nil {
		return core.CreateImageResponse{}, err
	}

	return a.compute.CreateImage(context.Background(), core.CreateImageRequest{
		CreateImageDetails: core.CreateImageDetails{
			CompartmentId: &a.opts.CompartmentID,
			DisplayName:   &name,
			ImageSourceDetails: core.ImageSourceViaObjectStorageTupleDetails{
				BucketName:    &bucketName,
				NamespaceName: &namespace,
				ObjectName:    &name,
			},
		},
	})
}

func (a *API) CreateImage(bucketName, name string) (string, error) {
	image, err := a.createImage(bucketName, name)
	if err != nil {
		return "", fmt.Errorf("creating image: %v", err)
	}
	if image.Image.Id == nil {
		return "", fmt.Errorf("received image id nil")
	}

	err = util.WaitUntilReady(10*time.Minute, 10*time.Second, func() (bool, error) {
		img, err := a.GetImage(*image.Image.Id)
		if err != nil {
			return false, fmt.Errorf("retrieving image: %v", err)
		}

		if img.Image.LifecycleState == core.ImageLifecycleStateAvailable {
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return "", fmt.Errorf("waiting for image to be available: %v", err)
	}

	return *image.Image.Id, nil
}

func (a *API) GetImage(id string) (core.GetImageResponse, error) {
	return a.compute.GetImage(context.Background(), core.GetImageRequest{
		ImageId: &id,
	})
}
