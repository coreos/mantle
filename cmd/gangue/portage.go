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

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// Read default parameters from the host for building PORTAGE_BINHOST
func getReleaseInfo() (board, version string) {
	contents, err := ioutil.ReadFile("/usr/share/coreos/release")
	if err != nil {
		return
	}

	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "COREOS_RELEASE_VERSION=") {
			version = strings.SplitAfterN(line, "=", 2)[1]
		} else if strings.HasPrefix(line, "COREOS_RELEASE_BOARD=") {
			board = strings.SplitAfterN(line, "=", 2)[1]
		}
	}

	// Drop the timestamp suffix while in the SDK
	if strings.Contains(version, "+") {
		version = strings.SplitN(version, "+", 2)[0]
	}

	// The board is empty while in the SDK, which is always amd64
	if board == "" {
		board = "amd64-usr"
	}

	return
}

// Test whether a short package name matches a full PVR string
func packageNameMatches(name, pvr string) bool {
	matchStart := pvr
	if !strings.Contains(name, "/") && strings.Contains(pvr, "/") {
		matchStart = strings.SplitAfterN(pvr, "/", 2)[1]
	}

	// When searching for a version number, match the complete string
	if match, _ := regexp.MatchString("-[0-9]", name); match {
		return name == matchStart
	}

	// If the start of the name matches, ensure it ends with the version
	if strings.HasPrefix(matchStart, name) {
		match, _ := regexp.MatchString("^-[0-9]", matchStart[len(name):])
		return match
	}

	return false
}

// Search through Packages, returning the full URL to a matching binary package
func findPackageURL(client *http.Client, name string) (string, error) {
	urlPrefix := fmt.Sprintf("gs://builds.developer.core-os.net/boards/%s/%s/pkgs/", coreosBoard, coreosVersion)

	file, err := ioutil.TempFile("", "Packages")
	if err != nil {
		return "", err
	}
	file.Close() // Close the file while its content is written
	tempFile := file.Name()
	defer os.Remove(tempFile)

	if err = downloadURL(client, urlPrefix+"Packages", tempFile); err != nil {
		return "", err
	}

	file, err = os.Open(tempFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var matches []string
	scanner := bufio.NewScanner(bufio.NewReader(file))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "CPV: ") {
			if pvr := strings.TrimSpace(strings.TrimPrefix(line, "CPV: ")); packageNameMatches(name, pvr) {
				matches = append(matches, pvr)
			}
		}
	}
	if err = scanner.Err(); err != nil {
		return "", err
	}

	if len(matches) < 1 {
		return "", fmt.Errorf("No package matching \"%s\" was found", name)
	} else if len(matches) > 1 {
		return "", fmt.Errorf("Several packages match \"%s\", pick one: %v", name, matches)
	}
	return urlPrefix + matches[0] + ".tbz2", nil
}
