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

type imageFlags struct {
	ImageUrl   string
	Component  string
	Namespace  string
	DeployName string
	ImageKey   string
}

var imageCMDFlags imageFlags

// newImageCommand represents the configure commands to configure the image for Knative
func newImageCommand(p *pkg.OperatorParams) *cobra.Command {
	var configureImagesCmd = &cobra.Command{
		Use:   "images",
		Short: "Configure the images for Knative",
		Example: `
  # Configure the images for Knative
  kn operation configure images --component eventing --deployName eventing-controller --imageKey key --imageURL value --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateImagesFlags(imageCMDFlags); err != nil {
				return err
			}

			rootPath, err := os.Getwd()
			if err != nil {
				return err
			}

			err = configureImages(imageCMDFlags, rootPath, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified images has been configured.\n")
			return nil
		},
	}

	configureImagesCmd.Flags().StringVar(&imageCMDFlags.ImageKey, "imageKey", "", "The image key")
	configureImagesCmd.Flags().StringVar(&imageCMDFlags.ImageUrl, "imageURL", "", "The image URL")
	configureImagesCmd.Flags().StringVar(&imageCMDFlags.DeployName, "deployName", "", "The flag to specify the deployment name")
	configureImagesCmd.Flags().StringVarP(&imageCMDFlags.Component, "component", "c", "", "The flag to specify the component name")
	configureImagesCmd.Flags().StringVarP(&imageCMDFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return configureImagesCmd
}

func validateImagesFlags(imageCMDFlags imageFlags) error {
	if imageCMDFlags.ImageKey == "" {
		return fmt.Errorf("You need to specify the image key.")
	}
	if imageCMDFlags.ImageUrl == "" {
		return fmt.Errorf("You need to specify the image URL.")
	}

	if imageCMDFlags.Namespace == "" {
		return fmt.Errorf("You need to specify the namespace.")
	}
	if imageCMDFlags.Component != "" && !strings.EqualFold(imageCMDFlags.Component, common.ServingComponent) && !strings.EqualFold(imageCMDFlags.Component, common.EventingComponent) {
		return fmt.Errorf("You need to specify the component for Knative: serving or eventing.")
	}
	return nil
}

func configureImages(imageCMDFlags imageFlags, rootPath string, p *pkg.OperatorParams) error {
	component := common.ServingComponent
	if strings.EqualFold(imageCMDFlags.Component, common.EventingComponent) {
		component = common.EventingComponent
	}
	yamlTemplateString, err := common.GenerateOperatorCRString(component, imageCMDFlags.Namespace, p)
	if err != nil {
		return err
	}

	overlayContent := getOverlayYamlContentImage(rootPath, imageCMDFlags)
	valuesYaml := getYamlValuesContentImages(imageCMDFlags)
	if err := common.ApplyManifests(yamlTemplateString, overlayContent, valuesYaml, p); err != nil {
		return err
	}
	return nil
}

func getOverlayYamlContentImage(rootPath string, imageCMDFlags imageFlags) string {
	path := rootPath + "/overlay/ks_image.yaml"
	if strings.EqualFold(imageCMDFlags.Component, common.EventingComponent) {
		path = rootPath + "/overlay/ke_image.yaml"
	}
	baseOverlayContent, _ := common.ReadFile(path)
	resourceContent := getImageConfiguration(imageCMDFlags)
	baseOverlayContent = fmt.Sprintf("%s\n%s", baseOverlayContent, resourceContent)
	return baseOverlayContent
}

func getImageConfiguration(imageCMDFlags imageFlags) string {
	resourceArray := []string{}
	fieldFirstLine := "registry"
	fieldSecondLine := "override"
	imageKey := imageCMDFlags.ImageKey
	if imageCMDFlags.DeployName != "" && !strings.EqualFold(imageCMDFlags.ImageKey, "default") {
		imageKey = fmt.Sprintf("%s/%s", imageCMDFlags.DeployName, imageCMDFlags.ImageKey)
	}

	if strings.EqualFold(imageCMDFlags.ImageKey, "default") {
		tag := fmt.Sprintf("%s%s", common.Spaces(2), common.YttMatchingTag)
		resourceArray = append(resourceArray, tag)
		field := fmt.Sprintf("%s%s:", common.Spaces(2), "registry")
		resourceArray = append(resourceArray, field)
		tag = fmt.Sprintf("%s%s", common.Spaces(4), common.YttMatchingTag)
		resourceArray = append(resourceArray, tag)
		keyValueField := fmt.Sprintf("%s%s: %s", common.Spaces(4), imageCMDFlags.ImageKey, "#@ data.values.imageValue")
		resourceArray = append(resourceArray, keyValueField)
		return strings.Join(resourceArray, "\n")
	} else if strings.EqualFold(imageCMDFlags.ImageKey, "queue-sidecar-image") || imageCMDFlags.ImageKey == "queueSidecarImage" {
		fieldFirstLine = "config"
		fieldSecondLine = "deployment"
		imageKey = "queue-sidecar-image"
	}

	tag := fmt.Sprintf("%s%s", common.Spaces(2), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)
	field := fmt.Sprintf("%s%s:", common.Spaces(2), fieldFirstLine)
	resourceArray = append(resourceArray, field)
	tag = fmt.Sprintf("%s%s", common.Spaces(4), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)
	field = fmt.Sprintf("%s%s:", common.Spaces(4), fieldSecondLine)
	resourceArray = append(resourceArray, field)
	tag = fmt.Sprintf("%s%s", common.Spaces(6), common.YttMatchingTag)
	resourceArray = append(resourceArray, tag)
	keyValueField := fmt.Sprintf("%s%s: %s", common.Spaces(6), imageKey, "#@ data.values.imageValue")
	resourceArray = append(resourceArray, keyValueField)

	return strings.Join(resourceArray, "\n")
}

func getYamlValuesContentImages(imageCMDFlags imageFlags) string {
	contentArray := []string{}
	header := "#@data/values\n---"
	contentArray = append(contentArray, header)
	namespace := fmt.Sprintf("namespace: %s", imageCMDFlags.Namespace)
	contentArray = append(contentArray, namespace)
	deployName := fmt.Sprintf("deployName: %s", imageCMDFlags.DeployName)
	contentArray = append(contentArray, deployName)
	value := fmt.Sprintf("imageValue: %s", imageCMDFlags.ImageUrl)
	contentArray = append(contentArray, value)
	return strings.Join(contentArray, "\n")
}
