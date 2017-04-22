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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coreos/mantle/auth"
	"github.com/coreos/mantle/cli"
	"github.com/coreos/mantle/sdk"
)

var (
	root = &cobra.Command{
		Use:   "gangue",
		Short: "Google Storage download and verification tool",
	}

	get = &cobra.Command{
		Use:   "get [url] [path]",
		Short: "download and verify a file from Google Storage",
		Run:   run,
	}

	coreosBoard, coreosVersion   string
	gpgKeyFile, jsonKeyFile      string
	keepSig, serviceAuth, verify bool
)

func init() {
	bv := get.PersistentFlags().BoolVar
	sv := get.PersistentFlags().StringVar

	board, version := getReleaseInfo()
	sv(&coreosBoard, "coreos-board", board, "select a release board for searching packages")
	sv(&coreosVersion, "coreos-version", version, "select a release version for searching packages")

	bv(&serviceAuth, "service-auth", false, "use non-interactive auth when running within GCE")
	sv(&jsonKeyFile, "json-key", "", "use a service account's JSON key for authentication")
	bv(&verify, "verify", true, "use GPG verification")
	sv(&gpgKeyFile, "verify-key", "", "PGP public key file to verify signatures, or blank for the default key built into the program")
	bv(&keepSig, "keep-sig", false, "keep the detached signature file on disk when successful")
	root.AddCommand(get)
}

// Ensure the given URL correctly points to a file on Google Storage
func validateGSURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if parsedURL.Scheme != "gs" {
		return fmt.Errorf("URL missing gs:// scheme: %v", rawURL)
	}
	if parsedURL.Host == "" {
		return fmt.Errorf("URL missing bucket name %v", rawURL)
	}
	if parsedURL.Path == "" {
		return fmt.Errorf("URL missing file path %v", rawURL)
	}
	if parsedURL.Path[len(parsedURL.Path)-1] == '/' {
		return fmt.Errorf("URL must not be a directory path %v", rawURL)
	}
	return nil
}

// Download a file using all the given command-line settings
func downloadURL(client *http.Client, source, output string) (err error) {
	if verify {
		err = sdk.UpdateSignedFile(output, source, client, gpgKeyFile)
		if err == nil && !keepSig {
			err = os.Remove(output + ".sig")
		}
	} else {
		err = sdk.UpdateFile(output, source, client)
	}
	return
}

func run(cmd *cobra.Command, args []string) {
	var client *http.Client
	var output, source string
	var err error

	if len(args) == 2 {
		source = args[0]
		output = args[1]
		if output == "" {
			output = "."
		}
	} else if len(args) == 1 {
		source = args[0]
		output = "."
	} else {
		fmt.Fprintf(os.Stderr, "Expected one or two arguments\n")
		os.Exit(1)
	}

	// Authenticate with Google
	if serviceAuth {
		client = auth.GoogleServiceClient()
	} else if jsonKeyFile != "" {
		b, err := ioutil.ReadFile(jsonKeyFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		client, err = auth.GoogleClientFromJSONKey(b)
	} else {
		client, err = auth.GoogleClient()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Construct a URL when given <category>/<package> or just <package>
	if fs := strings.IndexByte(source, '/'); fs < 0 || fs > 0 && strings.LastIndexByte(source, '/') == fs {
		if source, err = findPackageURL(client, source); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	}

	// Validate the URL before using it
	err = validateGSURL(source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// If the output path exists and is a directory, keep the file name
	if stat, err := os.Stat(output); err == nil && stat.IsDir() {
		output = path.Join(output, path.Base(source))
	}

	// Download the file and verify it (unless disabled)
	if err = downloadURL(client, source, output); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func main() {
	cli.Execute(root)
}
