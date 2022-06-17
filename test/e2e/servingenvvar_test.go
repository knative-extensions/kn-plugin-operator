//go:build servingenvvar
// +build servingenvvar

/*
Copyright 2022 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"testing"

	"knative.dev/kn-plugin-operator/pkg/command/configure"
	"knative.dev/kn-plugin-operator/test/resources"
	"knative.dev/operator/test"
	"knative.dev/operator/test/client"
)

// TestServingEnvvarConfiguration verifies whether the operator plugin can configure the env vars of the deployment for Knative Serving
func TestServingEnvvarConfiguration(t *testing.T) {
	clients := client.Setup(t)

	names := test.ResourceNames{
		KnativeServing:  "knative-serving",
		KnativeEventing: "knative-eventing",
		Namespace:       resources.ServingOperatorNamespace,
	}

	test.CleanupOnInterrupt(func() { test.TearDown(clients, names) })
	defer test.TearDown(clients, names)

	for _, tt := range []struct {
		name                string
		expectedEnvVarFlags configure.EnvVarFlags
	}{{
		name: "Knative Serving verifying the env var configuration for the deployment controller",
		expectedEnvVarFlags: configure.EnvVarFlags{
			EnvName:       resources.TestEnvName,
			Component:     "serving",
			Namespace:     resources.ServingOperatorNamespace,
			DeployName:    "controller",
			ContainerName: "controller",
			EnvValue:      resources.TestEnvValue,
		},
	}, {
		name: "Knative Serving verifying the additional env var configuration for the deployment controller",
		expectedEnvVarFlags: configure.EnvVarFlags{
			EnvName:       resources.TestAddEnvName,
			Component:     "serving",
			Namespace:     resources.ServingOperatorNamespace,
			DeployName:    "controller",
			ContainerName: "controller",
			EnvValue:      resources.TestAddEnvValue,
		},
	}, {
		name: "Knative Serving verifying the env var configuration for the deployment activator",
		expectedEnvVarFlags: configure.EnvVarFlags{
			EnvName:       resources.TestEnvName,
			Component:     "serving",
			Namespace:     resources.ServingOperatorNamespace,
			DeployName:    "activator",
			ContainerName: "activator",
			EnvValue:      resources.TestEnvValue,
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			resources.VerifyKnativeServingEnvVars(t, clients.Operator.KnativeServings(resources.ServingOperatorNamespace),
				tt.expectedEnvVarFlags)
		})
	}
}
