// Copyright 2021 The Knative Authors
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

package pkg

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/operator/pkg/client/clientset/versioned"
)

// OperatorParams stores the configs for interacting with kube api
type OperatorParams struct {
	KubeCfgPath       string
	ClientConfig      clientcmd.ClientConfig
	NewKubeClient     func() (kubernetes.Interface, error)
	NewOperatorClient func() (*versioned.Clientset, error)
}

// Initialize generate the clientset for params
func (params *OperatorParams) Initialize() error {
	if params.NewKubeClient == nil {
		params.NewKubeClient = params.newKubeClient
	}
	if params.NewOperatorClient == nil {
		params.NewOperatorClient = params.newOperatorClient
	}
	return nil
}

// RestConfig returns REST config, which can be to use to create specific clientset
func (params *OperatorParams) RestConfig() (*rest.Config, error) {
	var err error

	if params.ClientConfig == nil {
		params.ClientConfig, err = params.GetClientConfig()
		if err != nil {
			return nil, err
		}
	}

	config, err := params.ClientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}

// GetClientConfig gets ClientConfig from KubeCfgPath
func (params *OperatorParams) GetClientConfig() (clientcmd.ClientConfig, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if len(params.KubeCfgPath) == 0 {
		return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}), nil
	}

	_, err := os.Stat(params.KubeCfgPath)
	if err == nil {
		loadingRules.ExplicitPath = params.KubeCfgPath
		return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}), nil
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	paths := filepath.SplitList(params.KubeCfgPath)
	if len(paths) > 1 {
		return nil, fmt.Errorf("Can not find config file. '%s' looks like a path. "+
			"Please use the env var KUBECONFIG if you want to check for multiple configuration files", params.KubeCfgPath)
	}
	return nil, fmt.Errorf("Config file '%s' can not be found", params.KubeCfgPath)
}

// newOperatorClient creates an operator clientset from kubenetes config
func (params *OperatorParams) newOperatorClient() (*versioned.Clientset, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	return versioned.NewForConfig(restConfig)
}

// newKubeClient creates a kubenetes clientset from kubenetes config
func (params *OperatorParams) newKubeClient() (kubernetes.Interface, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(restConfig)
}
