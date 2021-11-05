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
	"errors"
	"os"
	"testing"

	"knative.dev/kn-plugin-operator/pkg/command/testingUtil"
)

func TestReadWriteDeleteFile(t *testing.T) {
	expContent, err := ReadFile("testdata/test.txt")
	testingUtil.AssertEqual(t, err == nil, true)

	path := "temp.txt"
	err = WriteFile(path, expContent)
	testingUtil.AssertEqual(t, err == nil, true)

	content, err := ReadFile("testdata/test.txt")
	testingUtil.AssertEqual(t, err == nil, true)
	testingUtil.AssertEqual(t, content, expContent)

	// Make sure the file exists
	_, err = os.Stat(path)
	testingUtil.AssertEqual(t, err == nil, true)

	err = DeleteFile(path)
	testingUtil.AssertEqual(t, err == nil, true)

	// Makes sure the file is deleted
	_, err = os.Stat(path)
	testingUtil.AssertEqual(t, errors.Is(err, os.ErrNotExist), true)
}
