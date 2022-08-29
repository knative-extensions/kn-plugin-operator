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

package configure

import (
	"fmt"
	"testing"

	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
)

func TestValidateEnvVarsFlags(t *testing.T) {
	for _, tt := range []struct {
		name           string
		envVarFlags    EnvVarFlags
		expectedResult error
	}{{
		name: "Knative Eventing with the correct configuration for env vars",
		envVarFlags: EnvVarFlags{
			EnvName:       "test-key",
			EnvValue:      "test-value",
			Component:     "eventing",
			Namespace:     "test-eventing",
			DeployName:    "eventing-controller",
			ContainerName: "container",
		},
		expectedResult: nil,
	}, {
		name: "Knative Serving with the correct configuration for env vars",
		envVarFlags: EnvVarFlags{
			EnvName:       "test-key",
			EnvValue:      "test-value",
			Component:     "serving",
			Namespace:     "test-serving",
			DeployName:    "controller",
			ContainerName: "container",
		},
		expectedResult: nil,
	}, {
		name: "Knative Serving with no deploy name",
		envVarFlags: EnvVarFlags{
			EnvName:       "test-key",
			EnvValue:      "test-value",
			Component:     "serving",
			Namespace:     "test-serving",
			ContainerName: "container",
		},
		expectedResult: fmt.Errorf("You need to specify the name for the deployment resource."),
	}, {
		name: "Knative Eventing with no container name",
		envVarFlags: EnvVarFlags{
			EnvName:    "test-key",
			EnvValue:   "test-value",
			Component:  "eventing",
			Namespace:  "test-eventing",
			DeployName: "eventing-controller",
		},
		expectedResult: fmt.Errorf("You need to specify the name for the container."),
	}, {
		name: "Knative Eventing with no name for the env var",
		envVarFlags: EnvVarFlags{
			EnvValue:      "test-value",
			Component:     "serving",
			Namespace:     "test-serving",
			DeployName:    "controller",
			ContainerName: "container",
		},
		expectedResult: fmt.Errorf("You need to specify the name for the environment variable."),
	}, {
		name: "Knative Eventing with no value for the env var",
		envVarFlags: EnvVarFlags{
			EnvName:       "test-name",
			Component:     "serving",
			Namespace:     "test-serving",
			DeployName:    "controller",
			ContainerName: "container",
		},
		expectedResult: fmt.Errorf("You need to specify the value for the environment variable."),
	}, {
		name: "Knative Eventing with invalid component name",
		envVarFlags: EnvVarFlags{
			EnvName:       "test-key",
			EnvValue:      "test-value",
			Component:     "serving-test",
			Namespace:     "test-serving",
			DeployName:    "controller",
			ContainerName: "container",
		},
		expectedResult: fmt.Errorf("You need to specify the component for Knative: serving or eventing."),
	}, {
		name: "Knative Eventing with no namespace",
		envVarFlags: EnvVarFlags{
			EnvName:       "test-key",
			EnvValue:      "test-value",
			Component:     "serving",
			DeployName:    "controller",
			ContainerName: "container",
		},
		expectedResult: fmt.Errorf("You need to specify the namespace."),
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := validateEnvVarsFlags(tt.envVarFlags)
			if tt.expectedResult == nil {
				testingUtil.AssertEqual(t, result, nil)
			} else {
				testingUtil.AssertEqual(t, result.Error(), tt.expectedResult.Error())
			}
		})
	}
}

func TestGetOverlayYamlContentEnvvar(t *testing.T) {
	for _, tt := range []struct {
		name           string
		envVarFlags    EnvVarFlags
		expectedResult string
	}{{
		name: "Knative Eventing",
		envVarFlags: EnvVarFlags{
			EnvName:       "test-key",
			EnvValue:      "test-value",
			Component:     "eventing",
			Namespace:     "test-eventing",
			DeployName:    "eventing-controller",
			ContainerName: "container",
		},
		expectedResult: `#@ load("@ytt:overlay", "overlay")
#@ load("@ytt:data", "data")

#@overlay/match by=overlay.subset({"kind": "KnativeEventing"}),expects=1
---
apiVersion: operator.knative.dev/v1beta1
kind: KnativeEventing
metadata:
  #@overlay/match missing_ok=True
  namespace: #@ data.values.namespace
#@overlay/match missing_ok=True
spec:
  #@overlay/match missing_ok=True
  deployments:
  #@overlay/match by="name",missing_ok=True
  - name: #@ data.values.deployName
    #@overlay/match missing_ok=True
    env:
    #@overlay/match by="container",missing_ok=True
    - container: #@ data.values.containerName
      #@overlay/match missing_ok=True
      envVars:
      #@overlay/match by="name",missing_ok=True
      - name: #@ data.values.envVarName
        #@overlay/match missing_ok=True
        value: #@ data.values.envVarValue
`,
	}, {
		name: "Knative Serving",
		envVarFlags: EnvVarFlags{
			EnvName:       "test-key",
			EnvValue:      "test-value",
			Component:     "serving",
			Namespace:     "test-serving",
			DeployName:    "controller",
			ContainerName: "container",
		},
		expectedResult: `#@ load("@ytt:overlay", "overlay")
#@ load("@ytt:data", "data")

#@overlay/match by=overlay.subset({"kind": "KnativeServing"}),expects=1
---
apiVersion: operator.knative.dev/v1beta1
kind: KnativeServing
metadata:
  #@overlay/match missing_ok=True
  namespace: #@ data.values.namespace
#@overlay/match missing_ok=True
spec:
  #@overlay/match missing_ok=True
  deployments:
  #@overlay/match by="name",missing_ok=True
  - name: #@ data.values.deployName
    #@overlay/match missing_ok=True
    env:
    #@overlay/match by="container",missing_ok=True
    - container: #@ data.values.containerName
      #@overlay/match missing_ok=True
      envVars:
      #@overlay/match by="name",missing_ok=True
      - name: #@ data.values.envVarName
        #@overlay/match missing_ok=True
        value: #@ data.values.envVarValue
`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := getOverlayYamlContentEnvvar(tt.envVarFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}

func TestGetYamlValuesContentEnvvars(t *testing.T) {
	for _, tt := range []struct {
		name           string
		envVarFlags    EnvVarFlags
		expectedResult string
	}{{
		name: "Knative Eventing",
		envVarFlags: EnvVarFlags{
			EnvName:       "test-key",
			EnvValue:      "test-value",
			Component:     "eventing",
			Namespace:     "test-eventing",
			DeployName:    "eventing-controller",
			ContainerName: "container",
		},
		expectedResult: `#@data/values
---
namespace: test-eventing
deployName: eventing-controller
containerName: container
envVarName: test-key
envVarValue: test-value`,
	}, {
		name: "Knative Serving",
		envVarFlags: EnvVarFlags{
			EnvName:       "test-key",
			EnvValue:      "test-value",
			Component:     "serving",
			Namespace:     "test-serving",
			DeployName:    "controller",
			ContainerName: "container",
		},
		expectedResult: `#@data/values
---
namespace: test-serving
deployName: controller
containerName: container
envVarName: test-key
envVarValue: test-value`,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			result := getYamlValuesContentEnvvars(tt.envVarFlags)
			testingUtil.AssertEqual(t, result, tt.expectedResult)
		})
	}
}
