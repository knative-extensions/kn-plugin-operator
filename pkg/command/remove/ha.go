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

	"knative.dev/operator/pkg/apis/operator/base"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"k8s.io/client-go/util/retry"
	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/kn-plugin-operator/pkg/command/common"
)

type HAFlags struct {
	Replicas   string
	Component  string
	Namespace  string
	DeployName string
}

var haCMDFlags HAFlags

// removeHACommand represents the HA deletion commands for Serving or Eventing
func removeHACommand(p *pkg.OperatorParams) *cobra.Command {
	var removeHAsCmd = &cobra.Command{
		Use:   "replicas",
		Short: "Remove the replica configuration for Knative Serving and Eventing deployments",
		Example: `
  # Remove the replica configuration for Knative Serving and Eventing deployments
  kn operation remove replicas --component eventing --deployName eventing-controller --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateHAsFlags(haCMDFlags); err != nil {
				return err
			}

			err := removeHAs(haCMDFlags, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified replicas configiuration has been removed in the namespace '%s'.\n",
				haCMDFlags.Namespace)
			return nil
		},
	}

	removeHAsCmd.Flags().StringVar(&haCMDFlags.DeployName, "deployName", "", "The flag to specify the deployment name")
	removeHAsCmd.Flags().StringVarP(&haCMDFlags.Component, "component", "c", "", "The flag to specify the component name")
	removeHAsCmd.Flags().StringVarP(&haCMDFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return removeHAsCmd
}

func validateHAsFlags(haCMDFlags HAFlags) error {
	if haCMDFlags.Component == "" {
		return fmt.Errorf("You need to specify the component name.")
	}
	if haCMDFlags.Namespace == "" {
		return fmt.Errorf("You need to specify the namespace.")
	}
	return nil
}

func removeHAs(haCMDFlags HAFlags, p *pkg.OperatorParams) error {
	ksCR, err := common.GetKnativeOperatorCR(p)
	if err != nil {
		return err
	}

	commonSpec, err := ksCR.GetCommonSpec(haCMDFlags.Component, haCMDFlags.Namespace)
	if err != nil {
		return err
	}

	commonSpec = removeReplicasFields(commonSpec, haCMDFlags)
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return ksCR.UpdateCommonSpec(haCMDFlags.Component, haCMDFlags.Namespace, commonSpec)
	})

	if err != nil {
		return err
	}

	return nil
}

func removeReplicasFields(commonSpec *base.CommonSpec, haCMDFlags HAFlags) *base.CommonSpec {
	if haCMDFlags.DeployName == "" {
		for i := range commonSpec.DeploymentOverride {
			commonSpec.DeploymentOverride[i].Replicas = nil
		}
		commonSpec.HighAvailability = nil
	} else {
		for i := range commonSpec.DeploymentOverride {
			if commonSpec.DeploymentOverride[i].Name == haCMDFlags.DeployName {
				commonSpec.DeploymentOverride[i].Replicas = nil
			}
		}
	}

	return commonSpec
}
