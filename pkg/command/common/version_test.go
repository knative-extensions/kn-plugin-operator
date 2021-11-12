/*
Copyright 2021 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"testing"

	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
)

type VersionResult struct {
	Validity bool
	Major    string
}

func TestGetMajor(t *testing.T) {
	for _, tt := range []struct {
		name         string
		inputVersion string
		expected     VersionResult
	}{{
		name:         "ValidSemanticVersion",
		inputVersion: "v1.9.0",
		expected:     VersionResult{Validity: true, Major: "v1"},
	}, {
		name:         "InvalidSemanticVersion",
		inputVersion: "vIHaveNoIdea",
		expected:     VersionResult{Validity: false, Major: ""},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			validity, major := GetMajor(tt.inputVersion)
			testingUtil.AssertEqual(t, validity, tt.expected.Validity)
			testingUtil.AssertEqual(t, major, tt.expected.Major)
		})
	}
}
