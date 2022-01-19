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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const KnativeOperatorName = "knative-operator"

// Deployment is used to access the cluster to check if the deployment of the knative operator exists
type Deployment struct {
	Client kubernetes.Interface
}

// CheckIfOperatorInstalled checks if Knative Operator exists
func (d *Deployment) CheckIfOperatorInstalled() (bool, error) {
	namespaces, err := d.Client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to receive a namespace list: %s", err)
	}

	found := false
	for _, ns := range namespaces.Items {
		// Create if the Knative Operator deployment exists
		_, err := d.Client.AppsV1().Deployments(ns.ObjectMeta.Name).Get(context.TODO(), KnativeOperatorName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			continue
		} else if err != nil {
			return false, err
		}
		found = true
		break
	}

	return found, nil
}
