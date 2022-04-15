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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
	servingv1beta1 "knative.dev/operator/pkg/apis/operator/v1beta1"
)

func TestYamlGenaratorGenerateYamlOutput(t *testing.T) {
	expectedYAMLTplData := `apiVersion: operator.knative.dev/v1beta1
kind: KnativeServing
metadata:
  creationTimestamp: null
  name: knative-serving
  namespace: knative-serving
spec:
  controller-custom-certs:
    name: ""
    type: ""
  registry: {}
status: {}
`

	ks := &servingv1beta1.KnativeServing{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KnativeServing",
			APIVersion: "operator.knative.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "knative-serving",
			Namespace: "knative-serving",
		},
	}

	yamlGenerator := YamlGenarator{
		Input: ks,
	}

	finalContent, err := yamlGenerator.GenerateYamlOutput()
	testingUtil.AssertEqual(t, err == nil, true)
	testingUtil.AssertEqual(t, finalContent, expectedYAMLTplData)
}
