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

func TestGetOverlayYamlContent(t *testing.T) {
	for _, tt := range []struct {
		name         string
		installFlags installCmdFlags
		expectedFile string
	}{{
		name: "Knative Serving",
		installFlags: installCmdFlags{
			Component: "serving",
		},
		expectedFile: "testdata/overlay/ks.yaml",
	}, {
		name: "Knative Serving with istio namespace",
		installFlags: installCmdFlags{
			Component:      "serving",
			IstioNamespace: "test",
		},
		expectedFile: "testdata/overlay/ks_istio_ns.yaml",
	}, {
		name: "Knative Eventing",
		installFlags: installCmdFlags{
			Component: "eventing",
		},
		expectedFile: "testdata/overlay/ke.yaml",
	}, {
		name: "Knative Operator",
		installFlags: installCmdFlags{
			Version: "1.2.0",
		},
		expectedFile: "testdata/overlay/operator.yaml",
	}, {
		name: "Knative Operator of 1.3",
		installFlags: installCmdFlags{
			Version: "1.3.0",
		},
		expectedFile: "testdata/overlay/full_operator.yaml",
	}} {
		t.Run(tt.name, func(t *testing.T) {
			tt.installFlags.fill_defaults()
			rootPath := "testdata"
			result := getOverlayYamlContent(tt.installFlags, rootPath)
			expectedResult, err := common.ReadFile(tt.expectedFile)
			testingUtil.AssertEqual(t, err == nil, true)
			testingUtil.AssertEqual(t, result, expectedResult)
		})
	}
}

func TestGetYamlValuesContent(t *testing.T) {
	for _, tt := range []struct {
		name           string
		installFlags   installCmdFlags
		expectedResult string
	}{{
		name: "Knative Serving with all parameters",
		installFlags: installCmdFlags{
			Namespace:      "test-serving",
			Component:      "serving",
			Version:        "1.0",
			IstioNamespace: "istio-namespace",
		},
		expectedResult: `#@data/values
---
name: knative-serving
namespace: test-serving
version: '1.0'
local_gateway_value: knative-local-gateway.istio-namespace.svc.cluster.local`,
	}, {
		name: "Knative Serving with namespace and version",
		installFlags: installCmdFlags{
			Namespace: "test-serving-1",
			Component: "serving",
			Version:   "1.0.0",
		},
		expectedResult: `#@data/values
---
name: knative-serving
namespace: test-serving-1
version: '1.0.0'`,
	}, {
		name: "Knative Serving with namespace only",
		installFlags: installCmdFlags{
			Namespace: "test-serving-1",
			Component: "serving",
		},
		expectedResult: `#@data/values
---
name: knative-serving
namespace: test-serving-1
version: 'latest'`,
	}, {
		name: "Knative Serving with version only",
		installFlags: installCmdFlags{
			Version:   "1.0",
			Component: "serving",
		},
		expectedResult: `#@data/values
---
name: knative-serving
namespace: knative-serving
version: '1.0'`,
	}, {
		name: "Knative Eventing with namespace and version",
		installFlags: installCmdFlags{
			Namespace: "test-eventing",
			Component: "eventing",
			Version:   "1.0.0",
		},
		expectedResult: `#@data/values
---
name: knative-eventing
namespace: test-eventing
version: '1.0.0'`,
	}, {
		name: "Knative Eventing with namespace only",
		installFlags: installCmdFlags{
			Namespace: "test-eventing-1",
			Component: "eventing",
		},
		expectedResult: `#@data/values
---
name: knative-eventing
namespace: test-eventing-1
version: 'latest'`,
	}, {
		name: "Knative Eventing with version only",
		installFlags: installCmdFlags{
			Version:   "1.0",
			Component: "eventing",
		},
		expectedResult: `#@data/values
---
name: knative-eventing
namespace: knative-eventing
version: '1.0'`,
	}, {
		name: "Knative unknown component",
		installFlags: installCmdFlags{
			Namespace: "1.0",
			Component: "unknown",
		},
		expectedResult: "",
	}, {
		name: "Knative Operator",
		installFlags: installCmdFlags{
			Version: "1.0",
		},
		expectedResult: `#@data/values
---
namespace: default`,
	}, {
		name: "Knative Operator with a namespace",
		installFlags: installCmdFlags{
			Namespace: "test",
		},
		expectedResult: `#@data/values
---
namespace: test`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			tt.installFlags.fill_defaults()
			result := getYamlValuesContent(tt.installFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}

func TestVersionWebhook(t *testing.T) {
	for _, tt := range []struct {
		name           string
		version        string
		expectedResult bool
	}{{
		name:           "Version 1.3",
		version:        "1.3",
		expectedResult: true,
	}, {
		name:           "Version v1.3",
		version:        "v1.3",
		expectedResult: true,
	}, {
		name:           "Version 1.2",
		version:        "1.2",
		expectedResult: false,
	}, {
		name:           "Version v1.2",
		version:        "v1.2",
		expectedResult: false,
	}, {
		name:           "Version 1.4",
		version:        "1.4",
		expectedResult: true,
	}, {
		name:           "Version 2.0",
		version:        "2.0",
		expectedResult: true,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := versionWebhook(tt.version)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}
