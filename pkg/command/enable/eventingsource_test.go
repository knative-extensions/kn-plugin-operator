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
	"testing"

	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
)

func TestGetYamlValuesContentSource(t *testing.T) {
	for _, tt := range []struct {
		name                   string
		eventingSourceCmdFlags eventingSourceFlags
		expectedResult         string
	}{{
		name: "Knative Eventing with ceph and Kafka enabled",
		eventingSourceCmdFlags: eventingSourceFlags{
			Namespace: "test-eventing",
			Ceph:      true,
			Kafka:     true,
		},
		expectedResult: `#@data/values
---
namespace: test-eventing
redis: false
rabbitmq: false
gitlab: false
github: false
ceph: true
kafka: true`,
	}, {
		name: "Knative Eventing with redis and github enabled",
		eventingSourceCmdFlags: eventingSourceFlags{
			Namespace: "test-eventing",
			Github:    true,
			Redis:     true,
		},
		expectedResult: `#@data/values
---
namespace: test-eventing
redis: true
rabbitmq: false
gitlab: false
github: true
ceph: false
kafka: false`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := getYamlValuesContentSource(tt.eventingSourceCmdFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}
