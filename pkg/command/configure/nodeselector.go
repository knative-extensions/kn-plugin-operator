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

//go:embed overlay/ks_deploy_label.yaml
var servingNodeSelectorOverlay string

//go:embed overlay/ke_deploy_label.yaml
var eventingNodeSelectorOverlay string

var nodeSelectorCMDFlags common.KeyValueFlags

// newNodeSelectorCommand represents the configure commands to configure the nodeSelector for Knative deployment
func newNodeSelectorCommand(p *pkg.OperatorParams) *cobra.Command {
	var configureNodeSelectorsCmd = &cobra.Command{
		Use:   "nodeSelectors",
		Short: "Configure the node selectors for Knative Serving and Eventing deployments",
		Example: `
  # Configure the nodeSelectors for Knative Serving and Eventing deployments
  kn operator configure nodeSelectors --component eventing --deployName eventing-controller --key key --value value --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateNodeSelectorFlags(nodeSelectorCMDFlags); err != nil {
				return err
			}

			err := configureNodeSelectors(nodeSelectorCMDFlags, p)
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
	configureNodeSelectorsCmd.Flags().StringVar(&nodeSelectorCMDFlags.DeployName, "deployName", "", "The flag to specify the deployment name")
	configureNodeSelectorsCmd.Flags().StringVarP(&nodeSelectorCMDFlags.Component, "component", "c", "", "The flag to specify the component name")
	configureNodeSelectorsCmd.Flags().StringVarP(&nodeSelectorCMDFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return configureNodeSelectorsCmd
}

func validateNodeSelectorFlags(keyValuesCMDFlags common.KeyValueFlags) error {
	if err := validateKeyValuePairs(keyValuesCMDFlags); err != nil {
		return err
	}
	if keyValuesCMDFlags.DeployName == "" {
		return fmt.Errorf("You need to specify the name of the deployment.")
	}
	return nil
}

func configureNodeSelectors(nodeSelectorCMDFlags common.KeyValueFlags, p *pkg.OperatorParams) error {
	component := common.ServingComponent
	if strings.EqualFold(nodeSelectorCMDFlags.Component, common.EventingComponent) {
		component = common.EventingComponent
	}
	yamlTemplateString, err := common.GenerateOperatorCRString(component, nodeSelectorCMDFlags.Namespace, p)
	if err != nil {
		return err
	}

	overlayContent := getOverlayYamlContentNodeSelector(nodeSelectorCMDFlags)
	valuesYaml := getYamlValuesContent(nodeSelectorCMDFlags)
	if err := common.ApplyManifests(yamlTemplateString, overlayContent, valuesYaml, p); err != nil {
		return err
	}
	return nil
}

func getOverlayYamlContentNodeSelector(nodeSelectorCMDFlags common.KeyValueFlags) string {
	baseOverlayContent := servingNodeSelectorOverlay
	if strings.EqualFold(nodeSelectorCMDFlags.Component, common.EventingComponent) {
		baseOverlayContent = eventingNodeSelectorOverlay
	}
	resourceContent := getNodeSelectorConfiguration(nodeSelectorCMDFlags)
	baseOverlayContent = fmt.Sprintf("%s\n%s", baseOverlayContent, resourceContent)
	return baseOverlayContent
}

func getNodeSelectorConfiguration(annotationCMDFlags common.KeyValueFlags) string {
	resourceArray := []string{}

	tag := fmt.Sprintf("%s%s", common.Spaces(2), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)

	field := fmt.Sprintf("%s%s:", common.Spaces(2), "deployments")
	resourceArray = append(resourceArray, field)

	field = fmt.Sprintf("%s%s", common.Spaces(2), common.FieldByName("name"))
	resourceArray = append(resourceArray, field)

	deployName := fmt.Sprintf("%s- %s: %s", common.Spaces(2), "name", "#@ data.values.deployName")
	resourceArray = append(resourceArray, deployName)

	tag = fmt.Sprintf("%s%s", common.Spaces(4), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)

	field = fmt.Sprintf("%s%s:", common.Spaces(4), "nodeSelector")
	resourceArray = append(resourceArray, field)

	tag = fmt.Sprintf("%s%s", common.Spaces(6), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)

	keyValueField := fmt.Sprintf("%s%s: %s", common.Spaces(6), annotationCMDFlags.Key, "#@ data.values.value")
	resourceArray = append(resourceArray, keyValueField)

	return strings.Join(resourceArray, "\n")
}
