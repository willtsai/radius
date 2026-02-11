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

package graph_test

import (
	"testing"

	"github.com/radius-project/radius/pkg/cli/cmd/app/graph"
	"github.com/stretchr/testify/assert"
)

func TestDetectInputType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected graph.InputType
	}{
		{
			name:     "empty input",
			input:    "",
			expected: graph.InputTypeUnknown,
		},
		{
			name:     "bicep file lowercase",
			input:    "app.bicep",
			expected: graph.InputTypeBicepFile,
		},
		{
			name:     "bicep file uppercase",
			input:    "App.BICEP",
			expected: graph.InputTypeBicepFile,
		},
		{
			name:     "bicep file with path",
			input:    "/path/to/app.bicep",
			expected: graph.InputTypeBicepFile,
		},
		{
			name:     "bicep file relative path",
			input:    "./myapp/app.bicep",
			expected: graph.InputTypeBicepFile,
		},
		{
			name:     "application name simple",
			input:    "myapp",
			expected: graph.InputTypeAppName,
		},
		{
			name:     "application name with dashes",
			input:    "my-cool-app",
			expected: graph.InputTypeAppName,
		},
		{
			name:     "application name that looks like file",
			input:    "myapp.json",
			expected: graph.InputTypeAppName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := graph.DetectInputType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsBicepFile(t *testing.T) {
	assert.True(t, graph.IsBicepFile("app.bicep"))
	assert.True(t, graph.IsBicepFile("/path/to/app.bicep"))
	assert.False(t, graph.IsBicepFile("myapp"))
	assert.False(t, graph.IsBicepFile(""))
}

func TestIsAppName(t *testing.T) {
	assert.True(t, graph.IsAppName("myapp"))
	assert.True(t, graph.IsAppName("my-app"))
	assert.False(t, graph.IsAppName("app.bicep"))
	assert.False(t, graph.IsAppName(""))
}

func TestDefaultOutputPath(t *testing.T) {
	tests := []struct {
		name      string
		bicepPath string
		expected  string
	}{
		{
			name:      "simple file",
			bicepPath: "app.bicep",
			expected:  ".radius/app-graph.json",
		},
		{
			name:      "file in directory",
			bicepPath: "myapp/app.bicep",
			expected:  "myapp/.radius/app-graph.json",
		},
		{
			name:      "absolute path",
			bicepPath: "/home/user/project/app.bicep",
			expected:  "/home/user/project/.radius/app-graph.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := graph.DefaultOutputPath(tt.bicepPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultMarkdownOutputPath(t *testing.T) {
	tests := []struct {
		name      string
		bicepPath string
		expected  string
	}{
		{
			name:      "simple file",
			bicepPath: "app.bicep",
			expected:  ".radius/app-graph.md",
		},
		{
			name:      "file in directory",
			bicepPath: "myapp/app.bicep",
			expected:  "myapp/.radius/app-graph.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := graph.DefaultMarkdownOutputPath(tt.bicepPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}
