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

func TestValidateManifestsFlags(t *testing.T) {
	for _, tt := range []struct {
		name              string
		manifestsCMDFlags manifestsFlags
		expectedResult    error
	}{{
		name: "Knative Eventing with no file path",
		manifestsCMDFlags: manifestsFlags{
			Component:         "eventing",
			Namespace:         "test-eventing",
			OperatorNamespace: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the local path of the file containing the custom manifests."),
	}, {
		name: "Knative Eventing",
		manifestsCMDFlags: manifestsFlags{
			File:              "test-file.yaml",
			Component:         "eventing",
			Namespace:         "test-eventing",
			OperatorNamespace: "eventing-controller",
		},
		expectedResult: nil,
	}, {
		name: "Knative Eventing with no namespace",
		manifestsCMDFlags: manifestsFlags{
			File:              "file.yaml",
			Component:         "eventing",
			OperatorNamespace: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the namespace for the Knative component."),
	}, {
		name: "Knative Eventing with invalid component",
		manifestsCMDFlags: manifestsFlags{
			File:              "file.yaml",
			Namespace:         "test",
			Component:         "test",
			OperatorNamespace: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the component for Knative: serving or eventing."),
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := validateManifestsFlags(tt.manifestsCMDFlags)
			if tt.expectedResult == nil {
				testingUtil.AssertEqual(t, result, nil)
			} else {
				testingUtil.AssertEqual(t, result.Error(), tt.expectedResult.Error())
			}
		})
	}
}

func TestGetYamlValuesContentManifests(t *testing.T) {
	for _, tt := range []struct {
		name              string
		manifestsCMDFlags manifestsFlags
		expectedResult    string
	}{{
		name: "Knative Eventing",
		manifestsCMDFlags: manifestsFlags{
			Component: "eventing",
			Namespace: "test-eventing",
		},
		expectedResult: `#@data/values
---
namespace: test-eventing
manifestsPath: /knative-custom-manifest`,
	}, {
		name: "Knative Eventing with accessible file",
		manifestsCMDFlags: manifestsFlags{
			File:       "public-file-link",
			Component:  "eventing",
			Namespace:  "test-eventing",
			Accessible: true,
		},
		expectedResult: `#@data/values
---
namespace: test-eventing
manifestsPath: public-file-link`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := getYamlValuesContentManifests(tt.manifestsCMDFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}

func TestGetOverlayYamlContentManifests(t *testing.T) {
	for _, tt := range []struct {
		name              string
		manifestsCMDFlags manifestsFlags
		expectedResult    string
	}{{
		name: "Knative Eventing",
		manifestsCMDFlags: manifestsFlags{
			File:      "local-test-file",
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
  additionalManifests:
  #@overlay/match by="URL",missing_ok=True
  - URL: #@ data.values.manifestsPath`,
	}, {
		name: "Knative Serving with overwrite mode",
		manifestsCMDFlags: manifestsFlags{
			File:      "local-test-file",
			Component: "serving",
			Namespace: "test-eventing",
			Overwrite: true,
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
  #@overlay/replace or_add=True
  additionalManifests:
  #@overlay/match by="URL",missing_ok=True
  - URL: #@ data.values.manifestsPath`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			rootPath := "testdata/"
			result := getOverlayYamlContentManifest(rootPath, tt.manifestsCMDFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}
