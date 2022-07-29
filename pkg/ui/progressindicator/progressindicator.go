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

package progressindicator

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

var (
	DefaultRefreshRate = time.Millisecond * 300
	DefaultCharset     = spinner.CharSets[35]
	DefaultColor       = "green"
)

// ProgressIndicator is used to indicator the progress of the running command.
type ProgressIndicator struct {
	spiner *spinner.Spinner
}

// New creates a new progress indicator.
func New() *ProgressIndicator {
	sp := spinner.New(DefaultCharset, DefaultRefreshRate)
	sp.Color(DefaultColor)
	pi := &ProgressIndicator{
		spiner: sp,
	}
	return pi.SetText("Initializing...")
}

// SetText sets the text for progress indicator.
func (pi *ProgressIndicator) SetText(text string) *ProgressIndicator {
	pi.spiner.Lock()
	pi.spiner.Suffix = fmt.Sprintf(" %s", text)
	pi.spiner.Unlock()
	return pi
}

// SetPrefix sets the prefix for progress indicator.
func (pi *ProgressIndicator) SetPrefix(text string) *ProgressIndicator {
	pi.spiner.Lock()
	pi.spiner.Prefix = fmt.Sprintf("%s ", text)
	pi.spiner.Unlock()
	return pi
}

// SetCharset sets the prefix for progress indicator.
func (pi *ProgressIndicator) SetCharset(charset []string) *ProgressIndicator {
	pi.spiner.UpdateCharSet(charset)
	return pi
}

// SetColor sets the prefix for progress indicator.
func (pi *ProgressIndicator) SetColor(color string) *ProgressIndicator {
	pi.spiner.Color(color)
	return pi
}

// Start starts the progress indicator.
func (pi *ProgressIndicator) Start() *ProgressIndicator {
	pi.spiner.Start()
	return pi
}

// Stop stops the progress indicator.
func (pi *ProgressIndicator) Stop() *ProgressIndicator {
	pi.spiner.Stop()
	pi.spiner.Prefix = ""
	pi.spiner.Color(DefaultColor)
	pi.spiner.UpdateCharSet(DefaultCharset)
	pi.spiner.Stop()
	return pi
}

// IsActive returns whether the progress indicator is active or not.
func (pi *ProgressIndicator) IsActive() bool {
	return pi.spiner.Active()
}
