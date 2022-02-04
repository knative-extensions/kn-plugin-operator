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

package enable

import (
	"fmt"
	"os"

	"knative.dev/kn-plugin-operator/pkg/command/common"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"knative.dev/kn-plugin-operator/pkg"
)

type ingressFlags struct {
	Istio     bool
	Kourier   bool
	Contour   bool
	Namespace string
}

var ingressCmdFlags ingressFlags

// newIngressCommand represents the enable commands for sources or ingresses
func newIngressCommand(p *pkg.OperatorParams) *cobra.Command {
	var enableIngressCmd = &cobra.Command{
		Use:   "ingress",
		Short: "Enable the ingress for Knative Serving",
		Example: `
  # Enable the ingress istio for Knative Serving
  kn operation enable ingress --istio --namespace knative-serving`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := validateIngressFlags(ingressCmdFlags)
			if err != nil {
				return err
			}

			if ingressCmdFlags.Namespace == "" {
				ingressCmdFlags.Namespace = common.DefaultKnativeServingNamespace
			}

			rootPath, err := os.Getwd()
			if err != nil {
				return err
			}

			err = enableIngress(ingressCmdFlags, rootPath, p)
			if err != nil {
				return err
			}

			ingress := "istio"
			if ingressCmdFlags.Kourier {
				ingress = "Kourier"
			}

			if ingressCmdFlags.Contour {
				ingress = "Contour"
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The ingress %s was enabled in the namespace '%s'.\n", ingress, ingressCmdFlags.Namespace)
			return nil
		},
	}

	enableIngressCmd.Flags().BoolVar(&ingressCmdFlags.Istio, "istio", false, "The flag to enable the ingress istio")
	enableIngressCmd.Flags().BoolVar(&ingressCmdFlags.Kourier, "kourier", false, "The flag to enable the ingress kourier")
	enableIngressCmd.Flags().BoolVar(&ingressCmdFlags.Contour, "contour", false, "The flag to enable the ingress contour")
	enableIngressCmd.Flags().StringVarP(&ingressCmdFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return enableIngressCmd
}

func validateIngressFlags(ingressCMDFlags ingressFlags) error {
	count := 0

	if ingressCMDFlags.Istio {
		count++
	}

	if ingressCMDFlags.Kourier {
		count++
	}

	if ingressCMDFlags.Contour {
		count++
	}

	if count == 0 {
		return fmt.Errorf("You need to enable at least one ingress for Knative Serving.")
	}
	if count > 1 {
		return fmt.Errorf("You can specify only one ingress for Knative Serving.")
	}
	return nil
}

func enableIngress(ingressCMDFlags ingressFlags, rootPath string, p *pkg.OperatorParams) error {
	// Generate the CR template
	yamlTemplateString, err := common.GenerateOperatorCRString(common.ServingComponent, ingressCMDFlags.Namespace, p)
	if err != nil {
		return err
	}

	overlayContent := getOverlayYamlContent(rootPath)
	valuesYaml := getYamlValuesContent(ingressCMDFlags)

	if err := common.ApplyManifests(yamlTemplateString, overlayContent, valuesYaml, p); err != nil {
		return err
	}
	return nil
}

func getOverlayYamlContent(rootPath string) string {
	path := rootPath + "/overlay/ks_ingress.yaml"
	overlayContent, _ := common.ReadFile(path)
	return overlayContent
}

func getYamlValuesContent(ingressCMDFlags ingressFlags) string {
	ingressClass := "istio.ingress.networking.knative.dev"
	if ingressCMDFlags.Kourier {
		ingressClass = "kourier.ingress.networking.knative.dev"
	}

	if ingressCMDFlags.Contour {
		ingressClass = "contour.ingress.networking.knative.dev"
	}

	content := fmt.Sprintf("#@data/values\n---\nnamespace: %s\nkourier: %t\nistio: %t\ncontour: %t\ningressClass: %s",
		ingressCMDFlags.Namespace, ingressCMDFlags.Kourier, ingressCMDFlags.Istio, ingressCMDFlags.Contour, ingressClass)

	return content
}
