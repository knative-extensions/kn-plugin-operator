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

	"knative.dev/kn-plugin-operator/pkg/command/common"
	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
)

func TestValidateSelectorFlags(t *testing.T) {
	for _, tt := range []struct {
		name                 string
		nodeSelectorCMDFlags common.KeyValueFlags
		expectedResult       error
	}{{
		name: "Knative Eventing with no deployment aspect",
		nodeSelectorCMDFlags: common.KeyValueFlags{
			Key:         "test-key",
			Value:       "test-value",
			Component:   "eventing",
			Namespace:   "test-eventing",
			ServiceName: "eventing-controller",
		},
		expectedResult: nil,
	}, {
		name: "Knative Eventing with no deployment name or service name",
		nodeSelectorCMDFlags: common.KeyValueFlags{
			Key:       "test-key",
			Value:     "test-value",
			Component: "eventing",
			Namespace: "test-eventing",
		},
		expectedResult: fmt.Errorf("You need to specify the name of the service."),
	}, {
		name: "Knative Eventing with invalid component name",
		nodeSelectorCMDFlags: common.KeyValueFlags{
			Key:        "test-key",
			Value:      "test-value",
			Component:  "eventing-test",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the component for Knative: serving or eventing."),
	}, {
		name: "Knative Eventing with no namespace",
		nodeSelectorCMDFlags: common.KeyValueFlags{
			Key:        "test-key",
			Value:      "test-value",
			Component:  "eventing-test",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the namespace."),
	}, {
		name: "Knative Eventing with no key",
		nodeSelectorCMDFlags: common.KeyValueFlags{
			Value:      "test-value",
			Component:  "eventing-test",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the key."),
	}, {
		name: "Knative Eventing with no value",
		nodeSelectorCMDFlags: common.KeyValueFlags{
			Key:        "test-value",
			Component:  "eventing-test",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the value."),
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := validateSelectorFlags(tt.nodeSelectorCMDFlags)
			if tt.expectedResult == nil {
				testingUtil.AssertEqual(t, result, nil)
			} else {
				testingUtil.AssertEqual(t, result.Error(), tt.expectedResult.Error())
			}
		})
	}
}

func TestGetOverlayYamlContentSelector(t *testing.T) {
	for _, tt := range []struct {
		name                    string
		deploymentLabelCMDFlags common.KeyValueFlags
		expectedResult          string
	}{{
		name: "Knative Serving template for NodeSelector configuration",
		deploymentLabelCMDFlags: common.KeyValueFlags{
			Key:        "test-key",
			Value:      "test-value",
			Component:  "serving",
			Namespace:  "test-serving",
			DeployName: "network",
		},
		expectedResult: `#@ load("@ytt:overlay", "overlay")
#@ load("@ytt:data", "data")

#@overlay/match by=overlay.subset({"kind": "KnativeServing"}),expects=1
---
apiVersion: operator.knative.dev/v1beta1
kind: KnativeServing
metadata:
  #@overlay/match missing_ok=True
  namespace: #@ data.values.namespace
#@overlay/match missing_ok=True
spec:

  #@overlay/match missing_ok=True
  services:
  #@overlay/match by="name",missing_ok=True
  - name: #@ data.values.serviceName
    #@overlay/match missing_ok=True
    selector:
      #@overlay/match missing_ok=True
      test-key: #@ data.values.value`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := getOverlayYamlContentSelector(tt.deploymentLabelCMDFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}
