// Copyright 2022 The Knative Authors
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

package configure

import (
	"fmt"
	"testing"

	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
)

func TestValidateCMsFlags(t *testing.T) {
	for _, tt := range []struct {
		name           string
		cmsCMDFlags    CMsFlags
		expectedResult error
	}{{
		name: "Knative Eventing",
		cmsCMDFlags: CMsFlags{
			Key:       "test-key",
			Value:     "test-value",
			Component: "eventing",
			Namespace: "test-eventing",
			CMName:    "eventing-controller",
		},
		expectedResult: nil,
	}, {
		name: "Knative Serving",
		cmsCMDFlags: CMsFlags{
			Key:       "test-key",
			Value:     "test-value",
			Component: "serving",
			Namespace: "test-serving",
			CMName:    "network",
		},
		expectedResult: nil,
	}, {
		name: "Knative Serving with no namespace",
		cmsCMDFlags: CMsFlags{
			Key:       "test-key",
			Value:     "test-value",
			Component: "serving",
			CMName:    "network",
		},
		expectedResult: fmt.Errorf("You need to specify the namespace."),
	}, {
		name: "Knative Eventing with no ConfigMap name",
		cmsCMDFlags: CMsFlags{
			Key:       "test-key",
			Value:     "test-value",
			Component: "eventing",
			Namespace: "test-eventing",
		},
		expectedResult: fmt.Errorf("You need to specify the name of the ConfigMap."),
	}, {
		name: "Knative with invalid component name",
		cmsCMDFlags: CMsFlags{
			Key:       "test-key",
			Value:     "test-value",
			Component: "eventing-test",
			Namespace: "test-eventing",
			CMName:    "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the component for Knative: serving or eventing."),
	}, {
		name: "Knative Serving with empty key",
		cmsCMDFlags: CMsFlags{
			Value:     "test-value",
			Component: "eventing",
			Namespace: "test-eventing",
			CMName:    "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the key in the ConfigMap data."),
	}, {
		name: "Knative Serving with empty key",
		cmsCMDFlags: CMsFlags{
			Key:       "test-key",
			Component: "eventing",
			Namespace: "test-eventing",
			CMName:    "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the value in the ConfigMap data."),
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := validateCMsFlags(tt.cmsCMDFlags)
			if tt.expectedResult == nil {
				testingUtil.AssertEqual(t, result, nil)
			} else {
				testingUtil.AssertEqual(t, result.Error(), tt.expectedResult.Error())
			}
		})
	}
}

func TestGetOverlayYamlContentCM(t *testing.T) {
	for _, tt := range []struct {
		name           string
		cmsCMDFlags    CMsFlags
		expectedResult string
	}{{
		name: "Knative Eventing",
		cmsCMDFlags: CMsFlags{
			Key:       "test-key",
			Value:     "test-value",
			Component: "eventing",
			Namespace: "test-eventing",
			CMName:    "eventing-controller",
		},
		expectedResult: `#@ load("@ytt:overlay", "overlay")
#@ load("@ytt:data", "data")

#@overlay/match by=overlay.subset({"kind": "KnativeEventing"}),expects=1
---
apiVersion: operator.knative.dev/v1alpha1
kind: KnativeEventing
metadata:
  #@overlay/match missing_ok=True
  namespace: #@ data.values.namespace
#@overlay/match missing_ok=True
spec:
  #@overlay/match missing_ok=True
  config:

    #@overlay/match missing_ok=True
    eventing-controller:
      #@overlay/match missing_ok=True
      test-key: #@ data.values.value`,
	}, {
		name: "Knative Serving",
		cmsCMDFlags: CMsFlags{
			Key:       "test-key",
			Value:     "test-value",
			Component: "serving",
			Namespace: "test-serving",
			CMName:    "network",
		},
		expectedResult: `#@ load("@ytt:overlay", "overlay")
#@ load("@ytt:data", "data")

#@overlay/match by=overlay.subset({"kind": "KnativeServing"}),expects=1
---
apiVersion: operator.knative.dev/v1alpha1
kind: KnativeServing
metadata:
  #@overlay/match missing_ok=True
  namespace: #@ data.values.namespace
#@overlay/match missing_ok=True
spec:
  #@overlay/match missing_ok=True
  config:

    #@overlay/match missing_ok=True
    network:
      #@overlay/match missing_ok=True
      test-key: #@ data.values.value`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			rootPath := "testdata/"
			result := getOverlayYamlContentCM(rootPath, tt.cmsCMDFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}

func TestGetYamlValuesContentCMs(t *testing.T) {
	for _, tt := range []struct {
		name           string
		cmsCMDFlags    CMsFlags
		expectedResult string
	}{{
		name: "Knative Eventing",
		cmsCMDFlags: CMsFlags{
			Key:       "test-key",
			Value:     "test-value",
			Component: "eventing",
			Namespace: "test-eventing",
			CMName:    "network",
		},
		expectedResult: `#@data/values
---
namespace: test-eventing
value: test-value`,
	}, {
		name: "Knative Serving",
		cmsCMDFlags: CMsFlags{
			Key:       "test-key",
			Value:     "test-value",
			Component: "serving",
			Namespace: "test-serving",
			CMName:    "network",
		},
		expectedResult: `#@data/values
---
namespace: test-serving
value: test-value`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := getYamlValuesContentCMs(tt.cmsCMDFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}
