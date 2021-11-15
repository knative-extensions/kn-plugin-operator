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
	mfc "github.com/manifestival/client-go-client"
	"k8s.io/client-go/rest"
)

// Manifest applies the content of the yaml file against a Kubernetes cluster
type Manifest struct {
	// YttPro is an instance of the YttProcessor to generate the output yaml
	YttPro *YttProcessor
	// RestConfig is the rest configuration to access the Kubernetes cluster
	RestConfig *rest.Config
}

// Apply applies the content of the yaml file against the Kubernetes cluster
func (man *Manifest) Apply() error {
	content, err := man.YttPro.GenerateOutput()
	if err != nil {
		return err
	}
	path := "tempFile.yaml"
	defer DeleteFile(path)

	if err = WriteFile(path, content); err != nil {
		return err
	}
	manifest, err := mfc.NewManifest(path, man.RestConfig)
	if err != nil {
		return err
	}

	if err = manifest.Apply(); err != nil {
		return err
	}

	return nil
}
