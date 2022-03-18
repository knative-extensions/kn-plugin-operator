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
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/test/logging"
)

const (
	// Interval specifies the time between two polls.
	Interval = 10 * time.Second
	// Timeout specifies the timeout for the function PollImmediate to reach a certain status.
	Timeout = 5 * time.Minute
)

// WaitForKnativeOperatorDeployment waits and gets the status of the deployment for Knative Operator
func WaitForKnativeOperatorDeployment(clients kubernetes.Interface, name, namespace string,
	inState func(s *appsv1.Deployment, err error) (bool, error)) (*appsv1.Deployment, error) {
	span := logging.GetEmitableSpan(context.Background(), fmt.Sprintf("WaitForKnativeOperatorDeployment/%s/%s", name, "KnativeOperatorIsReady"))
	defer span.End()

	var lastState *appsv1.Deployment
	waitErr := wait.PollImmediate(Interval, Timeout, func() (bool, error) {
		state, err := clients.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		lastState = state
		return inState(lastState, err)
	})

	if waitErr != nil {
		return lastState, fmt.Errorf("Knative Operator deployment %s is not in desired state, got: %+v: %w", name, lastState, waitErr)
	}
	return lastState, nil
}

// WaitForServiceAccount waits and gets the status of the ServiceAccount
func WaitForServiceAccount(clients kubernetes.Interface, name, namespace string) (*corev1.ServiceAccount, error) {
	span := logging.GetEmitableSpan(context.Background(), fmt.Sprintf("WaitForServiceAccount/%s", name))
	defer span.End()

	var lastState *corev1.ServiceAccount
	waitErr := wait.PollImmediate(Interval, Timeout, func() (bool, error) {
		state, err := clients.CoreV1().ServiceAccounts(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		lastState = state
		return true, nil
	})

	if waitErr != nil {
		return lastState, fmt.Errorf("The Service Account %s is not ready, got: %+v: %w", name, lastState, waitErr)
	}
	return lastState, nil
}

// WaitForClusterRoleBinding waits and gets the status of the ClusterRoleBinding
func WaitForClusterRoleBinding(clients kubernetes.Interface, name, namespace string,
	inState func(s *rbacv1.ClusterRoleBinding, namespace string, err error) (bool, error)) (*rbacv1.ClusterRoleBinding, error) {
	span := logging.GetEmitableSpan(context.Background(), fmt.Sprintf("WaitForServiceAccount/%s", name))
	defer span.End()

	var lastState *rbacv1.ClusterRoleBinding
	waitErr := wait.PollImmediate(Interval, Timeout, func() (bool, error) {
		state, err := clients.RbacV1().ClusterRoleBindings().Get(context.TODO(), name, metav1.GetOptions{})
		lastState = state
		return inState(lastState, namespace, err)
	})

	if waitErr != nil {
		return lastState, fmt.Errorf("The Service Account %s is not ready, got: %+v: %w", name, lastState, waitErr)
	}
	return lastState, nil
}

// WaitForRoleBinding waits and gets the status of the RoleBinding
func WaitForRoleBinding(clients kubernetes.Interface, name, namespace string) (*rbacv1.RoleBinding, error) {
	span := logging.GetEmitableSpan(context.Background(), fmt.Sprintf("WaitForRoleBinding/%s", name))
	defer span.End()

	var lastState *rbacv1.RoleBinding
	waitErr := wait.PollImmediate(Interval, Timeout, func() (bool, error) {
		state, err := clients.RbacV1().RoleBindings(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		lastState = state
		return true, nil
	})

	if waitErr != nil {
		return lastState, fmt.Errorf("The RoleBinding %s is not ready, got: %+v: %w", name, lastState, waitErr)
	}
	return lastState, nil
}

// WaitForRole waits and gets the status of the Role
func WaitForRole(clients kubernetes.Interface, name, namespace string) (*rbacv1.Role, error) {
	span := logging.GetEmitableSpan(context.Background(), fmt.Sprintf("WaitForRole/%s", name))
	defer span.End()

	var lastState *rbacv1.Role
	waitErr := wait.PollImmediate(Interval, Timeout, func() (bool, error) {
		state, err := clients.RbacV1().Roles(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		lastState = state
		return true, nil
	})

	if waitErr != nil {
		return lastState, fmt.Errorf("The Role %s is not ready, got: %+v: %w", name, lastState, waitErr)
	}
	return lastState, nil
}

// WaitForSecret waits and gets the status of the Secret
func WaitForSecret(clients kubernetes.Interface, name, namespace string) (*corev1.Secret, error) {
	span := logging.GetEmitableSpan(context.Background(), fmt.Sprintf("WaitForSecret/%s", name))
	defer span.End()

	var lastState *corev1.Secret
	waitErr := wait.PollImmediate(Interval, Timeout, func() (bool, error) {
		state, err := clients.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		lastState = state
		return true, nil
	})

	if waitErr != nil {
		return lastState, fmt.Errorf("The Secret %s is not ready, got: %+v: %w", name, lastState, waitErr)
	}
	return lastState, nil
}

// WaitForService waits and gets the status of the Service
func WaitForService(clients kubernetes.Interface, name, namespace string) (*corev1.Service, error) {
	span := logging.GetEmitableSpan(context.Background(), fmt.Sprintf("WaitForService/%s", name))
	defer span.End()

	var lastState *corev1.Service
	waitErr := wait.PollImmediate(Interval, Timeout, func() (bool, error) {
		state, err := clients.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		lastState = state
		return true, nil
	})

	if waitErr != nil {
		return lastState, fmt.Errorf("The Service %s is not ready, got: %+v: %w", name, lastState, waitErr)
	}
	return lastState, nil
}

// WaitForJob waits and gets the status of the Job
func WaitForJob(clients kubernetes.Interface, prefix, namespace string,
	inState func(s *batchv1.Job, err error) (bool, error)) (*batchv1.Job, error) {
	span := logging.GetEmitableSpan(context.Background(), fmt.Sprintf("WaitForJob/%s", prefix))
	defer span.End()

	var lastState *batchv1.Job
	waitErr := wait.PollImmediate(Interval, Timeout, func() (bool, error) {
		jobList, err := clients.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		for _, job := range jobList.Items {
			if strings.HasPrefix(job.Name, prefix) {
				lastState = &job
				return inState(lastState, err)
			}
		}
		return false, fmt.Errorf("The Job %s is not found", prefix)
	})

	if waitErr != nil {
		return lastState, fmt.Errorf("The Job %s is not ready, got: %+v: %w", prefix, lastState, waitErr)
	}
	return lastState, nil
}

func IsKnativeOperatorJobComplete(s *batchv1.Job, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	return isJobStatusComplete(s.Status), err
}

func isJobStatusComplete(status batchv1.JobStatus) bool {
	for _, c := range status.Conditions {
		if c.Type == batchv1.JobComplete && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func IsKnativeOperatorReady(s *appsv1.Deployment, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	return isStatusReady(s.Status), err
}

func isStatusReady(status appsv1.DeploymentStatus) bool {
	for _, c := range status.Conditions {
		if c.Type == appsv1.DeploymentAvailable && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func IsClusterRoleBindingReady(s *rbacv1.ClusterRoleBinding, namespace string, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	return isSubjectNSCorrect(s.Subjects, namespace), err
}

func isSubjectNSCorrect(subjects []rbacv1.Subject, namespace string) bool {
	for _, s := range subjects {
		if s.Namespace == namespace {
			return true
		}
	}
	return false
}
