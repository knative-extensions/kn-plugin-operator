/*
Copyright 2022 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
)

func TestAddVerticalBar(t *testing.T) {
	for _, tt := range []struct {
		name           string
		input          string
		expectedResult string
	}{{
		name: "Create the ConfigMap data for custom manifests",
		input: `apiVersion: operator.knative.dev/v1beta1
kind: KnativeEventing
metadata:
  #@overlay/match missing_ok=True
  namespace: #@ data.values.namespace
#@overlay/match missing_ok=True
spec:
  #@overlay/match missing_ok=True
  config:

    #@overlay/match missing_ok=True
    eventing-controller:
      #@overlay/match missing_ok=True
      test-key: #@ data.values.value`,
		expectedResult: `|
apiVersion: operator.knative.dev/v1beta1
kind: KnativeEventing
metadata:
  #@overlay/match missing_ok=True
  namespace: #@ data.values.namespace
#@overlay/match missing_ok=True
spec:
  #@overlay/match missing_ok=True
  config:

    #@overlay/match missing_ok=True
    eventing-controller:
      #@overlay/match missing_ok=True
      test-key: #@ data.values.value`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := addVerticalBar(tt.input)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}

func TestAppendCMData(t *testing.T) {
	for _, tt := range []struct {
		name           string
		existingData   string
		input          string
		expectedResult string
	}{{
		name: "Create the ConfigMap data for custom manifests",
		existingData: `|
existing test data`,
		input: `new data`,
		expectedResult: `|
existing test data
---
new data`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := appendCMData(tt.existingData, tt.input)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}

func TestUpdateVolumes(t *testing.T) {
	for _, tt := range []struct {
		name           string
		input          []v1.Volume
		expectedResult []v1.Volume
	}{{
		name:  "Update the empty volumes",
		input: []v1.Volume{},
		expectedResult: []v1.Volume{{
			Name: CustomVolumeName,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: ConfigMapName,
					},
				},
			},
		}},
	}, {
		name: "Update the volumes with existing volume of other name",
		input: []v1.Volume{{
			Name: "test",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: ConfigMapName,
					},
				},
				HostPath: &v1.HostPathVolumeSource{
					Path: "test-path",
				},
			},
		}},
		expectedResult: []v1.Volume{{
			Name: "test",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: ConfigMapName,
					},
				},
				HostPath: &v1.HostPathVolumeSource{
					Path: "test-path",
				},
			},
		}, {
			Name: CustomVolumeName,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: ConfigMapName,
					},
				},
			},
		}},
	}, {
		name: "Update the volumes with existing volumes",
		input: []v1.Volume{{
			Name: CustomVolumeName,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "test",
					},
				},
				HostPath: &v1.HostPathVolumeSource{
					Path: "test-path",
				},
			},
		}},
		expectedResult: []v1.Volume{{
			Name: CustomVolumeName,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: ConfigMapName,
					},
				},
			},
		}},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := updateVolumes(tt.input)
			testingUtil.AssertDeepEqual(t, result, tt.expectedResult)
		})
	}
}

func TestUpdateVolumeMounts(t *testing.T) {
	for _, tt := range []struct {
		name           string
		input          []v1.VolumeMount
		expectedResult []v1.VolumeMount
	}{{
		name:  "Update the empty volume mounts",
		input: []v1.VolumeMount{},
		expectedResult: []v1.VolumeMount{{
			Name:      CustomVolumeName,
			MountPath: MountPath,
			ReadOnly:  true,
		}},
	}, {
		name: "Update the volumes with existing volume of other name",
		input: []v1.VolumeMount{{
			Name:      "test",
			MountPath: "test-mount-path",
		}},
		expectedResult: []v1.VolumeMount{{
			Name:      "test",
			MountPath: "test-mount-path",
		}, {
			Name:      CustomVolumeName,
			MountPath: MountPath,
			ReadOnly:  true,
		}},
	}, {
		name: "Update the volumes with existing volumes",
		input: []v1.VolumeMount{{
			Name:      CustomVolumeName,
			MountPath: "MountPath",
		}},
		expectedResult: []v1.VolumeMount{{
			Name:      CustomVolumeName,
			MountPath: MountPath,
			ReadOnly:  true,
		}},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := updateVolumeMounts(tt.input)
			testingUtil.AssertDeepEqual(t, result, tt.expectedResult)
		})
	}
}

func TestUpdateContainers(t *testing.T) {
	for _, tt := range []struct {
		name           string
		input          []v1.Container
		expectedResult []v1.Container
	}{{
		name:  "Update the empty containers",
		input: []v1.Container{},
		expectedResult: []v1.Container{{
			Name: KnativeOperatorName,
			VolumeMounts: []v1.VolumeMount{{
				Name:      CustomVolumeName,
				ReadOnly:  true,
				MountPath: MountPath,
			}},
		}},
	}, {
		name: "Update the containers with existing container of other name",
		input: []v1.Container{{
			Name: "KnativeOperatorName",
			VolumeMounts: []v1.VolumeMount{{
				Name:      "CustomVolumeName",
				ReadOnly:  false,
				MountPath: "MountPath",
			}},
		}},
		expectedResult: []v1.Container{{
			Name: "KnativeOperatorName",
			VolumeMounts: []v1.VolumeMount{{
				Name:      "CustomVolumeName",
				ReadOnly:  false,
				MountPath: "MountPath",
			}},
		}, {
			Name: KnativeOperatorName,
			VolumeMounts: []v1.VolumeMount{{
				Name:      CustomVolumeName,
				ReadOnly:  true,
				MountPath: MountPath,
			}},
		}},
	}, {
		name: "Update the containers with existing containers",
		input: []v1.Container{{
			Name: KnativeOperatorName,
			VolumeMounts: []v1.VolumeMount{{
				Name:      "CustomVolumeName",
				ReadOnly:  true,
				MountPath: "MountPath",
			}},
		}},
		expectedResult: []v1.Container{{
			Name: KnativeOperatorName,
			VolumeMounts: []v1.VolumeMount{{
				Name:      "CustomVolumeName",
				ReadOnly:  true,
				MountPath: "MountPath",
			}, {
				Name:      CustomVolumeName,
				ReadOnly:  true,
				MountPath: MountPath,
			}},
		}},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := updateContainers(tt.input)
			testingUtil.AssertDeepEqual(t, result, tt.expectedResult)
		})
	}
}
