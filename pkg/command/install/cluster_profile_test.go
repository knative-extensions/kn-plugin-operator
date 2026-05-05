/*
Copyright 2026 The Knative Authors

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

package install

import (
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/kn-plugin-operator/pkg/command/common"
	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
	"knative.dev/operator/pkg/apis/operator/base"
	operatorv1beta1 "knative.dev/operator/pkg/apis/operator/v1beta1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func TestValidateClusterProfileFlags(t *testing.T) {
	for _, tt := range []struct {
		name    string
		flags   installCmdFlags
		wantErr string
	}{{
		name:  "no flags",
		flags: installCmdFlags{},
	}, {
		name: "both flags",
		flags: installCmdFlags{
			Component:               common.ServingComponent,
			ClusterProfile:          " spoke ",
			ClusterProfileNamespace: " fleet-system ",
		},
	}, {
		name: "missing namespace",
		flags: installCmdFlags{
			Component:      common.ServingComponent,
			ClusterProfile: "spoke",
		},
		wantErr: "must be provided together",
	}, {
		name: "operator install",
		flags: installCmdFlags{
			ClusterProfile:          "spoke",
			ClusterProfileNamespace: "fleet-system",
		},
		wantErr: "require --component serving or --component eventing",
	}, {
		name: "blank profile",
		flags: installCmdFlags{
			Component:               common.ServingComponent,
			ClusterProfile:          " ",
			ClusterProfileNamespace: "fleet-system",
		},
		wantErr: "must be non-empty",
	}} {
		t.Run(tt.name, func(t *testing.T) {
			err := validateClusterProfileFlags(&tt.flags)
			if tt.wantErr == "" {
				testingUtil.AssertEqual(t, err, nil)
			} else if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
			if tt.name == "both flags" {
				testingUtil.AssertEqual(t, tt.flags.ClusterProfile, "spoke")
				testingUtil.AssertEqual(t, tt.flags.ClusterProfileNamespace, "fleet-system")
			}
		})
	}
}

func TestClusterProfileOverlayAndValues(t *testing.T) {
	flags := installCmdFlags{
		Component:               common.EventingComponent,
		Namespace:               "knative-eventing",
		Version:                 "1.18.0",
		ClusterProfile:          "spoke",
		ClusterProfileNamespace: "fleet-system",
	}
	overlay := getOverlayYamlContent(&flags)
	values := getYamlValuesContent(&flags)

	if !strings.Contains(overlay, "clusterProfileRef") {
		t.Fatalf("expected clusterProfileRef overlay, got:\n%s", overlay)
	}
	if !strings.Contains(values, "cluster_profile: spoke") || !strings.Contains(values, "cluster_profile_namespace: fleet-system") {
		t.Fatalf("expected cluster profile values, got:\n%s", values)
	}
}

func TestClusterProfileOverlayRenders(t *testing.T) {
	flags := installCmdFlags{
		Component:               common.ServingComponent,
		Namespace:               "knative-serving",
		Version:                 "1.18.0",
		ClusterProfile:          "spoke",
		ClusterProfileNamespace: "fleet-system",
		Istio:                   true,
	}
	yttp := common.YttProcessor{
		BaseData: []byte(`apiVersion: operator.knative.dev/v1beta1
kind: KnativeServing
metadata:
  name: knative-serving
  namespace: knative-serving
spec:
  version: latest
`),
		OverlayData: []byte(getOverlayYamlContent(&flags)),
		ValuesData:  []byte(getYamlValuesContent(&flags)),
	}
	result, err := yttp.GenerateOutput()
	if err != nil {
		t.Fatalf("expected cluster profile overlay to render: %v", err)
	}
	if !strings.Contains(result, "clusterProfileRef:") ||
		!strings.Contains(result, "name: spoke") ||
		!strings.Contains(result, "namespace: fleet-system") {
		t.Fatalf("expected rendered cluster profile ref, got:\n%s", result)
	}
}

func TestClusterProfileProviderFileArg(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		want bool
	}{{
		name: "inline",
		args: []string{"--clusterprofile-provider-file=/var/run/provider.json"},
		want: true,
	}, {
		name: "split",
		args: []string{"--clusterprofile-provider-file", "/var/run/provider.json"},
		want: true,
	}, {
		name: "empty inline",
		args: []string{"--clusterprofile-provider-file="},
	}, {
		name: "empty split",
		args: []string{"--clusterprofile-provider-file", ""},
	}, {
		name: "next flag",
		args: []string{"--clusterprofile-provider-file", "--other"},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			testingUtil.AssertEqual(t, hasClusterProfileProviderFileArg(tt.args), tt.want)
		})
	}
}

func TestValidateOperatorMulticlusterEnabled(t *testing.T) {
	for _, tt := range []struct {
		name      string
		container corev1.Container
	}{{
		name: "provider file in args",
		container: corev1.Container{
			Name: common.KnativeOperatorName,
			Args: []string{"--clusterprofile-provider-file=/var/run/provider.json"},
		},
	}, {
		name: "provider file in command",
		container: corev1.Container{
			Name:    common.KnativeOperatorName,
			Command: []string{"manager", "--clusterprofile-provider-file=/var/run/provider.json"},
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			p := &pkg.OperatorParams{
				NewKubeClient: func() (kubernetes.Interface, error) {
					return kubefake.NewSimpleClientset(
						&appsv1.Deployment{
							ObjectMeta: metav1.ObjectMeta{Name: common.KnativeOperatorName, Namespace: "operator-ns"},
							Spec:       appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{tt.container}}}},
						},
					), nil
				},
			}
			testingUtil.AssertEqual(t, validateOperatorMulticlusterEnabled("operator-ns", p), nil)
		})
	}

	missingProvider := &pkg.OperatorParams{
		NewKubeClient: func() (kubernetes.Interface, error) {
			return kubefake.NewSimpleClientset(
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: common.KnativeOperatorName, Namespace: "operator-ns"},
					Spec:       appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: common.KnativeOperatorName}}}}},
				},
			), nil
		},
	}
	if err := validateOperatorMulticlusterEnabled("operator-ns", missingProvider); err == nil || !strings.Contains(err.Error(), "--clusterprofile-provider-file") {
		t.Fatalf("expected provider-file error, got %v", err)
	}
}

func TestValidateClusterProfileCRDSupport(t *testing.T) {
	p := &pkg.OperatorParams{
		NewDynamicClient: func() (dynamic.Interface, error) {
			return dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), clusterProfileCRD("knativeservings.operator.knative.dev", true)), nil
		},
	}
	testingUtil.AssertEqual(t, validateClusterProfileCRDSupport(common.ServingComponent, p), nil)

	missingField := &pkg.OperatorParams{
		NewDynamicClient: func() (dynamic.Interface, error) {
			return dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), clusterProfileCRD("knativeservings.operator.knative.dev", false)), nil
		},
	}
	if err := validateClusterProfileCRDSupport(common.ServingComponent, missingField); err == nil || !strings.Contains(err.Error(), "does not support spec.clusterProfileRef") {
		t.Fatalf("expected CRD support error, got %v", err)
	}
}

func TestRemoteReadinessTargetClusterResolvedFalse(t *testing.T) {
	ks := &operatorv1beta1.KnativeServing{
		Status: operatorv1beta1.KnativeServingStatus{
			Status: duckv1.Status{Conditions: duckv1.Conditions{{
				Type:    base.TargetClusterResolved,
				Status:  corev1.ConditionFalse,
				Reason:  base.ReasonMulticlusterDisabled,
				Message: "multi-cluster is disabled",
			}}},
		},
	}
	ready, err := IsRemoteKnativeServingReady(ks, common.Latest, nil)
	testingUtil.AssertEqual(t, ready, false)
	if err != nil {
		t.Fatalf("expected TargetClusterResolved=False to keep polling, got %v", err)
	}

	diagnostics := servingRemoteDiagnostics(ks)
	if !strings.Contains(diagnostics, string(base.TargetClusterResolved)) || !strings.Contains(diagnostics, base.ReasonMulticlusterDisabled) {
		t.Fatalf("expected remote diagnostics, got %s", diagnostics)
	}

	ke := &operatorv1beta1.KnativeEventing{
		Status: operatorv1beta1.KnativeEventingStatus{
			Status: duckv1.Status{Conditions: duckv1.Conditions{{
				Type:    base.TargetClusterResolved,
				Status:  corev1.ConditionFalse,
				Reason:  base.ReasonMulticlusterDisabled,
				Message: "multi-cluster is disabled",
			}}},
		},
	}
	ready, err = IsRemoteKnativeEventingReady(ke, common.Latest, nil)
	testingUtil.AssertEqual(t, ready, false)
	if err != nil {
		t.Fatalf("expected TargetClusterResolved=False to keep polling, got %v", err)
	}

	diagnostics = eventingRemoteDiagnostics(ke)
	if !strings.Contains(diagnostics, string(base.TargetClusterResolved)) || !strings.Contains(diagnostics, base.ReasonMulticlusterDisabled) {
		t.Fatalf("expected remote diagnostics, got %s", diagnostics)
	}
}

func clusterProfileCRD(name string, includeClusterProfileRef bool) *unstructured.Unstructured {
	specProperties := map[string]interface{}{}
	if includeClusterProfileRef {
		specProperties["clusterProfileRef"] = map[string]interface{}{"type": "object"}
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec": map[string]interface{}{
				"versions": []interface{}{
					map[string]interface{}{
						"name":   "v1beta1",
						"served": true,
						"schema": map[string]interface{}{
							"openAPIV3Schema": map[string]interface{}{
								"properties": map[string]interface{}{
									"spec": map[string]interface{}{
										"type":       "object",
										"properties": specProperties,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
