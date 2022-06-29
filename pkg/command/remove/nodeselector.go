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

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"k8s.io/client-go/util/retry"

	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/kn-plugin-operator/pkg/command/common"
	"knative.dev/operator/pkg/apis/operator/base"
)

var nodeSelectorFlags common.KeyValueFlags

// removeNodeSelectorCommand represents the configure commands to delete the node selector for Knative deployments
func removeNodeSelectorCommand(p *pkg.OperatorParams) *cobra.Command {
	var removeNodeSelectorsCmd = &cobra.Command{
		Use:   "nodeSelectors",
		Short: "Remove the node selectors for Knative Serving and Eventing deployments",
		Example: `
  # Remove the node selectors for Knative Serving and Eventing deployments
  kn operation remove nodeSelectors --component eventing --deployName eventing-controller --key key --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateNodeSelectorFlags(nodeSelectorFlags); err != nil {
				return err
			}

			err := deleteNodeSelectors(nodeSelectorFlags, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified node selector has been deleted in the namespace '%s'.\n",
				nodeSelectorFlags.Namespace)
			return nil
		},
	}

	removeNodeSelectorsCmd.Flags().StringVar(&nodeSelectorFlags.Key, "key", "", "The key of the data in the configmap")
	removeNodeSelectorsCmd.Flags().StringVar(&nodeSelectorFlags.DeployName, "deployName", "", "The flag to specify the deployment name")
	removeNodeSelectorsCmd.Flags().StringVarP(&nodeSelectorFlags.Component, "component", "c", "", "The flag to specify the component name")
	removeNodeSelectorsCmd.Flags().StringVarP(&nodeSelectorFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return removeNodeSelectorsCmd
}

func validateNodeSelectorFlags(keyValuesCMDFlags common.KeyValueFlags) error {
	if keyValuesCMDFlags.Component == "" {
		return fmt.Errorf("You need to specify the component for Knative.")
	}
	if keyValuesCMDFlags.Namespace == "" {
		return fmt.Errorf("You need to specify the namespace.")
	}
	if keyValuesCMDFlags.Component != "" && !strings.EqualFold(keyValuesCMDFlags.Component, common.ServingComponent) && !strings.EqualFold(keyValuesCMDFlags.Component, common.EventingComponent) {
		return fmt.Errorf("You need to specify the component for Knative: serving or eventing.")
	}
	if keyValuesCMDFlags.Key != "" {
		if keyValuesCMDFlags.DeployName == "" {
			return fmt.Errorf("You need to specify the name of the deployment.")
		}
	}
	return nil
}

func deleteNodeSelectors(nodeSelectorFlags common.KeyValueFlags, p *pkg.OperatorParams) error {
	ksCR, err := common.GetKnativeOperatorCR(p)
	if err != nil {
		return err
	}

	deploymentOverrides, err := ksCR.GetDeployments(nodeSelectorFlags.Component, nodeSelectorFlags.Namespace)
	if err != nil {
		return err
	}

	deploymentOverrides = removeNodeSelectorsDeployFields(deploymentOverrides, nodeSelectorFlags)
	if err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return ksCR.UpdateDeployments(nodeSelectorFlags.Component, nodeSelectorFlags.Namespace, deploymentOverrides)
	}); err != nil {
		return err
	}
	return nil
}

func removeNodeSelectorsDeployFields(deploymentOverrides []base.DeploymentOverride, nodeSelectorFlags common.KeyValueFlags) []base.DeploymentOverride {
	if nodeSelectorFlags.DeployName == "" {
		for i := range deploymentOverrides {
			deploymentOverrides[i].NodeSelector = nil
		}
	} else if nodeSelectorFlags.Key == "" {
		for i, deploy := range deploymentOverrides {
			if deploy.Name == nodeSelectorFlags.DeployName {
				deploymentOverrides[i].NodeSelector = nil
			}
		}
	} else if nodeSelectorFlags.Key != "" {
		for i, deploy := range deploymentOverrides {
			if deploy.Name == nodeSelectorFlags.DeployName {
				nodeSelector := make(map[string]string)
				for key, value := range deploymentOverrides[i].NodeSelector {
					if key != nodeSelectorFlags.Key {
						nodeSelector[key] = value
					}
				}
				deploymentOverrides[i].NodeSelector = nodeSelector
			}
		}
	}

	return deploymentOverrides
}
