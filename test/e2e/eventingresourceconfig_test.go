//go:build eventingresourceconfig
// +build eventingresourceconfig

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

// TestEventingResourceConfiguration verifies whether the operator plugin can configure the resources for Knative Eventing
func TestEventingResourceConfiguration(t *testing.T) {
	clients := client.Setup(t)

	names := test.ResourceNames{
		KnativeServing:  "knative-eventing",
		KnativeEventing: "knative-eventing",
		Namespace:       resources.EventingOperatorNamespace,
	}

	test.CleanupOnInterrupt(func() { test.TearDown(clients, names) })
	defer test.TearDown(clients, names)

	expectedResourcesFlags := configure.ResourcesFlags{
		LimitCPU:      "2048m",
		LimitMemory:   "1001M",
		RequestCPU:    "1024m",
		RequestMemory: "999M",
		Component:     "eventing",
		Container:     "eventing-controller",
		DeployName:    "eventing-controller",
	}
	resources.VerifyKnativeEventingExistence(t, clients.Operator.KnativeEventings(resources.EventingOperatorNamespace),
		expectedResourcesFlags)
}
