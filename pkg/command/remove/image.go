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
	"strings"

	"k8s.io/client-go/util/retry"
	"knative.dev/operator/pkg/apis/operator/base"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/kn-plugin-operator/pkg/command/common"
)

type ImageFlags struct {
	ImageUrl   string
	Component  string
	Namespace  string
	DeployName string
	ImageKey   string
}

var imageCMDFlags ImageFlags

// removeImageCommand represents the configure commands to delete the image configuration for Knative
func removeImageCommand(p *pkg.OperatorParams) *cobra.Command {
	var removeImagesCmd = &cobra.Command{
		Use:   "images",
		Short: "Remove the images for Knative",
		Example: `
  # Delete the images for Knative
  kn operator remove images --component eventing --deployName eventing-controller --imageKey key --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateImagesFlags(imageCMDFlags); err != nil {
				return err
			}

			err := removeImages(imageCMDFlags, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified images has been removed.\n")
			return nil
		},
	}

	removeImagesCmd.Flags().StringVar(&imageCMDFlags.ImageKey, "imageKey", "", "The image key")
	removeImagesCmd.Flags().StringVar(&imageCMDFlags.DeployName, "deployName", "", "The flag to specify the deployment name")
	removeImagesCmd.Flags().StringVarP(&imageCMDFlags.Component, "component", "c", "", "The flag to specify the component name")
	removeImagesCmd.Flags().StringVarP(&imageCMDFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return removeImagesCmd
}

func validateImagesFlags(imageCMDFlags ImageFlags) error {
	if imageCMDFlags.Namespace == "" {
		return fmt.Errorf("You need to specify the namespace.")
	}
	if imageCMDFlags.Component == "" {
		return fmt.Errorf("You need to specify the component for Knative: serving or eventing.")
	}
	if imageCMDFlags.ImageKey == "" && imageCMDFlags.DeployName == "" {
		return fmt.Errorf("You need to specify the image name or the name of the deployment for the image configuration.")
	}
	return nil
}

func removeImages(imageCMDFlags ImageFlags, p *pkg.OperatorParams) error {
	ksCR, err := common.GetKnativeOperatorCR(p)
	if err != nil {
		return err
	}

	registry, err := ksCR.GetRegistry(imageCMDFlags.Component, imageCMDFlags.Namespace)
	if err != nil {
		return err
	}

	registry = removeImagesFields(registry, imageCMDFlags)
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return ksCR.UpdateRegistry(imageCMDFlags.Component, imageCMDFlags.Namespace, registry)
	})

	if err != nil {
		return err
	}
	return nil
}

func removeImagesFields(registry base.Registry, imageCMDFlags ImageFlags) base.Registry {
	override := make(map[string]string)
	if imageCMDFlags.DeployName == "" {
		if imageCMDFlags.ImageKey == "default" {
			registry.Default = ""
		} else {
			for key, element := range registry.Override {
				suffix := fmt.Sprintf("/%s", imageCMDFlags.ImageKey)
				if key != imageCMDFlags.ImageKey && !strings.HasSuffix(key, suffix) {
					override[key] = element
				}
			}
			registry.Override = override
		}
	} else if imageCMDFlags.ImageKey == "" {
		for key, element := range registry.Override {
			prefix := fmt.Sprintf("%s/", imageCMDFlags.DeployName)
			if !strings.HasPrefix(key, prefix) {
				override[key] = element
			}
		}
		registry.Override = override
	} else if imageCMDFlags.ImageKey != "" {
		for key, element := range registry.Override {
			imageKey := fmt.Sprintf("%s/%s", imageCMDFlags.DeployName, imageCMDFlags.ImageKey)
			if key != imageKey {
				override[key] = element
			}
		}
		registry.Override = override
	}

	return registry
}
