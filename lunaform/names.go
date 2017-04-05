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
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// regex for cleaning up host names, meets GCE's requirements.
var nameFilterRe *regexp.Regexp

func init() {
	nameFilterRe = regexp.MustCompile(`[^-a-z0-9]`)
}

// strip any invalid characters and truncate
func nameFilter(name string, limit int) string {
	name = strings.ToLower(name)
	name = nameFilterRe.ReplaceAllString(name, "")
	if len(name) > limit {
		name = name[:limit]
	}
	return name
}

// detect the process name
func getProc() string {
	var name string
	if len(os.Args) >= 1 {
		name = nameFilter(filepath.Base(os.Args[0]), 9)
	}
	// be extra fussy and require a letter at the beginning
	if name == "" || name[0] < 'a' || name[0] > 'z' {
		name = "lunaform"
	}
	return name
}

// detect build tag from the jenkins, otherwise use user name and time.
func getTag() string {
	tag := nameFilter(os.Getenv("BUILD_TAG"), 30)
	if tag != "" {
		return tag
	}

	tag = time.Now().UTC().Format("20060102-150405")
	name := nameFilter(os.Getenv("USER"), 14)
	if name != "" {
		tag = name + "-" + tag
	}
	return tag
}

// generate something random
func getRand() []byte {
	b := make([]byte, 8)
	// it is a pretty bad day if this fails, just bail.
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic(fmt.Errorf("random source failed: %v", err))
	}
	return b
}

// NamePrefix generates a hopefully unique resource name prefix that
// includes the current process and build tag or user and time for context.
// The result is at most 57 characters so a small suffix can be added and
// still fit into the 63 character host name limit.
func NamePrefix() string {
	return fmt.Sprintf("%s-%s-%x",
		getProc(), // 9 chars max
		getTag(),  // 30 chars max
		getRand()) // 16 chars
}
