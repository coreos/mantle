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
//
// Portions of this file licensed under:
//
// The MIT License (MIT)
//
// Copyright (c) 2015 Dalton Hubble
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package oci

import (
	"net/http"
)

type Transport struct {
	Base   http.RoundTripper
	signer *Signer
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// RoundTripper should not modify the given request, clone it
	req2 := cloneRequest(req)
	err := t.signer.SignRequest(req2)
	if err != nil {
		return nil, err
	}
	return t.base().RoundTrip(req2)
}

func (t *Transport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

// cloneRequest returns a clone of the given *http.Request with a shallow
// copy of struct fields and a deep copy of the Header map.
func cloneRequest(req *http.Request) *http.Request {
	// shallow copy the struct
	r2 := new(http.Request)
	*r2 = *req
	// deep copy Header so setting a header on the clone does not affect original
	r2.Header = make(http.Header, len(req.Header))
	for k, s := range req.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	return r2
}
