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

	"k8s.io/client-go/util/retry"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/kn-plugin-operator/pkg/command/common"
)

var annotationCMDFlags common.KeyValueFlags

// removeAnnotationCommand represents the configure commands to delete the annotations for Knative deployments or services
func removeAnnotationCommand(p *pkg.OperatorParams) *cobra.Command {
	var removeAnnotationsCmd = &cobra.Command{
		Use:   "annotations",
		Short: "Remove the annotations for Knative Serving and Eventing deployments or services",
		Example: `
  # Remove the annotations for Knative Serving and Eventing services
  kn operator remove annotations --component eventing --serviceName eventing-controller --key key --namespace knative-eventing
  # Remove the annotations for Knative Serving and Eventing deployments
  kn operator remove annotations --component eventing --deployName eventing-controller --key key --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateLabelAnnotationsFlags(annotationCMDFlags); err != nil {
				return err
			}

			err := deleteAnnotations(annotationCMDFlags, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified annotations has been configured in the namespace '%s'.\n",
				annotationCMDFlags.Namespace)
			return nil
		},
	}

	removeAnnotationsCmd.Flags().StringVar(&annotationCMDFlags.Key, "key", "", "The key of the data in the configmap")
	removeAnnotationsCmd.Flags().StringVar(&annotationCMDFlags.DeployName, "deployName", "", "The flag to specify the deployment name")
	removeAnnotationsCmd.Flags().StringVar(&annotationCMDFlags.ServiceName, "serviceName", "", "The flag to specify the service name")
	removeAnnotationsCmd.Flags().StringVarP(&annotationCMDFlags.Component, "component", "c", "", "The flag to specify the component name")
	removeAnnotationsCmd.Flags().StringVarP(&annotationCMDFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return removeAnnotationsCmd
}

func deleteAnnotations(annotationCMDFlags common.KeyValueFlags, p *pkg.OperatorParams) error {
	ksCR, err := common.GetKnativeOperatorCR(p)
	if err != nil {
		return err
	}

	if annotationCMDFlags.DeployName != "" {
		wordloads, err := ksCR.GetDeployments(annotationCMDFlags.Component, annotationCMDFlags.Namespace)
		if err != nil {
			return err
		}

		wordloads = removeAnnotationsDeployFields(wordloads, annotationCMDFlags)
		if err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return ksCR.UpdateDeployments(annotationCMDFlags.Component, annotationCMDFlags.Namespace, wordloads)
		}); err != nil {
			return err
		}
	} else if annotationCMDFlags.ServiceName != "" {
		serviceOverrides, err := ksCR.GetServices(annotationCMDFlags.Component, annotationCMDFlags.Namespace)
		if err != nil {
			return err
		}

		serviceOverrides = removeAnnotationsServiceFields(serviceOverrides, annotationCMDFlags)
		if err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return ksCR.UpdateServices(annotationCMDFlags.Component, annotationCMDFlags.Namespace, serviceOverrides)
		}); err != nil {
			return err
		}
	}
	return nil
}

func removeAnnotationsDeployFields(workloadOverrides []base.WorkloadOverride, annotationCMDFlags common.KeyValueFlags) []base.WorkloadOverride {
	if annotationCMDFlags.Key == "" {
		for i, deploy := range workloadOverrides {
			if deploy.Name == annotationCMDFlags.DeployName {
				workloadOverrides[i].Annotations = nil
			}
		}
	} else if annotationCMDFlags.Key != "" {
		for i, deploy := range workloadOverrides {
			if deploy.Name == annotationCMDFlags.DeployName {
				labels := make(map[string]string)
				for key, value := range workloadOverrides[i].Annotations {
					if key != annotationCMDFlags.Key {
						labels[key] = value
					}
				}
				workloadOverrides[i].Annotations = labels
			}
		}
	}

	return workloadOverrides
}

func removeAnnotationsServiceFields(serviceOverrides []base.ServiceOverride, annotationCMDFlags common.KeyValueFlags) []base.ServiceOverride {
	if annotationCMDFlags.Key == "" {
		for i, service := range serviceOverrides {
			if service.Name == annotationCMDFlags.ServiceName {
				serviceOverrides[i].Annotations = nil
			}
		}
	} else if annotationCMDFlags.Key != "" {
		for i, service := range serviceOverrides {
			if service.Name == annotationCMDFlags.ServiceName {
				labels := make(map[string]string)
				for key, value := range serviceOverrides[i].Annotations {
					if key != annotationCMDFlags.Key {
						labels[key] = value
					}
				}
				serviceOverrides[i].Annotations = labels
			}
		}
	}

	return serviceOverrides
}
