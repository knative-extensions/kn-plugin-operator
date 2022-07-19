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
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/kn-plugin-operator/pkg/command/common"
)

type installCmdFlags struct {
	Component      string
	IstioNamespace string
	Namespace      string
	KubeConfig     string
	Version        string
}

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

	return installCmd
}

func RunInstallationCommand(installFlags *installCmdFlags, p *pkg.OperatorParams) error {
	// Fill in the default values for the empty fields
	installFlags.fill_defaults()
	p.KubeCfgPath = installFlags.KubeConfig

	rootPath, err := os.Getwd()
	if err != nil {
		return err
	}

	if installFlags.Component != "" {
		currentVersion := ""
		if exists, ns, version, err := checkIfKnativeInstalled(p, installFlags.Component); err != nil {
			return err
		} else if exists {
			// Check if the namespace is consistent
			if strings.EqualFold(ns, installFlags.Namespace) {
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
			installFlags.Version = v
			err = installKnativeComponent(installFlags, rootPath, p)
			if err != nil {
				return err
			}
		}

	} else {
		if exists, ns, _, err := checkIfOperatorInstalled(p); err != nil {
			return err
		} else if exists {
			// Check if the namespace is consistent
			if strings.EqualFold(ns, installFlags.Namespace) {
				return fmt.Errorf("The namespace %s you specified is not consistent with the existing namespace for Knative Operator %s",
					installFlags.Namespace, ns)
			}
		}

		// Install the Knative Operator
		err = installOperator(installFlags, rootPath, p)
		if err != nil {
			return err
		}
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

func getOverlayYamlContent(installFlags *installCmdFlags, rootPath string) string {
	path := ""
	if strings.EqualFold(installFlags.Component, common.ServingComponent) {
		path = rootPath + "/overlay/ks.yaml"
		if installFlags.IstioNamespace != common.DefaultIstioNamespace {
			path = rootPath + "/overlay/ks_istio_ns.yaml"
		}
	} else if strings.EqualFold(installFlags.Component, common.EventingComponent) {
		path = rootPath + "/overlay/ke.yaml"
	} else if installFlags.Component == "" {
		path = rootPath + "/overlay/operator.yaml"
	}

	if path == "" {
		return ""
	}
	overlayContent, _ := common.ReadFile(path)
	if installFlags.Component == "" && (strings.EqualFold(installFlags.Version, common.Latest) || strings.EqualFold(installFlags.Version, common.Nightly) || versionWebhook(installFlags.Version)) {
		crdOverlay, _ := common.ReadFile(rootPath + "/overlay/operator_crds.yaml")
		overlayContent = fmt.Sprintf("%s\n%s", overlayContent, crdOverlay)
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

	return content
}

func checkIfKnativeInstalled(p *pkg.OperatorParams, component string) (bool, string, string, error) {
	client, err := p.NewKubeClient()
	if err != nil {
		return false, "", "", fmt.Errorf("cannot get source cluster kube config, please use --kubeconfig or export environment variable KUBECONFIG to set\n")
	}
	deploy := common.Deployment{
		Client: client,
	}

	return deploy.CheckIfKnativeInstalled(component)
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

func installKnativeComponent(installFlags *installCmdFlags, rootPath string, p *pkg.OperatorParams) error {
	// Check if the knative operator is installed
	if exists, _, _, err := checkIfOperatorInstalled(p); err != nil {
		return err
	} else if !exists {
		operatorInstallFlags := installCmdFlags{
			Namespace: "default",
			Version:   common.Latest,
		}
		err = installOperator(&operatorInstallFlags, rootPath, p)
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

	err = applyOverlayValuesOnTemplate(yamlTemplateString, installFlags, rootPath, p)
	if err != nil {
		return err
	}

	// Make sure all the deployment resources are up and running

	return nil
}

func installOperator(installFlags *installCmdFlags, rootPath string, p *pkg.OperatorParams) error {
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

	return applyOverlayValuesOnTemplate(yamlTemplateString, installFlags, rootPath, p)
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

func applyOverlayValuesOnTemplate(yamlTemplateString string, installFlags *installCmdFlags, rootPath string, p *pkg.OperatorParams) error {
	overlayContent := getOverlayYamlContent(installFlags, rootPath)
	yamlValuesContent := getYamlValuesContent(installFlags)

	if err := common.ApplyManifests(yamlTemplateString, overlayContent, yamlValuesContent, p); err != nil {
		return err
	}
	return nil
}

func generateVersionStages(source, target string) ([]string, error) {
	stringArray := ""
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

	sourceMajor := strings.Split(sourceVersion, ".")[0]
	targetMajor := strings.Split(targetVersion, ".")[0]
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
		stringArray = targetVersion
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
				stringArray = stringArray + "," + targetVersion
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
				stringArray = stringArray + "," + targetVersion
			}
		}
		return strings.Split(stringArray, ","), nil
	}
}
