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

//go:embed overlay/ks_cm_base.yaml
var servingCMOverlay string

//go:embed overlay/ke_cm_base.yaml
var eventingCMOverlay string

var cmsCMDFlags common.CMsFlags

// newConfigmapsCommand represents the configure commands to update the ConfigMaps in Knative Serving or Eventing
func newConfigmapsCommand(p *pkg.OperatorParams) *cobra.Command {
	var configureCMsCmd = &cobra.Command{
		Use:   "configmaps",
		Short: "Configure the configmap for Knative Serving and Eventing deployments",
		Example: `
  # Configure the CM for Knative Serving and Eventing
  kn operation configure configmaps --component eventing --cmName eventing-controller --key key --value value --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateCMsFlags(cmsCMDFlags); err != nil {
				return err
			}

			err := configureCMs(cmsCMDFlags, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified ConfigMap has been configured in the namespace '%s'.\n",
				cmsCMDFlags.Namespace)
			return nil
		},
	}

	configureCMsCmd.Flags().StringVar(&cmsCMDFlags.Key, "key", "", "The key of the data in the configmap")
	configureCMsCmd.Flags().StringVar(&cmsCMDFlags.Value, "value", "", "The value of the data in the configmap")
	configureCMsCmd.Flags().StringVar(&cmsCMDFlags.CMName, "cmName", "", "The flag to specify the configmap name")
	configureCMsCmd.Flags().StringVarP(&cmsCMDFlags.Component, "component", "c", "", "The flag to specify the component name")
	configureCMsCmd.Flags().StringVarP(&cmsCMDFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return configureCMsCmd
}

func validateCMsFlags(cmsCMDFlags common.CMsFlags) error {
	if cmsCMDFlags.Key == "" {
		return fmt.Errorf("You need to specify the key in the ConfigMap data.")
	}
	if cmsCMDFlags.Value == "" {
		return fmt.Errorf("You need to specify the value in the ConfigMap data.")
	}
	if cmsCMDFlags.CMName == "" {
		return fmt.Errorf("You need to specify the name of the ConfigMap.")
	}
	if cmsCMDFlags.Namespace == "" {
		return fmt.Errorf("You need to specify the namespace.")
	}
	if cmsCMDFlags.Component != "" && !strings.EqualFold(cmsCMDFlags.Component, common.ServingComponent) && !strings.EqualFold(cmsCMDFlags.Component, common.EventingComponent) {
		return fmt.Errorf("You need to specify the component for Knative: serving or eventing.")
	}
	return nil
}

func configureCMs(cmsCMDFlags common.CMsFlags, p *pkg.OperatorParams) error {
	component := common.ServingComponent
	if strings.EqualFold(cmsCMDFlags.Component, common.EventingComponent) {
		component = common.EventingComponent
	}
	yamlTemplateString, err := common.GenerateOperatorCRString(component, cmsCMDFlags.Namespace, p)
	if err != nil {
		return err
	}

	overlayContent := getOverlayYamlContentCM(cmsCMDFlags)
	valuesYaml := getYamlValuesContentCMs(cmsCMDFlags)
	if err := common.ApplyManifests(yamlTemplateString, overlayContent, valuesYaml, p); err != nil {
		return err
	}
	return nil
}

func getOverlayYamlContentCM(cmsCMDFlags common.CMsFlags) string {
	baseOverlayContent := servingCMOverlay
	if strings.EqualFold(cmsCMDFlags.Component, common.EventingComponent) {
		baseOverlayContent = eventingCMOverlay
	}
	resourceContent := getCMConfiguration(cmsCMDFlags)
	baseOverlayContent = fmt.Sprintf("%s\n%s", baseOverlayContent, resourceContent)
	return baseOverlayContent
}

func getCMConfiguration(cmsCMDFlags common.CMsFlags) string {
	resourceArray := []string{}
	tag := fmt.Sprintf("%s%s", common.Spaces(4), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)

	cmName := fmt.Sprintf("%s%s:", common.Spaces(4), cmsCMDFlags.CMName)
	resourceArray = append(resourceArray, cmName)

	tag = fmt.Sprintf("%s%s", common.Spaces(6), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)

	keyValueField := fmt.Sprintf("%s%s: %s", common.Spaces(6), cmsCMDFlags.Key, "#@ data.values.value")
	resourceArray = append(resourceArray, keyValueField)

	return strings.Join(resourceArray, "\n")
}

func getYamlValuesContentCMs(cmsCMDFlags common.CMsFlags) string {
	contentArray := []string{}
	header := "#@data/values\n---"
	contentArray = append(contentArray, header)
	namespace := fmt.Sprintf("namespace: %s", cmsCMDFlags.Namespace)
	contentArray = append(contentArray, namespace)
	value := fmt.Sprintf("value: %s", cmsCMDFlags.Value)
	contentArray = append(contentArray, value)
	return strings.Join(contentArray, "\n")
}
