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

package resources

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/kn-plugin-operator/pkg/command/configure"
	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
	"knative.dev/operator/pkg/apis/operator/base"
	operatorv1beta1 "knative.dev/operator/pkg/client/clientset/versioned/typed/operator/v1beta1"
	"knative.dev/operator/test"
)

// VerifyOperatorInstallationAlpha verifies all the Knative Operator Resources for alpha CRD version
func VerifyOperatorInstallationAlpha(t *testing.T, clients *test.Clients) {
	_, err := WaitForKnativeOperatorDeployment(clients.KubeClient, OperatorName,
		OperatorNamespace, IsKnativeOperatorReady)
	testingUtil.AssertEqual(t, err, nil)
	_, err = WaitForServiceAccount(clients.KubeClient, "knative-operator",
		OperatorNamespace)
	testingUtil.AssertEqual(t, err, nil)
	_, err = WaitForClusterRoleBinding(clients.KubeClient, "knative-serving-operator",
		OperatorNamespace, IsClusterRoleBindingReady)
	testingUtil.AssertEqual(t, err, nil)
	_, err = WaitForClusterRoleBinding(clients.KubeClient, "knative-serving-operator-aggregated",
		OperatorNamespace, IsClusterRoleBindingReady)
	testingUtil.AssertEqual(t, err, nil)
	_, err = WaitForClusterRoleBinding(clients.KubeClient, "knative-eventing-operator",
		OperatorNamespace, IsClusterRoleBindingReady)
	testingUtil.AssertEqual(t, err, nil)
	_, err = WaitForClusterRoleBinding(clients.KubeClient, "knative-eventing-operator-aggregated",
		OperatorNamespace, IsClusterRoleBindingReady)
	testingUtil.AssertEqual(t, err, nil)
}

// VerifyOperatorInstallationBeta verifies all the Knative Operator Resources for beta CRD version
func VerifyOperatorInstallationBeta(t *testing.T, clients *test.Clients) {
	VerifyOperatorInstallationAlpha(t, clients)
	_, err := WaitForKnativeOperatorDeployment(clients.KubeClient, "operator-webhook",
		OperatorNamespace, IsKnativeOperatorReady)
	testingUtil.AssertEqual(t, err, nil)

	_, err = WaitForServiceAccount(clients.KubeClient, "operator-webhook",
		OperatorNamespace)
	testingUtil.AssertEqual(t, err, nil)

	_, err = WaitForServiceAccount(clients.KubeClient, "knative-operator-post-install-job",
		OperatorNamespace)
	testingUtil.AssertEqual(t, err, nil)

	_, err = WaitForClusterRoleBinding(clients.KubeClient, "operator-webhook",
		OperatorNamespace, IsClusterRoleBindingReady)
	testingUtil.AssertEqual(t, err, nil)

	_, err = WaitForClusterRoleBinding(clients.KubeClient, "knative-operator-post-install-job-role-binding",
		OperatorNamespace, IsClusterRoleBindingReady)
	testingUtil.AssertEqual(t, err, nil)

	_, err = WaitForRoleBinding(clients.KubeClient, "operator-webhook", OperatorNamespace)
	testingUtil.AssertEqual(t, err, nil)

	_, err = WaitForRole(clients.KubeClient, "knative-operator-webhook", OperatorNamespace)
	testingUtil.AssertEqual(t, err, nil)

	_, err = WaitForSecret(clients.KubeClient, "operator-webhook-certs", OperatorNamespace)
	testingUtil.AssertEqual(t, err, nil)

	_, err = WaitForService(clients.KubeClient, "operator-webhook", OperatorNamespace)
	testingUtil.AssertEqual(t, err, nil)

	_, err = WaitForJob(clients.KubeClient, "storage-version-migration-operator", OperatorNamespace,
		IsKnativeOperatorJobComplete)
	testingUtil.AssertEqual(t, err, nil)
}

func VerifyKnativeServingExistence(t *testing.T, clients operatorv1beta1.KnativeServingInterface, resourcesFlags configure.ResourcesFlags) {
	ks, err := clients.Get(context.TODO(), "knative-serving", metav1.GetOptions{})
	testingUtil.AssertEqual(t, err, nil)
	VerifyDeploymentOverride(t, ks.Spec.DeploymentOverride, resourcesFlags)
}

func VerifyKnativeEventingExistence(t *testing.T, clients operatorv1beta1.KnativeEventingInterface, resourcesFlags configure.ResourcesFlags) {
	ke, err := clients.Get(context.TODO(), "knative-eventing", metav1.GetOptions{})
	testingUtil.AssertEqual(t, err, nil)
	VerifyDeploymentOverride(t, ke.Spec.DeploymentOverride, resourcesFlags)
}

func VerifyDeploymentOverride(t *testing.T, deploymentOverride []base.DeploymentOverride, resourcesFlags configure.ResourcesFlags) {
	testingUtil.AssertEqual(t, len(deploymentOverride), 1)

	deploy := findDeployment(resourcesFlags.DeployName, deploymentOverride)
	testingUtil.AssertEqual(t, deploy == nil, false)
	testingUtil.AssertEqual(t, len(deploy.Resources), 1)

	firstResource := deploy.Resources[0]
	testingUtil.AssertEqual(t, firstResource.Container, resourcesFlags.Container)

	resourceRequirements := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse(resourcesFlags.LimitCPU),
			corev1.ResourceMemory: resource.MustParse(resourcesFlags.LimitMemory)},
		Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse(resourcesFlags.RequestCPU),
			corev1.ResourceMemory: resource.MustParse(resourcesFlags.RequestMemory)},
	}
	testingUtil.AssertDeepEqual(t, firstResource.ResourceRequirements, resourceRequirements)
}

func VerifyKnativeServingLabelsExistence(t *testing.T, clients operatorv1beta1.KnativeServingInterface, deployLabelFlags configure.DeploymentLabelFlags) {
	ks, err := clients.Get(context.TODO(), "knative-serving", metav1.GetOptions{})
	testingUtil.AssertEqual(t, err, nil)
	VerifyDeploymentLabels(t, ks.Spec.DeploymentOverride, deployLabelFlags)
}

func VerifyKnativeEventingLabelsExistence(t *testing.T, clients operatorv1beta1.KnativeEventingInterface, deployLabelFlags configure.DeploymentLabelFlags) {
	ks, err := clients.Get(context.TODO(), "knative-eventing", metav1.GetOptions{})
	testingUtil.AssertEqual(t, err, nil)
	VerifyDeploymentLabels(t, ks.Spec.DeploymentOverride, deployLabelFlags)
}

func VerifyKnativeServingServiceLabelsExistence(t *testing.T, clients operatorv1beta1.KnativeServingInterface, deployLabelFlags configure.DeploymentLabelFlags) {
	ks, err := clients.Get(context.TODO(), "knative-serving", metav1.GetOptions{})
	testingUtil.AssertEqual(t, err, nil)
	VerifyServiceLabels(t, ks.Spec.ServiceOverride, deployLabelFlags)
}

func VerifyKnativeEventingServiceLabelsExistence(t *testing.T, clients operatorv1beta1.KnativeEventingInterface, deployLabelFlags configure.DeploymentLabelFlags) {
	ks, err := clients.Get(context.TODO(), "knative-eventing", metav1.GetOptions{})
	testingUtil.AssertEqual(t, err, nil)
	VerifyServiceLabels(t, ks.Spec.ServiceOverride, deployLabelFlags)
}

func VerifyDeploymentLabels(t *testing.T, deploymentOverride []base.DeploymentOverride, deployLabelFlags configure.DeploymentLabelFlags) {
	testingUtil.AssertEqual(t, len(deploymentOverride), 1)

	deploy := findDeployment(deployLabelFlags.DeployName, deploymentOverride)
	testingUtil.AssertEqual(t, deploy == nil, false)

	indicator := "label"
	if deployLabelFlags.Annotation {
		indicator = "annotation"
	} else if deployLabelFlags.NodeSelector {
		indicator = "nodeselector"
	}
	result := findKeyValue(t, deployLabelFlags.Key, deployLabelFlags.Value, indicator, deploy)
	testingUtil.AssertEqual(t, result, true)
}

func findDeployment(name string, deploymentOverride []base.DeploymentOverride) *base.DeploymentOverride {
	for _, deploy := range deploymentOverride {
		if deploy.Name == name {
			return &deploy
		}
	}
	return nil
}

func findKeyValue(t *testing.T, key, expectedValue, indicator string, deploy *base.DeploymentOverride) bool {
	if indicator == "label" {
		if data, ok := deploy.Labels[key]; ok && expectedValue == data {
			return true
		}
	} else if indicator == "annotation" {
		if data, ok := deploy.Annotations[key]; ok && expectedValue == data {
			return true
		}
	} else if indicator == "nodeselector" {
		if data, ok := deploy.NodeSelector[key]; ok && expectedValue == data {
			return true
		}
	}
	return false
}

func VerifyServiceLabels(t *testing.T, serviceOverride []base.ServiceOverride, deployLabelFlags configure.DeploymentLabelFlags) {
	testingUtil.AssertEqual(t, len(serviceOverride), 1)

	service := findService(deployLabelFlags.ServiceName, serviceOverride)
	testingUtil.AssertEqual(t, service == nil, false)

	indicator := "label"
	if deployLabelFlags.Annotation {
		indicator = "annotation"
	} else if deployLabelFlags.Selector {
		indicator = "selector"
	}
	result := findKeyValueService(t, deployLabelFlags.Key, deployLabelFlags.Value, indicator, service)
	testingUtil.AssertEqual(t, result, true)
}

func findService(name string, serviceOverride []base.ServiceOverride) *base.ServiceOverride {
	for _, service := range serviceOverride {
		if service.Name == name {
			return &service
		}
	}
	return nil
}

func findKeyValueService(t *testing.T, key, expectedValue, indicator string, deploy *base.ServiceOverride) bool {
	if indicator == "label" {
		if data, ok := deploy.Labels[key]; ok && expectedValue == data {
			return true
		}
	} else if indicator == "annotation" {
		if data, ok := deploy.Annotations[key]; ok && expectedValue == data {
			return true
		}
	} else if indicator == "selector" {
		if data, ok := deploy.Selector[key]; ok && expectedValue == data {
			return true
		}
	}
	return false
}

func VerifyKnativeServingConfigMaps(t *testing.T, clients operatorv1beta1.KnativeServingInterface, cmsFlags configure.CMsFlags) {
	ks, err := clients.Get(context.TODO(), "knative-serving", metav1.GetOptions{})
	testingUtil.AssertEqual(t, err, nil)
	VerifyConfigMaps(t, ks.Spec.GetConfig(), cmsFlags)
}
func VerifyKnativeEventingConfigMaps(t *testing.T, clients operatorv1beta1.KnativeEventingInterface, cmsFlags configure.CMsFlags) {
	ke, err := clients.Get(context.TODO(), "knative-eventing", metav1.GetOptions{})
	testingUtil.AssertEqual(t, err, nil)
	VerifyConfigMaps(t, ke.Spec.GetConfig(), cmsFlags)
}

func VerifyConfigMaps(t *testing.T, configMapData base.ConfigMapData, cmsFlags configure.CMsFlags) {
	data, cmExist := configMapData[cmsFlags.CMName]
	testingUtil.AssertEqual(t, cmExist, true)
	value, valueExist := data[cmsFlags.Key]
	testingUtil.AssertEqual(t, valueExist, true)
	testingUtil.AssertEqual(t, value, cmsFlags.Value)
}

func VerifyKnativeServingHAs(t *testing.T, clients operatorv1beta1.KnativeServingInterface, haFlags configure.HAFlags) {
	ks, err := clients.Get(context.TODO(), "knative-serving", metav1.GetOptions{})
	testingUtil.AssertEqual(t, err, nil)
	VerifyHAs(t, ks.Spec.CommonSpec, haFlags)
}

func VerifyKnativeEventingHAs(t *testing.T, clients operatorv1beta1.KnativeEventingInterface, haFlags configure.HAFlags) {
	ke, err := clients.Get(context.TODO(), "knative-eventing", metav1.GetOptions{})
	testingUtil.AssertEqual(t, err, nil)
	VerifyHAs(t, ke.Spec.CommonSpec, haFlags)
}

func VerifyHAs(t *testing.T, spec base.CommonSpec, haFlags configure.HAFlags) {
	if haFlags.DeployName != "" {
		deploy := findDeployment(haFlags.DeployName, spec.DeploymentOverride)
		testingUtil.AssertEqual(t, deploy == nil, false)
		stringValue := strconv.Itoa(int(*deploy.Replicas))
		testingUtil.AssertEqual(t, stringValue, haFlags.Replicas)
	} else {
		stringValue := strconv.Itoa(int(*spec.HighAvailability.Replicas))
		testingUtil.AssertEqual(t, stringValue, haFlags.Replicas)
	}
}

func VerifyKnativeServingTolerations(t *testing.T, clients operatorv1beta1.KnativeServingInterface, tolerationsFlags configure.TolerationsFlags) {
	ks, err := clients.Get(context.TODO(), "knative-serving", metav1.GetOptions{})
	testingUtil.AssertEqual(t, err, nil)
	VerifyTolerations(t, ks.Spec.DeploymentOverride, tolerationsFlags)
}

func VerifyKnativeEventingTolerations(t *testing.T, clients operatorv1beta1.KnativeEventingInterface, tolerationsFlags configure.TolerationsFlags) {
	ke, err := clients.Get(context.TODO(), "knative-eventing", metav1.GetOptions{})
	testingUtil.AssertEqual(t, err, nil)
	VerifyTolerations(t, ke.Spec.DeploymentOverride, tolerationsFlags)
}

func VerifyTolerations(t *testing.T, deploymentOverride []base.DeploymentOverride, tolerationsFlags configure.TolerationsFlags) {
	deploy := findDeployment(tolerationsFlags.DeployName, deploymentOverride)
	testingUtil.AssertEqual(t, deploy == nil, false)
	toleration := findToleration(tolerationsFlags.Key, deploy.Tolerations)
	testingUtil.AssertEqual(t, toleration == nil, false)
	testingUtil.AssertEqual(t, toleration.Value, tolerationsFlags.Value)

	var operator string
	if toleration.Operator == corev1.TolerationOpEqual {
		operator = "Equal"
	} else if toleration.Operator == corev1.TolerationOpExists {
		operator = "Exists"
	}

	var effect string
	if toleration.Effect == corev1.TaintEffectNoSchedule {
		effect = "NoSchedule"
	} else if toleration.Effect == corev1.TaintEffectNoExecute {
		effect = "NoExecute"
	} else if toleration.Effect == corev1.TaintEffectPreferNoSchedule {
		effect = "PreferNoSchedule"
	}

	testingUtil.AssertDeepEqual(t, operator, tolerationsFlags.Operator)
	testingUtil.AssertDeepEqual(t, effect, tolerationsFlags.Effect)
}

func findToleration(key string, tolerations []corev1.Toleration) *corev1.Toleration {
	for _, toleration := range tolerations {
		if toleration.Key == key {
			return &toleration
		}
	}
	return nil
}

func VerifyKnativeEventingImages(t *testing.T, clients operatorv1beta1.KnativeEventingInterface, imageFlags configure.ImageFlags) {
	ke, err := clients.Get(context.TODO(), "knative-eventing", metav1.GetOptions{})
	testingUtil.AssertEqual(t, err, nil)
	VerifyImages(t, ke.Spec.Registry, imageFlags)
}

func VerifyKnativeServingImages(t *testing.T, clients operatorv1beta1.KnativeServingInterface, imageFlags configure.ImageFlags) {
	ks, err := clients.Get(context.TODO(), "knative-serving", metav1.GetOptions{})
	testingUtil.AssertEqual(t, err, nil)
	VerifyImages(t, ks.Spec.Registry, imageFlags)
}

func VerifyImages(t *testing.T, registry base.Registry, imageFlags configure.ImageFlags) {
	overrideMap := registry.Override
	imageKey := imageFlags.ImageKey
	if imageFlags.DeployName != "" {
		imageKey = fmt.Sprintf("%s/%s", imageFlags.DeployName, imageFlags.ImageKey)
		val, ok := overrideMap[imageKey]
		testingUtil.AssertEqual(t, ok, true)
		testingUtil.AssertEqual(t, val, imageFlags.ImageUrl)
	}
	if imageKey == "default" {
		testingUtil.AssertEqual(t, registry.Default, imageFlags.ImageUrl)
	}
}

func VerifyKnativeEventingEnvVars(t *testing.T, clients operatorv1beta1.KnativeEventingInterface, envVarFlags configure.EnvVarFlags) {
	ke, err := clients.Get(context.TODO(), "knative-eventing", metav1.GetOptions{})
	testingUtil.AssertEqual(t, err, nil)
	VerifyEnvVars(t, ke.Spec.DeploymentOverride, envVarFlags)
}

func VerifyKnativeServingEnvVars(t *testing.T, clients operatorv1beta1.KnativeServingInterface, envVarFlags configure.EnvVarFlags) {
	ks, err := clients.Get(context.TODO(), "knative-serving", metav1.GetOptions{})
	testingUtil.AssertEqual(t, err, nil)
	VerifyEnvVars(t, ks.Spec.DeploymentOverride, envVarFlags)
}

func VerifyEnvVars(t *testing.T, deploymentOverride []base.DeploymentOverride, envVarFlags configure.EnvVarFlags) {
	deploy := findDeployment(envVarFlags.DeployName, deploymentOverride)
	testingUtil.AssertEqual(t, deploy == nil, false)
	envVar := findEnvVar(envVarFlags.ContainerName, deploy.Env)
	testingUtil.AssertEqual(t, envVar == nil, false)
	testingUtil.AssertEqual(t, includeEnvVar(envVarFlags.EnvName, envVarFlags.EnvValue, envVar.EnvVars), true)
}

func findEnvVar(name string, envVarOverrides []base.EnvRequirementsOverride) *base.EnvRequirementsOverride {
	for _, envVarOverride := range envVarOverrides {
		if envVarOverride.Container == name {
			return &envVarOverride
		}
	}
	return nil
}

func includeEnvVar(name, value string, envVars []corev1.EnvVar) bool {
	for _, envVar := range envVars {
		if envVar.Name == name && envVar.Value == value {
			return true
		}
	}
	return false
}
