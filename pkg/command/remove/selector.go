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

var selectorFlags common.KeyValueFlags

// removeSelectorCommand represents the configure commands to delete the selector for Knative services
func removeSelectorCommand(p *pkg.OperatorParams) *cobra.Command {
	var removeNodeSelectorsCmd = &cobra.Command{
		Use:   "selectors",
		Short: "Remove the selectors for Knative Serving and Eventing service",
		Example: `
  # Remove the selectors for Knative Serving and Eventing services
  kn operation remove selectors --component eventing --serviceName eventing-controller --key key --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateSelectorFlags(selectorFlags); err != nil {
				return err
			}

			err := deleteSelectors(selectorFlags, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified selector has been deleted in the namespace '%s'.\n",
				selectorFlags.Namespace)
			return nil
		},
	}

	removeNodeSelectorsCmd.Flags().StringVar(&selectorFlags.Key, "key", "", "The key of the data in the configmap")
	removeNodeSelectorsCmd.Flags().StringVar(&selectorFlags.ServiceName, "serviceName", "", "The flag to specify the service name")
	removeNodeSelectorsCmd.Flags().StringVarP(&selectorFlags.Component, "component", "c", "", "The flag to specify the component name")
	removeNodeSelectorsCmd.Flags().StringVarP(&selectorFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return removeNodeSelectorsCmd
}

func validateSelectorFlags(keyValuesCMDFlags common.KeyValueFlags) error {
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
		if keyValuesCMDFlags.ServiceName == "" {
			return fmt.Errorf("You need to specify the name of the service.")
		}
	}
	return nil
}

func deleteSelectors(selectorFlags common.KeyValueFlags, p *pkg.OperatorParams) error {
	ksCR, err := common.GetKnativeOperatorCR(p)
	if err != nil {
		return err
	}

	serviceOverrides, err := ksCR.GetServices(selectorFlags.Component, selectorFlags.Namespace)
	if err != nil {
		return err
	}

	serviceOverrides = removeSelectorsServiceFields(serviceOverrides, selectorFlags)
	if err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return ksCR.UpdateServices(selectorFlags.Component, selectorFlags.Namespace, serviceOverrides)
	}); err != nil {
		return err
	}
	return nil
}

func removeSelectorsServiceFields(serviceOverrides []base.ServiceOverride, selectorFlags common.KeyValueFlags) []base.ServiceOverride {
	if selectorFlags.ServiceName == "" {
		for i := range serviceOverrides {
			serviceOverrides[i].Selector = nil
		}
	} else if selectorFlags.Key == "" {
		for i, service := range serviceOverrides {
			if service.Name == selectorFlags.ServiceName {
				serviceOverrides[i].Selector = nil
			}
		}
	} else if selectorFlags.Key != "" {
		for i, service := range serviceOverrides {
			if service.Name == selectorFlags.ServiceName {
				selector := make(map[string]string)
				for key, value := range serviceOverrides[i].Selector {
					if key != selectorFlags.Key {
						selector[key] = value
					}
				}
				serviceOverrides[i].Selector = selector
			}
		}
	}

	return serviceOverrides
}
