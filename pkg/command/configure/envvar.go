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

//go:embed overlay/ks_envvar.yaml
var servingEnvVarOverlay string

//go:embed overlay/ke_envvar.yaml
var eventingEnvVarOverlay string

type EnvVarFlags struct {
	EnvName       string
	EnvValue      string
	Component     string
	Namespace     string
	DeployName    string
	ContainerName string
}

var envVarFlags EnvVarFlags

// newEnvVarCommand represents the configure commands to configure the env vars for Knative Deployment resources
func newEnvVarCommand(p *pkg.OperatorParams) *cobra.Command {
	var configureImagesCmd = &cobra.Command{
		Use:   "envvars",
		Short: "Configure the env vars for Knative",
		Example: `
  # Configure the env vars for Knative
  kn operation configure envvars --component eventing --deployName eventing-controller --container eventing-controller --name key --value value --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateEnvVarsFlags(envVarFlags); err != nil {
				return err
			}

			err := configureEnvVars(envVarFlags, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified images has been configured.\n")
			return nil
		},
	}

	configureImagesCmd.Flags().StringVar(&envVarFlags.EnvName, "name", "", "The name for the environment variable")
	configureImagesCmd.Flags().StringVar(&envVarFlags.EnvValue, "value", "", "The value for the environment variable")
	configureImagesCmd.Flags().StringVar(&envVarFlags.DeployName, "deployName", "", "The flag to specify the deployment name")
	configureImagesCmd.Flags().StringVarP(&envVarFlags.Component, "component", "c", "", "The flag to specify the component name")
	configureImagesCmd.Flags().StringVarP(&envVarFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")
	configureImagesCmd.Flags().StringVar(&envVarFlags.ContainerName, "container", "", "The name of the container")

	return configureImagesCmd
}

func validateEnvVarsFlags(envVarFlags EnvVarFlags) error {
	if envVarFlags.EnvName == "" {
		return fmt.Errorf("You need to specify the name for the environment variable.")
	}
	if envVarFlags.EnvValue == "" {
		return fmt.Errorf("You need to specify the value for the environment variable.")
	}

	if envVarFlags.DeployName == "" {
		return fmt.Errorf("You need to specify the name for the deployment resource.")
	}

	if envVarFlags.ContainerName == "" {
		return fmt.Errorf("You need to specify the name for the container.")
	}

	if envVarFlags.Namespace == "" {
		return fmt.Errorf("You need to specify the namespace.")
	}
	if envVarFlags.Component != "" && !strings.EqualFold(envVarFlags.Component, common.ServingComponent) && !strings.EqualFold(envVarFlags.Component, common.EventingComponent) {
		return fmt.Errorf("You need to specify the component for Knative: serving or eventing.")
	}
	return nil
}

func configureEnvVars(envVarFlags EnvVarFlags, p *pkg.OperatorParams) error {
	component := common.ServingComponent
	if strings.EqualFold(envVarFlags.Component, common.EventingComponent) {
		component = common.EventingComponent
	}
	yamlTemplateString, err := common.GenerateOperatorCRString(component, envVarFlags.Namespace, p)
	if err != nil {
		return err
	}

	overlayContent := getOverlayYamlContentEnvvar(envVarFlags)
	valuesYaml := getYamlValuesContentEnvvars(envVarFlags)
	if err := common.ApplyManifests(yamlTemplateString, overlayContent, valuesYaml, p); err != nil {
		return err
	}
	return nil
}

func getOverlayYamlContentEnvvar(envVarFlags EnvVarFlags) string {
	baseOverlayContent := servingEnvVarOverlay
	if strings.EqualFold(envVarFlags.Component, common.EventingComponent) {
		baseOverlayContent = eventingEnvVarOverlay
	}
	return baseOverlayContent
}

func getYamlValuesContentEnvvars(envVarFlags EnvVarFlags) string {
	contentArray := []string{}
	header := "#@data/values\n---"
	contentArray = append(contentArray, header)

	namespace := fmt.Sprintf("namespace: %s", envVarFlags.Namespace)
	contentArray = append(contentArray, namespace)

	deployName := fmt.Sprintf("deployName: %s", envVarFlags.DeployName)
	contentArray = append(contentArray, deployName)

	value := fmt.Sprintf("containerName: %s", envVarFlags.ContainerName)
	contentArray = append(contentArray, value)

	envName := fmt.Sprintf("envVarName: %s", envVarFlags.EnvName)
	contentArray = append(contentArray, envName)

	envValue := fmt.Sprintf("envVarValue: %s", envVarFlags.EnvValue)
	contentArray = append(contentArray, envValue)
	return strings.Join(contentArray, "\n")
}
