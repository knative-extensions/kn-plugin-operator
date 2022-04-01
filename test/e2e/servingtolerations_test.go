//go:build servingtolerations
// +build servingtolerations

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

// TestServingTolerationConfiguration verifies whether the operator plugin can configure the tolerations for Knative Serving
func TestServingTolerationConfiguration(t *testing.T) {
	clients := client.Setup(t)

	names := test.ResourceNames{
		KnativeServing:  "knative-serving",
		KnativeEventing: "knative-eventing",
		Namespace:       resources.ServingOperatorNamespace,
	}

	test.CleanupOnInterrupt(func() { test.TearDown(clients, names) })
	defer test.TearDown(clients, names)

	for _, tt := range []struct {
		name                     string
		expectedTolerationsFlags configure.TolerationsFlags
	}{{
		name: "Knative Serving verifying the toleration for the deployment with Exists operator",
		expectedTolerationsFlags: configure.TolerationsFlags{
			Key:        resources.TestTolerationKey,
			Operator:   resources.TestOperation,
			Effect:     resources.TestEffect,
			Component:  "serving",
			Namespace:  resources.ServingOperatorNamespace,
			DeployName: "autoscaler",
		},
	}, {
		name: "Knative Serving verifying the toleration for the deployment with Equal operator",
		expectedTolerationsFlags: configure.TolerationsFlags{
			Key:        resources.TestAdditionalTolerationKey,
			Operator:   resources.TestAdditionalOperation,
			Value:      resources.TestAdditionalTolerationValue,
			Effect:     resources.TestAdditionalEffect,
			Component:  "serving",
			Namespace:  resources.ServingOperatorNamespace,
			DeployName: "autoscaler",
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			resources.VerifyKnativeServingTolerations(t, clients.Operator.KnativeServings(resources.ServingOperatorNamespace),
				tt.expectedTolerationsFlags)
		})
	}
}
