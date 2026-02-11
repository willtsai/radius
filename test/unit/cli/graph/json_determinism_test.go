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
	"time"

	"github.com/radius-project/radius/pkg/cli/output"
	"github.com/radius-project/radius/pkg/corerp/api/v20231001preview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeterministicJSON_SimpleMap(t *testing.T) {
	// Test that map keys are sorted alphabetically
	input := map[string]any{
		"zebra":  1,
		"apple":  2,
		"monkey": 3,
		"banana": 4,
	}

	result1, err := output.DeterministicJSON(input)
	require.NoError(t, err)

	result2, err := output.DeterministicJSON(input)
	require.NoError(t, err)

	// Results must be byte-identical
	assert.Equal(t, result1, result2)

	// Verify key ordering
	json := string(result1)
	assert.Contains(t, json, `"apple"`)

	// 'apple' should appear before 'banana' in sorted order
	appleIdx := indexOf(json, `"apple"`)
	bananaIdx := indexOf(json, `"banana"`)
	monkeyIdx := indexOf(json, `"monkey"`)
	zebraIdx := indexOf(json, `"zebra"`)

	assert.True(t, appleIdx < bananaIdx, "apple should come before banana")
	assert.True(t, bananaIdx < monkeyIdx, "banana should come before monkey")
	assert.True(t, monkeyIdx < zebraIdx, "monkey should come before zebra")
}

func TestDeterministicJSON_NestedMaps(t *testing.T) {
	// Test nested maps also have sorted keys
	input := map[string]any{
		"outer_z": map[string]any{
			"inner_b": 1,
			"inner_a": 2,
		},
		"outer_a": map[string]any{
			"inner_y": 3,
			"inner_x": 4,
		},
	}

	result, err := output.DeterministicJSON(input)
	require.NoError(t, err)

	json := string(result)

	// Verify outer keys sorted
	outerAIdx := indexOf(json, `"outer_a"`)
	outerZIdx := indexOf(json, `"outer_z"`)
	assert.True(t, outerAIdx < outerZIdx, "outer_a should come before outer_z")

	// Verify inner keys in outer_a are sorted
	innerXIdx := indexOf(json, `"inner_x"`)
	innerYIdx := indexOf(json, `"inner_y"`)
	assert.True(t, innerXIdx < innerYIdx, "inner_x should come before inner_y")
}

func TestDeterministicJSON_Arrays(t *testing.T) {
	// Arrays should maintain order (not sorted)
	input := map[string]any{
		"items": []any{"zebra", "apple", "monkey"},
	}

	result, err := output.DeterministicJSON(input)
	require.NoError(t, err)

	json := string(result)
	zebraIdx := indexOf(json, `"zebra"`)
	appleIdx := indexOf(json, `"apple"`)
	monkeyIdx := indexOf(json, `"monkey"`)

	// Array order preserved: zebra, apple, monkey
	assert.True(t, zebraIdx < appleIdx, "array order should be preserved: zebra before apple")
	assert.True(t, appleIdx < monkeyIdx, "array order should be preserved: apple before monkey")
}

func TestDeterministicJSON_StaticAppGraph(t *testing.T) {
	// Test with actual StaticAppGraph structure
	fixedTime := time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC)

	graph := &v20231001preview.StaticAppGraph{
		Metadata: v20231001preview.StaticAppGraphMetadata{
			GeneratedAt:      fixedTime,
			RadiusCliVersion: "1.0.0",
			SourceFiles:      []string{"main.bicep", "app.bicep"},
			SourceHash:       "abc123",
			GitCommit:        "def456",
		},
		Resources: []v20231001preview.StaticAppGraphResource{
			{
				ID:   "/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/api",
				Name: "api",
				Type: "Applications.Core/containers",
				SourceLocation: v20231001preview.SourceLocation{
					File: "main.bicep",
					Line: 10,
				},
			},
			{
				ID:   "/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/frontend",
				Name: "frontend",
				Type: "Applications.Core/containers",
				SourceLocation: v20231001preview.SourceLocation{
					File: "main.bicep",
					Line: 20,
				},
			},
		},
		Connections: []v20231001preview.StaticAppGraphConnection{
			{
				SourceID: "/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/frontend",
				TargetID: "/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/api",
				Type:     v20231001preview.StaticConnectionTypeConnection,
			},
		},
	}

	// Generate JSON multiple times
	result1, err := output.DeterministicJSON(graph)
	require.NoError(t, err)

	result2, err := output.DeterministicJSON(graph)
	require.NoError(t, err)

	result3, err := output.DeterministicJSON(graph)
	require.NoError(t, err)

	// All results must be byte-identical
	assert.Equal(t, result1, result2, "first and second generation should be identical")
	assert.Equal(t, result2, result3, "second and third generation should be identical")
}

func TestDeterministicJSON_EmptyStructures(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{"empty map", map[string]any{}},
		{"empty slice", []any{}},
		{"nil", nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := output.DeterministicJSON(tc.input)
			require.NoError(t, err)
			assert.NotEmpty(t, result)
		})
	}
}

func TestDeterministicJSON_SpecialCharacters(t *testing.T) {
	input := map[string]any{
		"quotes":   `value with "quotes"`,
		"newlines": "line1\nline2",
		"unicode":  "💻 code emoji",
		"html":     "<script>alert('xss')</script>",
	}

	result, err := output.DeterministicJSON(input)
	require.NoError(t, err)

	// Should properly escape special characters
	json := string(result)
	assert.NotContains(t, json, `<script>`, "HTML should be escaped")
}

func TestMarshalDeterministicJSON_Alias(t *testing.T) {
	// Test that DeterministicJSON and MarshalDeterministicJSON produce same output
	input := map[string]any{"key": "value"}

	result1, err1 := output.DeterministicJSON(input)
	result2, err2 := output.MarshalDeterministicJSON(input)

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Equal(t, result1, result2)
}

// indexOf returns the index of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
