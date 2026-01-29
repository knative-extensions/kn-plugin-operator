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

	"knative.dev/operator/pkg/apis/operator/base"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"k8s.io/client-go/util/retry"
	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/kn-plugin-operator/pkg/command/common"
)

type EnvVarFlags struct {
	EnvName       string
	EnvValue      string
	Component     string
	Namespace     string
	DeployName    string
	ContainerName string
}

var envVarFlags EnvVarFlags

// removeEnvVarCommand represents the configure commands to delete the env vars configuration for Knative Deployment resources
func removeEnvVarCommand(p *pkg.OperatorParams) *cobra.Command {
	var removeEnvVarsCmd = &cobra.Command{
		Use:   "envvars",
		Short: "Delete the env vars for Knative",
		Example: `
  # Delete the env vars for Knative
  kn operator remove envvars --component eventing --deployName eventing-controller --container eventing-controller --name key --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateEnvVarsFlags(envVarFlags); err != nil {
				return err
			}

			err := removeEnvVars(envVarFlags, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified environment variable has been deleted.\n")
			return nil
		},
	}

	removeEnvVarsCmd.Flags().StringVar(&envVarFlags.EnvName, "name", "", "The name for the environment variable")
	removeEnvVarsCmd.Flags().StringVar(&envVarFlags.DeployName, "deployName", "", "The flag to specify the deployment name")
	removeEnvVarsCmd.Flags().StringVarP(&envVarFlags.Component, "component", "c", "", "The flag to specify the component name")
	removeEnvVarsCmd.Flags().StringVarP(&envVarFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")
	removeEnvVarsCmd.Flags().StringVar(&envVarFlags.ContainerName, "container", "", "The name of the container")

	return removeEnvVarsCmd
}

func validateEnvVarsFlags(envVarFlags EnvVarFlags) error {
	if envVarFlags.Namespace == "" {
		return fmt.Errorf("You need to specify the namespace.")
	}

	if envVarFlags.Component == "" {
		return fmt.Errorf("You need to specify the component name.")
	}

	if envVarFlags.DeployName == "" && envVarFlags.ContainerName != "" {
		return fmt.Errorf("You need to specify the name for the deployment resource.")
	}

	if envVarFlags.EnvName != "" {
		if envVarFlags.DeployName == "" {
			return fmt.Errorf("You need to specify the name for the deployment resource.")
		}
		if envVarFlags.ContainerName == "" {
			return fmt.Errorf("You need to specify the name for the container.")
		}
	}
	return nil
}

func removeEnvVars(envVarFlags EnvVarFlags, p *pkg.OperatorParams) error {
	ksCR, err := common.GetKnativeOperatorCR(p)
	if err != nil {
		return err
	}

	workloadOverrides, err := ksCR.GetDeployments(envVarFlags.Component, envVarFlags.Namespace)
	if err != nil {
		return err
	}

	workloadOverrides = removeEnvVarsFields(workloadOverrides, envVarFlags)

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return ksCR.UpdateDeployments(envVarFlags.Component, envVarFlags.Namespace, workloadOverrides)
	})

	if err != nil {
		return err
	}

	return nil
}

func removeEnvVarsFields(workloadOverrides []base.WorkloadOverride, envVarFlags EnvVarFlags) []base.WorkloadOverride {
	if envVarFlags.DeployName == "" {
		for i := range workloadOverrides {
			workloadOverrides[i].Env = nil
		}
	} else {
		if envVarFlags.ContainerName == "" {
			for i := range workloadOverrides {
				if workloadOverrides[i].Name == envVarFlags.DeployName {
					workloadOverrides[i].Env = nil
				}
			}
		} else {
			if envVarFlags.EnvName == "" {
				for i := range workloadOverrides {
					if workloadOverrides[i].Name == envVarFlags.DeployName {
						env := make([]base.EnvRequirementsOverride, 0, len(workloadOverrides[i].Env))
						for _, val := range workloadOverrides[i].Env {
							if val.Container != envVarFlags.ContainerName {
								env = append(env, val)
							}
						}
						workloadOverrides[i].Env = env
					}
				}
			} else {
				for i := range workloadOverrides {
					if workloadOverrides[i].Name == envVarFlags.DeployName {
						env := make([]base.EnvRequirementsOverride, 0, len(workloadOverrides[i].Env))
						for _, val := range workloadOverrides[i].Env {
							if val.Container == envVarFlags.ContainerName {
								vars := make([]corev1.EnvVar, 0, len(val.EnvVars))
								for _, envVal := range val.EnvVars {
									if envVal.Name != envVarFlags.EnvName {
										vars = append(vars, envVal)
									}
								}
								val.EnvVars = vars
							}
							env = append(env, val)
						}
						workloadOverrides[i].Env = env
					}
				}
			}
		}
	}

	return workloadOverrides
}
