/*
Copyright 2023 The Radius Authors.

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

package graph

import (
	"path/filepath"
	"strings"
)

// InputType represents the type of input provided to the graph command.
type InputType int

const (
	// InputTypeAppName indicates the input is an application name (existing behavior)
	InputTypeAppName InputType = iota

	// InputTypeBicepFile indicates the input is a Bicep file path
	InputTypeBicepFile

	// InputTypeUnknown indicates the input type could not be determined
	InputTypeUnknown
)

// DetectInputType determines whether the input is an application name or a Bicep file.
//
// The detection logic:
// - If the input ends with ".bicep", it's a Bicep file
// - Otherwise, it's treated as an application name
//
// This allows for intuitive disambiguation:
//
//	rad app graph myapp       -> Application name
//	rad app graph app.bicep   -> Bicep file
func DetectInputType(input string) InputType {
	if input == "" {
		return InputTypeUnknown
	}

	// Check for .bicep extension (case-insensitive)
	ext := strings.ToLower(filepath.Ext(input))
	if ext == ".bicep" {
		return InputTypeBicepFile
	}

	// Otherwise, treat as application name
	return InputTypeAppName
}

// IsBicepFile returns true if the input appears to be a Bicep file path.
func IsBicepFile(input string) bool {
	return DetectInputType(input) == InputTypeBicepFile
}

// IsAppName returns true if the input appears to be an application name.
func IsAppName(input string) bool {
	return DetectInputType(input) == InputTypeAppName
}

// DefaultOutputPath returns the default output path for a given Bicep file.
// The default is .radius/app-graph.json in the same directory as the Bicep file.
func DefaultOutputPath(bicepFilePath string) string {
	dir := filepath.Dir(bicepFilePath)
	return filepath.Join(dir, ".radius", "app-graph.json")
}

// DefaultMarkdownOutputPath returns the default Markdown output path for a given Bicep file.
// The default is .radius/app-graph.md in the same directory as the Bicep file.
func DefaultMarkdownOutputPath(bicepFilePath string) string {
	dir := filepath.Dir(bicepFilePath)
	return filepath.Join(dir, ".radius", "app-graph.md")
}

// EnsureRadiusDir ensures the .radius directory exists relative to the given path.
func EnsureRadiusDir(basePath string) (string, error) {
	dir := filepath.Dir(basePath)
	radiusDir := filepath.Join(dir, ".radius")
	return radiusDir, nil
}
