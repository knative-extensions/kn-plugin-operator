/*
Copyright 2022 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package enable

import (
	"fmt"
	"testing"

	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
)

func TestValidateIngressFlags(t *testing.T) {
	for _, tt := range []struct {
		name            string
		ingressCMDFlags ingressFlags
		expectedError   error
	}{{
		name:            "Only Istio enabled",
		ingressCMDFlags: ingressFlags{Istio: true},
		expectedError:   nil,
	}, {
		name:            "No ingress enabled",
		ingressCMDFlags: ingressFlags{Istio: false, Kourier: false, Contour: false},
		expectedError:   fmt.Errorf("You need to enable at least one ingress for Knative Serving."),
	}, {
		name:            "Istio and Kourier enabled",
		ingressCMDFlags: ingressFlags{Istio: true, Kourier: true},
		expectedError:   fmt.Errorf("You can specify only one ingress for Knative Serving."),
	}, {
		name:            "Istio, Contour and Kourier enabled",
		ingressCMDFlags: ingressFlags{Istio: true, Kourier: true, Contour: true},
		expectedError:   fmt.Errorf("You can specify only one ingress for Knative Serving."),
	}, {
		name:            "Only Kourier enabled",
		ingressCMDFlags: ingressFlags{Istio: false, Kourier: true, Contour: false},
		expectedError:   nil,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIngressFlags(tt.ingressCMDFlags)
			if tt.expectedError == nil {
				testingUtil.AssertEqual(t, err, nil)
			} else {
				testingUtil.AssertEqual(t, err.Error(), tt.expectedError.Error())
			}
		})
	}
}

func TestGetYamlValuesContent(t *testing.T) {
	for _, tt := range []struct {
		name            string
		ingressCMDFlags ingressFlags
		expectedResult  string
	}{{
		name: "Knative Serving with istio enabled",
		ingressCMDFlags: ingressFlags{
			Namespace: "test-serving",
			Istio:     true,
		},
		expectedResult: `#@data/values
---
namespace: test-serving
kourier: false
istio: true
contour: false
ingressClass: istio.ingress.networking.knative.dev`,
	}, {
		name: "Knative Serving with Kourier enabled",
		ingressCMDFlags: ingressFlags{
			Namespace: "test-serving",
			Kourier:   true,
		},
		expectedResult: `#@data/values
---
namespace: test-serving
kourier: true
istio: false
contour: false
ingressClass: kourier.ingress.networking.knative.dev`,
	}, {
		name: "Knative Serving with Contour enabled",
		ingressCMDFlags: ingressFlags{
			Namespace: "test-serving",
			Contour:   true,
		},
		expectedResult: `#@data/values
---
namespace: test-serving
kourier: false
istio: false
contour: true
ingressClass: contour.ingress.networking.knative.dev`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := getYamlValuesContent(tt.ingressCMDFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}
