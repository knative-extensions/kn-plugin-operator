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

func TestValidateAnnotationsFlags(t *testing.T) {
	for _, tt := range []struct {
		name                string
		annotationsCMDFlags common.KeyValueFlags
		expectedResult      error
	}{{
		name: "Knative Eventing with no deployment aspect",
		annotationsCMDFlags: common.KeyValueFlags{
			Key:        "test-key",
			Value:      "test-value",
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: nil,
	}, {
		name: "Knative Eventing with service name and nodeSelector",
		annotationsCMDFlags: common.KeyValueFlags{
			Key:         "test-key",
			Value:       "test-value",
			Component:   "eventing",
			Namespace:   "test-eventing",
			ServiceName: "eventing-controller",
		},
		expectedResult: nil,
	}, {
		name: "Knative Eventing with no deployment name or service name",
		annotationsCMDFlags: common.KeyValueFlags{
			Key:       "test-key",
			Value:     "test-value",
			Component: "eventing",
			Namespace: "test-eventing",
		},
		expectedResult: fmt.Errorf("You need to specify the name of the deployment or the service."),
	}, {
		name: "Knative Eventing with invalid component name",
		annotationsCMDFlags: common.KeyValueFlags{
			Key:        "test-key",
			Value:      "test-value",
			Component:  "eventing-test",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the component for Knative: serving or eventing."),
	}, {
		name: "Knative Eventing with no namespace",
		annotationsCMDFlags: common.KeyValueFlags{
			Key:        "test-key",
			Value:      "test-value",
			Component:  "eventing-test",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the namespace."),
	}, {
		name: "Knative Eventing with no key",
		annotationsCMDFlags: common.KeyValueFlags{
			Value:      "test-value",
			Component:  "eventing-test",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the key for the deployment."),
	}, {
		name: "Knative Eventing with no value",
		annotationsCMDFlags: common.KeyValueFlags{
			Key:        "test-value",
			Component:  "eventing-test",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the value for the deployment."),
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := validateAnnotationsFlags(tt.annotationsCMDFlags)
			if tt.expectedResult == nil {
				testingUtil.AssertEqual(t, result, nil)
			} else {
				testingUtil.AssertEqual(t, result.Error(), tt.expectedResult.Error())
			}
		})
	}
}

func TestGetOverlayYamlContentAnnotation(t *testing.T) {
	for _, tt := range []struct {
		name                    string
		deploymentLabelCMDFlags common.KeyValueFlags
		expectedResult          string
	}{{
		name: "Knative Serving template for annotation configuration",
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
  deployments:
  #@overlay/match by="name",missing_ok=True
  - name: #@ data.values.deployName
    #@overlay/match missing_ok=True
    annotations:
      #@overlay/match missing_ok=True
      test-key: #@ data.values.value`,
	}, {
		name: "Knative Serving template for annotation configuration for the service",
		deploymentLabelCMDFlags: common.KeyValueFlags{
			Key:         "test-key",
			Value:       "test-value",
			Component:   "serving",
			Namespace:   "test-serving",
			ServiceName: "network",
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
    annotations:
      #@overlay/match missing_ok=True
      test-key: #@ data.values.value`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			rootPath := "testdata/"
			result := getOverlayYamlContentAnnotation(rootPath, tt.deploymentLabelCMDFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}
