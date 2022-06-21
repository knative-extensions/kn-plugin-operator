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
	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
	"knative.dev/operator/pkg/apis/operator/base"
)

func TestValidateResourcesFlags(t *testing.T) {
	for _, tt := range []struct {
		name           string
		resourcesFlags ResourcesFlags
		expectedResult error
	}{{
		name: "Resource flags with correct component and namespace",
		resourcesFlags: ResourcesFlags{
			Component: "eventing",
			Namespace: "test-eventing",
		},
		expectedResult: nil,
	}, {
		name: "Resource flags with correct component, namespace and the deploy name",
		resourcesFlags: ResourcesFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "test-deploy",
		},
		expectedResult: nil,
	}, {
		name: "Resource flags with correct component, namespace, deploy name and container name",
		resourcesFlags: ResourcesFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "test-deploy",
			Container:  "test-container",
		},
		expectedResult: nil,
	}, {
		name: "Resource flags without component",
		resourcesFlags: ResourcesFlags{
			Namespace:  "test-eventing",
			DeployName: "test-deploy",
			Container:  "test-container",
		},
		expectedResult: fmt.Errorf("You need to specify the component name."),
	}, {
		name: "Resource flags without namespace",
		resourcesFlags: ResourcesFlags{
			Component:  "eventing",
			DeployName: "test-deploy",
			Container:  "test-container",
		},
		expectedResult: fmt.Errorf("You need to specify the namespace."),
	}, {
		name: "Resource flags without deploy name for the container",
		resourcesFlags: ResourcesFlags{
			Component: "eventing",
			Namespace: "test-eventing",
			Container: "test-container",
		},
		expectedResult: fmt.Errorf("You need to specify the deployment name for the container."),
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := validateResourcesFlags(tt.resourcesFlags)
			if tt.expectedResult == nil {
				testingUtil.AssertEqual(t, result, nil)
			} else {
				testingUtil.AssertEqual(t, result.Error(), tt.expectedResult.Error())
			}
		})
	}
}

func testDeploymentOverride() []base.DeploymentOverride {
	return []base.DeploymentOverride{
		{
			Name: "net-istio-controller",
			Resources: []base.ResourceRequirementsOverride{{
				Container: "controller",
				ResourceRequirements: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("999m"),
						corev1.ResourceMemory: resource.MustParse("999Mi")},
					Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("999m"),
						corev1.ResourceMemory: resource.MustParse("999Mi")},
				},
			}, {
				Container: "controller-1",
				ResourceRequirements: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("999m"),
						corev1.ResourceMemory: resource.MustParse("999Mi")},
					Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("999m"),
						corev1.ResourceMemory: resource.MustParse("999Mi")},
				},
			}},
		},
		{
			Name: "net-istio-controller-1",
			Resources: []base.ResourceRequirementsOverride{{
				Container: "controller",
				ResourceRequirements: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("999m"),
						corev1.ResourceMemory: resource.MustParse("999Mi")},
					Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("999m"),
						corev1.ResourceMemory: resource.MustParse("999Mi")},
				},
			}, {
				Container: "controller-1",
				ResourceRequirements: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("999m"),
						corev1.ResourceMemory: resource.MustParse("999Mi")},
					Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("999m"),
						corev1.ResourceMemory: resource.MustParse("999Mi")},
				},
			}},
		},
	}
}

func TestRemoveResourcesFields(t *testing.T) {
	for _, tt := range []struct {
		name           string
		resourcesFlags ResourcesFlags
		input          []base.DeploymentOverride
		expectedResult []base.DeploymentOverride
	}{{
		name: "Resource flags with correct component and namespace",
		resourcesFlags: ResourcesFlags{
			Component: "eventing",
			Namespace: "test-eventing",
		},
		input: testDeploymentOverride(),
		expectedResult: []base.DeploymentOverride{
			{
				Name: "net-istio-controller",
			},
			{
				Name: "net-istio-controller-1",
			},
		},
	}, {
		name: "Resource flags with correct deploy, component and namespace",
		resourcesFlags: ResourcesFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "net-istio-controller",
		},
		input: testDeploymentOverride(),
		expectedResult: []base.DeploymentOverride{
			{
				Name: "net-istio-controller",
			},
			{
				Name: "net-istio-controller-1",
				Resources: []base.ResourceRequirementsOverride{{
					Container: "controller",
					ResourceRequirements: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("999m"),
							corev1.ResourceMemory: resource.MustParse("999Mi")},
						Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("999m"),
							corev1.ResourceMemory: resource.MustParse("999Mi")},
					},
				}, {
					Container: "controller-1",
					ResourceRequirements: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("999m"),
							corev1.ResourceMemory: resource.MustParse("999Mi")},
						Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("999m"),
							corev1.ResourceMemory: resource.MustParse("999Mi")},
					},
				}},
			},
		},
	}, {
		name: "Resource flags with correct container, deploy, component and namespace",
		resourcesFlags: ResourcesFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "net-istio-controller",
			Container:  "controller",
		},
		input: testDeploymentOverride(),
		expectedResult: []base.DeploymentOverride{
			{
				Name: "net-istio-controller",
				Resources: []base.ResourceRequirementsOverride{{
					Container: "controller-1",
					ResourceRequirements: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("999m"),
							corev1.ResourceMemory: resource.MustParse("999Mi")},
						Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("999m"),
							corev1.ResourceMemory: resource.MustParse("999Mi")},
					},
				}},
			},
			{
				Name: "net-istio-controller-1",
				Resources: []base.ResourceRequirementsOverride{{
					Container: "controller",
					ResourceRequirements: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("999m"),
							corev1.ResourceMemory: resource.MustParse("999Mi")},
						Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("999m"),
							corev1.ResourceMemory: resource.MustParse("999Mi")},
					},
				}, {
					Container: "controller-1",
					ResourceRequirements: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("999m"),
							corev1.ResourceMemory: resource.MustParse("999Mi")},
						Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("999m"),
							corev1.ResourceMemory: resource.MustParse("999Mi")},
					},
				}},
			},
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := removeResourcesFields(tt.input, tt.resourcesFlags)
			testingUtil.AssertDeepEqual(t, result, tt.expectedResult)
		})
	}
}
