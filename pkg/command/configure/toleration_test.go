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

func TestGetOverlayYamlContent(t *testing.T) {
	for _, tt := range []struct {
		name                string
		tolerationsCMDFlags TolerationsFlags
		expectedResult      string
	}{{
		name: "Knative Eventing",
		tolerationsCMDFlags: TolerationsFlags{
			Key:        "test",
			Operator:   "Exists",
			Effect:     "test-effect",
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
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
  deployments:
  #@overlay/match by="name",missing_ok=True
  - name: #@ data.values.deployName
    #@overlay/match missing_ok=True
    tolerations:

    #@overlay/match by="key",missing_ok=True
    - key: #@ data.values.key
      #@overlay/match missing_ok=True
      operator: #@ data.values.operator
      #@overlay/match missing_ok=True
      effect: #@ data.values.effect`,
	}, {
		name: "Knative Serving",
		tolerationsCMDFlags: TolerationsFlags{
			Key:        "test",
			Operator:   "Exists",
			Effect:     "test-effect",
			Namespace:  "test-serving",
			DeployName: "activator",
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
  deployments:
  #@overlay/match by="name",missing_ok=True
  - name: #@ data.values.deployName
    #@overlay/match missing_ok=True
    tolerations:

    #@overlay/match by="key",missing_ok=True
    - key: #@ data.values.key
      #@overlay/match missing_ok=True
      operator: #@ data.values.operator
      #@overlay/match missing_ok=True
      effect: #@ data.values.effect`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			rootPath := "testdata/"
			result := getOverlayYamlContent(rootPath, tt.tolerationsCMDFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}

func TestGetYamlValuesContentTolerations(t *testing.T) {
	for _, tt := range []struct {
		name                string
		tolerationsCMDFlags TolerationsFlags
		expectedResult      string
	}{{
		name: "Knative Eventing",
		tolerationsCMDFlags: TolerationsFlags{
			Key:        "test",
			Operator:   "Exists",
			Effect:     "test-effect",
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: `#@data/values
---
namespace: test-eventing
deployName: eventing-controller
key: test
operator: Exists
effect: test-effect`,
	}, {
		name: "Knative Serving",
		tolerationsCMDFlags: TolerationsFlags{
			Key:        "test",
			Operator:   "Exists",
			Effect:     "test-effect",
			Component:  "serving",
			Namespace:  "test-serving",
			DeployName: "activator",
		},
		expectedResult: `#@data/values
---
namespace: test-serving
deployName: activator
key: test
operator: Exists
effect: test-effect`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := getYamlValuesContentTolerations(tt.tolerationsCMDFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}

func TestValidateTolerationsFlags(t *testing.T) {
	for _, tt := range []struct {
		name                string
		tolerationsCMDFlags TolerationsFlags
		expectedResult      error
	}{{
		name: "Knative Eventing",
		tolerationsCMDFlags: TolerationsFlags{
			Key:        "test-key",
			Operator:   "Equal",
			Effect:     "test-effect",
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the effect to one of the following values: NoSchedule, PreferNoSchedule or NoExecute."),
	}, {
		name: "Knative Serving",
		tolerationsCMDFlags: TolerationsFlags{
			Key:        "test-key",
			Operator:   "Equal",
			Value:      "test-value",
			Effect:     "NoSchedule",
			Component:  "serving",
			Namespace:  "test-serving",
			DeployName: "activator",
		},
		expectedResult: nil,
	}, {
		name: "Knative Serving with invalid operator",
		tolerationsCMDFlags: TolerationsFlags{
			Key:        "test-key",
			Operator:   "Equals",
			Effect:     "test-effect",
			Component:  "serving",
			Namespace:  "test-serving",
			DeployName: "activator",
		},
		expectedResult: fmt.Errorf("You need to specify the operator to one of the following values: Exists or Equal."),
	}, {
		name: "Knative Serving with no deployment name",
		tolerationsCMDFlags: TolerationsFlags{
			Key:        "test-key",
			Operator:   "Equal",
			Effect:     "NoSchedule",
			Value:      "test-value",
			Component:  "serving",
			Namespace:  "test-serving",
			DeployName: "",
		},
		expectedResult: fmt.Errorf("You need to specify the name of the deployment."),
	}, {
		name: "Knative Serving with no value",
		tolerationsCMDFlags: TolerationsFlags{
			Key:        "test-key",
			Operator:   "Equal",
			Effect:     "NoSchedule",
			Component:  "serving",
			Namespace:  "test-serving",
			DeployName: "test",
		},
		expectedResult: fmt.Errorf("You need to specify the value, if the Operator is Equal."),
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := validateTolerationsFlags(tt.tolerationsCMDFlags)
			if tt.expectedResult == nil {
				testingUtil.AssertEqual(t, result, nil)
			} else {
				testingUtil.AssertEqual(t, result.Error(), tt.expectedResult.Error())
			}
		})
	}
}
