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

type tolerationsFlags struct {
	Key        string
	Operator   string
	Effect     string
	Component  string
	Namespace  string
	DeployName string
}

var tolerationsCMDFlags tolerationsFlags

func getValidOperators() []string {
	return []string{"Exists", "Equal"}
}

func getValidEffects() []string {
	return []string{"NoSchedule", "NoExecute", "PreferNoSchedule"}
}

// newTolerationsCommand represents the configure commands for Knative Serving or Eventing
func newTolerationsCommand(p *pkg.OperatorParams) *cobra.Command {
	var configureTolerationsCmd = &cobra.Command{
		Use:   "tolerations",
		Short: "Configure the tolerations for Knative Serving and Eventing deployments",
		Example: `
  # Configure the tolerations for Knative Serving and Eventing deployments
  kn operation configure tolerations --component eventing --deployName eventing-controller --key example-key --operator Exists --effect NoSchedule --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateTolerationsFlags(tolerationsCMDFlags); err != nil {
				return err
			}

			rootPath, err := os.Getwd()
			if err != nil {
				return err
			}

			err = configureTolerations(tolerationsCMDFlags, rootPath, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified tolerations have been configured in the namespace '%s'.\n",
				tolerationsCMDFlags.Namespace)
			return nil
		},
	}

	configureTolerationsCmd.Flags().StringVar(&tolerationsCMDFlags.Key, "key", "", "The flag to specify the key")
	configureTolerationsCmd.Flags().StringVar(&tolerationsCMDFlags.Operator, "operator", "", "The flag to specify the operator")
	configureTolerationsCmd.Flags().StringVar(&tolerationsCMDFlags.Effect, "effect", "", "The flag to specify the effect")
	configureTolerationsCmd.Flags().StringVar(&tolerationsCMDFlags.DeployName, "deployName", "", "The flag to specify the deployment name")
	configureTolerationsCmd.Flags().StringVarP(&tolerationsCMDFlags.Component, "component", "c", "", "The flag to specify the component name")
	configureTolerationsCmd.Flags().StringVarP(&tolerationsCMDFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return configureTolerationsCmd
}

func validateTolerationsFlags(tolerationsCMDFlags tolerationsFlags) error {
	if tolerationsCMDFlags.Key == "" {
		return fmt.Errorf("You need to specify the key for the toleration.")
	}
	if !common.Contains(getValidOperators(), tolerationsCMDFlags.Operator) {
		return fmt.Errorf("You need to specify the operator to one of the following values: Exists or Equal.")
	}
	if !common.Contains(getValidEffects(), tolerationsCMDFlags.Effect) {
		return fmt.Errorf("You need to specify the effect to one of the following values: NoSchedule, PreferNoSchedule or NoExecute.")
	}
	if tolerationsCMDFlags.Component == "" {
		return fmt.Errorf("You need to specify the component name.")
	}
	if tolerationsCMDFlags.DeployName == "" {
		return fmt.Errorf("You need to specify the name of the deployment.")
	}
	if tolerationsCMDFlags.Namespace == "" {
		return fmt.Errorf("You need to specify the namespace.")
	}
	return nil
}

func configureTolerations(tolerationsCMDFlags tolerationsFlags, rootPath string, p *pkg.OperatorParams) error {
	component := common.ServingComponent
	if strings.EqualFold(tolerationsCMDFlags.Component, common.EventingComponent) {
		component = common.EventingComponent
	}
	yamlTemplateString, err := common.GenerateOperatorCRString(component, tolerationsCMDFlags.Namespace, p)
	if err != nil {
		return err
	}

	overlayContent := getOverlayYamlContent(rootPath, tolerationsCMDFlags)
	valuesYaml := getYamlValuesContentTolerations(tolerationsCMDFlags)

	if err := common.ApplyManifests(yamlTemplateString, overlayContent, valuesYaml, p); err != nil {
		return err
	}
	return nil
}

func getOverlayYamlContent(rootPath string, tolerationsCMDFlags tolerationsFlags) string {
	path := rootPath + "/overlay/ks_toleration.yaml"
	if strings.EqualFold(tolerationsCMDFlags.Component, common.EventingComponent) {
		path = rootPath + "/overlay/ke_toleration.yaml"
	}
	baseOverlayContent, _ := common.ReadFile(path)
	resourceContent := getTolerationConfiguration()
	baseOverlayContent = fmt.Sprintf("%s\n%s", baseOverlayContent, resourceContent)
	return baseOverlayContent
}

func getTolerationConfiguration() string {
	resourceArray := []string{}
	tag := fmt.Sprintf("%s%s", common.Spaces(4), common.FieldByName("key"))
	resourceArray = append(resourceArray, tag)

	keyField := fmt.Sprintf("%s%s", common.Spaces(4), "- key: #@ data.values.key")
	resourceArray = append(resourceArray, keyField)

	tag = fmt.Sprintf("%s%s", common.Spaces(6), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)
	operatorField := fmt.Sprintf("%s%s", common.Spaces(6), "operator: #@ data.values.operator")
	resourceArray = append(resourceArray, operatorField)

	tag = fmt.Sprintf("%s%s", common.Spaces(6), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)
	tolerationField := fmt.Sprintf("%s%s", common.Spaces(6), "effect: #@ data.values.effect")
	resourceArray = append(resourceArray, tolerationField)

	return strings.Join(resourceArray, "\n")
}

func getYamlValuesContentTolerations(tolerationsCMDFlags tolerationsFlags) string {
	contentArray := []string{}
	header := "#@data/values\n---"
	contentArray = append(contentArray, header)
	namespace := fmt.Sprintf("namespace: %s", tolerationsCMDFlags.Namespace)
	contentArray = append(contentArray, namespace)

	deploy := fmt.Sprintf("deployName: %s", tolerationsCMDFlags.DeployName)
	contentArray = append(contentArray, deploy)

	if tolerationsCMDFlags.Key != "" {
		key := fmt.Sprintf("key: %s", tolerationsCMDFlags.Key)
		contentArray = append(contentArray, key)
	}

	if tolerationsCMDFlags.Operator != "" {
		operator := fmt.Sprintf("operator: %s", tolerationsCMDFlags.Operator)
		contentArray = append(contentArray, operator)
	}

	if tolerationsCMDFlags.Effect != "" {
		effect := fmt.Sprintf("effect: %s", tolerationsCMDFlags.Effect)
		contentArray = append(contentArray, effect)
	}

	return strings.Join(contentArray, "\n")
}
