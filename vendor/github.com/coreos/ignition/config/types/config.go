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

package types

import (
	"reflect"

	"github.com/coreos/go-semver/semver"
)

var (
	MaxVersion = semver.Version{
		Major: 2,
		Minor: 0,
	}
)

type Config struct {
	Ignition Ignition `json:"ignition"`
	Storage  Storage  `json:"storage,omitempty"`
	Systemd  Systemd  `json:"systemd,omitempty"`
	Networkd Networkd `json:"networkd,omitempty"`
	Passwd   Passwd   `json:"passwd,omitempty"`
}

func (c Config) AssertValid() error {
	return assertStructValid(reflect.ValueOf(c))
}

func assertValid(vObj reflect.Value) error {
	if obj, ok := vObj.Interface().(interface {
		AssertValid() error
	}); ok {
		if err := obj.AssertValid(); err != nil {
			return err
		}
	}

	switch vObj.Kind() {
	case reflect.Struct:
		return assertStructValid(vObj)
	case reflect.Slice:
		for i := 0; i < vObj.Len(); i++ {
			if err := assertValid(vObj.Index(i)); err != nil {
				return err
			}
		}
	}

	return nil
}

func assertStructValid(vObj reflect.Value) error {
	for i := 0; i < vObj.Type().NumField(); i++ {
		if err := assertValid(vObj.Field(i)); err != nil {
			return err
		}
	}
	return nil
}
