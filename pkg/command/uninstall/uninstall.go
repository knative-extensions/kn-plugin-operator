// Copyright 2021 The Knative Authors
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

package uninstall

import (
	"context"
	"fmt"
	"strings"

	"knative.dev/kn-plugin-operator/pkg/command/common"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"knative.dev/kn-plugin-operator/pkg"
)

type uninstallCmdFlags struct {
	Component      string
	Namespace      string
	CRName         string
	KubeConfig     string
	CRNameExplicit bool
}

var (
	uninstallFlags uninstallCmdFlags
)

// installCmd represents the install commands for the operation
func NewUninstallCommand(p *pkg.OperatorParams) *cobra.Command {
	var uninstallCmd = &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Knative Operator or Knative components",
		Example: `
  # Uninstall Knative Serving under the namespace knative-serving
  kn operator uninstall -c serving --namespace knative-serving`,

		RunE: func(cmd *cobra.Command, args []string) error {
			uninstallFlags.CRNameExplicit = cmd.Flags().Changed(common.CRNameFlag)
			if strings.ToLower(uninstallFlags.Component) == common.ServingComponent {
				if err := validateComponentCRName(&uninstallFlags); err != nil {
					return err
				}
				// Uninstall the serving
				if err := uninstallKnativeServing(uninstallFlags, p); err != nil {
					return err
				}
				if uninstallFlags.CRNameExplicit {
					fmt.Fprintf(cmd.OutOrStdout(), "Knative Serving custom resource '%s/%s' was removed.\n", uninstallFlags.Namespace, uninstallFlags.CRName)
					return nil
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Knative Serving was removed in the namespace '%s'.\n", uninstallFlags.Namespace)
			} else if strings.ToLower(uninstallFlags.Component) == common.EventingComponent {
				if err := validateComponentCRName(&uninstallFlags); err != nil {
					return err
				}
				// Uninstall the eventing
				if err := uninstallKnativeEventing(uninstallFlags, p); err != nil {
					return err
				}
				if uninstallFlags.CRNameExplicit {
					fmt.Fprintf(cmd.OutOrStdout(), "Knative Eventing custom resource '%s/%s' was removed.\n", uninstallFlags.Namespace, uninstallFlags.CRName)
					return nil
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Knative Eventing was removed in the namespace '%s'.\n", uninstallFlags.Namespace)
			} else if uninstallFlags.Component != "" {
				if uninstallFlags.CRNameExplicit {
					return fmt.Errorf("--%s requires --component serving or --component eventing", common.CRNameFlag)
				}
				return fmt.Errorf("Unknown component name: you need to set component name to serving or eventing.")
			} else {
				if uninstallFlags.CRNameExplicit {
					return fmt.Errorf("--%s is not valid for Operator uninstall", common.CRNameFlag)
				}
				// Uninstall the Knative Operator
				if err := uninstallOperator(uninstallFlags, p); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Knative operator was removed in the namespace '%s'.\n", uninstallFlags.Namespace)
			}

			return nil
		},
	}

	uninstallCmd.Flags().StringVarP(&uninstallFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")
	uninstallCmd.Flags().StringVarP(&uninstallFlags.Component, "component", "c", "", "The name of the Knative Component to install")
	uninstallCmd.Flags().StringVar(&uninstallFlags.CRName, common.CRNameFlag, "", "The name of the hub Knative Serving or Eventing custom resource to uninstall")

	return uninstallCmd
}

func validateComponentCRName(uninstallFlags *uninstallCmdFlags) error {
	if !uninstallFlags.CRNameExplicit {
		return nil
	}
	name, err := common.NormalizeExplicitComponentName(uninstallFlags.Component, uninstallFlags.CRName)
	if err != nil {
		return err
	}
	uninstallFlags.CRName = name
	return nil
}

func uninstallKnativeServing(uninstallFlags uninstallCmdFlags, p *pkg.OperatorParams) error {
	operatorClient, err := p.NewOperatorClient()
	if err != nil {
		return fmt.Errorf("cannot get source cluster kube config, please use --kubeconfig or export environment variable KUBECONFIG to set\n")
	}

	if uninstallFlags.CRNameExplicit {
		name, err := common.NormalizeExplicitComponentName(uninstallFlags.Component, uninstallFlags.CRName)
		if err != nil {
			return err
		}
		return operatorClient.OperatorV1beta1().KnativeServings(uninstallFlags.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	}

	list, err := operatorClient.OperatorV1beta1().KnativeServings(uninstallFlags.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	var errstrings []string
	for _, ks := range list.Items {
		if err = operatorClient.OperatorV1beta1().KnativeServings(uninstallFlags.Namespace).Delete(context.TODO(),
			ks.Name, metav1.DeleteOptions{}); err != nil {
			errstrings = append(errstrings, err.Error())
		}
	}

	if len(errstrings) != 0 {
		return fmt.Errorf("%s", strings.Join(errstrings, "\n"))
	}
	return nil
}

func uninstallKnativeEventing(uninstallFlags uninstallCmdFlags, p *pkg.OperatorParams) error {
	operatorClient, err := p.NewOperatorClient()
	if err != nil {
		return fmt.Errorf("cannot get source cluster kube config, please use --kubeconfig or export environment variable KUBECONFIG to set\n")
	}

	if uninstallFlags.CRNameExplicit {
		name, err := common.NormalizeExplicitComponentName(uninstallFlags.Component, uninstallFlags.CRName)
		if err != nil {
			return err
		}
		return operatorClient.OperatorV1beta1().KnativeEventings(uninstallFlags.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	}

	list, err := operatorClient.OperatorV1beta1().KnativeEventings(uninstallFlags.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	var errstrings []string
	for _, ke := range list.Items {
		if err = operatorClient.OperatorV1beta1().KnativeEventings(uninstallFlags.Namespace).Delete(context.TODO(),
			ke.Name, metav1.DeleteOptions{}); err != nil {
			errstrings = append(errstrings, err.Error())
		}
	}

	if len(errstrings) != 0 {
		return fmt.Errorf("%s", strings.Join(errstrings, "\n"))
	}
	return nil
}

func uninstallOperator(uninstallFlags uninstallCmdFlags, p *pkg.OperatorParams) error {
	client, err := p.NewKubeClient()
	if err != nil {
		return fmt.Errorf("cannot get source cluster kube config, please use --kubeconfig or export environment variable KUBECONFIG to set\n")
	}

	return client.AppsV1().Deployments(uninstallFlags.Namespace).Delete(context.TODO(), common.KnativeOperatorName, metav1.DeleteOptions{})
}
