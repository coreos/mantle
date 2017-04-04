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

package tfdata

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

const (
	testOutputData = `{
	"l": {
		"sensitive": false,
		"type": "list",
		"value": [ "a", "b" ]
	},
	"m": {
		"sensitive": false,
		"type": "map",
		"value": { "A": "a", "B": "b" }
	},
	"s": {
		"sensitive": false,
		"type": "string",
		"value": "a"
	}
}`
)

type testOutput struct {
	L OutputList
	M OutputMap
	S OutputString
}

var (
	testOutputExpect = testOutput{
		L: OutputList{
			Output: Output{Type: List},
			Value:  []string{"a", "b"},
		},
		M: OutputMap{
			Output: Output{Type: Map},
			Value:  map[string]string{"A": "a", "B": "b"},
		},
		S: OutputString{
			Output: Output{Type: String},
			Value:  "a",
		},
	}
)

func TestOutput(t *testing.T) {
	var parsed testOutput
	if err := ParseOutput([]byte(testOutputData), &parsed); err != nil {
		t.Fatal(err)
	}
	if diff := pretty.Compare(parsed, testOutputExpect); diff != "" {
		t.Fatal(diff)
	}
}
