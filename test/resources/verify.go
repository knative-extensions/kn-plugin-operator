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

	deploy := deploymentOverride[0]
	testingUtil.AssertEqual(t, deploy.Name, resourcesFlags.DeployName)
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
