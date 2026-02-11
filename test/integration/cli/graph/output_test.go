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
	"os"
	"path/filepath"
	"testing"

	"github.com/radius-project/radius/pkg/cli/cmd/app/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultOutputPath_SameDirectory(t *testing.T) {
	bicepPath := "/project/app.bicep"
	expected := "/project/.radius/app-graph.json"

	result := graph.DefaultOutputPath(bicepPath)
	assert.Equal(t, expected, result)
}

func TestDefaultOutputPath_NestedDirectory(t *testing.T) {
	bicepPath := "/project/src/infra/main.bicep"
	expected := "/project/src/infra/.radius/app-graph.json"

	result := graph.DefaultOutputPath(bicepPath)
	assert.Equal(t, expected, result)
}

func TestDefaultOutputPath_RelativePath(t *testing.T) {
	bicepPath := "app.bicep"
	result := graph.DefaultOutputPath(bicepPath)

	// Should be relative path with .radius subdirectory
	assert.Contains(t, result, ".radius")
	assert.Contains(t, result, "app-graph.json")
}

func TestDefaultMarkdownOutputPath(t *testing.T) {
	bicepPath := "/project/app.bicep"
	expected := "/project/.radius/app-graph.md"

	result := graph.DefaultMarkdownOutputPath(bicepPath)
	assert.Equal(t, expected, result)
}

func TestInputType_Detection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected graph.InputType
	}{
		{
			name:     "bicep file lowercase",
			input:    "app.bicep",
			expected: graph.InputTypeBicepFile,
		},
		{
			name:     "bicep file uppercase",
			input:    "APP.BICEP",
			expected: graph.InputTypeBicepFile,
		},
		{
			name:     "bicep file with path",
			input:    "/path/to/main.bicep",
			expected: graph.InputTypeBicepFile,
		},
		{
			name:     "app name simple",
			input:    "myapp",
			expected: graph.InputTypeAppName,
		},
		{
			name:     "app name with dashes",
			input:    "my-cool-app",
			expected: graph.InputTypeAppName,
		},
		{
			name:     "empty input",
			input:    "",
			expected: graph.InputTypeUnknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := graph.DetectInputType(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsBicepFile(t *testing.T) {
	assert.True(t, graph.IsBicepFile("app.bicep"))
	assert.True(t, graph.IsBicepFile("main.BICEP"))
	assert.True(t, graph.IsBicepFile("/path/to/file.bicep"))

	assert.False(t, graph.IsBicepFile("myapp"))
	assert.False(t, graph.IsBicepFile("app.json"))
	assert.False(t, graph.IsBicepFile(""))
}

func TestIsAppName(t *testing.T) {
	assert.True(t, graph.IsAppName("myapp"))
	assert.True(t, graph.IsAppName("my-app"))
	assert.True(t, graph.IsAppName("app123"))

	assert.False(t, graph.IsAppName("app.bicep"))
	assert.False(t, graph.IsAppName(""))
}

func TestOutputFileCreation(t *testing.T) {
	// Create a temp directory for testing
	tempDir, err := os.MkdirTemp("", "radius-graph-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Simulate creating output directory structure
	bicepPath := filepath.Join(tempDir, "app.bicep")
	outputPath := graph.DefaultOutputPath(bicepPath)

	// Create the .radius directory
	outputDir := filepath.Dir(outputPath)
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err)

	// Verify the directory was created
	info, err := os.Stat(outputDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Write test output
	testContent := []byte(`{"test": "data"}`)
	err = os.WriteFile(outputPath, testContent, 0644)
	require.NoError(t, err)

	// Verify file exists and has correct content
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, testContent, content)
}

func TestOutputPathsAreValid(t *testing.T) {
	// Test that generated paths are valid filesystem paths
	testCases := []string{
		"/absolute/path/app.bicep",
		"relative/path/app.bicep",
		"./local/app.bicep",
		"../parent/app.bicep",
		"app.bicep",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			jsonPath := graph.DefaultOutputPath(tc)
			mdPath := graph.DefaultMarkdownOutputPath(tc)

			// Paths should end with expected filenames
			assert.True(t, filepath.Base(jsonPath) == "app-graph.json")
			assert.True(t, filepath.Base(mdPath) == "app-graph.md")

			// Parent directory should be .radius
			assert.True(t, filepath.Base(filepath.Dir(jsonPath)) == ".radius")
			assert.True(t, filepath.Base(filepath.Dir(mdPath)) == ".radius")
		})
	}
}

func TestOutputPathHandlesSpecialCharacters(t *testing.T) {
	// Test paths with special characters (platform-dependent)
	bicepPath := "/path/with spaces/my app.bicep"
	jsonPath := graph.DefaultOutputPath(bicepPath)

	// Should contain the same directory structure
	assert.Contains(t, jsonPath, "with spaces")
	assert.True(t, filepath.Base(jsonPath) == "app-graph.json")
}
