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

func TestValidateImagesFlags(t *testing.T) {
	for _, tt := range []struct {
		name           string
		imageCMDFlags  ImageFlags
		expectedResult error
	}{{
		name: "Knative Eventing",
		imageCMDFlags: ImageFlags{
			ImageKey:   "test-key",
			ImageUrl:   "test-value",
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: nil,
	}, {
		name: "Knative Serving with multiple aspects",
		imageCMDFlags: ImageFlags{
			ImageKey:   "test-key",
			ImageUrl:   "test-value",
			Component:  "serving",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: nil,
	}, {
		name: "Knative Serving with no deploy name",
		imageCMDFlags: ImageFlags{
			ImageKey:  "test-key",
			ImageUrl:  "test-value",
			Component: "serving",
			Namespace: "test-eventing",
		},
		expectedResult: nil,
	}, {
		name: "Knative Eventing with no image key",
		imageCMDFlags: ImageFlags{
			ImageUrl:   "test-value",
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the image key."),
	}, {
		name: "Knative Eventing with no image value name",
		imageCMDFlags: ImageFlags{
			ImageKey:   "test-key",
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the image URL."),
	}, {
		name: "Knative Eventing with invalid component name",
		imageCMDFlags: ImageFlags{
			ImageKey:   "test-key",
			ImageUrl:   "test-value",
			Component:  "eventing-test",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the component for Knative: serving or eventing."),
	}, {
		name: "Knative Eventing with no namespace",
		imageCMDFlags: ImageFlags{
			ImageKey:   "test-key",
			ImageUrl:   "test-value",
			Component:  "eventing-test",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the namespace."),
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := validateImagesFlags(tt.imageCMDFlags)
			if tt.expectedResult == nil {
				testingUtil.AssertEqual(t, result, nil)
			} else {
				testingUtil.AssertEqual(t, result.Error(), tt.expectedResult.Error())
			}
		})
	}
}

func TestGetOverlayYamlContentImage(t *testing.T) {
	for _, tt := range []struct {
		name           string
		imageCMDFlags  ImageFlags
		expectedResult string
	}{{
		name: "Knative Eventing",
		imageCMDFlags: ImageFlags{
			ImageKey:  "test-key",
			ImageUrl:  "test-value",
			Component: "eventing",
			Namespace: "test-eventing",
		},
		expectedResult: `#@ load("@ytt:overlay", "overlay")
#@ load("@ytt:data", "data")

#@overlay/match by=overlay.subset({"kind": "KnativeEventing"}),expects=1
---
apiVersion: operator.knative.dev/v1beta1
kind: KnativeEventing
metadata:
  #@overlay/match missing_ok=True
  namespace: #@ data.values.namespace
#@overlay/match missing_ok=True
spec:

  #@overlay/match missing_ok=True
  registry:
    #@overlay/match missing_ok=True
    override:
      #@overlay/match missing_ok=True
      test-key: #@ data.values.imageValue`,
	}, {
		name: "Knative Serving with the image key default",
		imageCMDFlags: ImageFlags{
			ImageKey:   "default",
			ImageUrl:   "test-value",
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
  registry:
    #@overlay/match missing_ok=True
    default: #@ data.values.imageValue`,
	}, {
		name: "Knative Serving",
		imageCMDFlags: ImageFlags{
			ImageKey:   "test-key",
			ImageUrl:   "test-value",
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
  registry:
    #@overlay/match missing_ok=True
    override:
      #@overlay/match missing_ok=True
      network/test-key: #@ data.values.imageValue`,
	}, {
		name: "Knative Serving for queue-sidecar-image",
		imageCMDFlags: ImageFlags{
			ImageKey:  "queue-sidecar-image",
			ImageUrl:  "test-value",
			Component: "serving",
			Namespace: "test-serving",
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
  config:
    #@overlay/match missing_ok=True
    deployment:
      #@overlay/match missing_ok=True
      queue-sidecar-image: #@ data.values.imageValue`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			rootPath := "testdata/"
			result := getOverlayYamlContentImage(rootPath, tt.imageCMDFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}

func TestGetYamlValuesContentImages(t *testing.T) {
	for _, tt := range []struct {
		name           string
		imageCMDFlags  ImageFlags
		expectedResult string
	}{{
		name: "Knative Eventing",
		imageCMDFlags: ImageFlags{
			ImageKey:   "test-key",
			ImageUrl:   "test-value",
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "network",
		},
		expectedResult: `#@data/values
---
namespace: test-eventing
deployName: network
imageValue: test-value`,
	}, {
		name: "Knative Serving",
		imageCMDFlags: ImageFlags{
			ImageKey:   "test-key",
			ImageUrl:   "test-value",
			Component:  "serving",
			Namespace:  "test-serving",
			DeployName: "network",
		},
		expectedResult: `#@data/values
---
namespace: test-serving
deployName: network
imageValue: test-value`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := getYamlValuesContentImages(tt.imageCMDFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}
