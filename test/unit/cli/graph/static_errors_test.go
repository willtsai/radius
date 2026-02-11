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

	"github.com/radius-project/radius/pkg/cli/bicep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticGraphErrors_InvalidJSON(t *testing.T) {
	// Test handling of invalid JSON structure
	template := map[string]any{
		"resources": "not-an-array", // Invalid: resources should be an array
	}

	result, err := bicep.ParseARMTemplate(template)
	// ParseARMTemplate should handle gracefully - return empty resources, not error
	require.NoError(t, err)
	assert.Empty(t, result.Resources)
}

func TestStaticGraphErrors_MissingSchema(t *testing.T) {
	// Test handling of template without schema
	template := map[string]any{
		"contentVersion": "1.0.0.0",
		"resources":      []any{},
	}

	result, err := bicep.ParseARMTemplate(template)
	require.NoError(t, err)
	assert.Empty(t, result.Schema) // Schema not required, just empty
}

func TestStaticGraphErrors_MalformedResource(t *testing.T) {
	// Test handling of malformed resource definitions
	template := map[string]any{
		"resources": []any{
			map[string]any{
				// Missing "type" field
				"apiVersion": "2023-10-01-preview",
				"name":       "incomplete",
			},
		},
	}

	result, err := bicep.ParseARMTemplate(template)
	require.NoError(t, err)
	require.Len(t, result.Resources, 1)
	assert.Empty(t, result.Resources[0].Type) // Type will be empty
}

func TestStaticGraphErrors_NullProperties(t *testing.T) {
	// Test handling of null properties
	template := map[string]any{
		"resources": []any{
			map[string]any{
				"type":       "Applications.Core/containers",
				"apiVersion": "2023-10-01-preview",
				"name":       "mycontainer",
				"properties": nil,
			},
		},
	}

	result, err := bicep.ParseARMTemplate(template)
	require.NoError(t, err)
	require.Len(t, result.Resources, 1)
	assert.Nil(t, result.Resources[0].Properties)
}

func TestStaticGraphErrors_RequiredParametersMissing(t *testing.T) {
	// Test detection of required parameters without values
	template := map[string]any{
		"parameters": map[string]any{
			"environment": map[string]any{
				"type": "string",
				// No defaultValue - this is required
			},
			"application": map[string]any{
				"type": "string",
				// No defaultValue - this is required
			},
		},
		"resources": []any{},
	}

	result, err := bicep.ParseARMTemplate(template)
	require.NoError(t, err)

	required := result.GetRequiredParameterNames()
	assert.Len(t, required, 2)
	assert.Contains(t, required, "environment")
	assert.Contains(t, required, "application")
}

func TestStaticGraphErrors_ResourceIDConstruction(t *testing.T) {
	// Test resource ID construction handles edge cases
	extractor := bicep.NewResourceExtractor("")
	// Empty resource group should default to "default"
	assert.Equal(t, "default", extractor.ResourceGroup)

	extractor2 := bicep.NewResourceExtractor("custom-rg")
	assert.Equal(t, "custom-rg", extractor2.ResourceGroup)
}

func TestStaticGraphErrors_EmptyInputFile(t *testing.T) {
	// Test handling of empty template
	template := map[string]any{}

	result, err := bicep.ParseARMTemplate(template)
	require.NoError(t, err)
	assert.Empty(t, result.Resources)
	assert.Empty(t, result.Parameters)
}

func TestStaticGraphErrors_NestedResourcesIgnored(t *testing.T) {
	// Current implementation doesn't handle nested resources
	// This test documents the expected behavior
	template := map[string]any{
		"resources": []any{
			map[string]any{
				"type":       "Microsoft.Web/sites",
				"apiVersion": "2021-02-01",
				"name":       "mysite",
				"resources": []any{ // Nested resources
					map[string]any{
						"type":       "extensions",
						"apiVersion": "2021-02-01",
						"name":       "myextension",
					},
				},
			},
		},
	}

	result, err := bicep.ParseARMTemplate(template)
	require.NoError(t, err)
	// Only top-level resources are parsed
	require.Len(t, result.Resources, 1)
	assert.Equal(t, "mysite", result.Resources[0].Name)
}

func TestStaticGraphErrors_CopyIterationHandled(t *testing.T) {
	// Test handling of resources with copy/iteration
	template := map[string]any{
		"resources": []any{
			map[string]any{
				"type":       "Applications.Core/containers",
				"apiVersion": "2023-10-01-preview",
				"name":       "[concat('container-', copyIndex())]",
				"copy": map[string]any{
					"name":  "containerLoop",
					"count": 3,
				},
			},
		},
	}

	result, err := bicep.ParseARMTemplate(template)
	require.NoError(t, err)
	require.Len(t, result.Resources, 1)
	// Name contains expression, should be marked as dynamic
	assert.Contains(t, result.Resources[0].Name, "copyIndex")
}

func TestStaticGraphErrors_ConditionalResourceHandled(t *testing.T) {
	// Test handling of conditional resources
	template := map[string]any{
		"resources": []any{
			map[string]any{
				"type":       "Applications.Core/containers",
				"apiVersion": "2023-10-01-preview",
				"name":       "maybecontainer",
				"condition":  "[parameters('deployContainer')]",
			},
		},
	}

	result, err := bicep.ParseARMTemplate(template)
	require.NoError(t, err)
	require.Len(t, result.Resources, 1)
	assert.NotNil(t, result.Resources[0].Condition)
}
