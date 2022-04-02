//go:build eventingimage
// +build eventingimage

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

// TestEventingImageConfiguration verifies whether the operator plugin can configure the image of the deployment for Knative Eventing
func TestEventingImageConfiguration(t *testing.T) {
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
		expectedImageFlags configure.ImageFlags
	}{{
		name: "Knative Eventing verifying the image configuration for the deployment",
		expectedImageFlags: configure.ImageFlags{
			ImageUrl:   resources.TestImageUrl,
			Component:  "eventing",
			Namespace:  resources.EventingOperatorNamespace,
			DeployName: "eventing-controller",
			ImageKey:   resources.TestImageKey,
		},
	}, {
		name: "Knative Eventing verifying the image configuration for the default",
		expectedImageFlags: configure.ImageFlags{
			ImageUrl:  resources.TestDefaultEventingImageUrl,
			Component: "eventing",
			Namespace: resources.EventingOperatorNamespace,
			ImageKey:  "default",
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			resources.VerifyKnativeEventingImages(t, clients.Operator.KnativeEventings(resources.EventingOperatorNamespace),
				tt.expectedImageFlags)
		})
	}
}
