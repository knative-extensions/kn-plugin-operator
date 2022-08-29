// Copyright 2022 The Knative Authors
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

package configure

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345

	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/kn-plugin-operator/pkg/command/common"
)

//go:embed overlay/ks_custom_manifests.yaml
var servingManifestsOverlay string

//go:embed overlay/ke_custom_manifests.yaml
var eventingManifestsOverlay string

type manifestsFlags struct {
	File              string
	OperatorNamespace string
	Namespace         string
	Component         string
	Overwrite         bool
	Accessible        bool
}

var manifestsCMDFlags manifestsFlags

// newManifestsCommand represents the configure commands to configure the additional manifests
func newManifestsCommand(p *pkg.OperatorParams) *cobra.Command {
	var configureManifestsCmd = &cobra.Command{
		Use:   "manifests",
		Short: "Configure the custom manifests for Knative",
		Example: `
  # Configure the custom manifests for Knative
  kn operation configure manifests --component eventing --namespace knative-eventing --operatorNamespace default --file filePath`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateManifestsFlags(manifestsCMDFlags); err != nil {
				return err
			}

			err := configureManifests(manifestsCMDFlags, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified custom manifests has been configured.\n")
			return nil
		},
	}

	configureManifestsCmd.Flags().BoolVar(&manifestsCMDFlags.Overwrite, "overwrite", false, "The flag to specify the mode of the custom manifests")
	configureManifestsCmd.Flags().StringVar(&manifestsCMDFlags.File, "file", "", "The path to the local file with the custom manifests")
	configureManifestsCmd.Flags().StringVar(&manifestsCMDFlags.OperatorNamespace, "operatorNamespace", "default", "The flag to specify the configmap name")
	configureManifestsCmd.Flags().StringVarP(&manifestsCMDFlags.Component, "component", "c", "", "The flag to specify the component name")
	configureManifestsCmd.Flags().StringVarP(&manifestsCMDFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")
	configureManifestsCmd.Flags().BoolVar(&manifestsCMDFlags.Accessible, "accessible", false, "The flag to indicate wehther the link is accessible by Knative in the Kubernetes cluster")

	return configureManifestsCmd
}

func validateManifestsFlags(manifestsCMDFlags manifestsFlags) error {
	if manifestsCMDFlags.File == "" {
		return fmt.Errorf("You need to specify the local path of the file containing the custom manifests.")
	}
	if manifestsCMDFlags.Namespace == "" {
		return fmt.Errorf("You need to specify the namespace for the Knative component.")
	}
	if manifestsCMDFlags.Component != "" && !strings.EqualFold(manifestsCMDFlags.Component, common.ServingComponent) && !strings.EqualFold(manifestsCMDFlags.Component, common.EventingComponent) {
		return fmt.Errorf("You need to specify the component for Knative: serving or eventing.")
	}
	return nil
}

func UpdateOperatorForCustomManifests(manifestsCMDFlags manifestsFlags, p *pkg.OperatorParams) error {
	kubeClient, err := p.NewKubeClient()
	if err != nil {
		return fmt.Errorf("cannot get source cluster kube config, please use --kubeconfig or export environment variable KUBECONFIG to set\n")
	}

	kubeResource := common.KubeResource{
		KubeClient: kubeClient,
	}

	data, err := common.ReadFile(manifestsCMDFlags.File)
	if err != nil {
		return err
	}

	if err = kubeResource.CreateOrUpdateConfigMap(common.ConfigMapName, manifestsCMDFlags.Namespace, data, manifestsCMDFlags.Overwrite); err != nil {
		return err
	}

	if err = kubeResource.UpdateOperatorDeployment(common.KnativeOperatorName, manifestsCMDFlags.OperatorNamespace); err != nil {
		return err
	}

	return nil
}

func configureManifests(manifestsCMDFlags manifestsFlags, p *pkg.OperatorParams) error {
	if !manifestsCMDFlags.Accessible {
		if err := UpdateOperatorForCustomManifests(manifestsCMDFlags, p); err != nil {
			return err
		}
	}

	// Update the custom resource
	component := common.ServingComponent
	if strings.EqualFold(manifestsCMDFlags.Component, common.EventingComponent) {
		component = common.EventingComponent
	}
	yamlTemplateString, err := common.GenerateOperatorCRString(component, manifestsCMDFlags.Namespace, p)
	if err != nil {
		return err
	}

	overlayContent := getOverlayYamlContentManifest(manifestsCMDFlags)
	valuesYaml := getYamlValuesContentManifests(manifestsCMDFlags)
	if err = common.ApplyManifests(yamlTemplateString, overlayContent, valuesYaml, p); err != nil {
		return err
	}
	return nil
}

func getOverlayYamlContentManifest(manifestsCMDFlags manifestsFlags) string {
	baseOverlayContent := servingManifestsOverlay
	if strings.EqualFold(manifestsCMDFlags.Component, common.EventingComponent) {
		baseOverlayContent = eventingManifestsOverlay
	}
	resourceContent := getManifestsConfiguration(manifestsCMDFlags)
	baseOverlayContent = fmt.Sprintf("%s\n%s", baseOverlayContent, resourceContent)
	return baseOverlayContent
}

func getManifestsConfiguration(manifestsCMDFlags manifestsFlags) string {
	resourceArray := []string{}
	tag := fmt.Sprintf("%s%s", common.Spaces(2), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)
	if manifestsCMDFlags.Overwrite {
		tag = fmt.Sprintf("%s%s", common.Spaces(2), common.YttReplaceTag)
		resourceArray = append(resourceArray, tag)
	}

	field := fmt.Sprintf("%s%s:", common.Spaces(2), "additionalManifests")
	resourceArray = append(resourceArray, field)

	tag = fmt.Sprintf("%s%s", common.Spaces(2), common.FieldByName("URL"))
	resourceArray = append(resourceArray, tag)

	field = fmt.Sprintf("%s- %s: %s", common.Spaces(2), "URL", "#@ data.values.manifestsPath")
	resourceArray = append(resourceArray, field)

	return strings.Join(resourceArray, "\n")
}

func getYamlValuesContentManifests(manifestsCMDFlags manifestsFlags) string {
	contentArray := []string{}
	header := "#@data/values\n---"
	contentArray = append(contentArray, header)
	namespace := fmt.Sprintf("namespace: %s", manifestsCMDFlags.Namespace)
	contentArray = append(contentArray, namespace)
	deployName := fmt.Sprintf("manifestsPath: %s", common.MountPath)
	if manifestsCMDFlags.Accessible {
		deployName = fmt.Sprintf("manifestsPath: %s", manifestsCMDFlags.File)
	}
	contentArray = append(contentArray, deployName)
	return strings.Join(contentArray, "\n")
}
