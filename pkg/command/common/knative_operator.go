/*
Copyright 2021 The Knative Authors
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

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	eventingv1beta1 "knative.dev/operator/pkg/apis/operator/v1beta1"
	servingv1beta1 "knative.dev/operator/pkg/apis/operator/v1beta1"

	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/operator/pkg/apis/operator/base"
	"knative.dev/operator/pkg/client/clientset/versioned"
)

const (
	KnativeServingName  = "knative-serving"
	KnativeEventingName = "knative-eventing"
)

var kOperatorCR *KnativeOperatorCR

func GetKnativeOperatorCR(p *pkg.OperatorParams) (*KnativeOperatorCR, error) {
	if kOperatorCR != nil {
		return kOperatorCR, nil
	}
	operatorClient, err := p.NewOperatorClient()
	if err != nil {
		return nil, fmt.Errorf("cannot get source cluster kube config, please use --kubeconfig or export environment variable KUBECONFIG to set\n")
	}
	return &KnativeOperatorCR{
		KnativeOperatorClient: operatorClient,
	}, nil
}

// KnativeOperatorCR is used to access the knative custom resource in the Kubernetes cluster.
type KnativeOperatorCR struct {
	KnativeOperatorClient *versioned.Clientset
}

// GetCRInterface gets the Knative custom resource under a certain namespace
func (ko *KnativeOperatorCR) GetCRInterface(component, namespace string) (interface{}, error) {
	return ko.GetCRInterfaceForName(component, namespace, DefaultComponentName(component))
}

// GetCRInterfaceForName gets the named Knative custom resource under a certain namespace.
func (ko *KnativeOperatorCR) GetCRInterfaceForName(component, namespace, name string) (interface{}, error) {
	if strings.EqualFold(component, ServingComponent) {
		return ko.GetKnativeServingByName(namespace, name)
	} else if strings.EqualFold(component, EventingComponent) {
		return ko.GetKnativeEventingByName(namespace, name)
	}
	return nil, fmt.Errorf("unknow component is set in --component or -c\n")
}

// GetKnativeServing gets the Knative Serving custom resource under a certain namespace
func (ko *KnativeOperatorCR) GetKnativeServing(namespace string) (interface{}, error) {
	return ko.GetKnativeServingByName(namespace, KnativeServingName)
}

// GetKnativeServingByName gets the named Knative Serving custom resource under a certain namespace.
func (ko *KnativeOperatorCR) GetKnativeServingByName(namespace, name string) (interface{}, error) {
	knativeServing, err := ko.GetKnativeServingInClusterByName(namespace, name)

	serving := &servingv1beta1.KnativeServing{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KnativeServing",
			APIVersion: "operator.knative.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	if apierrs.IsNotFound(err) {
		return serving, nil
	} else if err != nil {
		return nil, err
	}

	serving.Spec = knativeServing.Spec
	return serving, nil
}

func (ko *KnativeOperatorCR) GetConfigMaps(component, namespace string) (base.ConfigMapData, error) {
	return ko.GetConfigMapsForName(component, namespace, DefaultComponentName(component))
}

func (ko *KnativeOperatorCR) GetConfigMapsForName(component, namespace, name string) (base.ConfigMapData, error) {
	var cmData base.ConfigMapData
	if strings.EqualFold(component, ServingComponent) {
		ks, err := ko.GetKnativeServingInClusterByName(namespace, name)
		if err != nil {
			return cmData, err
		}
		cmData = ks.Spec.Config
	} else if strings.EqualFold(component, EventingComponent) {
		ke, err := ko.GetKnativeEventingInClusterByName(namespace, name)
		if err != nil {
			return cmData, err
		}
		cmData = ke.Spec.Config
	}

	return cmData, nil
}

func (ko *KnativeOperatorCR) GetRegistry(component, namespace string) (base.Registry, error) {
	return ko.GetRegistryForName(component, namespace, DefaultComponentName(component))
}

func (ko *KnativeOperatorCR) GetRegistryForName(component, namespace, name string) (base.Registry, error) {
	var registry base.Registry
	if strings.EqualFold(component, ServingComponent) {
		ks, err := ko.GetKnativeServingInClusterByName(namespace, name)
		if err != nil {
			return registry, err
		}
		registry = ks.Spec.Registry
	} else if strings.EqualFold(component, EventingComponent) {
		ke, err := ko.GetKnativeEventingInClusterByName(namespace, name)
		if err != nil {
			return registry, err
		}
		registry = ke.Spec.Registry
	}

	return registry, nil
}

func (ko *KnativeOperatorCR) UpdateRegistry(component, namespace string, registry base.Registry) error {
	return ko.UpdateRegistryForName(component, namespace, DefaultComponentName(component), registry)
}

func (ko *KnativeOperatorCR) UpdateRegistryForName(component, namespace, name string, registry base.Registry) error {
	commonSpec, err := ko.GetCommonSpecForName(component, namespace, name)
	if err != nil {
		return err
	}
	commonSpec.Registry = registry
	return ko.UpdateCommonSpecForName(component, namespace, name, commonSpec)
}

func (ko *KnativeOperatorCR) UpdateConfigMaps(component, namespace string, cmData base.ConfigMapData) error {
	return ko.UpdateConfigMapsForName(component, namespace, DefaultComponentName(component), cmData)
}

func (ko *KnativeOperatorCR) UpdateConfigMapsForName(component, namespace, name string, cmData base.ConfigMapData) error {
	commonSpec, err := ko.GetCommonSpecForName(component, namespace, name)
	if err != nil {
		return err
	}
	commonSpec.Config = cmData
	return ko.UpdateCommonSpecForName(component, namespace, name, commonSpec)
}

func (ko *KnativeOperatorCR) GetDeployments(component, namespace string) ([]base.WorkloadOverride, error) {
	return ko.GetDeploymentsForName(component, namespace, DefaultComponentName(component))
}

func (ko *KnativeOperatorCR) GetDeploymentsForName(component, namespace, name string) ([]base.WorkloadOverride, error) {
	var workloadOverrides []base.WorkloadOverride
	if strings.EqualFold(component, ServingComponent) {
		ks, err := ko.GetKnativeServingInClusterByName(namespace, name)
		if err != nil {
			return workloadOverrides, err
		}
		workloadOverrides = ks.Spec.DeploymentOverride
	} else if strings.EqualFold(component, EventingComponent) {
		ke, err := ko.GetKnativeEventingInClusterByName(namespace, name)
		if err != nil {
			return workloadOverrides, err
		}
		workloadOverrides = ke.Spec.DeploymentOverride
	}

	return workloadOverrides, nil
}

func (ko *KnativeOperatorCR) GetServices(component, namespace string) ([]base.ServiceOverride, error) {
	return ko.GetServicesForName(component, namespace, DefaultComponentName(component))
}

func (ko *KnativeOperatorCR) GetServicesForName(component, namespace, name string) ([]base.ServiceOverride, error) {
	var serviceOverrides []base.ServiceOverride
	if strings.EqualFold(component, ServingComponent) {
		ks, err := ko.GetKnativeServingInClusterByName(namespace, name)
		if err != nil {
			return serviceOverrides, err
		}
		serviceOverrides = ks.Spec.ServiceOverride
	} else if strings.EqualFold(component, EventingComponent) {
		ke, err := ko.GetKnativeEventingInClusterByName(namespace, name)
		if err != nil {
			return serviceOverrides, err
		}
		serviceOverrides = ke.Spec.ServiceOverride
	}

	return serviceOverrides, nil
}

func (ko *KnativeOperatorCR) UpdateDeployments(component, namespace string, workloadOverrides []base.WorkloadOverride) error {
	return ko.UpdateDeploymentsForName(component, namespace, DefaultComponentName(component), workloadOverrides)
}

func (ko *KnativeOperatorCR) UpdateDeploymentsForName(component, namespace, name string, workloadOverrides []base.WorkloadOverride) error {
	commonSpec, err := ko.GetCommonSpecForName(component, namespace, name)
	if err != nil {
		return err
	}
	commonSpec.DeploymentOverride = workloadOverrides
	return ko.UpdateCommonSpecForName(component, namespace, name, commonSpec)
}

func (ko *KnativeOperatorCR) UpdateServices(component, namespace string, serviceOverrides []base.ServiceOverride) error {
	return ko.UpdateServicesForName(component, namespace, DefaultComponentName(component), serviceOverrides)
}

func (ko *KnativeOperatorCR) UpdateServicesForName(component, namespace, name string, serviceOverrides []base.ServiceOverride) error {
	commonSpec, err := ko.GetCommonSpecForName(component, namespace, name)
	if err != nil {
		return err
	}
	commonSpec.ServiceOverride = serviceOverrides
	return ko.UpdateCommonSpecForName(component, namespace, name, commonSpec)
}

func (ko *KnativeOperatorCR) GetCommonSpec(component, namespace string) (*base.CommonSpec, error) {
	return ko.GetCommonSpecForName(component, namespace, DefaultComponentName(component))
}

func (ko *KnativeOperatorCR) GetCommonSpecForName(component, namespace, name string) (*base.CommonSpec, error) {
	var commonSpec base.CommonSpec
	if strings.EqualFold(component, ServingComponent) {
		ks, err := ko.GetKnativeServingInClusterByName(namespace, name)
		if err != nil {
			return nil, err
		}
		commonSpec = ks.Spec.CommonSpec

	} else if strings.EqualFold(component, EventingComponent) {
		ke, err := ko.GetKnativeEventingInClusterByName(namespace, name)
		if err != nil {
			return nil, err
		}
		commonSpec = ke.Spec.CommonSpec
	}

	return &commonSpec, nil
}

func (ko *KnativeOperatorCR) UpdateCommonSpec(component, namespace string, commonSpec *base.CommonSpec) error {
	return ko.UpdateCommonSpecForName(component, namespace, DefaultComponentName(component), commonSpec)
}

func (ko *KnativeOperatorCR) UpdateCommonSpecForName(component, namespace, name string, commonSpec *base.CommonSpec) error {
	if strings.EqualFold(component, ServingComponent) {
		ks, err := ko.GetKnativeServingInClusterByName(namespace, name)
		if err != nil {
			return err
		}
		ks.Spec.CommonSpec = *commonSpec
		_, err = ko.UpdateKnativeServing(ks)
		if err != nil {
			return err
		}

	} else if strings.EqualFold(component, EventingComponent) {
		ke, err := ko.GetKnativeEventingInClusterByName(namespace, name)
		if err != nil {
			return err
		}
		ke.Spec.CommonSpec = *commonSpec
		_, err = ko.UpdateKnativeEventing(ke)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetKnativeServingInCluster gets the Knative Serving custom resource in the cluster under a certain namespace
func (ko *KnativeOperatorCR) GetKnativeServingInCluster(namespace string) (*servingv1beta1.KnativeServing, error) {
	return ko.GetKnativeServingInClusterByName(namespace, KnativeServingName)
}

// GetKnativeServingInClusterByName gets the named Knative Serving custom resource in the cluster under a certain namespace.
func (ko *KnativeOperatorCR) GetKnativeServingInClusterByName(namespace, name string) (*servingv1beta1.KnativeServing, error) {
	return ko.KnativeOperatorClient.OperatorV1beta1().KnativeServings(namespace).Get(context.TODO(),
		name, metav1.GetOptions{})
}

// UpdateKnativeServing updates the Knative Serving custom resource in the cluster based on the provided Knative Serving
func (ko *KnativeOperatorCR) UpdateKnativeServing(ks *servingv1beta1.KnativeServing) (*servingv1beta1.KnativeServing, error) {
	return ko.KnativeOperatorClient.OperatorV1beta1().KnativeServings(ks.Namespace).Update(context.TODO(), ks,
		metav1.UpdateOptions{})
}

// GetKnativeEventingInCluster gets the Knative Eventing custom resource in the cluster under a certain namespace
func (ko *KnativeOperatorCR) GetKnativeEventingInCluster(namespace string) (*eventingv1beta1.KnativeEventing, error) {
	return ko.GetKnativeEventingInClusterByName(namespace, KnativeEventingName)
}

// GetKnativeEventingInClusterByName gets the named Knative Eventing custom resource in the cluster under a certain namespace.
func (ko *KnativeOperatorCR) GetKnativeEventingInClusterByName(namespace, name string) (*eventingv1beta1.KnativeEventing, error) {
	return ko.KnativeOperatorClient.OperatorV1beta1().KnativeEventings(namespace).Get(context.TODO(),
		name, metav1.GetOptions{})
}

// UpdateKnativeEventing updates the Knative Eventing custom resource in the cluster based on the provided Knative Eventing
func (ko *KnativeOperatorCR) UpdateKnativeEventing(ks *eventingv1beta1.KnativeEventing) (*eventingv1beta1.KnativeEventing, error) {
	return ko.KnativeOperatorClient.OperatorV1beta1().KnativeEventings(ks.Namespace).Update(context.TODO(), ks,
		metav1.UpdateOptions{})
}

// GetKnativeEventing gets the Knative Eventing custom resource under a certain namespace
func (ko *KnativeOperatorCR) GetKnativeEventing(namespace string) (interface{}, error) {
	return ko.GetKnativeEventingByName(namespace, KnativeEventingName)
}

// GetKnativeEventingByName gets the named Knative Eventing custom resource under a certain namespace.
func (ko *KnativeOperatorCR) GetKnativeEventingByName(namespace, name string) (interface{}, error) {
	knativeEventing, err := ko.GetKnativeEventingInClusterByName(namespace, name)

	eventing := &eventingv1beta1.KnativeEventing{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KnativeEventing",
			APIVersion: "operator.knative.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	if apierrs.IsNotFound(err) {
		return eventing, nil
	} else if err != nil {
		return nil, err
	}

	eventing.Spec = knativeEventing.Spec
	return eventing, nil
}

func GenerateOperatorCRString(component, namespace string, p *pkg.OperatorParams) (string, error) {
	return GenerateOperatorCRStringForName(component, namespace, DefaultComponentName(component), p)
}

func GenerateOperatorCRStringForName(component, namespace, name string, p *pkg.OperatorParams) (string, error) {
	return GenerateOperatorCRStringForRef(ComponentRef{Component: component, Namespace: namespace, Name: name}, p)
}

func GenerateOperatorCRStringForRef(ref ComponentRef, p *pkg.OperatorParams) (string, error) {
	output := ""
	ksCR, err := GetKnativeOperatorCR(p)
	if err != nil {
		return output, err
	}

	kCR, err := ksCR.GetCRInterfaceForName(ref.Component, ref.Namespace, ref.Name)
	if err != nil {
		return output, err
	}

	yamlGenerator := YamlGenarator{
		Input: kCR,
	}

	// Generate the CR template
	return yamlGenerator.GenerateYamlOutput()
}

func ApplyManifests(yamlTemplateString, overlayContent, yamlValuesContent string, p *pkg.OperatorParams) error {
	restConfig, err := p.RestConfig()
	if err != nil {
		return err
	}

	yttp := YttProcessor{
		BaseData:    []byte(yamlTemplateString),
		OverlayData: []byte(overlayContent),
		ValuesData:  []byte(yamlValuesContent),
	}

	manifest := Manifest{
		YttPro:     &yttp,
		RestConfig: restConfig,
	}

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return manifest.Apply()
	})

	if err != nil {
		return err
	}

	return nil
}
