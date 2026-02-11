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

func TestParseARMTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template map[string]any
		validate func(t *testing.T, result *bicep.ARMTemplate)
	}{
		{
			name: "empty template",
			template: map[string]any{
				"$schema":        "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
				"contentVersion": "1.0.0.0",
			},
			validate: func(t *testing.T, result *bicep.ARMTemplate) {
				assert.Equal(t, "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#", result.Schema)
				assert.Equal(t, "1.0.0.0", result.ContentVersion)
				assert.Empty(t, result.Resources)
			},
		},
		{
			name: "template with resources",
			template: map[string]any{
				"$schema":        "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
				"contentVersion": "1.0.0.0",
				"resources": []any{
					map[string]any{
						"type":       "Applications.Core/containers",
						"apiVersion": "2023-10-01-preview",
						"name":       "frontend",
						"properties": map[string]any{
							"container": map[string]any{
								"image": "myapp/frontend:v1",
							},
						},
					},
					map[string]any{
						"type":       "Applications.Core/containers",
						"apiVersion": "2023-10-01-preview",
						"name":       "backend",
						"dependsOn": []any{
							"frontend",
						},
					},
				},
			},
			validate: func(t *testing.T, result *bicep.ARMTemplate) {
				require.Len(t, result.Resources, 2)
				assert.Equal(t, "Applications.Core/containers", result.Resources[0].Type)
				assert.Equal(t, "frontend", result.Resources[0].Name)
				assert.Equal(t, "backend", result.Resources[1].Name)
				assert.Contains(t, result.Resources[1].DependsOn, "frontend")
			},
		},
		{
			name: "template with parameters",
			template: map[string]any{
				"$schema":        "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
				"contentVersion": "1.0.0.0",
				"parameters": map[string]any{
					"appName": map[string]any{
						"type": "string",
					},
					"environment": map[string]any{
						"type":         "string",
						"defaultValue": "dev",
					},
				},
			},
			validate: func(t *testing.T, result *bicep.ARMTemplate) {
				require.Len(t, result.Parameters, 2)
				assert.Equal(t, "string", result.Parameters["appName"].Type)
				assert.Nil(t, result.Parameters["appName"].DefaultValue)
				assert.Equal(t, "dev", result.Parameters["environment"].DefaultValue)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bicep.ParseARMTemplate(tt.template)
			require.NoError(t, err)
			tt.validate(t, result)
		})
	}
}

func TestGetRequiredParameterNames(t *testing.T) {
	template := map[string]any{
		"parameters": map[string]any{
			"required1": map[string]any{
				"type": "string",
			},
			"optional1": map[string]any{
				"type":         "string",
				"defaultValue": "default",
			},
			"required2": map[string]any{
				"type": "int",
			},
		},
	}

	parsed, err := bicep.ParseARMTemplate(template)
	require.NoError(t, err)

	required := parsed.GetRequiredParameterNames()
	assert.Len(t, required, 2)
	assert.Contains(t, required, "required1")
	assert.Contains(t, required, "required2")
	assert.NotContains(t, required, "optional1")
}

func TestExtractResources(t *testing.T) {
	tests := []struct {
		name         string
		template     map[string]any
		expectedLen  int
		validateFunc func(t *testing.T, resources []bicep.ARMResource)
	}{
		{
			name: "simple container resources",
			template: map[string]any{
				"resources": []any{
					map[string]any{
						"type":       "Applications.Core/containers",
						"apiVersion": "2023-10-01-preview",
						"name":       "frontend",
					},
					map[string]any{
						"type":       "Applications.Core/containers",
						"apiVersion": "2023-10-01-preview",
						"name":       "backend",
					},
				},
			},
			expectedLen: 2,
			validateFunc: func(t *testing.T, resources []bicep.ARMResource) {
				assert.Equal(t, "frontend", resources[0].Name)
				assert.Equal(t, "backend", resources[1].Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := bicep.ParseARMTemplate(tt.template)
			require.NoError(t, err)
			resources := parsed.GetResources()
			assert.Len(t, resources, tt.expectedLen)
			if tt.validateFunc != nil {
				tt.validateFunc(t, resources)
			}
		})
	}
}
