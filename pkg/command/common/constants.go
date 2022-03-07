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

const (
	DefaultIstioNamespace           = "istio-system"
	DefaultKnativeServingNamespace  = "knative-serving"
	DefaultKnativeEventingNamespace = "knative-eventing"
	DefaultNamespace                = "default"
	Latest                          = "latest"
	ServingComponent                = "serving"
	EventingComponent               = "eventing"
	YttMatchingTag                  = "#@overlay/match missing_ok=True"
	YttReplaceTag                   = "#@overlay/replace or_add=True"
	Space                           = " "
)

// Spaces returns series of spaces based on the input number
func Spaces(num int) string {
	value := ""
	for i := 0; i < num; i++ {
		value += Space
	}
	return value
}
