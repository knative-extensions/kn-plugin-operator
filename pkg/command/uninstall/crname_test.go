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

package uninstall

import (
	"strings"
	"testing"

	"knative.dev/kn-plugin-operator/pkg/command/common"
	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
)

func TestValidateComponentCRName(t *testing.T) {
	flags := uninstallCmdFlags{
		Component:      common.ServingComponent,
		CRName:         " custom-serving ",
		CRNameExplicit: true,
	}
	testingUtil.AssertEqual(t, validateComponentCRName(&flags), nil)
	testingUtil.AssertEqual(t, flags.CRName, "custom-serving")

	blank := uninstallCmdFlags{
		Component:      common.EventingComponent,
		CRName:         " ",
		CRNameExplicit: true,
	}
	if err := validateComponentCRName(&blank); err == nil || !strings.Contains(err.Error(), "non-empty") {
		t.Fatalf("expected blank --cr-name error, got %v", err)
	}
}
