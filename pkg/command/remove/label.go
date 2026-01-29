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

var deploymentLabelCMDFlags common.KeyValueFlags

// removeLabelCommand represents the configure commands to delete the labels for Knative deployments or services
func removeLabelCommand(p *pkg.OperatorParams) *cobra.Command {
	var removeLabelsCmd = &cobra.Command{
		Use:   "labels",
		Short: "Remove the labels for Knative Serving and Eventing deployments or services",
		Example: `
  # Remove the labels for Knative Serving and Eventing services
  kn operator remove labels --component eventing --serviceName eventing-controller --key key --namespace knative-eventing
  # Remove the labels for Knative Serving and Eventing deployments
  kn operator remove labels --component eventing --deployName eventing-controller --key key --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateLabelAnnotationsFlags(deploymentLabelCMDFlags); err != nil {
				return err
			}

			err := deleteLabels(deploymentLabelCMDFlags, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified labels has been configured in the namespace '%s'.\n",
				deploymentLabelCMDFlags.Namespace)
			return nil
		},
	}

	removeLabelsCmd.Flags().StringVar(&deploymentLabelCMDFlags.Key, "key", "", "The key of the data in the configmap")
	removeLabelsCmd.Flags().StringVar(&deploymentLabelCMDFlags.DeployName, "deployName", "", "The flag to specify the deployment name")
	removeLabelsCmd.Flags().StringVar(&deploymentLabelCMDFlags.ServiceName, "serviceName", "", "The flag to specify the service name")
	removeLabelsCmd.Flags().StringVarP(&deploymentLabelCMDFlags.Component, "component", "c", "", "The flag to specify the component name")
	removeLabelsCmd.Flags().StringVarP(&deploymentLabelCMDFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return removeLabelsCmd
}

func validateLabelAnnotationsFlags(deploymentLabelCMDFlags common.KeyValueFlags) error {
	if deploymentLabelCMDFlags.Namespace == "" {
		return fmt.Errorf("You need to specify the namespace.")
	}
	if deploymentLabelCMDFlags.Component == "" {
		return fmt.Errorf("You need to specify the component for Knative.")
	}
	if deploymentLabelCMDFlags.Component != "" && !strings.EqualFold(deploymentLabelCMDFlags.Component, common.ServingComponent) && !strings.EqualFold(deploymentLabelCMDFlags.Component, common.EventingComponent) {
		return fmt.Errorf("You need to specify the component for Knative: serving or eventing.")
	}

	if deploymentLabelCMDFlags.DeployName == "" && deploymentLabelCMDFlags.ServiceName == "" {
		return fmt.Errorf("You need to specify the name of the deployment or the service.")
	}

	if deploymentLabelCMDFlags.DeployName != "" && deploymentLabelCMDFlags.ServiceName != "" {
		return fmt.Errorf("You are only allowed to specify either --deployName or --serviceName.")
	}

	return nil
}

func deleteLabels(labelCMDFlags common.KeyValueFlags, p *pkg.OperatorParams) error {
	ksCR, err := common.GetKnativeOperatorCR(p)
	if err != nil {
		return err
	}

	if labelCMDFlags.DeployName != "" {
		workloadOverrides, err := ksCR.GetDeployments(labelCMDFlags.Component, labelCMDFlags.Namespace)
		if err != nil {
			return err
		}

		workloadOverrides = removeLabelsDeployFields(workloadOverrides, labelCMDFlags)
		if err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return ksCR.UpdateDeployments(labelCMDFlags.Component, labelCMDFlags.Namespace, workloadOverrides)
		}); err != nil {
			return err
		}
	} else if labelCMDFlags.ServiceName != "" {
		serviceOverrides, err := ksCR.GetServices(labelCMDFlags.Component, labelCMDFlags.Namespace)
		if err != nil {
			return err
		}

		serviceOverrides = removeLabelsServiceFields(serviceOverrides, labelCMDFlags)
		if err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return ksCR.UpdateServices(labelCMDFlags.Component, labelCMDFlags.Namespace, serviceOverrides)
		}); err != nil {
			return err
		}
	}
	return nil
}

func removeLabelsDeployFields(workOverrides []base.WorkloadOverride, labelCMDFlags common.KeyValueFlags) []base.WorkloadOverride {
	if labelCMDFlags.Key == "" {
		for i, deploy := range workOverrides {
			if deploy.Name == labelCMDFlags.DeployName {
				workOverrides[i].Labels = nil
			}
		}
	} else if labelCMDFlags.Key != "" {
		for i, deploy := range workOverrides {
			if deploy.Name == labelCMDFlags.DeployName {
				labels := make(map[string]string)
				for key, value := range workOverrides[i].Labels {
					if key != labelCMDFlags.Key {
						labels[key] = value
					}
				}
				workOverrides[i].Labels = labels
			}
		}
	}

	return workOverrides
}

func removeLabelsServiceFields(serviceOverrides []base.ServiceOverride, labelCMDFlags common.KeyValueFlags) []base.ServiceOverride {
	if labelCMDFlags.Key == "" {
		for i, service := range serviceOverrides {
			if service.Name == labelCMDFlags.ServiceName {
				serviceOverrides[i].Labels = nil
			}
		}
	} else if labelCMDFlags.Key != "" {
		for i, service := range serviceOverrides {
			if service.Name == labelCMDFlags.ServiceName {
				labels := make(map[string]string)
				for key, value := range serviceOverrides[i].Labels {
					if key != labelCMDFlags.Key {
						labels[key] = value
					}
				}
				serviceOverrides[i].Labels = labels
			}
		}
	}

	return serviceOverrides
}
