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

func TestGetOverlayYamlContentHA(t *testing.T) {
	for _, tt := range []struct {
		name           string
		haCMDFlags     haFlags
		expectedResult string
	}{{
		name: "Knative Eventing",
		haCMDFlags: haFlags{
			Replicas:   "4",
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
  - name: #@ data.values.name
    #@overlay/match missing_ok=True
    replicas: #@ data.values.replicas`,
	}, {
		name: "Knative Serving",
		haCMDFlags: haFlags{
			Replicas:   "2",
			Component:  "serving",
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
  #@overlay/match by="name"
  - name: #@ data.values.name
    #@overlay/match missing_ok=True
    replicas: #@ data.values.replicas`,
	}, {
		name: "Knative Eventing with no deployment name",
		haCMDFlags: haFlags{
			Replicas:  "4",
			Component: "eventing",
			Namespace: "test-eventing",
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
  high-availability:
    #@overlay/match missing_ok=True
    replicas: #@ data.values.replicas`,
	}, {
		name: "Knative Serving with no deployment name",
		haCMDFlags: haFlags{
			Replicas:  "2",
			Component: "serving",
			Namespace: "test-serving",
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
  high-availability:
    #@overlay/match missing_ok=True
    replicas: #@ data.values.replicas`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			rootPath := "testdata/"
			result := getOverlayYamlContentHA(rootPath, tt.haCMDFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}

func TestGetYamlValuesContentHAs(t *testing.T) {
	for _, tt := range []struct {
		name           string
		haCMDFlags     haFlags
		expectedResult string
	}{{
		name: "Knative Eventing",
		haCMDFlags: haFlags{
			Replicas:   "2",
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: `#@data/values
---
namespace: test-eventing
name: eventing-controller
replicas: 2`,
	}, {
		name: "Knative Serving",
		haCMDFlags: haFlags{
			Replicas:   "2",
			Component:  "serving",
			Namespace:  "test-serving",
			DeployName: "activator",
		},
		expectedResult: `#@data/values
---
namespace: test-serving
name: activator
replicas: 2`,
	}, {
		name: "Knative Eventing with no deployment name",
		haCMDFlags: haFlags{
			Replicas:  "2",
			Component: "eventing",
			Namespace: "test-eventing",
		},
		expectedResult: `#@data/values
---
namespace: test-eventing
replicas: 2`,
	}, {
		name: "Knative Serving with no deployment name",
		haCMDFlags: haFlags{
			Replicas:  "2",
			Component: "serving",
			Namespace: "test-serving",
		},
		expectedResult: `#@data/values
---
namespace: test-serving
replicas: 2`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := getYamlValuesContentHAs(tt.haCMDFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}

func TestValidateHAsFlags(t *testing.T) {
	for _, tt := range []struct {
		name           string
		haCMDFlags     haFlags
		expectedResult error
	}{{
		name: "Knative Eventing",
		haCMDFlags: haFlags{
			Replicas:   "4",
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: nil,
	}, {
		name: "Knative Serving",
		haCMDFlags: haFlags{
			Replicas:   "3",
			Component:  "serving",
			Namespace:  "test-serving",
			DeployName: "activator",
		},
		expectedResult: nil,
	}, {
		name: "Knative Serving with the invalid component",
		haCMDFlags: haFlags{
			Replicas:   "3",
			Component:  "serving-invalid",
			Namespace:  "test-serving",
			DeployName: "activator",
		},
		expectedResult: fmt.Errorf("You need to specify the component for Knative: serving or eventing."),
	}, {
		name: "Knative Serving with no deployment name",
		haCMDFlags: haFlags{
			Replicas:  "3",
			Component: "serving",
			Namespace: "test-serving",
		},
		expectedResult: nil,
	}, {
		name: "Knative Serving with no namespace",
		haCMDFlags: haFlags{
			Replicas:   "3",
			Component:  "serving",
			DeployName: "activator",
		},
		expectedResult: fmt.Errorf("You need to specify the namespace."),
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := validateHAsFlags(tt.haCMDFlags)
			if tt.expectedResult == nil {
				testingUtil.AssertEqual(t, result, nil)
			} else {
				testingUtil.AssertEqual(t, result.Error(), tt.expectedResult.Error())
			}
		})
	}
}
