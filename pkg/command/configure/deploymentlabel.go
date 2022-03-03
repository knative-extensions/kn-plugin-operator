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

type deploymentLabelFlags struct {
	Value        string
	Key          string
	Component    string
	Namespace    string
	DeployName   string
	NodeSelector bool
	Annotation   bool
	Label        bool
}

var deploymentLabelCMDFlags deploymentLabelFlags

// newDeploymentLabelCommand represents the configure commands to configure the labels for Knative deployment
func newDeploymentLabelCommand(p *pkg.OperatorParams) *cobra.Command {
	var configureLabelsCmd = &cobra.Command{
		Use:   "labels",
		Short: "Configure the labels for Knative Serving and Eventing deployments",
		Example: `
  # Configure the labels for Knative Serving and Eventing deployments
  kn operation configure labels --component eventing --deployName eventing-controller --key key --value value--namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateLabelsFlags(deploymentLabelCMDFlags); err != nil {
				return err
			}

			rootPath, err := os.Getwd()
			if err != nil {
				return err
			}

			err = configureLabels(deploymentLabelCMDFlags, rootPath, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified labels has been configured for the deployment %s in the deployment '%s'.\n",
				deploymentLabelCMDFlags.DeployName, deploymentLabelCMDFlags.Namespace)
			return nil
		},
	}

	configureLabelsCmd.Flags().BoolVar(&deploymentLabelCMDFlags.Label, "label", false, "The flag to enable the label configuration")
	configureLabelsCmd.Flags().BoolVar(&deploymentLabelCMDFlags.Annotation, "annotation", false, "The flag to enable the annotation configuration")
	configureLabelsCmd.Flags().BoolVar(&deploymentLabelCMDFlags.NodeSelector, "nodeSelector", false, "The flag to enable the nodeSelector configuration")
	configureLabelsCmd.Flags().StringVar(&deploymentLabelCMDFlags.Key, "key", "", "The key of the data in the configmap")
	configureLabelsCmd.Flags().StringVar(&deploymentLabelCMDFlags.Value, "value", "", "The value of the data in the configmap")
	configureLabelsCmd.Flags().StringVar(&deploymentLabelCMDFlags.DeployName, "deployName", "", "The flag to specify the configmap name")
	configureLabelsCmd.Flags().StringVarP(&deploymentLabelCMDFlags.Component, "component", "c", "", "The flag to specify the component name")
	configureLabelsCmd.Flags().StringVarP(&deploymentLabelCMDFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return configureLabelsCmd
}

func validateLabelsFlags(deploymentLabelCMDFlags deploymentLabelFlags) error {
	count := 0

	if deploymentLabelCMDFlags.Label {
		count++
	}

	if deploymentLabelCMDFlags.NodeSelector {
		count++
	}

	if deploymentLabelCMDFlags.Annotation {
		count++
	}

	if count == 0 {
		return fmt.Errorf("You need to enable at least one deployment aspect for Knative: NodeSelector, Annotation or Label.")
	}
	if count > 1 {
		return fmt.Errorf("You can specify only one deployment aspect for Knative: NodeSelector, Annotation or Label.")
	}

	if deploymentLabelCMDFlags.Key == "" {
		return fmt.Errorf("You need to specify the key for the deployment.")
	}
	if deploymentLabelCMDFlags.Value == "" {
		return fmt.Errorf("You need to specify the value for the deployment.")
	}
	if deploymentLabelCMDFlags.DeployName == "" {
		return fmt.Errorf("You need to specify the name of the deployment.")
	}
	if deploymentLabelCMDFlags.Namespace == "" {
		return fmt.Errorf("You need to specify the namespace.")
	}
	if deploymentLabelCMDFlags.Component != "" && !strings.EqualFold(deploymentLabelCMDFlags.Component, common.ServingComponent) && !strings.EqualFold(deploymentLabelCMDFlags.Component, common.EventingComponent) {
		return fmt.Errorf("You need to specify the component for Knative: serving or eventing.")
	}
	return nil
}

func configureLabels(deploymentLabelCMDFlags deploymentLabelFlags, rootPath string, p *pkg.OperatorParams) error {
	component := common.ServingComponent
	if strings.EqualFold(deploymentLabelCMDFlags.Component, common.EventingComponent) {
		component = common.EventingComponent
	}
	yamlTemplateString, err := common.GenerateOperatorCRString(component, deploymentLabelCMDFlags.Namespace, p)
	if err != nil {
		return err
	}

	overlayContent := getOverlayYamlContentLabel(rootPath, deploymentLabelCMDFlags)
	valuesYaml := getYamlValuesContentLabels(deploymentLabelCMDFlags)
	if err := common.ApplyManifests(yamlTemplateString, overlayContent, valuesYaml, p); err != nil {
		return err
	}
	return nil
}

func getOverlayYamlContentLabel(rootPath string, deploymentLabelCMDFlags deploymentLabelFlags) string {
	path := rootPath + "/overlay/ks_deploy_label.yaml"
	if strings.EqualFold(deploymentLabelCMDFlags.Component, common.EventingComponent) {
		path = rootPath + "/overlay/ke_deploy_label.yaml"
	}
	baseOverlayContent, _ := common.ReadFile(path)
	resourceContent := getLabelConfiguration(deploymentLabelCMDFlags)
	baseOverlayContent = fmt.Sprintf("%s\n%s", baseOverlayContent, resourceContent)
	return baseOverlayContent
}

func getLabelConfiguration(deploymentLabelCMDFlags deploymentLabelFlags) string {
	resourceArray := []string{}
	tag := fmt.Sprintf("%s%s", common.Spaces(4), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)

	if deploymentLabelCMDFlags.Label {
		field := fmt.Sprintf("%s%s:", common.Spaces(4), "labels")
		resourceArray = append(resourceArray, field)
	}

	if deploymentLabelCMDFlags.Annotation {
		field := fmt.Sprintf("%s%s:", common.Spaces(4), "annotations")
		resourceArray = append(resourceArray, field)
	}

	if deploymentLabelCMDFlags.NodeSelector {
		field := fmt.Sprintf("%s%s:", common.Spaces(4), "nodeSelector")
		resourceArray = append(resourceArray, field)
	}

	tag = fmt.Sprintf("%s%s", common.Spaces(6), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)

	keyValueField := fmt.Sprintf("%s%s: %s", common.Spaces(6), deploymentLabelCMDFlags.Key, "#@ data.values.value")
	resourceArray = append(resourceArray, keyValueField)

	return strings.Join(resourceArray, "\n")
}

func getYamlValuesContentLabels(deploymentLabelCMDFlags deploymentLabelFlags) string {
	contentArray := []string{}
	header := "#@data/values\n---"
	contentArray = append(contentArray, header)
	namespace := fmt.Sprintf("namespace: %s", deploymentLabelCMDFlags.Namespace)
	contentArray = append(contentArray, namespace)
	deployName := fmt.Sprintf("deployName: %s", deploymentLabelCMDFlags.DeployName)
	contentArray = append(contentArray, deployName)
	value := fmt.Sprintf("value: %s", deploymentLabelCMDFlags.Value)
	contentArray = append(contentArray, value)
	return strings.Join(contentArray, "\n")
}
