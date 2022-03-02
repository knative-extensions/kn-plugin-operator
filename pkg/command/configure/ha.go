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

type haFlags struct {
	Replicas   string
	Component  string
	Namespace  string
	DeployName string
}

var haCMDFlags haFlags

// newHACommand represents the HA configure commands for Serving or Eventing
func newHACommand(p *pkg.OperatorParams) *cobra.Command {
	var configureHAsCmd = &cobra.Command{
		Use:   "replicas",
		Short: "Configure the number of replicas for Knative Serving and Eventing deployments",
		Example: `
  # Configure the number of replicas for Knative Serving and Eventing deployments
  kn operation configure replicas --component eventing --deployName eventing-controller --replicas 3 --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateHAsFlags(haCMDFlags); err != nil {
				return err
			}

			rootPath, err := os.Getwd()
			if err != nil {
				return err
			}

			err = configureHAs(haCMDFlags, rootPath, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified number of replicas has been configured in the namespace '%s'.\n",
				haCMDFlags.Namespace)
			return nil
		},
	}

	configureHAsCmd.Flags().StringVar(&haCMDFlags.Replicas, "replicas", "", "The flag to specify the minimum number of replicas")
	configureHAsCmd.Flags().StringVar(&haCMDFlags.DeployName, "deployName", "", "The flag to specify the deployment name")
	configureHAsCmd.Flags().StringVarP(&haCMDFlags.Component, "component", "c", "", "The flag to specify the component name")
	configureHAsCmd.Flags().StringVarP(&haCMDFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return configureHAsCmd
}

func validateHAsFlags(haCMDFlags haFlags) error {
	if haCMDFlags.Replicas == "" {
		return fmt.Errorf("You need to specify the number of the replicas.")
	}
	if haCMDFlags.Component == "" {
		return fmt.Errorf("You need to specify the component name.")
	}
	if haCMDFlags.Namespace == "" {
		return fmt.Errorf("You need to specify the namespace.")
	}
	if haCMDFlags.Component != "" && !strings.EqualFold(haCMDFlags.Component, common.ServingComponent) && !strings.EqualFold(haCMDFlags.Component, common.EventingComponent) {
		return fmt.Errorf("You need to specify the component for Knative: serving or eventing.")
	}
	return nil
}

func configureHAs(haCMDFlags haFlags, rootPath string, p *pkg.OperatorParams) error {
	component := common.ServingComponent
	if strings.EqualFold(haCMDFlags.Component, common.EventingComponent) {
		component = common.EventingComponent
	}
	yamlTemplateString, err := common.GenerateOperatorCRString(component, haCMDFlags.Namespace, p)
	if err != nil {
		return err
	}

	overlayContent := getOverlayYamlContentHA(rootPath, haCMDFlags)
	valuesYaml := getYamlValuesContentHAs(haCMDFlags)

	if err := common.ApplyManifests(yamlTemplateString, overlayContent, valuesYaml, p); err != nil {
		return err
	}
	return nil
}

func getOverlayYamlContentHA(rootPath string, haCMDFlags haFlags) string {
	path := rootPath + "/overlay/ks_replica.yaml"
	if strings.EqualFold(haCMDFlags.Component, common.EventingComponent) {
		path = rootPath + "/overlay/ke_replica.yaml"
	}
	baseOverlayContent, _ := common.ReadFile(path)
	haContent := getHAConfiguration(haCMDFlags)
	baseOverlayContent = fmt.Sprintf("%s\n%s", baseOverlayContent, haContent)
	return baseOverlayContent
}

func getHAConfiguration(haCMDFlags haFlags) string {
	resourceArray := []string{}
	if haCMDFlags.DeployName == "" {
		// Set the HA number of replicas globally to all deployments
		tag := fmt.Sprintf("%s%s", common.Spaces(2), common.YttMatchingTag)
		resourceArray = append(resourceArray, tag)
		haField := fmt.Sprintf("%s%s", common.Spaces(2), "high-availability:")
		resourceArray = append(resourceArray, haField)
		tag = fmt.Sprintf("%s%s", common.Spaces(4), common.YttMatchingTag)
		resourceArray = append(resourceArray, tag)
		replicaField := fmt.Sprintf("%s%s", common.Spaces(4), "replicas: #@ data.values.replicas")
		resourceArray = append(resourceArray, replicaField)
	} else {
		tag := fmt.Sprintf("%s%s", common.Spaces(2), common.YttMatchingTag)
		resourceArray = append(resourceArray, tag)
		deploymentField := fmt.Sprintf("%s%s", common.Spaces(2), "deployments:")
		resourceArray = append(resourceArray, deploymentField)
		tag = fmt.Sprintf("%s%s", common.Spaces(2), common.FieldByName("name"))
		resourceArray = append(resourceArray, tag)
		nameField := fmt.Sprintf("%s%s", common.Spaces(2), "- name: #@ data.values.name")
		resourceArray = append(resourceArray, nameField)
		tag = fmt.Sprintf("%s%s", common.Spaces(4), common.YttMatchingTag)
		resourceArray = append(resourceArray, tag)
		replicaField := fmt.Sprintf("%s%s", common.Spaces(4), "replicas: #@ data.values.replicas")
		resourceArray = append(resourceArray, replicaField)
	}

	return strings.Join(resourceArray, "\n")
}

func getYamlValuesContentHAs(haCMDFlags haFlags) string {
	contentArray := []string{}
	header := "#@data/values\n---"
	contentArray = append(contentArray, header)

	namespace := fmt.Sprintf("namespace: %s", haCMDFlags.Namespace)
	contentArray = append(contentArray, namespace)

	if haCMDFlags.DeployName == "" {
		replicas := fmt.Sprintf("replicas: %s", haCMDFlags.Replicas)
		contentArray = append(contentArray, replicas)
	} else {
		deployName := fmt.Sprintf("name: %s", haCMDFlags.DeployName)
		contentArray = append(contentArray, deployName)
		replicas := fmt.Sprintf("replicas: %s", haCMDFlags.Replicas)
		contentArray = append(contentArray, replicas)
	}

	return strings.Join(contentArray, "\n")
}
