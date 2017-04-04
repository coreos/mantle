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

// Package lunaform provides a Terraform based test harness that works
// similarly to the standard "testing" Go package. Each top level test is
// registered with the harness with a struct defining the requirements of
// the test. During execution a fresh set of virtual machines are created
// using Terraform before calling each test function and then destroyed
// afterwards.
package lunaform
