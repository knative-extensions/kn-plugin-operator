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

	"knative.dev/kn-plugin-operator/pkg/command/common"
	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
	"knative.dev/operator/pkg/apis/operator/base"
)

func TestValidateLabelAnnotationsFlags(t *testing.T) {
	for _, tt := range []struct {
		name           string
		labelCMDFlags  common.KeyValueFlags
		expectedResult error
	}{{
		name: "Label flags with correct component and namespace",
		labelCMDFlags: common.KeyValueFlags{
			Component: "eventing",
			Namespace: "test-eventing",
		},
		expectedResult: fmt.Errorf("You need to specify the name of the deployment or the service."),
	}, {
		name: "Label flags with correct component, namespace and the deploy name",
		labelCMDFlags: common.KeyValueFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "test-deploy",
		},
		expectedResult: nil,
	}, {
		name: "Label flags without component",
		labelCMDFlags: common.KeyValueFlags{
			Namespace:  "test-eventing",
			DeployName: "test-deploy",
		},
		expectedResult: fmt.Errorf("You need to specify the component for Knative."),
	}, {
		name: "Label flags without namespace",
		labelCMDFlags: common.KeyValueFlags{
			Component:  "eventing",
			DeployName: "test-deploy",
		},
		expectedResult: fmt.Errorf("You need to specify the namespace."),
	}, {
		name: "Label flags with both service and deploy",
		labelCMDFlags: common.KeyValueFlags{
			Component:   "eventing",
			Namespace:   "test-eventing",
			DeployName:  "test-deploy",
			ServiceName: "test-deploy",
		},
		expectedResult: fmt.Errorf("You are only allowed to specify either --deployName or --serviceName."),
	}, {
		name: "Label flags with invalid component",
		labelCMDFlags: common.KeyValueFlags{
			Component:  "eventing1",
			Namespace:  "test-eventing",
			DeployName: "test-deploy",
		},
		expectedResult: fmt.Errorf("You need to specify the component for Knative: serving or eventing."),
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := validateLabelAnnotationsFlags(tt.labelCMDFlags)
			if tt.expectedResult == nil {
				testingUtil.AssertEqual(t, result, nil)
			} else {
				testingUtil.AssertEqual(t, result.Error(), tt.expectedResult.Error())
			}
		})
	}
}

func testDeploymentForLabels() []base.DeploymentOverride {
	return []base.DeploymentOverride{
		{
			Name: "net-istio-controller",
			Labels: map[string]string{"test-key": "v0.13.0",
				"test-key-1": "test-val-1"},
		},
		{
			Name: "net-istio-controller-1",
			Labels: map[string]string{"test-key": "v0.13.0",
				"test-key-1": "test-val-1"},
		},
	}
}

func TestRemoveLabelsDeployFields(t *testing.T) {
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
		input: testDeploymentForLabels(),
		expectedResult: []base.DeploymentOverride{
			{
				Name: "net-istio-controller",
			},
			{
				Name: "net-istio-controller-1",
				Labels: map[string]string{"test-key": "v0.13.0",
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
		input: testDeploymentForLabels(),
		expectedResult: []base.DeploymentOverride{
			{
				Name:   "net-istio-controller",
				Labels: map[string]string{"test-key-1": "test-val-1"},
			},
			{
				Name: "net-istio-controller-1",
				Labels: map[string]string{"test-key": "v0.13.0",
					"test-key-1": "test-val-1"},
			},
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := removeLabelsDeployFields(tt.input, tt.labelCMDFlags)
			testingUtil.AssertDeepEqual(t, result, tt.expectedResult)
		})
	}
}

func testServiceForLabels() []base.ServiceOverride {
	return []base.ServiceOverride{
		{
			Name: "net-istio-controller",
			Labels: map[string]string{"test-key": "v0.13.0",
				"test-key-1": "test-val-1"},
		},
		{
			Name: "net-istio-controller-1",
			Labels: map[string]string{"test-key": "v0.13.0",
				"test-key-1": "test-val-1"},
		},
	}
}

func TestRemoveLabelsServiceFields(t *testing.T) {
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
		input: testServiceForLabels(),
		expectedResult: []base.ServiceOverride{
			{
				Name: "net-istio-controller",
			},
			{
				Name: "net-istio-controller-1",
				Labels: map[string]string{"test-key": "v0.13.0",
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
		input: testServiceForLabels(),
		expectedResult: []base.ServiceOverride{
			{
				Name:   "net-istio-controller",
				Labels: map[string]string{"test-key-1": "test-val-1"},
			},
			{
				Name: "net-istio-controller-1",
				Labels: map[string]string{"test-key": "v0.13.0",
					"test-key-1": "test-val-1"},
			},
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := removeLabelsServiceFields(tt.input, tt.labelCMDFlags)
			testingUtil.AssertDeepEqual(t, result, tt.expectedResult)
		})
	}
}
