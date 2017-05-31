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

package maxint

import (
	"math"
	"testing"
)

func TestMaxUint(t *testing.T) {
	if MaxUint != math.MaxUint32 && MaxUint != math.MaxUint64 {
		t.Errorf("Bad value for MaxUint: %d", MaxUint)
	}
}

func TestMaxInt(t *testing.T) {
	if MaxInt != math.MaxInt32 && MaxInt != math.MaxInt64 {
		t.Errorf("Bad value for MaxInt: %d", MaxInt)
	}
}

func TestMinInt(t *testing.T) {
	if MinInt != math.MinInt32 && MinInt != math.MinInt64 {
		t.Errorf("Bad value for MinInt: %d", MinInt)
	}
}
