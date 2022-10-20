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

func TestValidateEnvVarsFlags(t *testing.T) {
	for _, tt := range []struct {
		name           string
		envVarFlags    EnvVarFlags
		expectedResult error
	}{{
		name: "EnvVar flags with correct component and namespace",
		envVarFlags: EnvVarFlags{
			Component: "eventing",
			Namespace: "test-eventing",
		},
		expectedResult: nil,
	}, {
		name: "EnvVar flags with correct component, namespace and the deploy name",
		envVarFlags: EnvVarFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "test-deploy",
		},
		expectedResult: nil,
	}, {
		name: "EnvVar flags with correct component, namespace, deploy name and container name",
		envVarFlags: EnvVarFlags{
			Component:     "eventing",
			Namespace:     "test-eventing",
			DeployName:    "test-deploy",
			ContainerName: "test-container",
		},
		expectedResult: nil,
	}, {
		name: "EnvVar flags without component",
		envVarFlags: EnvVarFlags{
			Namespace:     "test-eventing",
			DeployName:    "test-deploy",
			ContainerName: "test-container",
		},
		expectedResult: fmt.Errorf("You need to specify the component name."),
	}, {
		name: "EnvVar flags without namespace",
		envVarFlags: EnvVarFlags{
			Component:     "eventing",
			DeployName:    "test-deploy",
			ContainerName: "test-container",
		},
		expectedResult: fmt.Errorf("You need to specify the namespace."),
	}, {
		name: "EnvVar flags without deploy name for the container",
		envVarFlags: EnvVarFlags{
			Component:     "eventing",
			Namespace:     "test-eventing",
			ContainerName: "test-container",
		},
		expectedResult: fmt.Errorf("You need to specify the name for the deployment resource."),
	}, {
		name: "EnvVar flags without container name",
		envVarFlags: EnvVarFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "test-deploy",
			EnvName:    "test-name",
		},
		expectedResult: fmt.Errorf("You need to specify the name for the container."),
	}, {
		name: "EnvVar flags without container name",
		envVarFlags: EnvVarFlags{
			Component:     "eventing",
			Namespace:     "test-eventing",
			ContainerName: "test-deploy",
			EnvName:       "test-name",
		},
		expectedResult: fmt.Errorf("You need to specify the name for the deployment resource."),
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := validateEnvVarsFlags(tt.envVarFlags)
			if tt.expectedResult == nil {
				testingUtil.AssertEqual(t, result, nil)
			} else {
				testingUtil.AssertEqual(t, result.Error(), tt.expectedResult.Error())
			}
		})
	}
}

func testDeploymentForEnvVars() []base.WorkloadOverride {
	return []base.WorkloadOverride{
		{
			Name: "net-istio-controller",
			Env: []base.EnvRequirementsOverride{{
				Container: "controller",
				EnvVars: []corev1.EnvVar{{
					Name:  "METRICS_DOMAIN",
					Value: "test",
				}, {
					Name:  "METRICS_DOMAIN_TEST",
					Value: "test",
				}},
			}, {
				Container: "controller1",
				EnvVars: []corev1.EnvVar{{
					Name:  "METRICS_DOMAIN",
					Value: "test",
				}, {
					Name:  "METRICS_DOMAIN_TEST",
					Value: "test",
				}},
			}},
		},
		{
			Name: "net-istio-controller-1",
			Env: []base.EnvRequirementsOverride{{
				Container: "controller",
				EnvVars: []corev1.EnvVar{{
					Name:  "METRICS_DOMAIN",
					Value: "test",
				}, {
					Name:  "METRICS_DOMAIN_TEST",
					Value: "test",
				}},
			}, {
				Container: "controller1",
				EnvVars: []corev1.EnvVar{{
					Name:  "METRICS_DOMAIN",
					Value: "test",
				}, {
					Name:  "METRICS_DOMAIN_TEST",
					Value: "test",
				}},
			}},
		},
	}
}

func TestRemoveEnvVarsFields(t *testing.T) {
	for _, tt := range []struct {
		name           string
		envVarFlags    EnvVarFlags
		input          []base.WorkloadOverride
		expectedResult []base.WorkloadOverride
	}{{
		name: "Env Vars flags with correct component and namespace",
		envVarFlags: EnvVarFlags{
			Component: "eventing",
			Namespace: "test-eventing",
		},
		input: testDeploymentForEnvVars(),
		expectedResult: []base.WorkloadOverride{
			{
				Name: "net-istio-controller",
				Env:  nil,
			},
			{
				Name: "net-istio-controller-1",
				Env:  nil,
			},
		},
	}, {
		name: "Env Vars flags with correct deploy, component and namespace",
		envVarFlags: EnvVarFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "net-istio-controller",
		},
		input: testDeploymentForEnvVars(),
		expectedResult: []base.WorkloadOverride{
			{
				Name: "net-istio-controller",
				Env:  nil,
			},
			{
				Name: "net-istio-controller-1",
				Env: []base.EnvRequirementsOverride{{
					Container: "controller",
					EnvVars: []corev1.EnvVar{{
						Name:  "METRICS_DOMAIN",
						Value: "test",
					}, {
						Name:  "METRICS_DOMAIN_TEST",
						Value: "test",
					}},
				}, {
					Container: "controller1",
					EnvVars: []corev1.EnvVar{{
						Name:  "METRICS_DOMAIN",
						Value: "test",
					}, {
						Name:  "METRICS_DOMAIN_TEST",
						Value: "test",
					}},
				}},
			},
		},
	}, {
		name: "Env Vars flags with correct container, deploy, component and namespace",
		envVarFlags: EnvVarFlags{
			Component:     "eventing",
			Namespace:     "test-eventing",
			DeployName:    "net-istio-controller",
			ContainerName: "controller",
		},
		input: testDeploymentForEnvVars(),
		expectedResult: []base.WorkloadOverride{
			{
				Name: "net-istio-controller",
				Env: []base.EnvRequirementsOverride{{
					Container: "controller1",
					EnvVars: []corev1.EnvVar{{
						Name:  "METRICS_DOMAIN",
						Value: "test",
					}, {
						Name:  "METRICS_DOMAIN_TEST",
						Value: "test",
					}},
				}},
			},
			{
				Name: "net-istio-controller-1",
				Env: []base.EnvRequirementsOverride{{
					Container: "controller",
					EnvVars: []corev1.EnvVar{{
						Name:  "METRICS_DOMAIN",
						Value: "test",
					}, {
						Name:  "METRICS_DOMAIN_TEST",
						Value: "test",
					}},
				}, {
					Container: "controller1",
					EnvVars: []corev1.EnvVar{{
						Name:  "METRICS_DOMAIN",
						Value: "test",
					}, {
						Name:  "METRICS_DOMAIN_TEST",
						Value: "test",
					}},
				}},
			},
		},
	}, {
		name: "Env Vars flags with correct env name, container, deploy, component and namespace",
		envVarFlags: EnvVarFlags{
			Component:     "eventing",
			Namespace:     "test-eventing",
			DeployName:    "net-istio-controller",
			ContainerName: "controller",
			EnvName:       "METRICS_DOMAIN",
		},
		input: testDeploymentForEnvVars(),
		expectedResult: []base.WorkloadOverride{
			{
				Name: "net-istio-controller",
				Env: []base.EnvRequirementsOverride{{
					Container: "controller",
					EnvVars: []corev1.EnvVar{{
						Name:  "METRICS_DOMAIN_TEST",
						Value: "test",
					}},
				}, {
					Container: "controller1",
					EnvVars: []corev1.EnvVar{{
						Name:  "METRICS_DOMAIN",
						Value: "test",
					}, {
						Name:  "METRICS_DOMAIN_TEST",
						Value: "test",
					}},
				}},
			},
			{
				Name: "net-istio-controller-1",
				Env: []base.EnvRequirementsOverride{{
					Container: "controller",
					EnvVars: []corev1.EnvVar{{
						Name:  "METRICS_DOMAIN",
						Value: "test",
					}, {
						Name:  "METRICS_DOMAIN_TEST",
						Value: "test",
					}},
				}, {
					Container: "controller1",
					EnvVars: []corev1.EnvVar{{
						Name:  "METRICS_DOMAIN",
						Value: "test",
					}, {
						Name:  "METRICS_DOMAIN_TEST",
						Value: "test",
					}},
				}},
			},
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := removeEnvVarsFields(tt.input, tt.envVarFlags)
			testingUtil.AssertDeepEqual(t, result, tt.expectedResult)
		})
	}
}
