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

package oci

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coreos/mantle/auth"
)

type MockTransport struct{}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	response := &http.Response{
		Header:     make(http.Header),
		Request:    req,
		StatusCode: http.StatusOK,
	}
	return response, nil
}

func getDefaultAuthProfile(t *testing.T, tempDir string) auth.OCIProfile {
	exampleRSAPrivateKey := `-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDCFENGw33yGihy92pDjZQhl0C36rPJj+CvfSC8+q28hxA161QF
NUd13wuCTUcq0Qd2qsBe/2hFyc2DCJJg0h1L78+6Z4UMR7EOcpfdUE9Hf3m/hs+F
UR45uBJeDK1HSFHD8bHKD6kv8FPGfJTotc+2xjJwoYi+1hqp1fIekaxsyQIDAQAB
AoGBAJR8ZkCUvx5kzv+utdl7T5MnordT1TvoXXJGXK7ZZ+UuvMNUCdN2QPc4sBiA
QWvLw1cSKt5DsKZ8UETpYPy8pPYnnDEz2dDYiaew9+xEpubyeW2oH4Zx71wqBtOK
kqwrXa/pzdpiucRRjk6vE6YY7EBBs/g7uanVpGibOVAEsqH1AkEA7DkjVH28WDUg
f1nqvfn2Kj6CT7nIcE3jGJsZZ7zlZmBmHFDONMLUrXR/Zm3pR5m0tCmBqa5RK95u
412jt1dPIwJBANJT3v8pnkth48bQo/fKel6uEYyboRtA5/uHuHkZ6FQF7OUkGogc
mSJluOdc5t6hI1VsLn0QZEjQZMEOWr+wKSMCQQCC4kXJEsHAve77oP6HtG/IiEn7
kpyUXRNvFsDE0czpJJBvL/aRFUJxuRK91jhjC68sA7NsKMGg5OXb5I5Jj36xAkEA
gIT7aFOYBFwGgQAQkWNKLvySgKbAZRTeLBacpHMuQdl1DfdntvAyqpAZ0lY0RKmW
G6aFKaqQfOXKCyWoUiVknQJAXrlgySFci/2ueKlIE1QqIiLSZ8V8OlpFLRnb1pzI
7U1yQXnTAEFYM560yJlzUpOb1V4cScGd365tiSMvxLOvTA==
-----END RSA PRIVATE KEY-----`
	rsaPath := filepath.Join(tempDir, "tempRSA.pem")
	file, err := os.Create(rsaPath)
	if err != nil {
		t.Fatalf("creating temp rsa file: %v", err)
	}
	defer file.Close()
	file.WriteString(exampleRSAPrivateKey)

	exampleTenancy := "ocid1.tenancy.oc1..aaaaaaaaba3pv6wkcr4jqae5f15p2b2m2yt2j6rx32uzr4h25vqstifsfdsq"
	exampleUser := "ocid1.user.oc1..aaaaaaaat5nvwcna5j6aqzjcaty5eqbb6qt2jvpkanghtgdaqedqw3rynjq"
	exampleFingerprint := "20:3b:97:13:55:1c:5b:0d:d3:37:d8:50:4e:c5:3a:34"
	exampleRegion := "us-phoenix-1"

	profile := auth.OCIProfile{
		TenancyID:   exampleTenancy,
		UserID:      exampleUser,
		Fingerprint: exampleFingerprint,
		Region:      exampleRegion,
		KeyFile:     rsaPath,
	}

	return profile
}

func TestAuthOCIHelperMethods(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "mantle-oci-tests")
	if err != nil {
		t.Fatalf("creating tempdir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	profile := getDefaultAuthProfile(t, tempDir)

	t.Run("Key ID", func(t *testing.T) {
		expectedKeyID := strings.Join([]string{profile.TenancyID, profile.UserID, profile.Fingerprint}, "/")

		if expectedKeyID != profile.GetKeyID() {
			t.Fatalf("KeyID differs:\n\texpected: %s\n\treceived: %s", expectedKeyID, profile.GetKeyID())
		}
	})

	t.Run("Private Key", func(t *testing.T) {
		// The actual Fingeprint of the example RSA key
		profile.Fingerprint = "73:61:a2:21:67:e0:df:be:7e:4b:93:1e:15:98:a5:b7"
		_, err := profile.GetPrivateKey()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Private Key Wrong Fingerprint", func(t *testing.T) {
		profile.Fingerprint = "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00"
		_, err := profile.GetPrivateKey()
		if err == nil || err.Error() != "fingerprint given differs from actual key fingerprint" {
			t.Fatalf("didn't detect wrong fingerprint: %v", err)
		}
	})
}

func TestHTTPClient(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "mantle-oci-tests")
	if err != nil {
		t.Fatalf("creating tempdir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	profile := getDefaultAuthProfile(t, tempDir)

	ctx := context.WithValue(context.TODO(), HTTPClient, &http.Client{Transport: &MockTransport{}})

	// build *rsa.PrivateKey manually to bypass the fingerprint check
	// the actual fingerprint of the example key differs from the one
	// used by the example for the signing.
	keyData, err := ioutil.ReadFile(profile.KeyFile)
	if err != nil {
		t.Fatalf("reading RSA key: %v", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		t.Fatalf("parsing PEM block")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("parsing RSA key: %v", err)
	}

	client := NewHTTPClient(ctx, priv, profile.GetKeyID())

	t.Run("GET Call", func(t *testing.T) {
		clientGetCall(t, client)
	})

	t.Run("POST Call", func(t *testing.T) {
		clientPostCall(t, client)
	})
}

func clientGetCall(t *testing.T, client *http.Client) {
	uri := "https://iaas.us-phoenix-1.oraclecloud.com/20160918/instances?availabilityDomain=Pjwf%3A%20PHX-AD-1&compartmentId=ocid1.compartment.oc1..aaaaaaaam3we6vgnherjq5q2idnccdflvjsnog7mlr6rtdb25gilchfeyjxa&displayName=TeamXInstances&volumeId=ocid1.volume.oc1.phx.abyhqljrgvttnlx73nmrwfaux7kcvzfs3s66izvxf2h4lgvyndsdsnoiwr5q"

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	req.Header.Set("date", "Thu, 05 Jan 2014 21:31:40 GMT")

	resp, _ := client.Do(req)

	auth := resp.Request.Header.Get("Authorization")

	expectedValues := map[string]string{
		"keyId":     "keyId=\"ocid1.tenancy.oc1..aaaaaaaaba3pv6wkcr4jqae5f15p2b2m2yt2j6rx32uzr4h25vqstifsfdsq/ocid1.user.oc1..aaaaaaaat5nvwcna5j6aqzjcaty5eqbb6qt2jvpkanghtgdaqedqw3rynjq/20:3b:97:13:55:1c:5b:0d:d3:37:d8:50:4e:c5:3a:34\"",
		"algorithm": "algorithm=\"rsa-sha256\"",
		"version":   "version=\"1\"",
		"headers":   "headers=\"date (request-target) host\"",
		"signature": "signature=\"GBas7grhyrhSKHP6AVIj/h5/Vp8bd/peM79H9Wv8kjoaCivujVXlpbKLjMPeDUhxkFIWtTtLBj3sUzaFj34XE6YZAHc9r2DmE4pMwOAy/kiITcZxa1oHPOeRheC0jP2dqbTll8fmTZVwKZOKHYPtrLJIJQHJjNvxFWeHQjMaR7M=\"",
	}

	values := ParseAuthorization(t, auth)
	for key, expectedVal := range expectedValues {
		if actualVal, ok := values[key]; !ok {
			t.Fatalf("couldn't find %s in authorization", key)
		} else if expectedVal != actualVal {
			t.Fatalf("%s is incorrect:\n\texpected: %s\n\treceived: %s", key, expectedVal, actualVal)
		}
	}
}

func clientPostCall(t *testing.T, client *http.Client) {
	uri := "https://iaas.us-phoenix-1.oraclecloud.com/20160918/volumeAttachments"

	req, err := http.NewRequest("POST", uri, nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	req.Body = ioutil.NopCloser(bytes.NewReader([]byte(`{
    "compartmentId": "ocid1.compartment.oc1..aaaaaaaam3we6vgnherjq5q2idnccdflvjsnog7mlr6rtdb25gilchfeyjxa",
    "instanceId": "ocid1.instance.oc1.phx.abuw4ljrlsfiqw6vzzxb43vyypt4pkodawglp3wqxjqofakrwvou52gb6s5a",
    "volumeId": "ocid1.volume.oc1.phx.abyhqljrgvttnlx73nmrwfaux7kcvzfs3s66izvxf2h4lgvyndsdsnoiwr5q"
}`)))

	req.Header.Set("date", "Thu, 05 Jan 2014 21:31:40 GMT")

	resp, _ := client.Do(req)

	auth := resp.Request.Header.Get("Authorization")

	expectedValues := map[string]string{
		"keyId":     "keyId=\"ocid1.tenancy.oc1..aaaaaaaaba3pv6wkcr4jqae5f15p2b2m2yt2j6rx32uzr4h25vqstifsfdsq/ocid1.user.oc1..aaaaaaaat5nvwcna5j6aqzjcaty5eqbb6qt2jvpkanghtgdaqedqw3rynjq/20:3b:97:13:55:1c:5b:0d:d3:37:d8:50:4e:c5:3a:34\"",
		"algorithm": "algorithm=\"rsa-sha256\"",
		"version":   "version=\"1\"",
		"headers":   "headers=\"date (request-target) host content-length content-type x-content-sha256\"",
		"signature": "signature=\"Mje8vIDPlwIHmD/cTDwRxE7HaAvBg16JnVcsuqaNRim23fFPgQfLoOOxae6WqKb1uPjYEl0qIdazWaBy/Ml8DRhqlocMwoSXv0fbukP8J5N80LCmzT/FFBvIvTB91XuXI3hYfP9Zt1l7S6ieVadHUfqBedWH0itrtPJBgKmrWso=\"",
	}

	values := ParseAuthorization(t, auth)
	for key, expectedVal := range expectedValues {
		if actualVal, ok := values[key]; !ok {
			t.Fatalf("couldn't find %s in authorization", key)
		} else if expectedVal != actualVal {
			t.Fatalf("%s is incorrect:\n\texpected: %s\n\treceived: %s", key, expectedVal, actualVal)
		}
	}

	contentHash := "V9Z20UJTvkvpJ50flBzKE32+6m2zJjweHpDMX/U4Uy0="
	if resp.Request.Header.Get("x-content-sha256") != contentHash {
		t.Fatalf("x-content-sha256 wrong:\nexpected: %s\nreceived: %s", contentHash, resp.Request.Header.Get("x-content-sha256"))
	}

	if resp.Request.Header.Get("content-length") != "316" {
		t.Fatalf("content length wrong:\nexpected: %s\nreceived: %s", "316", resp.Request.Header.Get("content-length"))
	}
}
