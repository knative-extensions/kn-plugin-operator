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
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"knative.dev/kn-plugin-operator/pkg"
)

// NewRemoveCommand represents the remove commands for Knative Serving and Eventing
func NewRemoveCommand(p *pkg.OperatorParams) *cobra.Command {
	var removeCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove the ingress for Knative Serving",
		Example: `
  # Remove the configuration of the resources for Knative Serving
  kn operation remove resources --component serving --namespace knative-serving
  # Remove the configuration of the resources for the container activator in the deployment activator
  kn operation remove resources --component serving --deployName activator --namespace knative-serving
  # Remove the configuration of the resources for the deployment activator
  kn operation remove resources --component serving --deployName activator --namespace knative-serving`,
	}

	removeCmd.AddCommand(removeResourcesCommand(p))
	removeCmd.AddCommand(removeConfigMapsCommand(p))
	removeCmd.AddCommand(removeTolerationsCommand(p))
	removeCmd.AddCommand(removeImageCommand(p))
	removeCmd.AddCommand(removeEnvVarCommand(p))
	removeCmd.AddCommand(removeHACommand(p))
	removeCmd.AddCommand(removeLabelCommand(p))
	removeCmd.AddCommand(removeAnnotationCommand(p))
	removeCmd.AddCommand(removeNodeSelectorCommand(p))
	removeCmd.AddCommand(removeSelectorCommand(p))

	return removeCmd
}
