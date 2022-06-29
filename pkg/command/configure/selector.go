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
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/kn-plugin-operator/pkg/command/common"
)

var selectorCMDFlags common.KeyValueFlags

// newSelectorCommand represents the configure commands to configure the nodeSelector for Knative service
func newSelectorCommand(p *pkg.OperatorParams) *cobra.Command {
	var configureNodeSelectorsCmd = &cobra.Command{
		Use:   "selectors",
		Short: "Configure the selectors for Knative Serving and Eventing services",
		Example: `
  # Configure the selectors for Knative Serving and Eventing services
  kn operation selectors --component eventing --serviceName eventing-controller --key key --value value --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateSelectorFlags(nodeSelectorCMDFlags); err != nil {
				return err
			}

			rootPath, err := os.Getwd()
			if err != nil {
				return err
			}

			err = configureSelectors(nodeSelectorCMDFlags, rootPath, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified annotation has been configured for the deployment %s in the deployment '%s'.\n",
				nodeSelectorCMDFlags.DeployName, nodeSelectorCMDFlags.Namespace)
			return nil
		},
	}

	configureNodeSelectorsCmd.Flags().StringVar(&nodeSelectorCMDFlags.Key, "key", "", "The key of the data in the configmap")
	configureNodeSelectorsCmd.Flags().StringVar(&nodeSelectorCMDFlags.Value, "value", "", "The value of the data in the configmap")
	configureNodeSelectorsCmd.Flags().StringVar(&nodeSelectorCMDFlags.ServiceName, "serviceName", "", "The flag to specify the service name")
	configureNodeSelectorsCmd.Flags().StringVarP(&nodeSelectorCMDFlags.Component, "component", "c", "", "The flag to specify the component name")
	configureNodeSelectorsCmd.Flags().StringVarP(&nodeSelectorCMDFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return configureNodeSelectorsCmd
}

func validateSelectorFlags(keyValuesCMDFlags common.KeyValueFlags) error {
	if err := validateKeyValuePairs(keyValuesCMDFlags); err != nil {
		return err
	}
	if keyValuesCMDFlags.ServiceName == "" {
		return fmt.Errorf("You need to specify the name of the service.")
	}
	return nil
}

func configureSelectors(nodeSelectorCMDFlags common.KeyValueFlags, rootPath string, p *pkg.OperatorParams) error {
	component := common.ServingComponent
	if strings.EqualFold(nodeSelectorCMDFlags.Component, common.EventingComponent) {
		component = common.EventingComponent
	}
	yamlTemplateString, err := common.GenerateOperatorCRString(component, nodeSelectorCMDFlags.Namespace, p)
	if err != nil {
		return err
	}

	overlayContent := getOverlayYamlContentSelector(rootPath, nodeSelectorCMDFlags)
	valuesYaml := getYamlValuesContent(nodeSelectorCMDFlags)
	if err := common.ApplyManifests(yamlTemplateString, overlayContent, valuesYaml, p); err != nil {
		return err
	}
	return nil
}

func getOverlayYamlContentSelector(rootPath string, nodeSelectorCMDFlags common.KeyValueFlags) string {
	path := rootPath + "/overlay/ks_deploy_label.yaml"
	if strings.EqualFold(nodeSelectorCMDFlags.Component, common.EventingComponent) {
		path = rootPath + "/overlay/ke_deploy_label.yaml"
	}
	baseOverlayContent, _ := common.ReadFile(path)
	resourceContent := getSelectorConfiguration(nodeSelectorCMDFlags)
	baseOverlayContent = fmt.Sprintf("%s\n%s", baseOverlayContent, resourceContent)
	return baseOverlayContent
}

func getSelectorConfiguration(annotationCMDFlags common.KeyValueFlags) string {
	resourceArray := []string{}

	tag := fmt.Sprintf("%s%s", common.Spaces(2), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)

	field := fmt.Sprintf("%s%s:", common.Spaces(2), "services")
	resourceArray = append(resourceArray, field)

	field = fmt.Sprintf("%s%s", common.Spaces(2), common.FieldByName("name"))
	resourceArray = append(resourceArray, field)

	serviceName := fmt.Sprintf("%s- %s: %s", common.Spaces(2), "name", "#@ data.values.serviceName")
	resourceArray = append(resourceArray, serviceName)

	tag = fmt.Sprintf("%s%s", common.Spaces(4), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)

	field = fmt.Sprintf("%s%s:", common.Spaces(4), "selector")
	resourceArray = append(resourceArray, field)

	tag = fmt.Sprintf("%s%s", common.Spaces(6), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)

	keyValueField := fmt.Sprintf("%s%s: %s", common.Spaces(6), annotationCMDFlags.Key, "#@ data.values.value")
	resourceArray = append(resourceArray, keyValueField)

	return strings.Join(resourceArray, "\n")
}
