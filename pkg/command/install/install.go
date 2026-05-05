// Copyright 2021 The Knative Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package install

import (
	"context"
	_ "embed"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"knative.dev/kn-plugin-operator/pkg/ui/progressindicator"

	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345

	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/kn-plugin-operator/pkg/command/common"
	operatorapi "knative.dev/operator/pkg/apis/operator"
	"knative.dev/operator/pkg/apis/operator/base"
	"knative.dev/operator/pkg/apis/operator/v1beta1"
	operatorv1beta1 "knative.dev/operator/pkg/client/clientset/versioned/typed/operator/v1beta1"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/test/logging"
)

const clusterProfileProviderFileFlag = "--clusterprofile-provider-file"

var customResourceDefinitionGVR = schema.GroupVersionResource{
	Group:    "apiextensions.k8s.io",
	Version:  "v1",
	Resource: "customresourcedefinitions",
}

//go:embed overlay/ks.yaml
var servingOverlay string

//go:embed overlay/ke.yaml
var eventingOverlay string

//go:embed overlay/ks_istio_ns.yaml
var servingIstioOverlay string

//go:embed overlay/operator.yaml
var operatorOverlay string

//go:embed overlay/operator_crds.yaml
var operatorCRDsOverlay string

//go:embed overlay/ks_ingress.yaml
var servingWithIngressOverlay string

//go:embed overlay/cluster_profile.yaml
var clusterProfileOverlay string

type installCmdFlags struct {
	Component               string
	IstioNamespace          string
	Namespace               string
	CRName                  string
	KubeConfig              string
	Version                 string
	ClusterProfile          string
	ClusterProfileNamespace string
	CRNameExplicit          bool
	Istio                   bool
	Kourier                 bool
	Contour                 bool
}

var (
	ServingKeyDeployments  = []string{"activator", "autoscaler", "autoscaler-hpa", "controller", "webhook"}
	EventingKeyDeployments = []string{"eventing-controller", "eventing-webhook", "imc-controller", "imc-dispatcher",
		"mt-broker-controller", "mt-broker-filter", "mt-broker-ingress", "pingsource-mt-adapter"}
	// Interval specifies the time between two polls.
	Interval = 10 * time.Second
	// Timeout specifies the timeout for the function PollImmediate to reach a certain status.
	Timeout = 5 * time.Minute
)

func (flags *installCmdFlags) fill_defaults() {
	if flags.Version == "" {
		flags.Version = common.Latest
	}

	if flags.IstioNamespace == "" && strings.EqualFold(flags.Component, common.ServingComponent) {
		flags.IstioNamespace = common.DefaultIstioNamespace
	}

	if flags.Namespace == "" {
		if strings.EqualFold(flags.Component, common.ServingComponent) {
			flags.Namespace = common.DefaultKnativeServingNamespace
		} else if strings.EqualFold(flags.Component, common.EventingComponent) {
			flags.Namespace = common.DefaultKnativeEventingNamespace
		} else if flags.Component == "" {
			flags.Namespace = common.DefaultNamespace
		}
	}

	if flags.CRName == "" && flags.Component != "" {
		flags.CRName = common.DefaultComponentName(flags.Component)
	}

	// Set the default ingress istio to true
	if strings.EqualFold(flags.Component, common.ServingComponent) && !flags.Kourier && !flags.Contour && !flags.Istio {
		flags.Istio = true
		flags.Kourier = false
		flags.Contour = false
	}
}

var (
	installFlags installCmdFlags
)

// NewInstallCommand represents the install commands for the operation
func NewInstallCommand(p *pkg.OperatorParams) *cobra.Command {
	var installCmd = &cobra.Command{
		Use:   "install",
		Short: "Install Knative Operator or Knative components",
		Example: `
  # Install Knative Serving under the namespace knative-serving
  kn-operator install -c serving --namespace knative-serving`,

		RunE: func(cmd *cobra.Command, args []string) error {
			installFlags.CRNameExplicit = cmd.Flags().Changed(common.CRNameFlag)
			// Fill in the default values for the empty fields
			err := RunInstallationCommand(&installFlags, p)
			if err != nil {
				return err
			}

			component := "Operator"
			if strings.EqualFold(installFlags.Component, common.ServingComponent) {
				component = common.ServingComponent
			} else if strings.EqualFold(installFlags.Component, common.EventingComponent) {
				component = common.EventingComponent
			}

			if hasClusterProfileFlags(&installFlags) && installFlags.Component != "" {
				ref := common.ComponentRef{Component: installFlags.Component, Namespace: installFlags.Namespace, Name: installFlags.CRName}
				fmt.Fprintf(cmd.OutOrStdout(), "Knative %s of the '%s' version was created as %s for ClusterProfile '%s/%s'.\n",
					component, installFlags.Version, ref.String(), installFlags.ClusterProfileNamespace, installFlags.ClusterProfile)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Knative %s of the '%s' version was created in the namespace '%s'.\n",
				component, installFlags.Version, installFlags.Namespace)
			return nil
		},
	}

	installCmd.Flags().StringVar(&installFlags.KubeConfig, "kubeconfig", "", "The kubeconfig of the Knative resources (default is KUBECONFIG from environment variable)")
	installCmd.Flags().StringVarP(&installFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")
	installCmd.Flags().StringVarP(&installFlags.Component, "component", "c", "", "The name of the Knative Component to install")
	installCmd.Flags().StringVar(&installFlags.CRName, common.CRNameFlag, "", "The name of the hub Knative Serving or Eventing custom resource for remote component installation")
	installCmd.Flags().StringVarP(&installFlags.Version, "version", "v", common.Latest, "The version of the the Knative Operator or the Knative component")
	installCmd.Flags().StringVar(&installFlags.IstioNamespace, "istio-namespace", "", "The namespace of istio")
	installCmd.Flags().StringVar(&installFlags.ClusterProfile, "cluster-profile", "", "The ClusterProfile name for remote component installation")
	installCmd.Flags().StringVar(&installFlags.ClusterProfileNamespace, "cluster-profile-namespace", "", "The ClusterProfile namespace for remote component installation")
	installCmd.Flags().BoolVar(&installFlags.Istio, "istio", false, "The flag to enable the ingress istio")
	installCmd.Flags().BoolVar(&installFlags.Kourier, "kourier", false, "The flag to enable the ingress kourier")
	installCmd.Flags().BoolVar(&installFlags.Contour, "contour", false, "The flag to enable the ingress contour")

	return installCmd
}

func RunInstallationCommand(installFlags *installCmdFlags, p *pkg.OperatorParams) error {
	pi := progressindicator.New().SetText("Installing...")
	pi.Start()
	defer pi.Stop()

	err := validateIngressFlags(installFlags)
	if err != nil {
		return err
	}

	if err := validateClusterProfileFlags(installFlags); err != nil {
		return err
	}

	if err := validateInstallCRNameFlags(installFlags); err != nil {
		return err
	}

	// Fill in the default values for the empty fields
	installFlags.fill_defaults()

	p.KubeCfgPath = installFlags.KubeConfig

	client, err := p.NewKubeClient()
	if err != nil {
		return fmt.Errorf("cannot get source cluster kube config, please use --kubeconfig or export environment variable KUBECONFIG to set\n")
	}
	deploy := common.Deployment{
		Client: client,
	}

	if installFlags.Component != "" {
		component := common.EventingComponent
		if strings.EqualFold(installFlags.Component, common.ServingComponent) {
			component = common.ServingComponent
		}

		currentVersion, err := prepareComponentInstall(installFlags, deploy, p)
		if err != nil {
			return err
		}

		// Install serving or eventing
		versions, err := generateVersionStages(currentVersion, installFlags.Version)
		if err != nil {
			return err
		}

		for _, v := range versions {
			text := fmt.Sprintf("Installing Knative %s, Version %s...", component, v)
			if currentVersion != "" {
				text = fmt.Sprintf("Migrating Knative %s to Version %s...", component, v)
			}
			pi.SetText(text)

			installFlags.Version = v
			err = installKnativeComponent(installFlags, p)
			if err != nil {
				return err
			}
		}

	} else {
		if exists, ns, _, err := checkIfOperatorInstalled(p); err != nil {
			return err
		} else if exists {
			// Check if the namespace is consistent
			if !strings.EqualFold(ns, installFlags.Namespace) {
				return fmt.Errorf("The namespace %s you specified is not consistent with the existing namespace for Knative Operator %s",
					installFlags.Namespace, ns)
			}
		}

		// Install the Knative Operator
		text := fmt.Sprintf("Installing Knative Operator, Version %s...", installFlags.Version)
		pi.SetText(text)
		err = installOperator(installFlags, p)
		if err != nil {
			return err
		}
	}

	pi.Stop()
	return nil
}

func validateIngressFlags(installFlags *installCmdFlags) error {
	count := 0

	if installFlags.Istio {
		count++
	}

	if installFlags.Kourier {
		count++
	}

	if installFlags.Contour {
		count++
	}

	if strings.EqualFold(installFlags.Component, common.ServingComponent) {
		if count > 1 {
			return fmt.Errorf("You can specify only one ingress for Knative Serving.")
		}
	} else if count > 0 {
		return fmt.Errorf("You can only specify the ingress for Knative Serving.")
	}
	return nil
}

// hasClusterProfileFlags reports whether either ClusterProfile flag is set.
// Callers must run validateClusterProfileFlags first; that function trims the
// values, so this helper relies on already-trimmed input.
func hasClusterProfileFlags(flags *installCmdFlags) bool {
	return flags.ClusterProfile != "" || flags.ClusterProfileNamespace != ""
}

func validateClusterProfileFlags(flags *installCmdFlags) error {
	profile := strings.TrimSpace(flags.ClusterProfile)
	profileNamespace := strings.TrimSpace(flags.ClusterProfileNamespace)
	if (flags.ClusterProfile != "" && profile == "") || (flags.ClusterProfileNamespace != "" && profileNamespace == "") {
		return fmt.Errorf("--cluster-profile and --cluster-profile-namespace must be non-empty after trimming whitespace")
	}
	if profile == "" && profileNamespace == "" {
		return nil
	}
	if profile == "" || profileNamespace == "" {
		return fmt.Errorf("--cluster-profile and --cluster-profile-namespace must be provided together")
	}
	if !strings.EqualFold(flags.Component, common.ServingComponent) && !strings.EqualFold(flags.Component, common.EventingComponent) {
		return fmt.Errorf("--cluster-profile and --cluster-profile-namespace require --component serving or --component eventing")
	}
	flags.ClusterProfile = profile
	flags.ClusterProfileNamespace = profileNamespace
	return nil
}

func validateInstallCRNameFlags(flags *installCmdFlags) error {
	if !flags.CRNameExplicit {
		return nil
	}
	name, err := common.NormalizeExplicitComponentName(flags.Component, flags.CRName)
	if err != nil {
		return err
	}
	if !hasClusterProfileFlags(flags) {
		return fmt.Errorf("--%s is valid only for remote component installs with --cluster-profile and --cluster-profile-namespace", common.CRNameFlag)
	}
	flags.CRName = name
	return nil
}

func prepareComponentInstall(installFlags *installCmdFlags, deploy common.Deployment, p *pkg.OperatorParams) (string, error) {
	remote, statusVersion, specVersion, err := detectRemoteComponentInstall(installFlags, p)
	if err != nil {
		return "", err
	}
	if !remote {
		if err := ensureOperatorInstalled(p); err != nil {
			return "", err
		}
		return currentLocalComponentVersion(installFlags, deploy)
	}

	if err := validateRemoteInstallPrereqs(installFlags.Component, p); err != nil {
		return "", err
	}
	if statusVersion == "" && specVersion == "" {
		_, statusVersion, specVersion, err = getExistingComponentCR(installFlags.Component, installFlags.Namespace, installFlags.CRName, p)
		if err != nil {
			return "", err
		}
	}
	return currentRemoteComponentVersion(statusVersion, specVersion), nil
}

func detectRemoteComponentInstall(installFlags *installCmdFlags, p *pkg.OperatorParams) (bool, string, string, error) {
	if hasClusterProfileFlags(installFlags) {
		return true, "", "", nil
	}

	operatorExists, _, _, err := checkIfOperatorInstalled(p)
	if err != nil || !operatorExists {
		return false, "", "", err
	}

	existingRef, statusVersion, specVersion, err := getExistingComponentCR(installFlags.Component, installFlags.Namespace, installFlags.CRName, p)
	if err != nil || existingRef == nil {
		return false, "", "", err
	}

	installFlags.ClusterProfile = existingRef.Name
	installFlags.ClusterProfileNamespace = existingRef.Namespace
	return true, statusVersion, specVersion, nil
}

func currentRemoteComponentVersion(statusVersion, specVersion string) string {
	if currentVersion := concreteVersionOrEmpty(statusVersion); currentVersion != "" {
		return currentVersion
	}
	return concreteVersionOrEmpty(specVersion)
}

func currentLocalComponentVersion(installFlags *installCmdFlags, deploy common.Deployment) (string, error) {
	exists, ns, version, err := deploy.CheckIfKnativeInstalled(installFlags.Component)
	if err != nil || !exists {
		return "", err
	}
	// Check if the namespace is consistent
	if !strings.EqualFold(ns, installFlags.Namespace) {
		return "", fmt.Errorf("The namespace %s you specified is not consistent with the existing namespace for Knative Component %s",
			installFlags.Namespace, ns)
	}
	return version, nil
}

func getExistingComponentCR(component, namespace, name string, p *pkg.OperatorParams) (*base.ClusterProfileReference, string, string, error) {
	operatorClient, err := p.NewOperatorClient()
	if err != nil {
		return nil, "", "", fmt.Errorf("cannot get source cluster kube config, please use --kubeconfig or export environment variable KUBECONFIG to set")
	}

	if strings.EqualFold(component, common.ServingComponent) {
		ks, err := operatorClient.OperatorV1beta1().KnativeServings(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil, "", "", nil
		}
		if err != nil {
			return nil, "", "", err
		}
		return ks.Spec.ClusterProfileRef, ks.Status.Version, ks.Spec.Version, nil
	}

	ke, err := operatorClient.OperatorV1beta1().KnativeEventings(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil, "", "", nil
	}
	if err != nil {
		return nil, "", "", err
	}
	return ke.Spec.ClusterProfileRef, ke.Status.Version, ke.Spec.Version, nil
}

func concreteVersionOrEmpty(version string) string {
	if version == "" || version == common.Latest || version == common.Nightly {
		return ""
	}
	targetVersion := version
	if !strings.HasPrefix(targetVersion, "v") {
		targetVersion = fmt.Sprintf("v%s", targetVersion)
	}
	if !semver.IsValid(targetVersion) {
		return ""
	}
	return version
}

func validateOperatorMulticlusterEnabled(namespace string, p *pkg.OperatorParams) error {
	client, err := p.NewKubeClient()
	if err != nil {
		return fmt.Errorf("cannot get source cluster kube config, please use --kubeconfig or export environment variable KUBECONFIG to set")
	}
	deploy, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), common.KnativeOperatorName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	container := operatorContainer(deploy)
	if container == nil || !hasClusterProfileProviderFileArg(operatorContainerArgs(container)) {
		return fmt.Errorf("Knative Operator Deployment %s/%s is not configured with --clusterprofile-provider-file; configure the Operator for ClusterProfile access before remote installs", namespace, common.KnativeOperatorName)
	}
	return nil
}

func operatorContainer(deploy *v1.Deployment) *corev1.Container {
	for i := range deploy.Spec.Template.Spec.Containers {
		if deploy.Spec.Template.Spec.Containers[i].Name == common.KnativeOperatorName {
			return &deploy.Spec.Template.Spec.Containers[i]
		}
	}
	return nil
}

func operatorContainerArgs(container *corev1.Container) []string {
	args := make([]string, 0, len(container.Command)+len(container.Args))
	args = append(args, container.Command...)
	args = append(args, container.Args...)
	return args
}

func hasClusterProfileProviderFileArg(args []string) bool {
	for i, arg := range args {
		if value, ok := strings.CutPrefix(arg, clusterProfileProviderFileFlag+"="); ok {
			return strings.TrimSpace(value) != ""
		}
		if arg == clusterProfileProviderFileFlag {
			if i+1 >= len(args) {
				return false
			}
			next := strings.TrimSpace(args[i+1])
			return next != "" && !strings.HasPrefix(next, "--")
		}
	}
	return false
}

func validateClusterProfileCRDSupport(component string, p *pkg.OperatorParams) error {
	client, err := p.NewDynamicClient()
	if err != nil {
		return fmt.Errorf("cannot get source cluster kube config, please use --kubeconfig or export environment variable KUBECONFIG to set")
	}

	crdName := operatorapi.KnativeEventingResource.String()
	if strings.EqualFold(component, common.ServingComponent) {
		crdName = operatorapi.KnativeServingResource.String()
	}
	crd, err := client.Resource(customResourceDefinitionGVR).Get(context.TODO(), crdName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return fmt.Errorf("%s CRD is not installed; upgrade the Knative Operator CRDs before remote installs", crdName)
	}
	if err != nil {
		return fmt.Errorf("failed to get %s CRD: %w", crdName, err)
	}
	if !crdSupportsClusterProfileRef(crd) {
		return fmt.Errorf("installed %s CRD does not support spec.clusterProfileRef; upgrade the Knative Operator CRDs before remote installs", crdName)
	}
	return nil
}

func crdSupportsClusterProfileRef(crd *unstructured.Unstructured) bool {
	versions, found, err := unstructured.NestedSlice(crd.Object, "spec", "versions")
	if err != nil || !found {
		return false
	}

	for _, version := range versions {
		versionMap, ok := version.(map[string]interface{})
		if !ok {
			continue
		}
		name, _, _ := unstructured.NestedString(versionMap, "name")
		served, _, _ := unstructured.NestedBool(versionMap, "served")
		if name != v1beta1.SchemaVersion || !served {
			continue
		}
		_, found, err := unstructured.NestedMap(versionMap, "schema", "openAPIV3Schema", "properties", "spec", "properties", "clusterProfileRef")
		return err == nil && found
	}
	return false
}

func getBaseURL(version, base string) (string, error) {
	versionSanitized := strings.ToLower(version)
	URL := "https://github.com/knative/operator/releases/latest/download/" + base
	if version != common.Latest && version != common.Nightly {
		if !strings.HasPrefix(version, "v") {
			versionSanitized = fmt.Sprintf("v%s", versionSanitized)
		}
		validity, major := common.GetMajor(versionSanitized)
		if !validity {
			return "", fmt.Errorf("%v is not a semantic version", version)
		}
		prefix := ""
		if semver.Compare(major, "v0") == 1 {
			prefix = "knative-"
		}
		URL = fmt.Sprintf("https://github.com/knative/operator/releases/download/%s%s/%s", prefix, versionSanitized, base)
	}
	if version == common.Nightly {
		URL = "https://storage.googleapis.com/knative-nightly/operator/latest/" + base
	}
	return URL, nil
}

func getPostInstallURL(version string) (string, error) {
	return getBaseURL(version, "operator-post-install.yaml")
}

func getOperatorURL(version string) (string, error) {
	return getBaseURL(version, "operator.yaml")
}

func getOverlayYamlContent(installFlags *installCmdFlags) string {
	overlayContent := ""
	if strings.EqualFold(installFlags.Component, common.ServingComponent) {
		if installFlags.Istio {
			overlayContent = servingOverlay
			if installFlags.IstioNamespace != common.DefaultIstioNamespace {
				overlayContent = servingIstioOverlay
			}
		} else {
			overlayContent = servingWithIngressOverlay
		}
	} else if strings.EqualFold(installFlags.Component, common.EventingComponent) {
		overlayContent = eventingOverlay
	} else if installFlags.Component == "" {
		overlayContent = operatorOverlay
	}

	if overlayContent == "" {
		return ""
	}
	if installFlags.Component == "" && (strings.EqualFold(installFlags.Version, common.Latest) || strings.EqualFold(installFlags.Version, common.Nightly) || versionWebhook(installFlags.Version)) {
		overlayContent = fmt.Sprintf("%s\n%s", overlayContent, operatorCRDsOverlay)
	}
	if hasClusterProfileFlags(installFlags) && installFlags.Component != "" {
		overlayContent = fmt.Sprintf("%s\n%s", overlayContent, clusterProfileOverlay)
	}

	return overlayContent
}

func versionWebhook(version string) bool {
	targetVersion := version
	if !strings.HasPrefix(version, "v") {
		targetVersion = fmt.Sprintf("v%s", targetVersion)
	}
	semver.MajorMinor(targetVersion)
	return semver.Compare(targetVersion, "v1.3") >= 0
}

func getYamlValuesContent(installFlags *installCmdFlags) string {
	content := ""
	if strings.EqualFold(installFlags.Component, common.ServingComponent) {
		name := installFlags.CRName
		if name == "" {
			name = common.KnativeServingName
		}
		content = fmt.Sprintf("#@data/values\n---\nname: %s\nnamespace: %s\nversion: '%s'",
			name, installFlags.Namespace, installFlags.Version)
		if installFlags.IstioNamespace != common.DefaultIstioNamespace {
			myslice := []string{content, fmt.Sprintf("local_gateway_value: knative-local-gateway.%s.svc.cluster.local", installFlags.IstioNamespace)}
			content = strings.Join(myslice, "\n")
		}
	} else if strings.EqualFold(installFlags.Component, common.EventingComponent) {
		name := installFlags.CRName
		if name == "" {
			name = common.KnativeEventingName
		}
		content = fmt.Sprintf("#@data/values\n---\nname: %s\nnamespace: %s\nversion: '%s'",
			name, installFlags.Namespace, installFlags.Version)
	} else if installFlags.Component == "" {
		content = fmt.Sprintf("#@data/values\n---\nnamespace: %s", installFlags.Namespace)
	}

	if !strings.EqualFold(installFlags.Component, common.ServingComponent) || installFlags.Istio {
		return appendClusterProfileValues(content, installFlags)
	}

	ingressClass := "istio.ingress.networking.knative.dev"
	if installFlags.Kourier {
		ingressClass = "kourier.ingress.networking.knative.dev"
	}

	if installFlags.Contour {
		ingressClass = "contour.ingress.networking.knative.dev"
	}

	content = fmt.Sprintf("%s\nkourier: %t\nistio: %t\ncontour: %t\ningressClass: %s",
		content, installFlags.Kourier, installFlags.Istio, installFlags.Contour, ingressClass)

	return appendClusterProfileValues(content, installFlags)
}

func appendClusterProfileValues(content string, installFlags *installCmdFlags) string {
	if installFlags.Component == "" || !hasClusterProfileFlags(installFlags) {
		return content
	}
	return fmt.Sprintf("%s\ncluster_profile: %s\ncluster_profile_namespace: %s",
		content, installFlags.ClusterProfile, installFlags.ClusterProfileNamespace)
}

func checkIfOperatorInstalled(p *pkg.OperatorParams) (bool, string, string, error) {
	client, err := p.NewKubeClient()
	if err != nil {
		return false, "", "", fmt.Errorf("cannot get source cluster kube config, please use --kubeconfig or export environment variable KUBECONFIG to set\n")
	}
	deploy := common.Deployment{
		Client: client,
	}
	return deploy.CheckIfOperatorInstalled()
}

// ensureOperatorInstalled installs the Knative Operator with default flags if
// it is not already present.
func ensureOperatorInstalled(p *pkg.OperatorParams) error {
	exists, _, _, err := checkIfOperatorInstalled(p)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	operatorInstallFlags := installCmdFlags{
		Namespace: common.DefaultNamespace,
		Version:   common.Latest,
	}
	return installOperator(&operatorInstallFlags, p)
}

// validateRemoteInstallPrereqs verifies that the Knative Operator is installed,
// configured for ClusterProfile-driven multicluster installs, and that the
// component CRD on the hub supports spec.clusterProfileRef.
func validateRemoteInstallPrereqs(component string, p *pkg.OperatorParams) error {
	exists, operatorNamespace, _, err := checkIfOperatorInstalled(p)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("remote component install requires an existing Knative Operator configured with %s", clusterProfileProviderFileFlag)
	}
	if err := validateOperatorMulticlusterEnabled(operatorNamespace, p); err != nil {
		return err
	}
	return validateClusterProfileCRDSupport(component, p)
}

func installKnativeComponent(installFlags *installCmdFlags, p *pkg.OperatorParams) error {
	if err := createNamspaceIfNecessary(installFlags.Namespace, p); err != nil {
		return err
	}

	// Generate the CR template
	yamlTemplateString, err := common.GenerateOperatorCRStringForName(installFlags.Component, installFlags.Namespace, installFlags.CRName, p)
	if err != nil {
		return err
	}

	if err := applyOverlayValuesOnTemplate(yamlTemplateString, installFlags, p); err != nil {
		return err
	}

	if hasClusterProfileFlags(installFlags) {
		return ensureRemoteKnativeComponentReady(installFlags, p)
	}
	return ensureKnativeComponentReady(installFlags, p)
}

func ensureKnativeComponentReady(installFlags *installCmdFlags, p *pkg.OperatorParams) error {
	client, err := p.NewKubeClient()
	if err != nil {
		return fmt.Errorf("cannot get source cluster kube config, please use --kubeconfig or export environment variable KUBECONFIG to set\n")
	}

	operatorClient, err := p.NewOperatorClient()
	if err != nil {
		return fmt.Errorf("cannot get source cluster kube config, please use --kubeconfig or export environment variable KUBECONFIG to set\n")
	}

	if strings.EqualFold(installFlags.Component, common.ServingComponent) {
		err := WaitForKnativeDeploymentState(client, installFlags.Namespace, installFlags.Version, ServingKeyDeployments,
			IsKnativeDeploymentReady)
		if err != nil {
			return err
		}
		_, err = WaitForKnativeServingState(operatorClient.OperatorV1beta1().KnativeServings(installFlags.Namespace), common.KnativeServingName,
			installFlags.Version, IsKnativeServingReady)

		if err != nil {
			return err
		}
	} else if strings.EqualFold(installFlags.Component, common.EventingComponent) {
		err := WaitForKnativeDeploymentState(client, installFlags.Namespace, installFlags.Version, EventingKeyDeployments,
			IsKnativeDeploymentReady)
		if err != nil {
			return err
		}
		_, err = WaitForKnativeEventingState(operatorClient.OperatorV1beta1().KnativeEventings(installFlags.Namespace), common.KnativeEventingName,
			installFlags.Version, IsKnativeEventingReady)

		if err != nil {
			return err
		}
	}

	return nil
}

func ensureRemoteKnativeComponentReady(installFlags *installCmdFlags, p *pkg.OperatorParams) error {
	operatorClient, err := p.NewOperatorClient()
	if err != nil {
		return fmt.Errorf("cannot get source cluster kube config, please use --kubeconfig or export environment variable KUBECONFIG to set")
	}

	if strings.EqualFold(installFlags.Component, common.ServingComponent) {
		_, err = WaitForRemoteKnativeServingState(operatorClient.OperatorV1beta1().KnativeServings(installFlags.Namespace),
			installFlags.CRName, installFlags.Version, IsRemoteKnativeServingReady)
		return err
	}
	if strings.EqualFold(installFlags.Component, common.EventingComponent) {
		_, err = WaitForRemoteKnativeEventingState(operatorClient.OperatorV1beta1().KnativeEventings(installFlags.Namespace),
			installFlags.CRName, installFlags.Version, IsRemoteKnativeEventingReady)
		return err
	}
	return nil
}

func installOperator(installFlags *installCmdFlags, p *pkg.OperatorParams) error {
	err := createNamspaceIfNecessary(installFlags.Namespace, p)
	if err != nil {
		return err
	}

	URL, err := getOperatorURL(installFlags.Version)
	if err != nil {
		return err
	}

	postInstallURL, err := getPostInstallURL(installFlags.Version)
	if err != nil {
		return err
	}

	// Generate the CR template by downloading the operator yaml
	yamlTemplateString, err := common.DownloadFile(URL)
	if err != nil {
		return err
	}

	yamlTemplateStringPostInstall, err := common.DownloadFile(postInstallURL)
	if err == nil && yamlTemplateStringPostInstall != "" {
		// If operator-post-install.yaml exists, append the content to the template content
		yamlTemplateString = fmt.Sprintf("%s\n%s", yamlTemplateString, yamlTemplateStringPostInstall)
	}

	return applyOverlayValuesOnTemplate(yamlTemplateString, installFlags, p)
}

func createNamspaceIfNecessary(namespace string, p *pkg.OperatorParams) error {
	client, err := p.NewKubeClient()
	if err != nil {
		return fmt.Errorf("cannot get source cluster kube config, please use --kubeconfig or export environment variable KUBECONFIG to set\n")
	}

	ns := common.Namespace{
		Client:    client,
		Component: namespace,
	}
	if err = ns.CreateNamespace(namespace); err != nil {
		return err
	}
	return nil
}

func applyOverlayValuesOnTemplate(yamlTemplateString string, installFlags *installCmdFlags, p *pkg.OperatorParams) error {
	overlayContent := getOverlayYamlContent(installFlags)
	yamlValuesContent := getYamlValuesContent(installFlags)

	if err := common.ApplyManifests(yamlTemplateString, overlayContent, yamlValuesContent, p); err != nil {
		return err
	}
	return nil
}

func generateVersionStages(source, target string) ([]string, error) {
	stringArray := ""

	if strings.HasPrefix(source, "v") {
		source = source[1:]
	}
	if strings.HasPrefix(target, "v") {
		target = target[1:]
	}
	if source == "" {
		stringArray = target
		return strings.Split(stringArray, ","), nil
	}

	targetVersion := target
	if targetVersion == common.Latest || targetVersion == common.Nightly {
		targetVersion = common.LatestVersion
	}

	if !strings.HasPrefix(targetVersion, "v") {
		targetVersion = fmt.Sprintf("v%s", targetVersion)
	}

	sourceVersion := source
	if !strings.HasPrefix(sourceVersion, "v") {
		sourceVersion = fmt.Sprintf("v%s", sourceVersion)
	}

	sourceMajor := strings.Split(sourceVersion, ".")[0][1:]
	targetMajor := strings.Split(targetVersion, ".")[0][1:]

	if targetMajor != sourceMajor {
		return nil, fmt.Errorf("Unable to migrate from the source version %s to the target version %s", sourceVersion,
			targetVersion)
	}
	sourceMinor, err := strconv.Atoi(strings.Split(sourceVersion, ".")[1])
	if err != nil {
		return nil, fmt.Errorf("minor number of the current version %v should be an integer", sourceVersion)
	}
	targetMinor, err := strconv.Atoi(strings.Split(targetVersion, ".")[1])
	if err != nil {
		return nil, fmt.Errorf("minor number of the target version %v should be an integer", targetVersion)
	}
	if math.Abs(float64(targetMinor-sourceMinor)) < 2 {
		stringArray = target
		return strings.Split(stringArray, ","), nil
	}

	if targetMinor > sourceMinor {
		minorStart := sourceMinor + 1
		stringArray = targetMajor + "." + fmt.Sprintf("%d", minorStart) + ".0"
		for i := 1; i < targetMinor-sourceMinor; i++ {
			minor := sourceMinor + i + 1
			if minor != targetMinor {
				stringArray = stringArray + "," + targetMajor + "." + fmt.Sprintf("%d", minor) + ".0"
			} else {
				if target == common.Latest || target == common.Nightly {
					stringArray = stringArray + "," + targetMajor + "." + fmt.Sprintf("%d", minor) + ".0"
				}
				stringArray = stringArray + "," + target
			}
		}
		return strings.Split(stringArray, ","), nil

	} else {
		minorStart := sourceMinor - 1
		stringArray = targetMajor + "." + fmt.Sprintf("%d", minorStart) + ".0"
		for i := 1; i < sourceMinor-targetMinor; i++ {
			minor := sourceMinor - i - 1
			if minor != targetMinor {
				stringArray = stringArray + "," + targetMajor + "." + fmt.Sprintf("%d", minor) + ".0"
			} else {
				stringArray = stringArray + "," + target
			}
		}
		return strings.Split(stringArray, ","), nil
	}
}

// WaitForKnativeDeploymentState polls the status of the Knative deployments every `interval`
// until `inState` returns `true` indicating the deployments match the desired deployments.
func WaitForKnativeDeploymentState(client kubernetes.Interface, namespace string, version string, expectedDeployments []string,
	inState func(deps *v1.DeploymentList, expectedDeployments []string, version string, err error) (bool, error)) error {
	span := logging.GetEmitableSpan(context.Background(), fmt.Sprintf("WaitForKnativeDeploymentState/%s/%s", expectedDeployments, "KnativeDeploymentIsReady"))
	defer span.End()

	waitErr := wait.PollImmediate(Interval, Timeout, func() (bool, error) {
		dpList, err := client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
		return inState(dpList, expectedDeployments, version, err)
	})

	return waitErr
}

// IsKnativeDeploymentReady will check the status conditions of the deployments and return true if the deployments meet the desired status.
func IsKnativeDeploymentReady(dpList *v1.DeploymentList, expectedDeployments []string, version string, err error) (bool, error) {
	if err != nil {
		return false, err
	}

	findDeployment := func(name string, deployments []v1.Deployment) *v1.Deployment {
		for _, deployment := range deployments {
			if deployment.Name == name {
				return &deployment
			}
		}
		return nil
	}

	isStatusReady := func(status v1.DeploymentStatus) bool {
		for _, c := range status.Conditions {
			if c.Type == v1.DeploymentAvailable && c.Status == corev1.ConditionTrue {
				return true
			}
		}
		return false
	}

	isReady := func(d *v1.Deployment) bool {
		for key, val := range d.GetObjectMeta().GetLabels() {
			// Check if the version matches. As long as we find a value equals to the version, we can determine
			// the deployment is for the specific version. The key "networking.knative.dev/ingress-provider" is
			// used to indicate the network ingress resource.
			// Currently, the network ingress resource is still specified together with the knative serving.
			// It is possible that network ingress resource is not using the same version as knative serving.
			// This is the reason why we skip the version checking for network ingress resource.

			// The parameter version means the target version of Knative component to be installed.
			// The parameter existingVersion means the installed version of Knative component. It is set to empty, if
			// there is no Knative installation.

			// If the deployment resource is for ingress, we will check the status of the deployment.
			if key == "networking.knative.dev/ingress-provider" {
				return isStatusReady(d.Status)
			}

			if key == "app.kubernetes.io/version" || key == "serving.knative.dev/release" || key == "eventing.knative.dev/release" {
				if val == fmt.Sprintf("v%s", version) || val == version {
					// When on of the following conditions is met:
					// * spec.version is set to latest, but operator returns an actual semantic version
					// * spec.version is set to a valid semantic version
					// we need to verify the value of the key serving.knative.dev/release or eventing.knative.dev/release
					// matches the version.
					return isStatusReady(d.Status)
				}
			}

			// If spec.version is set to latest and operator bundles a directory called latest, it is possible that both
			// the version and the existing version are latest. In this case, the knative component to be installed is the
			// same as the existing one, and we will check the status of the deployment.
			if version == common.Latest {
				return isStatusReady(d.Status)
			}
		}
		return false
	}

	for _, name := range expectedDeployments {
		dep := findDeployment(name, dpList.Items)
		if dep == nil {
			return false, nil
		}
		if !isReady(dep) {
			return false, nil
		}
	}

	return true, nil
}

// WaitForKnativeServingState polls the status of the KnativeServing called name
// from client every `interval` until `inState` returns `true` indicating it
// is done, returns an error or timeout.
func WaitForKnativeServingState(clients operatorv1beta1.KnativeServingInterface, name string, version string,
	inState func(s *v1beta1.KnativeServing, version string, err error) (bool, error)) (*v1beta1.KnativeServing, error) {
	span := logging.GetEmitableSpan(context.Background(), fmt.Sprintf("WaitForKnativeServingState/%s/%s", name, "KnativeServingIsReady"))
	defer span.End()

	var lastState *v1beta1.KnativeServing
	waitErr := wait.PollImmediate(Interval, Timeout, func() (bool, error) {
		state, err := clients.Get(context.TODO(), name, metav1.GetOptions{})
		lastState = state
		return inState(lastState, version, err)
	})

	if waitErr != nil {
		return lastState, fmt.Errorf("knativeserving %s is not in desired state, got: %+v: %w", name, lastState, waitErr)
	}
	return lastState, nil
}

// IsKnativeServingReady will check the status conditions of the KnativeServing and return true if the KnativeServing is ready.
func IsKnativeServingReady(s *v1beta1.KnativeServing, version string, err error) (bool, error) {
	if version == common.Latest || version == common.Nightly {
		return s.Status.IsReady(), err
	}
	return s.Status.IsReady() && version == s.Status.Version, err
}

func WaitForRemoteKnativeServingState(clients operatorv1beta1.KnativeServingInterface, name string, version string,
	inState func(s *v1beta1.KnativeServing, version string, err error) (bool, error)) (*v1beta1.KnativeServing, error) {
	span := logging.GetEmitableSpan(context.Background(), fmt.Sprintf("WaitForRemoteKnativeServingState/%s/%s", name, "KnativeServingIsReady"))
	defer span.End()

	var lastState *v1beta1.KnativeServing
	waitErr := wait.PollImmediate(Interval, Timeout, func() (bool, error) {
		state, err := clients.Get(context.TODO(), name, metav1.GetOptions{})
		lastState = state
		return inState(lastState, version, err)
	})

	if waitErr != nil {
		return lastState, fmt.Errorf("knativeserving %s is not in desired remote state: %s: %w", name, servingRemoteDiagnostics(lastState), waitErr)
	}
	return lastState, nil
}

func IsRemoteKnativeServingReady(s *v1beta1.KnativeServing, version string, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	if s == nil {
		return false, nil
	}
	if targetCluster := s.Status.GetCondition(base.TargetClusterResolved); targetCluster != nil && targetCluster.IsFalse() {
		return false, nil
	}
	return IsKnativeServingReady(s, version, nil)
}

// WaitForKnativeEventingState polls the status of the KnativeEventing called name
// from client every `interval` until `inState` returns `true` indicating it
// is done, returns an error or timeout.
func WaitForKnativeEventingState(clients operatorv1beta1.KnativeEventingInterface, name string, version string,
	inState func(s *v1beta1.KnativeEventing, version string, err error) (bool, error)) (*v1beta1.KnativeEventing, error) {
	span := logging.GetEmitableSpan(context.Background(), fmt.Sprintf("WaitForKnativeEventingState/%s/%s", name, "KnativeEventingIsReady"))
	defer span.End()

	var lastState *v1beta1.KnativeEventing
	waitErr := wait.PollImmediate(Interval, Timeout, func() (bool, error) {
		state, err := clients.Get(context.TODO(), name, metav1.GetOptions{})
		lastState = state
		return inState(lastState, version, err)
	})

	if waitErr != nil {
		return lastState, fmt.Errorf("KnativeEventing %s is not in desired state, got: %+v: %w", name, lastState, waitErr)
	}
	return lastState, nil
}

// IsKnativeEventingReady will check the status conditions of the KnativeEventing and return true if the KnativeEventing is ready.
func IsKnativeEventingReady(s *v1beta1.KnativeEventing, version string, err error) (bool, error) {
	if version == common.Latest || version == common.Nightly {
		return s.Status.IsReady(), err
	}
	return s.Status.IsReady() && version == s.Status.Version, err
}

func WaitForRemoteKnativeEventingState(clients operatorv1beta1.KnativeEventingInterface, name string, version string,
	inState func(s *v1beta1.KnativeEventing, version string, err error) (bool, error)) (*v1beta1.KnativeEventing, error) {
	span := logging.GetEmitableSpan(context.Background(), fmt.Sprintf("WaitForRemoteKnativeEventingState/%s/%s", name, "KnativeEventingIsReady"))
	defer span.End()

	var lastState *v1beta1.KnativeEventing
	waitErr := wait.PollImmediate(Interval, Timeout, func() (bool, error) {
		state, err := clients.Get(context.TODO(), name, metav1.GetOptions{})
		lastState = state
		return inState(lastState, version, err)
	})

	if waitErr != nil {
		return lastState, fmt.Errorf("knativeeventing %s is not in desired remote state: %s: %w", name, eventingRemoteDiagnostics(lastState), waitErr)
	}
	return lastState, nil
}

func IsRemoteKnativeEventingReady(s *v1beta1.KnativeEventing, version string, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	if s == nil {
		return false, nil
	}
	if targetCluster := s.Status.GetCondition(base.TargetClusterResolved); targetCluster != nil && targetCluster.IsFalse() {
		return false, nil
	}
	return IsKnativeEventingReady(s, version, nil)
}

func servingRemoteDiagnostics(s *v1beta1.KnativeServing) string {
	if s == nil {
		return "latest status is unavailable"
	}
	return remoteConditionDiagnostics(func(t apis.ConditionType) *apis.Condition {
		return s.Status.GetCondition(t)
	})
}

func eventingRemoteDiagnostics(s *v1beta1.KnativeEventing) string {
	if s == nil {
		return "latest status is unavailable"
	}
	return remoteConditionDiagnostics(func(t apis.ConditionType) *apis.Condition {
		return s.Status.GetCondition(t)
	})
}

func remoteConditionDiagnostics(getCondition func(apis.ConditionType) *apis.Condition) string {
	conditionTypes := []apis.ConditionType{
		base.TargetClusterResolved,
		base.InstallSucceeded,
		base.DeploymentsAvailable,
		base.VersionMigrationEligible,
	}
	parts := make([]string, 0, len(conditionTypes))
	for _, conditionType := range conditionTypes {
		parts = append(parts, fmt.Sprintf("%s=%s", conditionType, conditionSummary(getCondition(conditionType))))
	}
	return strings.Join(parts, "; ")
}

func conditionSummary(condition *apis.Condition) string {
	if condition == nil {
		return "Unknown"
	}
	parts := []string{string(condition.Status)}
	if condition.Reason != "" {
		parts = append(parts, fmt.Sprintf("reason=%s", condition.Reason))
	}
	if condition.Message != "" {
		parts = append(parts, fmt.Sprintf("message=%q", condition.Message))
	}
	return strings.Join(parts, " ")
}
