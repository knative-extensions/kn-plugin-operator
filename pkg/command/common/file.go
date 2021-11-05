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
	"io/ioutil"
	"os"
)

// WriteFile creates a file with the string content
func WriteFile(path, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	data := []byte(content)
	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// ReadFile returns the content of the file as a string
func ReadFile(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// DeleteFile deletes the file
func DeleteFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		return err
	}

	return nil
}
