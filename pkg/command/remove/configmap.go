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

	"k8s.io/client-go/util/retry"
	"knative.dev/operator/pkg/apis/operator/base"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/kn-plugin-operator/pkg/command/common"
)

var cmsCMDFlags common.CMsFlags

// removeConfigMapsCommand represents the remove commands to delete the ConfigMaps in Knative Serving or Eventing
func removeConfigMapsCommand(p *pkg.OperatorParams) *cobra.Command {
	var configureCMsCmd = &cobra.Command{
		Use:   "configmaps",
		Short: "Delete the configmap configurations for Knative Serving and Eventing deployments",
		Example: `
  # Delete the CM for Knative Serving and Eventing
  kn operator remove configmaps --component eventing --cmName eventing-controller --key key --value value --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateCMsFlags(cmsCMDFlags); err != nil {
				return err
			}

			err := removeCMs(cmsCMDFlags, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The configuration for the specified ConfigMap has been removed in the namespace '%s'.\n",
				cmsCMDFlags.Namespace)
			return nil
		},
	}

	configureCMsCmd.Flags().StringVar(&cmsCMDFlags.Key, "key", "", "The key of the data in the configmap")
	configureCMsCmd.Flags().StringVar(&cmsCMDFlags.CMName, "cmName", "", "The flag to specify the configmap name")
	configureCMsCmd.Flags().StringVarP(&cmsCMDFlags.Component, "component", "c", "", "The flag to specify the component name")
	configureCMsCmd.Flags().StringVarP(&cmsCMDFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return configureCMsCmd
}

func validateCMsFlags(cmsCMDFlags common.CMsFlags) error {
	if cmsCMDFlags.Namespace == "" {
		return fmt.Errorf("You need to specify the namespace.")
	}
	if cmsCMDFlags.Component == "" {
		return fmt.Errorf("You need to specify the component name.")
	}
	if cmsCMDFlags.Key != "" && cmsCMDFlags.CMName == "" {
		return fmt.Errorf("You need to specify the name for the ConfigMap.")
	}

	return nil
}

func removeCMs(cmsCMDFlags common.CMsFlags, p *pkg.OperatorParams) error {
	ksCR, err := common.GetKnativeOperatorCR(p)
	if err != nil {
		return err
	}

	cmData, err := ksCR.GetConfigMaps(cmsCMDFlags.Component, cmsCMDFlags.Namespace)
	if err != nil {
		return err
	}

	cmData = removeCMsFields(cmData, cmsCMDFlags)
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return ksCR.UpdateConfigMaps(cmsCMDFlags.Component, cmsCMDFlags.Namespace, cmData)
	})

	if err != nil {
		return err
	}

	return nil
}

func removeCMsFields(cmData base.ConfigMapData, cmsCMDFlags common.CMsFlags) base.ConfigMapData {
	if cmsCMDFlags.CMName == "" {
		// If no name for the CM is specified, we will remove all config maps.
		cmData = nil
	} else if cmsCMDFlags.Key == "" {
		// Remove the configurations for the CM named cmsCMDFlags.CMName.
		dropMapKey(cmData, cmsCMDFlags.CMName)
		key := addOrRemovePrefix(cmsCMDFlags.CMName, "config-")
		dropMapKey(cmData, key)
	} else if cmsCMDFlags.Key != "" {
		dropMapConfigKey(cmData, cmsCMDFlags.CMName, cmsCMDFlags.Key)
		key := addOrRemovePrefix(cmsCMDFlags.CMName, "config-")
		dropMapConfigKey(cmData, key, cmsCMDFlags.Key)
	}

	return cmData
}

func dropMapKey(cmData base.ConfigMapData, key string) {
	if _, ok := cmData[key]; ok {
		delete(cmData, key)
	}
}

func dropMapConfigKey(cmData base.ConfigMapData, key, configKey string) {
	if val, ok := cmData[key]; ok {
		if _, ok := val[configKey]; ok {
			delete(val, configKey)
		}
	}
}

func addOrRemovePrefix(val, prefix string) string {
	result := ""
	if strings.HasPrefix(val, prefix) {
		result = val[len(prefix):]
	} else {
		result = fmt.Sprintf("%s%s", prefix, val)
	}
	return result
}
