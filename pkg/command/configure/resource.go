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

//go:embed overlay/ks_resource_base.yaml
var servingResourceOverlay string

//go:embed overlay/ke_resource_base.yaml
var eventingResourceOverlay string

type ResourcesFlags struct {
	LimitCPU      string
	LimitMemory   string
	RequestCPU    string
	RequestMemory string
	Component     string
	Namespace     string
	Container     string
	DeployName    string
}

var resourcesCMDFlags ResourcesFlags

// newResourcesCommand represents the configure commands for Knative Serving or Eventing
func newResourcesCommand(p *pkg.OperatorParams) *cobra.Command {
	var configureResourcesCmd = &cobra.Command{
		Use:   "resources",
		Short: "Configure the resource for Knative Serving and Eventing deployments",
		Example: `
  # Configure the resource for Knative Serving and Eventing deployments
  kn operation configure resources --component eventing --deployName eventing-controller --container eventing-controller --requestMemory 200Mi --requestCPU 200m --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateResourcesFlags(resourcesCMDFlags); err != nil {
				return err
			}

			err := configureResources(resourcesCMDFlags, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified resources have been configured in the namespace '%s'.\n",
				resourcesCMDFlags.Namespace)
			return nil
		},
	}

	configureResourcesCmd.Flags().StringVar(&resourcesCMDFlags.LimitCPU, "limitCPU", "", "The flag to specify the limit CPU")
	configureResourcesCmd.Flags().StringVar(&resourcesCMDFlags.LimitMemory, "limitMemory", "", "The flag to specify the limit memory")
	configureResourcesCmd.Flags().StringVar(&resourcesCMDFlags.RequestCPU, "requestCPU", "", "The flag to specify the request CPU")
	configureResourcesCmd.Flags().StringVar(&resourcesCMDFlags.RequestMemory, "requestMemory", "", "The flag to specify the request memory")
	configureResourcesCmd.Flags().StringVar(&resourcesCMDFlags.DeployName, "deployName", "", "The flag to specify the deployment name")
	configureResourcesCmd.Flags().StringVarP(&resourcesCMDFlags.Component, "component", "c", "", "The flag to specify the component name")
	configureResourcesCmd.Flags().StringVar(&resourcesCMDFlags.Container, "container", "", "The flag to specify the container name")
	configureResourcesCmd.Flags().StringVarP(&resourcesCMDFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return configureResourcesCmd
}

func validateResourcesFlags(resourcesCMDFlags ResourcesFlags) error {
	count := 0

	if resourcesCMDFlags.LimitCPU != "" {
		count++
	}

	if resourcesCMDFlags.LimitMemory != "" {
		count++
	}

	if resourcesCMDFlags.RequestCPU != "" {
		count++
	}

	if resourcesCMDFlags.RequestMemory != "" {
		count++
	}

	if count == 0 {
		return fmt.Errorf("You need to specify at least one resource parameter: limitCPU, limitMemory, requestCPU or requestMemory.")
	}

	if resourcesCMDFlags.Container == "" {
		return fmt.Errorf("You need to specify the container name.")
	}
	if resourcesCMDFlags.Component == "" {
		return fmt.Errorf("You need to specify the component name.")
	}
	if resourcesCMDFlags.DeployName == "" {
		return fmt.Errorf("You need to specify the name of the deployment.")
	}
	if resourcesCMDFlags.Namespace == "" {
		return fmt.Errorf("You need to specify the namespace.")
	}
	return nil
}

func configureResources(resourcesCMDFlags ResourcesFlags, p *pkg.OperatorParams) error {
	component := common.ServingComponent
	if strings.EqualFold(resourcesCMDFlags.Component, common.EventingComponent) {
		component = common.EventingComponent
	}
	yamlTemplateString, err := common.GenerateOperatorCRString(component, resourcesCMDFlags.Namespace, p)
	if err != nil {
		return err
	}

	overlayContent := getOverlayYamlContentResource(resourcesCMDFlags)
	valuesYaml := getYamlValuesContentResources(resourcesCMDFlags)

	if err := common.ApplyManifests(yamlTemplateString, overlayContent, valuesYaml, p); err != nil {
		return err
	}
	return nil
}

func getOverlayYamlContentResource(resourcesCMDFlags ResourcesFlags) string {
	baseOverlayContent := servingResourceOverlay
	if strings.EqualFold(resourcesCMDFlags.Component, common.EventingComponent) {
		baseOverlayContent = eventingResourceOverlay
	}
	resourceContent := getResourceConfiguration(resourcesCMDFlags)
	baseOverlayContent = fmt.Sprintf("%s\n%s", baseOverlayContent, resourceContent)
	return baseOverlayContent
}

func getResourceConfiguration(resourcesCMDFlags ResourcesFlags) string {
	resourceArray := []string{}
	tag := fmt.Sprintf("%s%s", common.Spaces(6), common.FieldByName("container"))
	resourceArray = append(resourceArray, tag)

	containerField := fmt.Sprintf("%s%s", common.Spaces(4), "- container: #@ data.values.container")
	resourceArray = append(resourceArray, containerField)

	if resourcesCMDFlags.RequestCPU != "" || resourcesCMDFlags.RequestMemory != "" {
		tag = fmt.Sprintf("%s%s", common.Spaces(6), common.YttMatchingTag)
		resourceArray = append(resourceArray, tag)
		requestField := fmt.Sprintf("%s%s", common.Spaces(6), "requests:")
		resourceArray = append(resourceArray, requestField)
	}

	if resourcesCMDFlags.RequestCPU != "" {
		tag = fmt.Sprintf("%s%s", common.Spaces(8), common.YttMatchingTag)
		resourceArray = append(resourceArray, tag)
		requestCPUField := fmt.Sprintf("%s%s", common.Spaces(8), "cpu: #@ data.values.requestCPU")
		resourceArray = append(resourceArray, requestCPUField)
	}

	if resourcesCMDFlags.RequestMemory != "" {
		tag = fmt.Sprintf("%s%s", common.Spaces(8), common.YttMatchingTag)
		resourceArray = append(resourceArray, tag)
		requestMemoryField := fmt.Sprintf("%s%s", common.Spaces(8), "memory: #@ data.values.requestMemory")
		resourceArray = append(resourceArray, requestMemoryField)
	}

	if resourcesCMDFlags.LimitCPU != "" || resourcesCMDFlags.LimitMemory != "" {
		tag = fmt.Sprintf("%s%s", common.Spaces(6), common.YttMatchingTag)
		resourceArray = append(resourceArray, tag)
		limitField := fmt.Sprintf("%s%s", common.Spaces(6), "limits:")
		resourceArray = append(resourceArray, limitField)
	}

	if resourcesCMDFlags.LimitCPU != "" {
		tag = fmt.Sprintf("%s%s", common.Spaces(8), common.YttMatchingTag)
		resourceArray = append(resourceArray, tag)
		limitCPUField := fmt.Sprintf("%s%s", common.Spaces(8), "cpu: #@ data.values.limitCPU")
		resourceArray = append(resourceArray, limitCPUField)
	}

	if resourcesCMDFlags.LimitMemory != "" {
		tag = fmt.Sprintf("%s%s", common.Spaces(8), common.YttMatchingTag)
		resourceArray = append(resourceArray, tag)
		limitMemoryField := fmt.Sprintf("%s%s", common.Spaces(8), "memory: #@ data.values.limitMemory")
		resourceArray = append(resourceArray, limitMemoryField)
	}

	return strings.Join(resourceArray, "\n")
}

func getYamlValuesContentResources(resourcesCMDFlags ResourcesFlags) string {
	contentArray := []string{}
	header := "#@data/values\n---"
	contentArray = append(contentArray, header)
	namespace := fmt.Sprintf("namespace: %s", resourcesCMDFlags.Namespace)
	contentArray = append(contentArray, namespace)

	container := fmt.Sprintf("container: %s", resourcesCMDFlags.Container)
	contentArray = append(contentArray, container)

	deploy := fmt.Sprintf("deployName: %s", resourcesCMDFlags.DeployName)
	contentArray = append(contentArray, deploy)

	if resourcesCMDFlags.RequestCPU != "" {
		requestCPU := fmt.Sprintf("requestCPU: %s", resourcesCMDFlags.RequestCPU)
		contentArray = append(contentArray, requestCPU)
	}

	if resourcesCMDFlags.RequestMemory != "" {
		requestMemory := fmt.Sprintf("requestMemory: %s", resourcesCMDFlags.RequestMemory)
		contentArray = append(contentArray, requestMemory)
	}

	if resourcesCMDFlags.LimitCPU != "" {
		limitCPU := fmt.Sprintf("limitCPU: %s", resourcesCMDFlags.LimitCPU)
		contentArray = append(contentArray, limitCPU)
	}

	if resourcesCMDFlags.LimitMemory != "" {
		limitMemory := fmt.Sprintf("limitMemory: %s", resourcesCMDFlags.LimitMemory)
		contentArray = append(contentArray, limitMemory)
	}

	return strings.Join(contentArray, "\n")
}
