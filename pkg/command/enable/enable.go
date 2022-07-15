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

package enable

import (
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"knative.dev/kn-plugin-operator/pkg"
)

// NewEnableCommand represents the enable commands for sources or ingresses
func NewEnableCommand(p *pkg.OperatorParams) *cobra.Command {
	var enableCmd = &cobra.Command{
		Use:   "enable",
		Short: "Enable the ingress for Knative Serving and the eventing sources for Knative Eventing",
		Example: `
  # Enable the ingress istio for Knative Serving
  kn-operator enable ingress --istio --namespace knative-serving
  # Enable the eventing source github for Knative Eventing
  kn-operator enable eventing-source --github --namespace knative-eventing`,
	}

	enableCmd.AddCommand(newIngressCommand(p))
	enableCmd.AddCommand(newEventingSourcesCommand(p))

	return enableCmd
}
