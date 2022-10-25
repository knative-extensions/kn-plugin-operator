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
	"fmt"
	"testing"

	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
	"knative.dev/operator/pkg/apis/operator/base"
)

var replica int32 = 3

func TestValidateHAsFlags(t *testing.T) {
	for _, tt := range []struct {
		name           string
		haCMDFlags     HAFlags
		expectedResult error
	}{{
		name: "HA flags with correct component and namespace",
		haCMDFlags: HAFlags{
			Component: "eventing",
			Namespace: "test-eventing",
		},
		expectedResult: nil,
	}, {
		name: "HA flags with correct component, namespace and the deploy name",
		haCMDFlags: HAFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "test-deploy",
		},
		expectedResult: nil,
	}, {
		name: "HA flags with correct component, namespace, deploy name and container name",
		haCMDFlags: HAFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "test-deploy",
		},
		expectedResult: nil,
	}, {
		name: "HA flags without component",
		haCMDFlags: HAFlags{
			Namespace:  "test-eventing",
			DeployName: "test-deploy",
		},
		expectedResult: fmt.Errorf("You need to specify the component name."),
	}, {
		name: "HA flags without namespace",
		haCMDFlags: HAFlags{
			Component:  "eventing",
			DeployName: "test-deploy",
		},
		expectedResult: fmt.Errorf("You need to specify the namespace."),
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := validateHAsFlags(tt.haCMDFlags)
			if tt.expectedResult == nil {
				testingUtil.AssertEqual(t, result, nil)
			} else {
				testingUtil.AssertEqual(t, result.Error(), tt.expectedResult.Error())
			}
		})
	}
}

func testCommonSpec() base.CommonSpec {
	return base.CommonSpec{
		HighAvailability: &base.HighAvailability{
			Replicas: &replica,
		},
		Workloads: []base.WorkloadOverride{
			{
				Name:     "net-istio-controller",
				Replicas: &replica,
			},
			{
				Name:     "net-istio-controller-1",
				Replicas: &replica,
			},
		},
	}
}

func TestRemoveReplicasFields(t *testing.T) {
	for _, tt := range []struct {
		name           string
		haCMDFlags     HAFlags
		input          base.CommonSpec
		expectedResult base.CommonSpec
	}{{
		name: "HA flags with correct component and namespace",
		haCMDFlags: HAFlags{
			Component: "eventing",
			Namespace: "test-eventing",
		},
		input: testCommonSpec(),
		expectedResult: base.CommonSpec{
			HighAvailability: nil,
			Workloads: []base.WorkloadOverride{
				{
					Name:     "net-istio-controller",
					Replicas: nil,
				},
				{
					Name:     "net-istio-controller-1",
					Replicas: nil,
				},
			},
		},
	}, {
		name: "HA flags with correct deploy, component and namespace",
		haCMDFlags: HAFlags{
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "net-istio-controller",
		},
		input: testCommonSpec(),
		expectedResult: base.CommonSpec{
			HighAvailability: &base.HighAvailability{
				Replicas: &replica,
			},
			Workloads: []base.WorkloadOverride{
				{
					Name:     "net-istio-controller",
					Replicas: nil,
				},
				{
					Name:     "net-istio-controller-1",
					Replicas: &replica,
				},
			},
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := removeReplicasFields(&tt.input, tt.haCMDFlags)
			testingUtil.AssertDeepEqual(t, *result, tt.expectedResult)
		})
	}
}
