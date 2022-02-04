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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // from https://github.com/kubernetes/client-go/issues/345
	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/kn-plugin-operator/pkg/command/common"
)

type eventingSourceFlags struct {
	Ceph      bool
	Github    bool
	Gitlab    bool
	Kafka     bool
	Natss     bool
	Rabbitmq  bool
	Redis     bool
	Namespace string
}

var eventingSourceCmdFlags eventingSourceFlags

// newEventingSourcesCommand represents the enable commands for eventing sources
func newEventingSourcesCommand(p *pkg.OperatorParams) *cobra.Command {
	var enableEventingSourceCmd = &cobra.Command{
		Use:   "eventing-source",
		Short: "Enable the eventing source for Knative Eventing",
		Example: `
  # Enable the eventing source github for Knative Serving
  kn operation enable eventing-source --github --namespace knative-eventing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if eventingSourceCmdFlags.Namespace == "" {
				eventingSourceCmdFlags.Namespace = common.DefaultKnativeEventingNamespace
			}

			rootPath, err := os.Getwd()
			if err != nil {
				return err
			}

			err = enableEventingSource(eventingSourceCmdFlags, rootPath, p)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "The specified eventing sources were enabled in the namespace '%s'.\n",
				eventingSourceCmdFlags.Namespace)
			return nil
		},
	}

	enableEventingSourceCmd.Flags().BoolVar(&eventingSourceCmdFlags.Kafka, "kafka", false, "The flag to enable the kafka source")
	enableEventingSourceCmd.Flags().BoolVar(&eventingSourceCmdFlags.Ceph, "ceph", false, "The flag to enable the ceph source")
	enableEventingSourceCmd.Flags().BoolVar(&eventingSourceCmdFlags.Github, "github", false, "The flag to enable the github source")
	enableEventingSourceCmd.Flags().BoolVar(&eventingSourceCmdFlags.Gitlab, "gitlab", false, "The flag to enable the gitlab source")
	enableEventingSourceCmd.Flags().BoolVar(&eventingSourceCmdFlags.Redis, "redis", false, "The flag to enable the redis source")
	enableEventingSourceCmd.Flags().BoolVar(&eventingSourceCmdFlags.Natss, "natss", false, "The flag to enable the natss source")
	enableEventingSourceCmd.Flags().BoolVar(&eventingSourceCmdFlags.Rabbitmq, "rabbitmq", false, "The flag to enable the rabbitmq source")
	enableEventingSourceCmd.Flags().StringVarP(&eventingSourceCmdFlags.Namespace, "namespace", "n", "", "The namespace of the Knative Operator or the Knative component")

	return enableEventingSourceCmd
}

func enableEventingSource(eventingSourceCmdFlags eventingSourceFlags, rootPath string, p *pkg.OperatorParams) error {
	// Generate the CR template
	yamlTemplateString, err := common.GenerateOperatorCRString(common.EventingComponent, eventingSourceCmdFlags.Namespace, p)
	if err != nil {
		return err
	}

	overlayContent := getOverlayYamlContentSource(rootPath)
	valuesYaml := getYamlValuesContentSource(eventingSourceCmdFlags)

	if err := common.ApplyManifests(yamlTemplateString, overlayContent, valuesYaml, p); err != nil {
		return err
	}
	return nil
}

func getOverlayYamlContentSource(rootPath string) string {
	path := rootPath + "/overlay/ke_source.yaml"
	overlayContent, _ := common.ReadFile(path)
	return overlayContent
}

func getYamlValuesContentSource(eventingSourceCmdFlags eventingSourceFlags) string {
	return fmt.Sprintf("#@data/values\n---\nnamespace: %s\nredis: %t\nrabbitmq: %t\ngitlab: %t\ngithub: %t\nceph: %t\nkafka: %t\nnatss: %t",
		eventingSourceCmdFlags.Namespace, eventingSourceCmdFlags.Redis, eventingSourceCmdFlags.Rabbitmq, eventingSourceCmdFlags.Gitlab,
		eventingSourceCmdFlags.Github, eventingSourceCmdFlags.Ceph, eventingSourceCmdFlags.Kafka, eventingSourceCmdFlags.Natss)
}
