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

func TestValidateSelectorFlags(t *testing.T) {
	for _, tt := range []struct {
		name           string
		labelCMDFlags  common.KeyValueFlags
		expectedResult error
	}{{
		name: "Selector flags with correct component and namespace",
		labelCMDFlags: common.KeyValueFlags{
			Component: "eventing",
			Namespace: "test-eventing",
		},
		expectedResult: nil,
	}, {
		name: "Selector flags with correct component, namespace and the deploy name",
		labelCMDFlags: common.KeyValueFlags{
			Component:   "eventing",
			Namespace:   "test-eventing",
			ServiceName: "test-deploy",
		},
		expectedResult: nil,
	}, {
		name: "Node selector flags without component",
		labelCMDFlags: common.KeyValueFlags{
			Namespace:   "test-eventing",
			ServiceName: "test-deploy",
		},
		expectedResult: fmt.Errorf("You need to specify the component for Knative."),
	}, {
		name: "Node selector flags without namespace",
		labelCMDFlags: common.KeyValueFlags{
			Component:   "eventing",
			ServiceName: "test-deploy",
		},
		expectedResult: fmt.Errorf("You need to specify the namespace."),
	}, {
		name: "Node selector flags with invalid component",
		labelCMDFlags: common.KeyValueFlags{
			Component:   "eventing1",
			Namespace:   "test-eventing",
			ServiceName: "test-deploy",
		},
		expectedResult: fmt.Errorf("You need to specify the component for Knative: serving or eventing."),
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := validateSelectorFlags(tt.labelCMDFlags)
			if tt.expectedResult == nil {
				testingUtil.AssertEqual(t, result, nil)
			} else {
				testingUtil.AssertEqual(t, result.Error(), tt.expectedResult.Error())
			}
		})
	}
}

func testServiceForSelectors() []base.ServiceOverride {
	return []base.ServiceOverride{
		{
			Name: "net-istio-controller",
			Selector: map[string]string{"test-key": "v0.13.0",
				"test-key-1": "test-val-1"},
		},
		{
			Name: "net-istio-controller-1",
			Selector: map[string]string{"test-key": "v0.13.0",
				"test-key-1": "test-val-1"},
		},
	}
}

func TestRemoveSelectorsServiceFields(t *testing.T) {
	for _, tt := range []struct {
		name           string
		labelCMDFlags  common.KeyValueFlags
		input          []base.ServiceOverride
		expectedResult []base.ServiceOverride
	}{{
		name: "Selector flags with correct component and namespace",
		labelCMDFlags: common.KeyValueFlags{
			Component: "eventing",
			Namespace: "test-eventing",
		},
		input: testServiceForSelectors(),
		expectedResult: []base.ServiceOverride{
			{
				Name: "net-istio-controller",
			},
			{
				Name: "net-istio-controller-1",
			},
		},
	}, {
		name: "Selector flags with correct service, component and namespace",
		labelCMDFlags: common.KeyValueFlags{
			Component:   "eventing",
			Namespace:   "test-eventing",
			ServiceName: "net-istio-controller",
		},
		input: testServiceForSelectors(),
		expectedResult: []base.ServiceOverride{
			{
				Name: "net-istio-controller",
			},
			{
				Name: "net-istio-controller-1",
				Selector: map[string]string{"test-key": "v0.13.0",
					"test-key-1": "test-val-1"},
			},
		},
	}, {
		name: "Node selector flags with correct service, component and namespace",
		labelCMDFlags: common.KeyValueFlags{
			Component:   "eventing",
			Namespace:   "test-eventing",
			ServiceName: "net-istio-controller",
			Key:         "test-key",
		},
		input: testServiceForSelectors(),
		expectedResult: []base.ServiceOverride{
			{
				Name:     "net-istio-controller",
				Selector: map[string]string{"test-key-1": "test-val-1"},
			},
			{
				Name: "net-istio-controller-1",
				Selector: map[string]string{"test-key": "v0.13.0",
					"test-key-1": "test-val-1"},
			},
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := removeSelectorsServiceFields(tt.input, tt.labelCMDFlags)
			testingUtil.AssertDeepEqual(t, result, tt.expectedResult)
		})
	}
}
