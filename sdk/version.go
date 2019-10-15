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

package sdk

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	coreosId = "{E96281A6-D1AF-4BDE-9A0A-97B76E56DC57}"
)

type Versions struct {
	Version    string
	VersionID  string
	BuildID    string
	SDKVersion string
}

func unquote(s string) string {
	if len(s) < 2 {
		return s
	}
	for _, q := range []byte{'\'', '"'} {
		if s[0] == q && s[len(s)-1] == q {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func parseVersions(f *os.File, prefix string) (ver Versions, err error) {
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.SplitN(scanner.Text(), "=", 2)
		if len(line) != 2 {
			continue
		}
		switch line[0] {
		case prefix + "VERSION":
			ver.Version = unquote(line[1])
		case prefix + "VERSION_ID":
			ver.VersionID = unquote(line[1])
		case prefix + "BUILD_ID":
			ver.BuildID = unquote(line[1])
		case prefix + "SDK_VERSION":
			ver.SDKVersion = unquote(line[1])
		}
	}
	if err = scanner.Err(); err != nil {
		return
	}

	if ver.VersionID == "" {
		err = fmt.Errorf("Missing %sVERSION_ID in %s", prefix, f.Name())
	} else if !strings.HasPrefix(ver.Version, ver.VersionID) {
		err = fmt.Errorf("Invalid %sVERSION in %s", prefix, f.Name())
	}

	return
}

func VersionsFromDir(dir string) (ver Versions, err error) {
	f, err := os.Open(filepath.Join(dir, "version.txt"))
	if err != nil {
		return
	}
	defer f.Close()

	ver, err = parseVersions(f, "COREOS_")
	if ver.SDKVersion == "" {
		err = fmt.Errorf("Missing COREOS_SDK_VERSION in %s", f.Name())
	}

	return
}
