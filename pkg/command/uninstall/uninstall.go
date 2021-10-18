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
	"fmt"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"knative.dev/kn-plugin-operator/pkg"
	"os"
)

type uninstallCmdFlags struct {
	Component      string
	IstioNamespace string
	Namespace      string
	KubeConfig     string
	Version        string
}

var (
	uninstallFlags   uninstallCmdFlags
)

// installCmd represents the install commands for the operation
func NewUninstallCommand(p *pkg.OperatorParams) *cobra.Command {
	var installCmd = &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Knative Operator or Knative components",
		Example: `
  # Uninstall Knative Serving under the namespace knative-serving
  kn operation uninstall -c serving --namespace knative-serving`,

		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := p.NewKubeClient()
			if err != nil {
				return fmt.Errorf("cannot get source cluster kube config, please use --kubeconfig or export environment variable KUBECONFIG to set\n")
			}

			_, err = os.Getwd()
			if err != nil {
				return err
			}

			// TODO: Need to implement how to uninstall the operator and knative components.

			return nil
		},
	}

	installCmd.Flags().StringVarP(&uninstallFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")
	installCmd.Flags().StringVarP(&uninstallFlags.Component, "component", "c", "", "The name of the Knative Component to install")

	return installCmd
}