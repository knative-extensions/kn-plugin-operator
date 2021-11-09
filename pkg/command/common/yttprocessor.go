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
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
)

// YttProcessor generates the final output yaml content, based on the base, overlay and values.
type YttProcessor struct {
	// baseData is a byte array to save the content of the base yaml
	BaseData []byte
	// overlayData is a byte array to save the content of the overlay yaml
	OverlayData []byte
	// valuesData is a byte array to save the content of the values
	ValuesData []byte
}

// GenerateOutput returns the generated content and path, based on the base, overlay and values.
func (yttp *YttProcessor) GenerateOutput() (string, error) {
	templatePath := "tpl.yml"
	overlayPath := "overlay.yml"
	valuesPath := "values.yml"

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource(templatePath, yttp.BaseData)),
		files.MustNewFileFromSource(files.NewBytesSource(overlayPath, yttp.OverlayData)),
		files.MustNewFileFromSource(files.NewBytesSource(valuesPath, yttp.ValuesData)),
	}

	ui := ui.NewTTY(false)
	opts := cmdtpl.NewOptions()
	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui)
	if out.Err != nil {
		return "", out.Err
	}

	finalFile := out.Files[0]
	finalContent := string(finalFile.Bytes())
	return finalContent, nil
}
