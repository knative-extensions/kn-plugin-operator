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

	corev1 "k8s.io/api/core/v1"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"k8s.io/client-go/util/retry"
	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/kn-plugin-operator/pkg/command/common"
	"knative.dev/operator/pkg/apis/operator/base"
)

type TolerationsFlags struct {
	Key        string
	Operator   string
	Value      string
	Effect     string
	Component  string
	Namespace  string
	DeployName string
}

var tolerationsCMDFlags TolerationsFlags

// removeTolerationsCommand represents the remove commands for the tolerations in Knative Serving or Eventing
func removeTolerationsCommand(p *pkg.OperatorParams) *cobra.Command {
	var configureTolerationsCmd = &cobra.Command{
		Use:   "tolerations",
		Short: "Remove the tolerations for Knative Serving and Eventing deployments",
		Example: `
  # Remove the tolerations for Knative Serving and Eventing deployments
  kn operation remove tolerations --component eventing --deployName eventing-controller --key example-key --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateTolerationsFlags(tolerationsCMDFlags); err != nil {
				return err
			}

			err := deleteTolerations(tolerationsCMDFlags, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified tolerations have been deleted in the namespace '%s'.\n",
				tolerationsCMDFlags.Namespace)
			return nil
		},
	}

	configureTolerationsCmd.Flags().StringVar(&tolerationsCMDFlags.Key, "key", "", "The flag to specify the key")
	configureTolerationsCmd.Flags().StringVar(&tolerationsCMDFlags.DeployName, "deployName", "", "The flag to specify the deployment name")
	configureTolerationsCmd.Flags().StringVarP(&tolerationsCMDFlags.Component, "component", "c", "", "The flag to specify the component name")
	configureTolerationsCmd.Flags().StringVarP(&tolerationsCMDFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return configureTolerationsCmd
}

func validateTolerationsFlags(tolerationsCMDFlags TolerationsFlags) error {
	if tolerationsCMDFlags.Component == "" {
		return fmt.Errorf("You need to specify the component name.")
	}
	if tolerationsCMDFlags.Namespace == "" {
		return fmt.Errorf("You need to specify the namespace.")
	}
	if tolerationsCMDFlags.Key != "" && tolerationsCMDFlags.DeployName == "" {
		return fmt.Errorf("You need to specify the deployment name for the toleration.")
	}

	return nil
}

func deleteTolerations(tolerationsCMDFlags TolerationsFlags, p *pkg.OperatorParams) error {
	ksCR, err := common.GetKnativeOperatorCR(p)
	if err != nil {
		return err
	}

	workloadOverrides, err := ksCR.GetDeployments(tolerationsCMDFlags.Component, tolerationsCMDFlags.Namespace)
	if err != nil {
		return err
	}

	workloadOverrides = removeTolerationsFields(workloadOverrides, tolerationsCMDFlags)
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return ksCR.UpdateDeployments(tolerationsCMDFlags.Component, tolerationsCMDFlags.Namespace, workloadOverrides)
	})

	if err != nil {
		return err
	}

	return nil
}

func removeTolerationsFields(workloadOverrides []base.WorkloadOverride, tolerationsCMDFlags TolerationsFlags) []base.WorkloadOverride {
	if tolerationsCMDFlags.DeployName == "" {
		// If no deploy is specified, we will iterate all the deployments to remove all toleration configurations.
		for i := range workloadOverrides {
			workloadOverrides[i].Tolerations = nil
		}
	} else if tolerationsCMDFlags.Key == "" {
		for i, deploy := range workloadOverrides {
			if deploy.Name == tolerationsCMDFlags.DeployName {
				workloadOverrides[i].Tolerations = nil
			}
		}
	} else if tolerationsCMDFlags.Key != "" {
		deployIndex := -1
		var tolerations []corev1.Toleration
		for i, deploy := range workloadOverrides {
			if deploy.Name == tolerationsCMDFlags.DeployName {
				tolerationsBack := make([]corev1.Toleration, 0, len(deploy.Tolerations))
				deployIndex = i
				for _, toleration := range deploy.Tolerations {
					if toleration.Key != tolerationsCMDFlags.Key {
						tolerationsBack = append(tolerationsBack, toleration)
					}
				}
				tolerations = tolerationsBack
				break
			}
		}

		if deployIndex != -1 {
			workloadOverrides[deployIndex].Tolerations = tolerations
		}
	}

	return workloadOverrides
}
