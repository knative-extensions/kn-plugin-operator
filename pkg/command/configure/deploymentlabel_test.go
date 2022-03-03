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

func TestValidateLabelsFlags(t *testing.T) {
	for _, tt := range []struct {
		name                    string
		deploymentLabelCMDFlags deploymentLabelFlags
		expectedResult          error
	}{{
		name: "Knative Eventing with no deployment aspect",
		deploymentLabelCMDFlags: deploymentLabelFlags{
			Key:        "test-key",
			Value:      "test-value",
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to enable at least one deployment aspect for Knative: NodeSelector, Annotation or Label."),
	}, {
		name: "Knative Eventing with multiple aspects",
		deploymentLabelCMDFlags: deploymentLabelFlags{
			Label:      true,
			Annotation: true,
			Key:        "test-key",
			Value:      "test-value",
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You can specify only one deployment aspect for Knative: NodeSelector, Annotation or Label."),
	}, {
		name: "Knative Eventing",
		deploymentLabelCMDFlags: deploymentLabelFlags{
			Label:      true,
			Key:        "test-key",
			Value:      "test-value",
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: nil,
	}, {
		name: "Knative Eventing with no deployment name",
		deploymentLabelCMDFlags: deploymentLabelFlags{
			Label:     true,
			Key:       "test-key",
			Value:     "test-value",
			Component: "eventing",
			Namespace: "test-eventing",
		},
		expectedResult: fmt.Errorf("You need to specify the name of the deployment."),
	}, {
		name: "Knative Eventing with invalid component name",
		deploymentLabelCMDFlags: deploymentLabelFlags{
			Label:      true,
			Key:        "test-key",
			Value:      "test-value",
			Component:  "eventing-test",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the component for Knative: serving or eventing."),
	}, {
		name: "Knative Eventing with no namespace",
		deploymentLabelCMDFlags: deploymentLabelFlags{
			Label:      true,
			Key:        "test-key",
			Value:      "test-value",
			Component:  "eventing-test",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the namespace."),
	}, {
		name: "Knative Eventing with no key",
		deploymentLabelCMDFlags: deploymentLabelFlags{
			Label:      true,
			Value:      "test-value",
			Component:  "eventing-test",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the key for the deployment."),
	}, {
		name: "Knative Eventing with no value",
		deploymentLabelCMDFlags: deploymentLabelFlags{
			Label:      true,
			Key:        "test-value",
			Component:  "eventing-test",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the value for the deployment."),
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := validateLabelsFlags(tt.deploymentLabelCMDFlags)
			if tt.expectedResult == nil {
				testingUtil.AssertEqual(t, result, nil)
			} else {
				testingUtil.AssertEqual(t, result.Error(), tt.expectedResult.Error())
			}
		})
	}
}

func TestGetOverlayYamlContentLabel(t *testing.T) {
	for _, tt := range []struct {
		name                    string
		deploymentLabelCMDFlags deploymentLabelFlags
		expectedResult          string
	}{{
		name: "Knative Eventing",
		deploymentLabelCMDFlags: deploymentLabelFlags{
			Label:      true,
			Key:        "test-key",
			Value:      "test-value",
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
  #@overlay/match by="name"
  - name: #@ data.values.deployName

    #@overlay/match missing_ok=True
    labels:
      #@overlay/match missing_ok=True
      test-key: #@ data.values.value`,
	}, {
		name: "Knative Serving",
		deploymentLabelCMDFlags: deploymentLabelFlags{
			Annotation: true,
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
apiVersion: operator.knative.dev/v1alpha1
kind: KnativeServing
metadata:
  #@overlay/match missing_ok=True
  namespace: #@ data.values.namespace
#@overlay/match missing_ok=True
spec:
  #@overlay/match missing_ok=True
  deployments:
  #@overlay/match by="name"
  - name: #@ data.values.deployName

    #@overlay/match missing_ok=True
    annotations:
      #@overlay/match missing_ok=True
      test-key: #@ data.values.value`,
	}, {
		name: "Knative Serving",
		deploymentLabelCMDFlags: deploymentLabelFlags{
			NodeSelector: true,
			Key:          "test-key",
			Value:        "test-value",
			Component:    "serving",
			Namespace:    "test-serving",
			DeployName:   "network",
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
  #@overlay/match by="name"
  - name: #@ data.values.deployName

    #@overlay/match missing_ok=True
    nodeSelector:
      #@overlay/match missing_ok=True
      test-key: #@ data.values.value`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			rootPath := "testdata/"
			result := getOverlayYamlContentLabel(rootPath, tt.deploymentLabelCMDFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}

func TestGetYamlValuesContentLabels(t *testing.T) {
	for _, tt := range []struct {
		name                    string
		deploymentLabelCMDFlags deploymentLabelFlags
		expectedResult          string
	}{{
		name: "Knative Eventing",
		deploymentLabelCMDFlags: deploymentLabelFlags{
			Key:        "test-key",
			Value:      "test-value",
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "network",
		},
		expectedResult: `#@data/values
---
namespace: test-eventing
deployName: network
value: test-value`,
	}, {
		name: "Knative Serving",
		deploymentLabelCMDFlags: deploymentLabelFlags{
			Key:        "test-key",
			Value:      "test-value",
			Component:  "serving",
			Namespace:  "test-serving",
			DeployName: "network",
		},
		expectedResult: `#@data/values
---
namespace: test-serving
deployName: network
value: test-value`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := getYamlValuesContentLabels(tt.deploymentLabelCMDFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}
