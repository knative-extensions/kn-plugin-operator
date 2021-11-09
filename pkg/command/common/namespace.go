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
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Namespace is used to access the namespace resource in the Kubernetes cluster.
type Namespace struct {
	client    kubernetes.Interface
	component string
}

// CreateNamespace creates the namespace if it is not available in the Kubernetes cluster
func (ns *Namespace) CreateNamespace(namespace string) error {
	if namespace != "default" {
		// Create the namespace if it is not available
		_, err := ns.client.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			nspace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}
			if strings.EqualFold(ns.component, "serving") {
				nspace = &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace,
					Labels: map[string]string{"istio-injection": "enabled"}}}
			}
			ns.client.CoreV1().Namespaces().Create(context.TODO(), nspace, metav1.CreateOptions{})
		} else if err != nil {
			return err
		}
	}
	return nil
}
