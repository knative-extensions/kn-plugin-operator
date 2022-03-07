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
	"context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	CustomDataKey     = "custom-manifests.yaml"
	VerticalDelimiter = "|"
	LineWrapper       = "\n"
	Separator         = "---"
	MountPath         = "/knative-custom-manifest"
	CustomVolumeName  = "config-manifest-volume"
	ConfigMapName     = "config-manifest"
)

// KubeResource is used to access the Kubernetes resources in the Kubernetes cluster.
type KubeResource struct {
	KubeClient kubernetes.Interface
}

// CreateOrUpdateConfigMap creates or updates the ConfigMap with the data under a certain namespace
func (kr *KubeResource) CreateOrUpdateConfigMap(name, namespace, data string, overwrite bool) error {
	cm, err := kr.getConfigMap(name, namespace)
	if err != nil {
		return err
	}

	cmData := addVerticalBar(data)
	if cm == nil {
		dataArray := []string{}
		dataArray = append(dataArray, VerticalDelimiter)
		dataArray = append(dataArray, data)

		// Create the ConfigMap
		configMap := &v1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Data: map[string]string{CustomDataKey: cmData},
		}

		if _, err := kr.KubeClient.CoreV1().ConfigMaps(namespace).Create(context.TODO(),
			configMap, metav1.CreateOptions{}); err != nil {
			return err
		}
	} else {
		// Update the ConfigMap
		if !overwrite {
			existingData := cm.Data[CustomDataKey]
			cmData = appendCMData(existingData, data)
		}
		cm.Data = map[string]string{CustomDataKey: cmData}
		if _, err := kr.KubeClient.CoreV1().ConfigMaps(namespace).Update(context.TODO(),
			cm, metav1.UpdateOptions{}); err != nil {
			return err
		}
	}
	return nil
}

// getConfigMap gets the ConfigMap under a certain namespace
func (kr *KubeResource) getConfigMap(name, namespace string) (*v1.ConfigMap, error) {
	cm, err := kr.KubeClient.CoreV1().ConfigMaps(namespace).Get(context.TODO(),
		name, metav1.GetOptions{})

	if apierrs.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return cm, nil
}

// UpdateOperatorDeployment updates the deployment of the operator
func (kr *KubeResource) UpdateOperatorDeployment(name, namespace string) error {
	deploy, err := kr.getDeployment(name, namespace)
	if err != nil {
		return err
	}
	if deploy == nil {
		return fmt.Errorf("The Knative Operator is not install.")
	}

	deploy.Spec.Template.Spec.Volumes = updateVolumes(deploy.Spec.Template.Spec.Volumes)
	deploy.Spec.Template.Spec.Containers = updateContainers(deploy.Spec.Template.Spec.Containers)
	if _, err := kr.KubeClient.AppsV1().Deployments(namespace).Update(context.TODO(),
		deploy, metav1.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

// getDeployment gets the deployment under a certain namespace
func (kr *KubeResource) getDeployment(name, namespace string) (*appsv1.Deployment, error) {
	deploy, err := kr.KubeClient.AppsV1().Deployments(namespace).Get(context.TODO(),
		name, metav1.GetOptions{})
	if apierrs.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return deploy, nil
}

func updateContainers(containers []v1.Container) []v1.Container {
	found := false
	for i := range containers {
		if containers[i].Name == KnativeOperatorName {
			found = true
			containers[i].VolumeMounts = updateVolumeMounts(containers[i].VolumeMounts)
		}
	}
	if !found {
		containers = append(containers, v1.Container{
			Name: KnativeOperatorName,
			VolumeMounts: []v1.VolumeMount{{
				Name:      CustomVolumeName,
				ReadOnly:  true,
				MountPath: MountPath,
			}},
		})
	}

	return containers
}

func updateVolumeMounts(volumeMounts []v1.VolumeMount) []v1.VolumeMount {
	found := false
	for i := range volumeMounts {
		if volumeMounts[i].Name == CustomVolumeName {
			found = true
			volumeMounts[i].MountPath = MountPath
			volumeMounts[i].ReadOnly = true
		}
	}
	if !found {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      CustomVolumeName,
			ReadOnly:  true,
			MountPath: MountPath,
		})
	}
	return volumeMounts
}

func updateVolumes(volumes []v1.Volume) []v1.Volume {
	found := false
	for i := range volumes {
		if volumes[i].Name == CustomVolumeName {
			found = true
			volumes[i].VolumeSource = v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: ConfigMapName,
					},
				},
			}
		}
	}
	if !found {
		volumes = append(volumes, v1.Volume{
			Name: CustomVolumeName,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: ConfigMapName,
					},
				},
			},
		})
	}
	return volumes
}

func addVerticalBar(data string) string {
	dataArray := []string{}
	dataArray = append(dataArray, VerticalDelimiter)
	dataArray = append(dataArray, data)
	return strings.Join(dataArray, LineWrapper)
}

func appendCMData(existingData, newData string) string {
	dataArray := []string{}
	dataArray = append(dataArray, existingData)
	dataArray = append(dataArray, Separator)
	dataArray = append(dataArray, newData)
	return strings.Join(dataArray, LineWrapper)
}
