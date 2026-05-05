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

package install

import (
	"strings"
	"testing"

	"knative.dev/kn-plugin-operator/pkg/command/common"
	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
)

func TestValidateInstallCRNameFlags(t *testing.T) {
	for _, tt := range []struct {
		name    string
		flags   installCmdFlags
		wantErr string
	}{{
		name: "omitted",
		flags: installCmdFlags{
			Component: common.ServingComponent,
		},
	}, {
		name: "remote explicit",
		flags: installCmdFlags{
			Component:               common.ServingComponent,
			CRName:                  " spoke-a.example ",
			CRNameExplicit:          true,
			ClusterProfile:          "spoke-a",
			ClusterProfileNamespace: "fleet-system",
		},
	}, {
		name: "local explicit",
		flags: installCmdFlags{
			Component:      common.ServingComponent,
			CRName:         "custom-serving",
			CRNameExplicit: true,
		},
		wantErr: "remote component installs",
	}, {
		name: "operator explicit",
		flags: installCmdFlags{
			CRName:         "custom-serving",
			CRNameExplicit: true,
		},
		wantErr: "requires --component",
	}, {
		name: "blank explicit",
		flags: installCmdFlags{
			Component:               common.ServingComponent,
			CRName:                  " ",
			CRNameExplicit:          true,
			ClusterProfile:          "spoke-a",
			ClusterProfileNamespace: "fleet-system",
		},
		wantErr: "non-empty",
	}, {
		name: "invalid explicit",
		flags: installCmdFlags{
			Component:               common.ServingComponent,
			CRName:                  "Invalid_Name",
			CRNameExplicit:          true,
			ClusterProfile:          "spoke-a",
			ClusterProfileNamespace: "fleet-system",
		},
		wantErr: "valid Kubernetes DNS subdomain",
	}} {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInstallCRNameFlags(&tt.flags)
			if tt.wantErr == "" {
				testingUtil.AssertEqual(t, err, nil)
			} else if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
			if tt.name == "remote explicit" {
				testingUtil.AssertEqual(t, tt.flags.CRName, "spoke-a.example")
			}
		})
	}
}

func TestGetYamlValuesContentUsesCRName(t *testing.T) {
	flags := installCmdFlags{
		Component: common.ServingComponent,
		Namespace: "knative-serving",
		CRName:    "custom-serving",
		Version:   "1.18.0",
		Istio:     true,
	}
	values := getYamlValuesContent(&flags)
	if !strings.Contains(values, "name: custom-serving") {
		t.Fatalf("expected custom CR name in values, got:\n%s", values)
	}
}
