//go:build eventingcmrremove
// +build eventingcmrremove

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

	"knative.dev/kn-plugin-operator/pkg/command/common"
	"knative.dev/kn-plugin-operator/test/resources"
	"knative.dev/operator/test"
	"knative.dev/operator/test/client"
)

// TestEventingConfigMapDeletion verifies whether the operator plugin can delete the ConfigMaps configuration for Knative Eventing
func TestEventingConfigMapDeletion(t *testing.T) {
	clients := client.Setup(t)

	names := test.ResourceNames{
		KnativeServing:  "knative-serving",
		KnativeEventing: "knative-eventing",
		Namespace:       resources.EventingOperatorNamespace,
	}

	test.CleanupOnInterrupt(func() { test.TearDown(clients, names) })
	defer test.TearDown(clients, names)

	for _, tt := range []struct {
		name               string
		expectedConfigMaps common.CMsFlags
	}{{
		name: "Knative Eventing verifying the key-value pair deletion for the ConfigMap",
		expectedConfigMaps: common.CMsFlags{
			Key:       resources.TestKey,
			Component: "eventing",
			CMName:    "config-tracing",
			Namespace: resources.EventingOperatorNamespace,
		},
	}, {
		name: "Knative Eventing verifying the key-value pair deletion for the ConfigMap",
		expectedConfigMaps: common.CMsFlags{
			Component: "eventing",
			CMName:    "config-features",
			Namespace: resources.EventingOperatorNamespace,
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			resources.VerifyKnativeEventingConfigMapsDeletion(t, clients.Operator.KnativeEventings(resources.EventingOperatorNamespace),
				tt.expectedConfigMaps)
		})
	}
}
