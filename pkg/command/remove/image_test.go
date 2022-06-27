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

package remove

import (
	"fmt"
	"testing"

	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
	"knative.dev/operator/pkg/apis/operator/base"
)

func TestValidateImagesFlags(t *testing.T) {
	for _, tt := range []struct {
		name           string
		imageCMDFlags  ImageFlags
		expectedResult error
	}{{
		name: "Image flags with only component and namespace",
		imageCMDFlags: ImageFlags{
			Component: "eventing",
			Namespace: "test-eventing",
		},
		expectedResult: fmt.Errorf("You need to specify the image name or the name of the deployment for the image configuration."),
	}, {
		name: "Image flags with correct component, namespace and the deploy name",
		imageCMDFlags: ImageFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "test-cm",
		},
		expectedResult: nil,
	}, {
		name: "Image flags with correct component, namespace, deploy name and image key",
		imageCMDFlags: ImageFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "test-cm",
			ImageKey:   "test-key",
		},
		expectedResult: nil,
	}, {
		name: "Image flags with correct component, namespace and image key",
		imageCMDFlags: ImageFlags{
			Component: "eventing",
			Namespace: "test-eventing",
			ImageKey:  "test-container",
		},
		expectedResult: nil,
	}, {
		name: "Image flags without namespace",
		imageCMDFlags: ImageFlags{
			Component:  "eventing",
			DeployName: "test-deploy",
			ImageKey:   "test-container",
		},
		expectedResult: fmt.Errorf("You need to specify the namespace."),
	}, {
		name: "Image flags without component namr",
		imageCMDFlags: ImageFlags{
			Namespace:  "eventing",
			DeployName: "test-deploy",
			ImageKey:   "test-container",
		},
		expectedResult: fmt.Errorf("You need to specify the component for Knative: serving or eventing."),
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

func testRegistry() base.Registry {
	return base.Registry{
		Default: "default-image_configuration",
		Override: map[string]string{
			"container1":              "new-registry.io/test/path/new-container-1:new-tag",
			"container2":              "new-registry.io/test/path/new-container-2:new-tag",
			"test-deploy/container1":  "new-registry.io/test/path/new-container-1:new-tag",
			"test-deploy/container2":  "new-registry.io/test/path/new-container-2:new-tag",
			"test-deploy1/container1": "new-registry.io/test/path/new-container-1:new-tag",
			"test-deploy1/container2": "new-registry.io/test/path/new-container-2:new-tag",
		},
	}
}

func TestRemoveImagesFields(t *testing.T) {
	for _, tt := range []struct {
		name           string
		imageCMDFlags  ImageFlags
		input          base.Registry
		expectedResult base.Registry
	}{{
		name: "Image flags with correct component and namespace",
		imageCMDFlags: ImageFlags{
			Component: "eventing",
			Namespace: "test-eventing",
			ImageKey:  "container1",
		},
		input: testRegistry(),
		expectedResult: base.Registry{
			Default: "default-image_configuration",
			Override: map[string]string{
				"container2":              "new-registry.io/test/path/new-container-2:new-tag",
				"test-deploy/container2":  "new-registry.io/test/path/new-container-2:new-tag",
				"test-deploy1/container2": "new-registry.io/test/path/new-container-2:new-tag",
			},
		},
	}, {
		name: "Image flags with correct component and namespace",
		imageCMDFlags: ImageFlags{
			Component: "eventing",
			Namespace: "test-eventing",
			ImageKey:  "default",
		},
		input: testRegistry(),
		expectedResult: base.Registry{
			Default: "",
			Override: map[string]string{
				"container1":              "new-registry.io/test/path/new-container-1:new-tag",
				"container2":              "new-registry.io/test/path/new-container-2:new-tag",
				"test-deploy/container1":  "new-registry.io/test/path/new-container-1:new-tag",
				"test-deploy/container2":  "new-registry.io/test/path/new-container-2:new-tag",
				"test-deploy1/container1": "new-registry.io/test/path/new-container-1:new-tag",
				"test-deploy1/container2": "new-registry.io/test/path/new-container-2:new-tag",
			},
		},
	}, {
		name: "Image flags with correct deploy, component and namespace",
		imageCMDFlags: ImageFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "test-deploy",
		},
		input: testRegistry(),
		expectedResult: base.Registry{
			Default: "default-image_configuration",
			Override: map[string]string{
				"container1":              "new-registry.io/test/path/new-container-1:new-tag",
				"container2":              "new-registry.io/test/path/new-container-2:new-tag",
				"test-deploy1/container1": "new-registry.io/test/path/new-container-1:new-tag",
				"test-deploy1/container2": "new-registry.io/test/path/new-container-2:new-tag",
			},
		},
	}, {
		name: "Image flags with correct container, deploy, component and namespace",
		imageCMDFlags: ImageFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "test-deploy",
			ImageKey:   "container1",
		},
		input: testRegistry(),
		expectedResult: base.Registry{
			Default: "default-image_configuration",
			Override: map[string]string{
				"container1":              "new-registry.io/test/path/new-container-1:new-tag",
				"container2":              "new-registry.io/test/path/new-container-2:new-tag",
				"test-deploy/container2":  "new-registry.io/test/path/new-container-2:new-tag",
				"test-deploy1/container1": "new-registry.io/test/path/new-container-1:new-tag",
				"test-deploy1/container2": "new-registry.io/test/path/new-container-2:new-tag",
			},
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := removeImagesFields(tt.input, tt.imageCMDFlags)
			testingUtil.AssertDeepEqual(t, result, tt.expectedResult)
		})
	}
}
