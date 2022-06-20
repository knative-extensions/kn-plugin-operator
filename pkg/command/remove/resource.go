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

package remove

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/kn-plugin-operator/pkg/command/common"
)

type ResourcesFlags struct {
	Component  string
	Namespace  string
	Container  string
	DeployName string
}

var resourcesCMDFlags ResourcesFlags

// newResourcesCommand represents the configure commands for Knative Serving or Eventing
func newResourcesCommand(p *pkg.OperatorParams) *cobra.Command {
	var deleteResourcesCmd = &cobra.Command{
		Use:   "resources",
		Short: "Remove the resource for Knative Serving and Eventing deployments",
		Example: `
  # Remove the configuration of the resources for Knative Serving
  kn operation remove resources --component serving --namespace knative-serving
  # Remove the configuration of the resources for the container activator in the deployment activator
  kn operation remove resources --component serving --deployName activator --namespace knative-serving
  # Remove the configuration of the resources for the deployment activator
  kn operation remove resources --component serving --deployName activator --namespace knative-serving`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateResourcesFlags(resourcesCMDFlags); err != nil {
				return err
			}

			rootPath, err := os.Getwd()
			if err != nil {
				return err
			}

			err = removeResources(resourcesCMDFlags, rootPath, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified resources have been removed in the namespace '%s'.\n",
				resourcesCMDFlags.Namespace)
			return nil
		},
	}

	deleteResourcesCmd.Flags().StringVar(&resourcesCMDFlags.DeployName, "deployName", "", "The flag to specify the deployment name")
	deleteResourcesCmd.Flags().StringVarP(&resourcesCMDFlags.Component, "component", "c", "", "The flag to specify the component name")
	deleteResourcesCmd.Flags().StringVar(&resourcesCMDFlags.Container, "container", "", "The flag to specify the container name")
	deleteResourcesCmd.Flags().StringVarP(&resourcesCMDFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return deleteResourcesCmd
}

func validateResourcesFlags(resourcesCMDFlags ResourcesFlags) error {
	if resourcesCMDFlags.Component == "" {
		return fmt.Errorf("You need to specify the component name.")
	}
	if resourcesCMDFlags.Namespace == "" {
		return fmt.Errorf("You need to specify the namespace.")
	}
	return nil
}

func removeResources(resourcesCMDFlags ResourcesFlags, rootPath string, p *pkg.OperatorParams) error {
	component := common.ServingComponent
	if strings.EqualFold(resourcesCMDFlags.Component, common.EventingComponent) {
		component = common.EventingComponent
	}
	yamlTemplateString, err := common.GenerateOperatorCRString(component, resourcesCMDFlags.Namespace, p)
	if err != nil {
		return err
	}

	overlayContent := getOverlayYamlContentResource(rootPath, resourcesCMDFlags)
	valuesYaml := getYamlValuesContentResources(resourcesCMDFlags)

	if err := common.ApplyManifests(yamlTemplateString, overlayContent, valuesYaml, p); err != nil {
		return err
	}
	return nil
}
