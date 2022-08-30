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
var servingAnnotationOverlay string

//go:embed overlay/ke_deploy_label.yaml
var eventingAnnotationOverlay string

var annotationCMDFlags common.KeyValueFlags

// newAnnotationCommand represents the configure commands to configure the annotations for Knative deployment
func newAnnotationCommand(p *pkg.OperatorParams) *cobra.Command {
	var configureLabelsCmd = &cobra.Command{
		Use:   "annotations",
		Short: "Configure the annotations for Knative Serving and Eventing deployments or services",
		Example: `
  # Configure the annotations for Knative Serving and Eventing deployments
  kn operation annotations --component eventing --deployName eventing-controller --key key --value value --namespace knative-eventing
  # Configure the annotations for Knative Serving and Eventing services
  kn operation annotations --component eventing --serviceName eventing-controller --key key --value value --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateLabelsAnnotationsFlags(annotationCMDFlags); err != nil {
				return err
			}

			err := configureAnnotations(annotationCMDFlags, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified annotation has been configured for the deployment %s in the deployment '%s'.\n",
				annotationCMDFlags.DeployName, annotationCMDFlags.Namespace)
			return nil
		},
	}

	configureLabelsCmd.Flags().StringVar(&annotationCMDFlags.Key, "key", "", "The key of the data in the configmap")
	configureLabelsCmd.Flags().StringVar(&annotationCMDFlags.Value, "value", "", "The value of the data in the configmap")
	configureLabelsCmd.Flags().StringVar(&annotationCMDFlags.DeployName, "deployName", "", "The flag to specify the deployment name")
	configureLabelsCmd.Flags().StringVar(&annotationCMDFlags.ServiceName, "serviceName", "", "The flag to specify the service name")
	configureLabelsCmd.Flags().StringVarP(&annotationCMDFlags.Component, "component", "c", "", "The flag to specify the component name")
	configureLabelsCmd.Flags().StringVarP(&annotationCMDFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return configureLabelsCmd
}

func configureAnnotations(annotationCMDFlags common.KeyValueFlags, p *pkg.OperatorParams) error {
	component := common.ServingComponent
	if strings.EqualFold(annotationCMDFlags.Component, common.EventingComponent) {
		component = common.EventingComponent
	}
	yamlTemplateString, err := common.GenerateOperatorCRString(component, annotationCMDFlags.Namespace, p)
	if err != nil {
		return err
	}

	annotationOverlayContent := getOverlayYamlContentAnnotation(annotationCMDFlags)
	valuesYaml := getYamlValuesContent(annotationCMDFlags)
	if err := common.ApplyManifests(yamlTemplateString, annotationOverlayContent, valuesYaml, p); err != nil {
		return err
	}
	return nil
}

func getOverlayYamlContentAnnotation(annotationCMDFlags common.KeyValueFlags) string {
	baseOverlayContent := servingAnnotationOverlay
	if strings.EqualFold(annotationCMDFlags.Component, common.EventingComponent) {
		baseOverlayContent = eventingAnnotationOverlay
	}
	resourceContent := getAnnotationConfiguration(annotationCMDFlags)
	baseOverlayContent = fmt.Sprintf("%s\n%s", baseOverlayContent, resourceContent)
	return baseOverlayContent
}

func getAnnotationConfiguration(annotationCMDFlags common.KeyValueFlags) string {
	resourceArray := []string{}

	tag := fmt.Sprintf("%s%s", common.Spaces(2), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)

	if annotationCMDFlags.DeployName != "" {
		field := fmt.Sprintf("%s%s:", common.Spaces(2), "deployments")
		resourceArray = append(resourceArray, field)
	} else {
		field := fmt.Sprintf("%s%s:", common.Spaces(2), "services")
		resourceArray = append(resourceArray, field)
	}

	field := fmt.Sprintf("%s%s", common.Spaces(2), common.FieldByName("name"))
	resourceArray = append(resourceArray, field)

	if annotationCMDFlags.DeployName != "" {
		deployName := fmt.Sprintf("%s- %s: %s", common.Spaces(2), "name", "#@ data.values.deployName")
		resourceArray = append(resourceArray, deployName)
	} else {
		serviceName := fmt.Sprintf("%s- %s: %s", common.Spaces(2), "name", "#@ data.values.serviceName")
		resourceArray = append(resourceArray, serviceName)
	}

	tag = fmt.Sprintf("%s%s", common.Spaces(4), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)

	field = fmt.Sprintf("%s%s:", common.Spaces(4), "annotations")
	resourceArray = append(resourceArray, field)

	tag = fmt.Sprintf("%s%s", common.Spaces(6), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)

	keyValueField := fmt.Sprintf("%s%s: %s", common.Spaces(6), annotationCMDFlags.Key, "#@ data.values.value")
	resourceArray = append(resourceArray, keyValueField)

	return strings.Join(resourceArray, "\n")
}
