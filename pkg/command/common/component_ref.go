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

package common

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/validation"
)

const CRNameFlag = "cr-name"

type ComponentRef struct {
	Component string
	Namespace string
	Name      string
}

func (r ComponentRef) String() string {
	return fmt.Sprintf("%s %s/%s", ComponentKind(r.Component), r.Namespace, r.Name)
}

func ComponentKind(component string) string {
	if strings.EqualFold(component, ServingComponent) {
		return "KnativeServing"
	}
	if strings.EqualFold(component, EventingComponent) {
		return "KnativeEventing"
	}
	return "Knative component"
}

func DefaultComponentName(component string) string {
	if strings.EqualFold(component, ServingComponent) {
		return KnativeServingName
	}
	if strings.EqualFold(component, EventingComponent) {
		return KnativeEventingName
	}
	return ""
}

func NormalizeComponentName(component, name string) (string, error) {
	normalized := strings.TrimSpace(name)
	if normalized == "" {
		normalized = DefaultComponentName(component)
	}
	if normalized == "" {
		return "", fmt.Errorf("--%s requires --component serving or --component eventing", CRNameFlag)
	}
	if err := ValidateComponentName(normalized); err != nil {
		return "", err
	}
	return normalized, nil
}

func NormalizeExplicitComponentName(component, name string) (string, error) {
	normalized := strings.TrimSpace(name)
	if normalized == "" {
		return "", fmt.Errorf("--%s must be non-empty after trimming whitespace", CRNameFlag)
	}
	if DefaultComponentName(component) == "" {
		return "", fmt.Errorf("--%s requires --component serving or --component eventing", CRNameFlag)
	}
	if err := ValidateComponentName(normalized); err != nil {
		return "", err
	}
	return normalized, nil
}

func ValidateComponentName(name string) error {
	if errs := validation.IsDNS1123Subdomain(name); len(errs) > 0 {
		return fmt.Errorf("--%s must be a valid Kubernetes DNS subdomain: %s", CRNameFlag, strings.Join(errs, "; "))
	}
	return nil
}

func SetComponentNameFromFlag(flags *pflag.FlagSet, component string, name *string) error {
	if flags.Changed(CRNameFlag) {
		normalized, err := NormalizeExplicitComponentName(component, *name)
		if err != nil {
			return err
		}
		*name = normalized
		return nil
	}
	normalized, err := NormalizeComponentName(component, *name)
	if err != nil {
		return err
	}
	*name = normalized
	return nil
}
