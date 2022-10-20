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

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"k8s.io/client-go/util/retry"

	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/kn-plugin-operator/pkg/command/common"
	"knative.dev/operator/pkg/apis/operator/base"
)

type ResourcesFlags struct {
	Component  string
	Namespace  string
	Container  string
	DeployName string
}

var resourcesCMDFlags ResourcesFlags

// removeResourcesCommand represents the remove commands for the resources in Knative Serving or Eventing
func removeResourcesCommand(p *pkg.OperatorParams) *cobra.Command {
	var deleteResourcesCmd = &cobra.Command{
		Use:   "resources",
		Short: "Remove the resource for Knative Serving and Eventing deployments",
		Example: `
  # Remove the configuration of the resources for Knative Serving
  kn operation remove resources --component serving --namespace knative-serving
  # Remove the configuration of the resources for the container activator in the deployment activator
  kn operation remove resources --component serving --deployName activator --namespace knative-serving
  # Remove the configuration of the resources for the deployment activator
  kn operation remove resources --component serving --deployName activator --namespace knative-serving`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateResourcesFlags(resourcesCMDFlags); err != nil {
				return err
			}

			err := removeResources(resourcesCMDFlags, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified resources have been removed in the namespace '%s'.\n",
				resourcesCMDFlags.Namespace)
			return nil
		},
	}

	deleteResourcesCmd.Flags().StringVar(&resourcesCMDFlags.DeployName, "deployName", "", "The flag to specify the deployment name")
	deleteResourcesCmd.Flags().StringVarP(&resourcesCMDFlags.Component, "component", "c", "", "The flag to specify the component name")
	deleteResourcesCmd.Flags().StringVar(&resourcesCMDFlags.Container, "container", "", "The flag to specify the container name")
	deleteResourcesCmd.Flags().StringVarP(&resourcesCMDFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return deleteResourcesCmd
}

func validateResourcesFlags(resourcesCMDFlags ResourcesFlags) error {
	if resourcesCMDFlags.Component == "" {
		return fmt.Errorf("You need to specify the component name.")
	}
	if resourcesCMDFlags.Namespace == "" {
		return fmt.Errorf("You need to specify the namespace.")
	}
	if resourcesCMDFlags.Container != "" && resourcesCMDFlags.DeployName == "" {
		return fmt.Errorf("You need to specify the deployment name for the container.")
	}

	return nil
}

func removeResources(resourcesCMDFlags ResourcesFlags, p *pkg.OperatorParams) error {
	ksCR, err := common.GetKnativeOperatorCR(p)
	if err != nil {
		return err
	}

	workloadOverrides, err := ksCR.GetDeployments(resourcesCMDFlags.Component, resourcesCMDFlags.Namespace)
	if err != nil {
		return err
	}

	workloadOverrides = removeResourcesFields(workloadOverrides, resourcesCMDFlags)
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return ksCR.UpdateDeployments(resourcesCMDFlags.Component, resourcesCMDFlags.Namespace, workloadOverrides)
	})

	if err != nil {
		return err
	}

	return nil
}

func removeResourcesFields(workloadOverrides []base.WorkloadOverride, resourcesCMDFlags ResourcesFlags) []base.WorkloadOverride {
	if resourcesCMDFlags.DeployName == "" {
		// If no deploy is specified, we will iterate all the deployments to remove all resource configurations.
		for i := range workloadOverrides {
			workloadOverrides[i].Resources = nil
		}
	} else if resourcesCMDFlags.Container == "" {
		for i, deploy := range workloadOverrides {
			if deploy.Name == resourcesCMDFlags.DeployName {
				workloadOverrides[i].Resources = nil
			}
		}
	} else if resourcesCMDFlags.Container != "" {
		deployIndex := -1
		containerIndex := -1
		var resourceRequirementsOverrides []base.ResourceRequirementsOverride
		for i, deploy := range workloadOverrides {
			if deploy.Name == resourcesCMDFlags.DeployName {
				deployIndex = i
				for j, resource := range deploy.Resources {
					if resource.Container == resourcesCMDFlags.Container {
						containerIndex = j
						resourceRequirementsOverrides = deploy.Resources
						break
					}
				}
				break
			}
		}

		if containerIndex != -1 {
			modifiedOverrides := append(resourceRequirementsOverrides[:containerIndex], resourceRequirementsOverrides[containerIndex+1:]...)
			workloadOverrides[deployIndex].Resources = modifiedOverrides
		}
	}

	return workloadOverrides
}
