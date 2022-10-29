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

	corev1 "k8s.io/api/core/v1"
	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
	"knative.dev/operator/pkg/apis/operator/base"
)

func TestValidateTolerationsFlags(t *testing.T) {
	for _, tt := range []struct {
		name                string
		tolerationsCMDFlags TolerationsFlags
		expectedResult      error
	}{{
		name: "Toleration flags with correct component and namespace",
		tolerationsCMDFlags: TolerationsFlags{
			Component: "eventing",
			Namespace: "test-eventing",
		},
		expectedResult: nil,
	}, {
		name: "Toleration flags with correct component, namespace and the deploy name",
		tolerationsCMDFlags: TolerationsFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "test-deploy",
		},
		expectedResult: nil,
	}, {
		name: "Toleration flags with correct component, namespace, deploy name and container name",
		tolerationsCMDFlags: TolerationsFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "test-deploy",
			Key:        "test-key",
		},
		expectedResult: nil,
	}, {
		name: "Toleration flags without component",
		tolerationsCMDFlags: TolerationsFlags{
			Namespace:  "test-eventing",
			DeployName: "test-deploy",
			Key:        "test-key",
		},
		expectedResult: fmt.Errorf("You need to specify the component name."),
	}, {
		name: "Toleration flags without namespace",
		tolerationsCMDFlags: TolerationsFlags{
			Component:  "eventing",
			DeployName: "test-deploy",
			Key:        "test-key",
		},
		expectedResult: fmt.Errorf("You need to specify the namespace."),
	}, {
		name: "Toleration flags without deploy name for the container",
		tolerationsCMDFlags: TolerationsFlags{
			Component: "eventing",
			Namespace: "test-eventing",
			Key:       "test-key",
		},
		expectedResult: fmt.Errorf("You need to specify the deployment name for the toleration."),
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := validateTolerationsFlags(tt.tolerationsCMDFlags)
			if tt.expectedResult == nil {
				testingUtil.AssertEqual(t, result, nil)
			} else {
				testingUtil.AssertEqual(t, result.Error(), tt.expectedResult.Error())
			}
		})
	}
}

func testDeploymentForToleration() []base.WorkloadOverride {
	return []base.WorkloadOverride{
		{
			Name: "net-istio-controller",
			Tolerations: []corev1.Toleration{{
				Key:      "test-key",
				Operator: corev1.TolerationOpExists,
				Effect:   corev1.TaintEffectNoSchedule,
			}, {
				Key:      "test-key",
				Operator: "Equal",
				Effect:   corev1.TaintEffectNoSchedule,
			}, {
				Key:      "test-key-1",
				Operator: "Equal",
				Effect:   corev1.TaintEffectNoSchedule,
			}},
		},
		{
			Name: "net-istio-controller-1",
			Tolerations: []corev1.Toleration{{
				Key:      "test-key",
				Operator: corev1.TolerationOpExists,
				Effect:   corev1.TaintEffectNoSchedule,
			}, {
				Key:      "test-key-1",
				Operator: "Equal",
				Effect:   corev1.TaintEffectNoSchedule,
			}},
		},
	}
}

func TestRemoveTolerationsFields(t *testing.T) {
	for _, tt := range []struct {
		name                string
		tolerationsCMDFlags TolerationsFlags
		input               []base.WorkloadOverride
		expectedResult      []base.WorkloadOverride
	}{{
		name: "Toleration flags with correct component and namespace",
		tolerationsCMDFlags: TolerationsFlags{
			Component: "eventing",
			Namespace: "test-eventing",
		},
		input: testDeploymentForToleration(),
		expectedResult: []base.WorkloadOverride{
			{
				Name: "net-istio-controller",
			},
			{
				Name: "net-istio-controller-1",
			},
		},
	}, {
		name: "Toleration flags with correct deploy, component and namespace",
		tolerationsCMDFlags: TolerationsFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "net-istio-controller",
		},
		input: testDeploymentForToleration(),
		expectedResult: []base.WorkloadOverride{
			{
				Name: "net-istio-controller",
			},
			{
				Name: "net-istio-controller-1",
				Tolerations: []corev1.Toleration{{
					Key:      "test-key",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				}, {
					Key:      "test-key-1",
					Operator: "Equal",
					Effect:   corev1.TaintEffectNoSchedule,
				}},
			},
		},
	}, {
		name: "Toleration flags with correct container, deploy, component and namespace",
		tolerationsCMDFlags: TolerationsFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "net-istio-controller",
			Key:        "test-key",
		},
		input: testDeploymentForToleration(),
		expectedResult: []base.WorkloadOverride{
			{
				Name: "net-istio-controller",
				Tolerations: []corev1.Toleration{{
					Key:      "test-key-1",
					Operator: "Equal",
					Effect:   corev1.TaintEffectNoSchedule,
				}},
			},
			{
				Name: "net-istio-controller-1",
				Tolerations: []corev1.Toleration{{
					Key:      "test-key",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				}, {
					Key:      "test-key-1",
					Operator: "Equal",
					Effect:   corev1.TaintEffectNoSchedule,
				}},
			},
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := removeTolerationsFields(tt.input, tt.tolerationsCMDFlags)
			testingUtil.AssertDeepEqual(t, result, tt.expectedResult)
		})
	}
}
