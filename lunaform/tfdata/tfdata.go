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

// Package tfdata provides data types for parsing Terraform output.
package tfdata

import (
	"encoding/json"
)

type Type string

const (
	String Type = "string"
	List   Type = "list"
	Map    Type = "map"
)

type Output struct {
	Sensitive bool `json:"sensitive"`
	Type      Type `json:"type"`
}

type OutputString struct {
	Output
	Value string
}

type OutputList struct {
	Output
	Value []string
}

type OutputMap struct {
	Output
	Value map[string]string
}

// ParseOutput should be given the output of `terraform output -json`
// and a struct containing Output* values.
func ParseOutput(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
