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

	"knative.dev/kn-plugin-operator/pkg"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingv1alpha1 "knative.dev/operator/pkg/apis/operator/v1alpha1"
	servingv1alpha1 "knative.dev/operator/pkg/apis/operator/v1alpha1"
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
	if strings.EqualFold(component, ServingComponent) {
		return ko.GetKnativeServing(namespace)
	} else if strings.EqualFold(component, EventingComponent) {
		return ko.GetKnativeEventing(namespace)
	}
	return nil, fmt.Errorf("unknow component is set in --component or -c\n")
}

// GetKnativeServing gets the Knative Serving custom resource under a certain namespace
func (ko *KnativeOperatorCR) GetKnativeServing(namespace string) (interface{}, error) {
	knativeServing, err := ko.KnativeOperatorClient.OperatorV1alpha1().KnativeServings(namespace).Get(context.TODO(),
		KnativeServingName, metav1.GetOptions{})

	serving := &servingv1alpha1.KnativeServing{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KnativeServing",
			APIVersion: "operator.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      KnativeServingName,
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

// GetKnativeEventing gets the Knative Eventing custom resource under a certain namespace
func (ko *KnativeOperatorCR) GetKnativeEventing(namespace string) (interface{}, error) {
	knativeEventing, err := ko.KnativeOperatorClient.OperatorV1alpha1().KnativeEventings(namespace).Get(context.TODO(),
		KnativeEventingName, metav1.GetOptions{})

	eventing := &eventingv1alpha1.KnativeEventing{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KnativeEventing",
			APIVersion: "operator.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      KnativeEventingName,
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
	output := ""
	ksCR, err := GetKnativeOperatorCR(p)
	if err != nil {
		return output, err
	}

	kCR, err := ksCR.GetCRInterface(component, namespace)
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
	if err := manifest.Apply(); err != nil {
		return err
	}
	return nil
}
