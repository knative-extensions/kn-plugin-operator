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

package install

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/kn-plugin-operator/pkg/command/common"
)

type installCmdFlags struct {
	Component      string
	IstioNamespace string
	Namespace      string
	KubeConfig     string
	Version        string
}

var (
	installFlags installCmdFlags
)

// installCmd represents the install commands for the operation
func NewInstallCommand(p *pkg.OperatorParams) *cobra.Command {
	var installCmd = &cobra.Command{
		Use:   "install",
		Short: "Install Knative Operator or Knative components",
		Example: `
  # Install Knative Serving under the namespace knative-serving
  kn operation install -c serving --namespace knative-serving`,

		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := p.NewKubeClient()
			if err != nil {
				return fmt.Errorf("cannot get source cluster kube config, please use --kubeconfig or export environment variable KUBECONFIG to set\n")
			}

			rootPath, err := os.Getwd()
			if err != nil {
				return err
			}

			URL, err := getOperatorURL(installFlags.Version)
			if err != nil {
				return err
			}

			ns := common.Namespace{
				Client:    client,
				Component: installFlags.Component,
			}
			if err = ns.CreateNamespace(installFlags.Namespace); err != nil {
				return err
			}

			yamlTemplateString, err := common.DownloadFile(URL)
			if err != nil {
				return err
			}

			operatorOverlay := rootPath + "/overlay/operator.yaml"
			overlayContent, err := common.ReadFile(operatorOverlay)
			if err != nil {
				return err
			}
			yamlValuesContent := fmt.Sprintf("#@data/values\n---\nnamespace: %s", installFlags.Namespace)

			yttp := common.YttProcessor{
				BaseData:    []byte(yamlTemplateString),
				OverlayData: []byte(overlayContent),
				ValuesData:  []byte(yamlValuesContent),
			}

			restConfig, err := p.RestConfig()
			if err != nil {
				return err
			}

			manifest := common.Manifest{
				YttPro:     &yttp,
				RestConfig: restConfig,
			}
			if err = manifest.Apply(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Knative operator of the '%s' version was created in the namespace '%s'.\n", installFlags.Version, installFlags.Namespace)
			return nil
		},
	}

	installCmd.Flags().StringVarP(&installFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")
	installCmd.Flags().StringVarP(&installFlags.Component, "component", "c", "", "The name of the Knative Component to install")
	installCmd.Flags().StringVarP(&installFlags.Version, "version", "v", "latest", "The version of the the Knative Operator or the Knative component")
	installCmd.Flags().StringVar(&installFlags.IstioNamespace, "istio-namespace", "", "The namespace of istio")

	return installCmd
}

func getOperatorURL(version string) (string, error) {
	versionSanitized := strings.ToLower(version)
	URL := "https://github.com/knative/operator/releases/latest/download/operator.yaml"
	if version != "latest" {
		if !strings.HasPrefix(version, "v") {
			versionSanitized = fmt.Sprintf("v%s", versionSanitized)
		}
		validity, major := common.GetMajor(versionSanitized)
		if !validity {
			return "", fmt.Errorf("%v is not a semantic version", version)
		}
		prefix := ""
		if semver.Compare(major, "v0") == 1 {
			prefix = "knative-"
		}
		URL = fmt.Sprintf("https://github.com/knative/operator/releases/download/%s%s/operator.yaml", prefix, versionSanitized)
	}
	return URL, nil
}
