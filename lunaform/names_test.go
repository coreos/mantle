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

package lunaform

import (
	"regexp"
	"testing"
)

func TestNamePrefix(t *testing.T) {
	prefix := NamePrefix()
	if len(prefix) < 22 {
		t.Errorf("too short: %q", prefix)
	} else if len(prefix) > 57 {
		t.Errorf("too long: %q", prefix)
	}
	ok, err := regexp.MatchString(`^[a-z][-a-z0-9]*$`, prefix)
	if err != nil {
		t.Error(err)
	} else if !ok {
		t.Errorf("invalid characters in %q", prefix)
	}
}
