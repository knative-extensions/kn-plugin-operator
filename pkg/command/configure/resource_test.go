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

func TestGetOverlayYamlContentSource(t *testing.T) {
	for _, tt := range []struct {
		name              string
		resourcesCMDFlags ResourcesFlags
		expectedResult    string
	}{{
		name: "Knative Eventing",
		resourcesCMDFlags: ResourcesFlags{
			LimitCPU:      "2.5G",
			LimitMemory:   "1001M",
			RequestCPU:    "2.2G",
			RequestMemory: "999M",
			Component:     "eventing",
			Namespace:     "test-eventing",
			Container:     "eventing-controller",
			DeployName:    "eventing-controller",
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
    resources:

      #@overlay/match by="container"
    - container: #@ data.values.container
      #@overlay/match missing_ok=True
      requests:
        #@overlay/match missing_ok=True
        cpu: #@ data.values.requestCPU
        #@overlay/match missing_ok=True
        memory: #@ data.values.requestMemory
      #@overlay/match missing_ok=True
      limits:
        #@overlay/match missing_ok=True
        cpu: #@ data.values.limitCPU
        #@overlay/match missing_ok=True
        memory: #@ data.values.limitMemory`,
	}, {
		name: "Knative Serving",
		resourcesCMDFlags: ResourcesFlags{
			LimitCPU:      "2.5G",
			LimitMemory:   "1001M",
			RequestCPU:    "2.2G",
			RequestMemory: "999M",
			Component:     "serving",
			Namespace:     "test-serving",
			Container:     "activator",
			DeployName:    "activator",
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
    resources:

      #@overlay/match by="container"
    - container: #@ data.values.container
      #@overlay/match missing_ok=True
      requests:
        #@overlay/match missing_ok=True
        cpu: #@ data.values.requestCPU
        #@overlay/match missing_ok=True
        memory: #@ data.values.requestMemory
      #@overlay/match missing_ok=True
      limits:
        #@overlay/match missing_ok=True
        cpu: #@ data.values.limitCPU
        #@overlay/match missing_ok=True
        memory: #@ data.values.limitMemory`,
	}, {
		name: "Knative Serving",
		resourcesCMDFlags: ResourcesFlags{
			LimitMemory: "1001M",
			RequestCPU:  "2.2G",
			Component:   "serving",
			Namespace:   "test-serving",
			Container:   "activator",
			DeployName:  "activator",
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
    resources:

      #@overlay/match by="container"
    - container: #@ data.values.container
      #@overlay/match missing_ok=True
      requests:
        #@overlay/match missing_ok=True
        cpu: #@ data.values.requestCPU
      #@overlay/match missing_ok=True
      limits:
        #@overlay/match missing_ok=True
        memory: #@ data.values.limitMemory`,
	}, {
		name: "Knative Serving",
		resourcesCMDFlags: ResourcesFlags{
			RequestCPU: "2.2G",
			Component:  "serving",
			Namespace:  "test-serving",
			Container:  "activator",
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
  - name: #@ data.values.deployName
    #@overlay/match missing_ok=True
    resources:

      #@overlay/match by="container"
    - container: #@ data.values.container
      #@overlay/match missing_ok=True
      requests:
        #@overlay/match missing_ok=True
        cpu: #@ data.values.requestCPU`,
	}, {
		name: "Knative Serving",
		resourcesCMDFlags: ResourcesFlags{
			LimitCPU:   "2.2G",
			Component:  "serving",
			Namespace:  "test-serving",
			Container:  "activator",
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
  - name: #@ data.values.deployName
    #@overlay/match missing_ok=True
    resources:

      #@overlay/match by="container"
    - container: #@ data.values.container
      #@overlay/match missing_ok=True
      limits:
        #@overlay/match missing_ok=True
        cpu: #@ data.values.limitCPU`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			rootPath := "testdata/"
			result := getOverlayYamlContentResource(rootPath, tt.resourcesCMDFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}

func TestGetYamlValuesContentResources(t *testing.T) {
	for _, tt := range []struct {
		name              string
		resourcesCMDFlags ResourcesFlags
		expectedResult    string
	}{{
		name: "Knative Eventing",
		resourcesCMDFlags: ResourcesFlags{
			LimitCPU:      "2.5G",
			LimitMemory:   "1001M",
			RequestCPU:    "2.2G",
			RequestMemory: "999M",
			Component:     "eventing",
			Namespace:     "test-eventing",
			Container:     "eventing-controller",
			DeployName:    "eventing-controller",
		},
		expectedResult: `#@data/values
---
namespace: test-eventing
container: eventing-controller
deployName: eventing-controller
requestCPU: 2.2G
requestMemory: 999M
limitCPU: 2.5G
limitMemory: 1001M`,
	}, {
		name: "Knative Serving",
		resourcesCMDFlags: ResourcesFlags{
			LimitCPU:      "2.5G",
			LimitMemory:   "1001M",
			RequestCPU:    "2.2G",
			RequestMemory: "999M",
			Component:     "serving",
			Namespace:     "test-serving",
			Container:     "activator",
			DeployName:    "activator",
		},
		expectedResult: `#@data/values
---
namespace: test-serving
container: activator
deployName: activator
requestCPU: 2.2G
requestMemory: 999M
limitCPU: 2.5G
limitMemory: 1001M`,
	}, {
		name: "Knative Serving",
		resourcesCMDFlags: ResourcesFlags{
			LimitMemory: "1001M",
			RequestCPU:  "2.2G",
			Component:   "serving",
			Namespace:   "test-serving",
			Container:   "activator",
			DeployName:  "activator1",
		},
		expectedResult: `#@data/values
---
namespace: test-serving
container: activator
deployName: activator1
requestCPU: 2.2G
limitMemory: 1001M`,
	}, {
		name: "Knative Serving",
		resourcesCMDFlags: ResourcesFlags{
			RequestCPU: "2.2G",
			Component:  "serving",
			Namespace:  "test-serving",
			Container:  "activator",
			DeployName: "activator",
		},
		expectedResult: `#@data/values
---
namespace: test-serving
container: activator
deployName: activator
requestCPU: 2.2G`,
	}, {
		name: "Knative Serving",
		resourcesCMDFlags: ResourcesFlags{
			LimitCPU:   "22G",
			Component:  "serving",
			Namespace:  "test-serving",
			Container:  "activator",
			DeployName: "activator",
		},
		expectedResult: `#@data/values
---
namespace: test-serving
container: activator
deployName: activator
limitCPU: 22G`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := getYamlValuesContentResources(tt.resourcesCMDFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}

func TestValidateResourcesFlags(t *testing.T) {
	for _, tt := range []struct {
		name              string
		resourcesCMDFlags ResourcesFlags
		expectedResult    error
	}{{
		name: "Knative Eventing",
		resourcesCMDFlags: ResourcesFlags{
			LimitCPU:      "2.5G",
			LimitMemory:   "1001M",
			RequestCPU:    "2.2G",
			RequestMemory: "999M",
			Component:     "eventing",
			Namespace:     "test-eventing",
			Container:     "eventing-controller",
			DeployName:    "eventing-controller",
		},
		expectedResult: nil,
	}, {
		name: "Knative Serving",
		resourcesCMDFlags: ResourcesFlags{
			Component:  "serving",
			Namespace:  "test-serving",
			Container:  "activator",
			DeployName: "activator",
		},
		expectedResult: fmt.Errorf("You need to specify at least one resource parameter: limitCPU, limitMemory, requestCPU or requestMemory."),
	}, {
		name: "Knative Serving with no component name",
		resourcesCMDFlags: ResourcesFlags{
			LimitMemory: "1001M",
			RequestCPU:  "2.2G",
			Namespace:   "test-serving",
			Container:   "activator",
			DeployName:  "activator1",
		},
		expectedResult: fmt.Errorf("You need to specify the component name."),
	}, {
		name: "Knative Serving with no container name",
		resourcesCMDFlags: ResourcesFlags{
			RequestCPU: "2.2G",
			Component:  "serving",
			Namespace:  "test-serving",
			DeployName: "activator",
		},
		expectedResult: fmt.Errorf("You need to specify the container name."),
	}, {
		name: "Knative Serving with no deployment name",
		resourcesCMDFlags: ResourcesFlags{
			LimitCPU:  "22G",
			Component: "serving",
			Namespace: "test-serving",
			Container: "activator",
		},
		expectedResult: fmt.Errorf("You need to specify the name of the deployment."),
	}, {
		name: "Knative Serving with no namespace",
		resourcesCMDFlags: ResourcesFlags{
			LimitCPU:   "22G",
			Component:  "serving",
			DeployName: "test-serving",
			Container:  "activator",
		},
		expectedResult: fmt.Errorf("You need to specify the namespace."),
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := validateResourcesFlags(tt.resourcesCMDFlags)
			if tt.expectedResult == nil {
				testingUtil.AssertEqual(t, result, nil)
			} else {
				testingUtil.AssertEqual(t, result.Error(), tt.expectedResult.Error())
			}
		})
	}
}
