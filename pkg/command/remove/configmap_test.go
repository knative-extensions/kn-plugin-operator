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

func TestValidateCMsFlags(t *testing.T) {
	for _, tt := range []struct {
		name           string
		cmsCMDFlags    common.CMsFlags
		expectedResult error
	}{{
		name: "CM flags with correct component and namespace",
		cmsCMDFlags: common.CMsFlags{
			Component: "eventing",
			Namespace: "test-eventing",
		},
		expectedResult: nil,
	}, {
		name: "CM flags with correct component, namespace and the deploy name",
		cmsCMDFlags: common.CMsFlags{
			Component: "eventing",
			Namespace: "test-eventing",
			CMName:    "test-cm",
		},
		expectedResult: nil,
	}, {
		name: "CM flags with correct component, namespace, deploy name and container name",
		cmsCMDFlags: common.CMsFlags{
			Component: "eventing",
			Namespace: "test-eventing",
			CMName:    "test-cm",
			Key:       "test-key",
		},
		expectedResult: nil,
	}, {
		name: "CM flags without component",
		cmsCMDFlags: common.CMsFlags{
			Namespace: "test-eventing",
			CMName:    "test-deploy",
			Key:       "test-container",
		},
		expectedResult: fmt.Errorf("You need to specify the component name."),
	}, {
		name: "CM flags without namespace",
		cmsCMDFlags: common.CMsFlags{
			Component: "eventing",
			CMName:    "test-deploy",
			Key:       "test-container",
		},
		expectedResult: fmt.Errorf("You need to specify the namespace."),
	}, {
		name: "CM flags without deploy name for the container",
		cmsCMDFlags: common.CMsFlags{
			Component: "eventing",
			Namespace: "test-eventing",
			Key:       "test-container",
		},
		expectedResult: fmt.Errorf("You need to specify the name for the ConfigMap."),
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := validateCMsFlags(tt.cmsCMDFlags)
			if tt.expectedResult == nil {
				testingUtil.AssertEqual(t, result, nil)
			} else {
				testingUtil.AssertEqual(t, result.Error(), tt.expectedResult.Error())
			}
		})
	}
}

func testCMData() base.ConfigMapData {
	return base.ConfigMapData{
		"test": {
			"brokerClass": "Foo",
			"apiVersion":  "v1",
			"kind":        "ConfigMap",
			"name":        "config-br-default-channel",
			"namespace":   "knative-eventing",
		},
		"config-test": {
			"brokerClass": "Foo",
			"apiVersion":  "v1",
			"kind":        "ConfigMap",
			"name":        "config-br-default-channel",
			"namespace":   "knative-eventing",
		},
		"config-other": {
			"brokerClass": "Foo",
			"apiVersion":  "v1",
			"kind":        "ConfigMap",
			"name":        "config-br-default-channel",
			"namespace":   "knative-eventing",
		},
	}
}

func TestRemoveCMsFields(t *testing.T) {
	for _, tt := range []struct {
		name           string
		cmsCMDFlags    common.CMsFlags
		input          base.ConfigMapData
		expectedResult base.ConfigMapData
	}{{
		name: "CM flags with correct component and namespace",
		cmsCMDFlags: common.CMsFlags{
			Component: "eventing",
			Namespace: "test-eventing",
		},
		input:          testCMData(),
		expectedResult: nil,
	}, {
		name: "CM flags with correct deploy, component and namespace",
		cmsCMDFlags: common.CMsFlags{
			Component: "eventing",
			Namespace: "test-eventing",
			CMName:    "test",
		},
		input: testCMData(),
		expectedResult: base.ConfigMapData{
			"config-other": {
				"brokerClass": "Foo",
				"apiVersion":  "v1",
				"kind":        "ConfigMap",
				"name":        "config-br-default-channel",
				"namespace":   "knative-eventing",
			},
		},
	}, {
		name: "CM flags with correct deploy, component and namespace",
		cmsCMDFlags: common.CMsFlags{
			Component: "eventing",
			Namespace: "test-eventing",
			CMName:    "config-test",
		},
		input: testCMData(),
		expectedResult: base.ConfigMapData{
			"config-other": {
				"brokerClass": "Foo",
				"apiVersion":  "v1",
				"kind":        "ConfigMap",
				"name":        "config-br-default-channel",
				"namespace":   "knative-eventing",
			},
		},
	}, {
		name: "CM flags with correct container, deploy, component and namespace",
		cmsCMDFlags: common.CMsFlags{
			Component: "eventing",
			Namespace: "test-eventing",
			CMName:    "test",
			Key:       "brokerClass",
		},
		input: testCMData(),
		expectedResult: base.ConfigMapData{
			"test": {
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"name":       "config-br-default-channel",
				"namespace":  "knative-eventing",
			},
			"config-test": {
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"name":       "config-br-default-channel",
				"namespace":  "knative-eventing",
			},
			"config-other": {
				"brokerClass": "Foo",
				"apiVersion":  "v1",
				"kind":        "ConfigMap",
				"name":        "config-br-default-channel",
				"namespace":   "knative-eventing",
			},
		},
	}, {
		name: "CM flags with correct container, deploy, component and namespace",
		cmsCMDFlags: common.CMsFlags{
			Component: "eventing",
			Namespace: "test-eventing",
			CMName:    "config-test",
			Key:       "brokerClass",
		},
		input: testCMData(),
		expectedResult: base.ConfigMapData{
			"test": {
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"name":       "config-br-default-channel",
				"namespace":  "knative-eventing",
			},
			"config-test": {
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"name":       "config-br-default-channel",
				"namespace":  "knative-eventing",
			},
			"config-other": {
				"brokerClass": "Foo",
				"apiVersion":  "v1",
				"kind":        "ConfigMap",
				"name":        "config-br-default-channel",
				"namespace":   "knative-eventing",
			},
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := removeCMsFields(tt.input, tt.cmsCMDFlags)
			testingUtil.AssertDeepEqual(t, result, tt.expectedResult)
		})
	}
}

func TestAddOrRemovePrefix(t *testing.T) {
	result := addOrRemovePrefix("test", "config-")
	testingUtil.AssertDeepEqual(t, result, "config-test")

	result = addOrRemovePrefix("config-test", "config-")
	testingUtil.AssertDeepEqual(t, result, "test")
}
