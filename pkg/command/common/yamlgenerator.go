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
	"github.com/ghodss/yaml"
)

// YamlGenarator generates the final output yaml content for Knative Eventing or Serving custom resource.
type YamlGenarator struct {
	// Input is a Kubernetes resource, either Knative Serving or Eventing
	Input interface{}
}

// GenerateYamlOutput returns the yaml content for Knative Serving or Eventing CR
func (yamlg *YamlGenarator) GenerateYamlOutput() (string, error) {
	d, err := yaml.Marshal(&yamlg.Input)
	if err != nil {
		return "", err
	}
	return string(d), nil
}
