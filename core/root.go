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

package core

import (
	"github.com/spf13/cobra"
	"knative.dev/kn-plugin-operator/pkg"
	"knative.dev/kn-plugin-operator/pkg/command/configure"
	"knative.dev/kn-plugin-operator/pkg/command/enable"
	"knative.dev/kn-plugin-operator/pkg/command/install"
	"knative.dev/kn-plugin-operator/pkg/command/remove"
	"knative.dev/kn-plugin-operator/pkg/command/uninstall"
)

var cfgFile string

// operationCmd represents the base command when called without any subcommands
func NewOperationCommand() *cobra.Command {
	p := &pkg.OperatorParams{}
	p.Initialize()
	rootCmd := &cobra.Command{
		Use:   "kn operation",
		Short: "A plugin of kn client to operate Knative components",
		Long: `kn operation: a plugin of kn client to operate Knative components.
For example:
kn operation install
kn operation install -c serving
kn operation install -c eventing
`,
	}

	rootCmd.AddCommand(install.NewInstallCommand(p))
	rootCmd.AddCommand(uninstall.NewUninstallCommand(p))
	rootCmd.AddCommand(enable.NewEnableCommand(p))
	rootCmd.AddCommand(configure.NewConfigureCommand(p))
	rootCmd.AddCommand(remove.NewRemoveCommand(p))
	return rootCmd
}
