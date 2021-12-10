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

package install

import (
	"fmt"
	"testing"

	"knative.dev/kn-plugin-operator/pkg/command/common"
	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
)

func TestGetOperatorURL(t *testing.T) {
	for _, tt := range []struct {
		name         string
		inputVersion string
		expected     string
	}{{
		name:         "GetLatestOperatorURL",
		inputVersion: "latest",
		expected:     "https://github.com/knative/operator/releases/latest/download/operator.yaml",
	}, {
		name:         "GetV1OperatorURL",
		inputVersion: "1.0.0",
		expected:     "https://github.com/knative/operator/releases/download/knative-v1.0.0/operator.yaml",
	}, {
		name:         "GetV1OperatorURLWithPrefix",
		inputVersion: "v1.0.0",
		expected:     "https://github.com/knative/operator/releases/download/knative-v1.0.0/operator.yaml",
	}, {
		name:         "GetV0OperatorURL",
		inputVersion: "0.26.0",
		expected:     "https://github.com/knative/operator/releases/download/v0.26.0/operator.yaml",
	}, {
		name:         "GetV0OperatorURLWithPrefix",
		inputVersion: "v0.26.0",
		expected:     "https://github.com/knative/operator/releases/download/v0.26.0/operator.yaml",
	}} {
		t.Run(tt.name, func(t *testing.T) {
			URL, err := getOperatorURL(tt.inputVersion)
			testingUtil.AssertEqual(t, err, nil)
			testingUtil.AssertEqual(t, URL, tt.expected)
		})
	}
}

func TestGetOperatorURLInvalidVersion(t *testing.T) {
	inputVersion := "invalidVersion"
	for _, tt := range []struct {
		name         string
		inputVersion string
		expectedErr  error
	}{{
		name:         "GetOperatorURLInvalidVersion",
		inputVersion: inputVersion,
		expectedErr:  fmt.Errorf("%v is not a semantic version", inputVersion),
	}} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getOperatorURL(tt.inputVersion)
			testingUtil.AssertEqual(t, err == nil, false)
			testingUtil.AssertEqual(t, err.Error(), tt.expectedErr.Error())
		})
	}
}

func TestFillDefaultsForInstallCmdFlags(t *testing.T) {
	for _, tt := range []struct {
		name          string
		inputFlags    installCmdFlags
		expectedFlags installCmdFlags
	}{{
		name:       "Empty namespace and version for operator",
		inputFlags: installCmdFlags{},
		expectedFlags: installCmdFlags{
			Namespace: common.DefaultNamespace,
			Version:   common.Latest,
		},
	}, {
		name: "Empty istio namespace, namespace and version for serving",
		inputFlags: installCmdFlags{
			Component: "serving",
		},
		expectedFlags: installCmdFlags{
			Component:      "serving",
			IstioNamespace: common.DefaultIstioNamespace,
			Namespace:      common.DefaultKnativeServingNamespace,
			Version:        common.Latest,
		},
	}, {
		name: "Empty namespace and version for eventing",
		inputFlags: installCmdFlags{
			Component: "eventing",
		},
		expectedFlags: installCmdFlags{
			Component: "eventing",
			Namespace: common.DefaultKnativeEventingNamespace,
			Version:   common.Latest,
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			tt.inputFlags.fill_defaults()
			testingUtil.AssertEqual(t, tt.inputFlags, tt.expectedFlags)
		})
	}
}
