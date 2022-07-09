//go:build servingservicelabeldelete
// +build servingservicelabeldelete

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

// TestServingServiceLabelDeletion verifies whether the operator plugin can delete the labels, annotations and selectors for the service in Knative Serving
func TestServingServiceLabelDeletion(t *testing.T) {
	clients := client.Setup(t)

	names := test.ResourceNames{
		KnativeServing:  "knative-serving",
		KnativeEventing: "knative-eventing",
		Namespace:       resources.ServingOperatorNamespace,
	}

	test.CleanupOnInterrupt(func() { test.TearDown(clients, names) })
	defer test.TearDown(clients, names)

	for _, tt := range []struct {
		name           string
		expectedLabels common.KeyValueFlags
	}{{
		name: "Knative Serving verifying the deletion of the first key-value pair for labels",
		expectedLabels: common.KeyValueFlags{
			Value:       resources.TestValue,
			Key:         resources.TestKey,
			Component:   "serving",
			ServiceName: "activator-service",
			Label:       true,
		},
	}, {
		name: "Knative Serving verifying the deletion of the additional key-value pair for labels",
		expectedLabels: common.KeyValueFlags{
			Value:       resources.TestValueAdditional,
			Key:         resources.TestKeyAdditional,
			Component:   "serving",
			ServiceName: "activator-service",
			Label:       true,
		},
	}, {
		name: "Knative Serving verifying the deletion of the first key-value pair for annotations",
		expectedLabels: common.KeyValueFlags{
			Value:       resources.TestValue,
			Key:         resources.TestKey,
			Component:   "serving",
			ServiceName: "activator-service",
			Annotation:  true,
		},
	}, {
		name: "Knative Serving verifying the deletion of the additional key-value pair for annotations",
		expectedLabels: common.KeyValueFlags{
			Value:       resources.TestValueAdditional,
			Key:         resources.TestKeyAdditional,
			Component:   "serving",
			ServiceName: "activator-service",
			Annotation:  true,
		},
	}, {
		name: "Knative Serving verifying the deletion of the first key-value pair for selector",
		expectedLabels: common.KeyValueFlags{
			Value:       resources.TestValue,
			Key:         resources.TestKey,
			Component:   "serving",
			ServiceName: "activator-service",
			Selector:    true,
		},
	}, {
		name: "Knative Serving verifying the deletion of the additional key-value pair for selector",
		expectedLabels: common.KeyValueFlags{
			Value:       resources.TestValueAdditional,
			Key:         resources.TestKeyAdditional,
			Component:   "serving",
			ServiceName: "activator-service",
			Selector:    true,
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			resources.VerifyKnativeServingServiceLabelsDelete(t, clients.Operator.KnativeServings(resources.ServingOperatorNamespace),
				tt.expectedLabels)
		})
	}
}
