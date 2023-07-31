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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345

	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/kn-plugin-operator/pkg/command/common"
	"knative.dev/operator/pkg/apis/operator/v1beta1"
	operatorv1beta1 "knative.dev/operator/pkg/client/clientset/versioned/typed/operator/v1beta1"
	"knative.dev/pkg/test/logging"
)

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

type installCmdFlags struct {
	Component      string
	IstioNamespace string
	Namespace      string
	KubeConfig     string
	Version        string
	Istio          bool
	Kourier        bool
	Contour        bool
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

			fmt.Fprintf(cmd.OutOrStdout(), "Knative %s of the '%s' version was created in the namespace '%s'.\n",
				component, installFlags.Version, installFlags.Namespace)
			return nil
		},
	}

	installCmd.Flags().StringVar(&installFlags.KubeConfig, "kubeconfig", "", "The kubeconfig of the Knative resources (default is KUBECONFIG from environment variable)")
	installCmd.Flags().StringVarP(&installFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")
	installCmd.Flags().StringVarP(&installFlags.Component, "component", "c", "", "The name of the Knative Component to install")
	installCmd.Flags().StringVarP(&installFlags.Version, "version", "v", common.Latest, "The version of the the Knative Operator or the Knative component")
	installCmd.Flags().StringVar(&installFlags.IstioNamespace, "istio-namespace", "", "The namespace of istio")
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

		currentVersion := ""
		if exists, ns, version, err := deploy.CheckIfKnativeInstalled(installFlags.Component); err != nil {
			return err
		} else if exists {
			// Check if the namespace is consistent
			if !strings.EqualFold(ns, installFlags.Namespace) {
				return fmt.Errorf("The namespace %s you specified is not consistent with the existing namespace for Knative Component %s",
					installFlags.Namespace, ns)
			}
			currentVersion = version
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
		content = fmt.Sprintf("#@data/values\n---\nname: %s\nnamespace: %s\nversion: '%s'",
			common.DefaultKnativeServingNamespace, installFlags.Namespace, installFlags.Version)
		if installFlags.IstioNamespace != common.DefaultIstioNamespace {
			myslice := []string{content, fmt.Sprintf("local_gateway_value: knative-local-gateway.%s.svc.cluster.local", installFlags.IstioNamespace)}
			content = strings.Join(myslice, "\n")
		}
	} else if strings.EqualFold(installFlags.Component, common.EventingComponent) {
		content = fmt.Sprintf("#@data/values\n---\nname: %s\nnamespace: %s\nversion: '%s'",
			common.DefaultKnativeEventingNamespace, installFlags.Namespace, installFlags.Version)
	} else if installFlags.Component == "" {
		content = fmt.Sprintf("#@data/values\n---\nnamespace: %s", installFlags.Namespace)
	}

	if !strings.EqualFold(installFlags.Component, common.ServingComponent) || installFlags.Istio {
		return content
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

	return content
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

func installKnativeComponent(installFlags *installCmdFlags, p *pkg.OperatorParams) error {
	// Check if the knative operator is installed
	if exists, _, _, err := checkIfOperatorInstalled(p); err != nil {
		return err
	} else if !exists {
		operatorInstallFlags := installCmdFlags{
			Namespace: "default",
			Version:   common.Latest,
		}
		err = installOperator(&operatorInstallFlags, p)
		if err != nil {
			return err
		}
	}

	err := createNamspaceIfNecessary(installFlags.Namespace, p)
	if err != nil {
		return err
	}

	// Generate the CR template
	yamlTemplateString, err := common.GenerateOperatorCRString(installFlags.Component, installFlags.Namespace, p)
	if err != nil {
		return err
	}

	err = applyOverlayValuesOnTemplate(yamlTemplateString, installFlags, p)
	if err != nil {
		return err
	}

	// Make sure all the deployment resources are up and running
	err = ensureKnativeComponentReady(installFlags, p)
	if err != nil {
		return err
	}

	return nil
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
