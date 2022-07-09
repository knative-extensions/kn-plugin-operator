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
	"testing"

	"knative.dev/kn-plugin-operator/pkg/command/common"
	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
	"knative.dev/operator/pkg/apis/operator/base"
)

func testDeploymentForAnnotations() []base.DeploymentOverride {
	return []base.DeploymentOverride{
		{
			Name: "net-istio-controller",
			Annotations: map[string]string{"test-key": "v0.13.0",
				"test-key-1": "test-val-1"},
		},
		{
			Name: "net-istio-controller-1",
			Annotations: map[string]string{"test-key": "v0.13.0",
				"test-key-1": "test-val-1"},
		},
	}
}

func TestRemoveAnnotationsDeployFields(t *testing.T) {
	for _, tt := range []struct {
		name           string
		labelCMDFlags  common.KeyValueFlags
		input          []base.DeploymentOverride
		expectedResult []base.DeploymentOverride
	}{{
		name: "Label flags with correct component and namespace",
		labelCMDFlags: common.KeyValueFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "net-istio-controller",
		},
		input: testDeploymentForAnnotations(),
		expectedResult: []base.DeploymentOverride{
			{
				Name: "net-istio-controller",
			},
			{
				Name: "net-istio-controller-1",
				Annotations: map[string]string{"test-key": "v0.13.0",
					"test-key-1": "test-val-1"},
			},
		},
	}, {
		name: "Label flags with correct deploy, component and namespace",
		labelCMDFlags: common.KeyValueFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "net-istio-controller",
			Key:        "test-key",
		},
		input: testDeploymentForAnnotations(),
		expectedResult: []base.DeploymentOverride{
			{
				Name:        "net-istio-controller",
				Annotations: map[string]string{"test-key-1": "test-val-1"},
			},
			{
				Name: "net-istio-controller-1",
				Annotations: map[string]string{"test-key": "v0.13.0",
					"test-key-1": "test-val-1"},
			},
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := removeAnnotationsDeployFields(tt.input, tt.labelCMDFlags)
			testingUtil.AssertDeepEqual(t, result, tt.expectedResult)
		})
	}
}

func testServiceForAnnotations() []base.ServiceOverride {
	return []base.ServiceOverride{
		{
			Name: "net-istio-controller",
			Annotations: map[string]string{"test-key": "v0.13.0",
				"test-key-1": "test-val-1"},
		},
		{
			Name: "net-istio-controller-1",
			Annotations: map[string]string{"test-key": "v0.13.0",
				"test-key-1": "test-val-1"},
		},
	}
}

func TestRemoveAnnotationsServiceFields(t *testing.T) {
	for _, tt := range []struct {
		name           string
		labelCMDFlags  common.KeyValueFlags
		input          []base.ServiceOverride
		expectedResult []base.ServiceOverride
	}{{
		name: "Label flags with correct component and namespace",
		labelCMDFlags: common.KeyValueFlags{
			Component:   "eventing",
			Namespace:   "test-eventing",
			ServiceName: "net-istio-controller",
		},
		input: testServiceForAnnotations(),
		expectedResult: []base.ServiceOverride{
			{
				Name: "net-istio-controller",
			},
			{
				Name: "net-istio-controller-1",
				Annotations: map[string]string{"test-key": "v0.13.0",
					"test-key-1": "test-val-1"},
			},
		},
	}, {
		name: "Label flags with correct deploy, component and namespace",
		labelCMDFlags: common.KeyValueFlags{
			Component:   "eventing",
			Namespace:   "test-eventing",
			ServiceName: "net-istio-controller",
			Key:         "test-key",
		},
		input: testServiceForAnnotations(),
		expectedResult: []base.ServiceOverride{
			{
				Name:        "net-istio-controller",
				Annotations: map[string]string{"test-key-1": "test-val-1"},
			},
			{
				Name: "net-istio-controller-1",
				Annotations: map[string]string{"test-key": "v0.13.0",
					"test-key-1": "test-val-1"},
			},
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := removeAnnotationsServiceFields(tt.input, tt.labelCMDFlags)
			testingUtil.AssertDeepEqual(t, result, tt.expectedResult)
		})
	}
}
