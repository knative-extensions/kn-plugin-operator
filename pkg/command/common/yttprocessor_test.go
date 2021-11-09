/*
Copyright 2021 The Knative Authors

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

package common

import (
	"testing"

	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
)

func TestYttProcessorGenerateOutput(t *testing.T) {
	yamlTplData := []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: knative-operator
  namespace: default
spec:
  replicas: 1
`)

	overlayData := []byte(`
#@ load("@ytt:overlay", "overlay")
#@ load("@ytt:data", "data")

#@overlay/match by=overlay.subset({"kind": "Deployment", "metadata":{"name":"knative-operator"}}),expects=1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: #@ data.values.namespace
`)

	valuesData := []byte(`
#@data/values
---
namespace: test-namespace
`)

	expectedYAMLTplData := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: knative-operator
  namespace: test-namespace
spec:
  replicas: 1
`

	yttp := YttProcessor{
		BaseData:    yamlTplData,
		OverlayData: overlayData,
		ValuesData:  valuesData,
	}

	finalContent, err := yttp.GenerateOutput()
	testingUtil.AssertEqual(t, err == nil, true)
	testingUtil.AssertEqual(t, finalContent, expectedYAMLTplData)
}
