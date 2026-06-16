/*
Copyright 2026 The Knative Authors

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

package core

import (
	"testing"

	"knative.dev/kn-plugin-operator/pkg/command/common"
)

func TestCRNameFlagRegistration(t *testing.T) {
	root := NewOperationCommand()
	commands := [][]string{
		{"install"},
		{"uninstall"},
		{"configure", "annotations"},
		{"configure", "configmaps"},
		{"configure", "envvars"},
		{"configure", "images"},
		{"configure", "labels"},
		{"configure", "manifests"},
		{"configure", "nodeSelectors"},
		{"configure", "replicas"},
		{"configure", "resources"},
		{"configure", "selectors"},
		{"configure", "tolerations"},
		{"remove", "annotations"},
		{"remove", "configmaps"},
		{"remove", "envvars"},
		{"remove", "images"},
		{"remove", "labels"},
		{"remove", "nodeSelectors"},
		{"remove", "replicas"},
		{"remove", "resources"},
		{"remove", "selectors"},
		{"remove", "tolerations"},
		{"enable", "ingress"},
		{"enable", "eventing-source"},
	}

	for _, path := range commands {
		cmd, _, err := root.Find(path)
		if err != nil {
			t.Fatalf("failed to find command %v: %v", path, err)
		}
		flag := cmd.Flags().Lookup(common.CRNameFlag)
		if flag == nil {
			t.Fatalf("expected --%s on command %v", common.CRNameFlag, path)
		}
		if flag.Shorthand != "" {
			t.Fatalf("expected --%s on command %v to have no shorthand, got %q", common.CRNameFlag, path, flag.Shorthand)
		}
	}
}
