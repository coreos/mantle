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

package auth

import (
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/go-ini/ini"
)

const ociConfigPath = ".oraclebmc/config"

// OCIProfile represents a parsed OCI profile.
type OCIProfile struct {
	TenancyID   string `ini:"tenancy"`
	UserID      string `ini:"user"`
	Fingerprint string `ini:"fingerprint"`
	KeyFile     string `ini:"key_file"`
	Region      string `ini:"region"`
	Compartment string `ini:"compartment"`
}

// ReadOCIConfig builds an OCIProfile from the OCIConfig files.
// It takes a path (which defaults to $HOME/.oraclebmc/config) and will parse
// the standard configuration, defined at:
// https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/sdkconfig.htm
// It will then attempt to parse a file named mantle in the same directory to
// allow overrides & other variable definitions not used in the standard
// configuration.
func ReadOCIConfig(path string) (map[string]OCIProfile, error) {
	if path == "" {
		user, err := user.Current()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(user.HomeDir, ociConfigPath)
	}

	profiles := make(map[string]OCIProfile)

	// first parse the standard oracle config
	cfg, err := ini.InsensitiveLoad(path)
	if err != nil {
		return nil, fmt.Errorf("Loading OCI config: %v", err)
	}

	for _, section := range cfg.Sections() {
		p := OCIProfile{}
		err = section.MapTo(&p)
		if err != nil {
			return nil, err
		}

		profiles[strings.ToLower(section.Name())] = p
	}

	// attempt to parse the mantle config
	cfg, err = ini.InsensitiveLoad(filepath.Join(filepath.Dir(path), "mantle"))
	if err == nil {
		for _, section := range cfg.Sections() {
			var p OCIProfile
			var ok bool
			if p, ok = profiles[strings.ToLower(section.Name())]; !ok {
				p = OCIProfile{}
			}

			err = section.MapTo(&p)
			profiles[strings.ToLower(section.Name())] = p
		}
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("OCI config %q contains no profiles", path)
	}

	return profiles, nil
}

func (profile *OCIProfile) GetPrivateKey() (*rsa.PrivateKey, error) {
	keyData, err := ioutil.ReadFile(profile.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("reading RSA key: %v", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("parsing PEM block")
	}

	if priv, err := x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		return nil, fmt.Errorf("parsing RSA key: %v", err)
	} else {
		// Calculate the public key fingerprint
		m, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("marshaling PKIX public key: %v", err)
		}

		h := md5.New()
		h.Write(m)
		fp := fmt.Sprintf("%x", h.Sum(nil))
		for i := 2; i < len(fp); i += 3 {
			fp = fp[:i] + ":" + fp[i:]
		}

		if fp != profile.Fingerprint {
			return nil, fmt.Errorf("fingerprint given differs from actual key fingerprint")
		}

		return priv, nil
	}
}

func (profile *OCIProfile) GetKeyID() string {
	return strings.Join([]string{
		profile.TenancyID, profile.UserID, profile.Fingerprint,
	}, "/")
}
