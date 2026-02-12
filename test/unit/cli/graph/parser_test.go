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
			name: "v2 template with resources as map",
			template: map[string]any{
				"$schema":         "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
				"contentVersion":  "1.0.0.0",
				"languageVersion": "2.0",
				"resources": map[string]any{
					"myapp": map[string]any{
						"type": "Applications.Core/applications@2023-10-01-preview",
						"properties": map[string]any{
							"name":        "my-app",
							"environment": "[parameters('environment')]",
						},
					},
					"frontend": map[string]any{
						"type": "Applications.Core/containers@2023-10-01-preview",
						"properties": map[string]any{
							"name":        "frontend",
							"application": "[resourceInfo('myapp').id]",
							"container": map[string]any{
								"image": "ghcr.io/myapp/frontend:latest",
							},
						},
						"dependsOn": []any{
							"myapp",
						},
					},
				},
			},
			validate: func(t *testing.T, result *bicep.ARMTemplate) {
				require.Len(t, result.Resources, 2)

				// Build a map by name for order-independent assertions
				resourcesByName := make(map[string]bicep.ARMResource)
				for _, r := range result.Resources {
					resourcesByName[r.Name] = r
				}

				app := resourcesByName["my-app"]
				assert.Equal(t, "Applications.Core/applications", app.Type)
				assert.Equal(t, "2023-10-01-preview", app.APIVersion)
				assert.Equal(t, "my-app", app.Name)
				assert.Equal(t, "myapp", app.SymbolicName)

				frontend := resourcesByName["frontend"]
				assert.Equal(t, "Applications.Core/containers", frontend.Type)
				assert.Equal(t, "2023-10-01-preview", frontend.APIVersion)
				assert.Equal(t, "frontend", frontend.Name)
				assert.Equal(t, "frontend", frontend.SymbolicName)
				assert.Contains(t, frontend.DependsOn, "myapp")
			},
		},
		{
			name: "v2 template with double-nested properties",
			template: map[string]any{
				"$schema":         "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
				"contentVersion":  "1.0.0.0",
				"languageVersion": "2.0",
				"resources": map[string]any{
					"app": map[string]any{
						"type": "Applications.Core/applications@2023-10-01-preview",
						"properties": map[string]any{
							"name": "testapp",
							"properties": map[string]any{
								"environment": "[parameters('environment')]",
							},
						},
					},
					"demo": map[string]any{
						"type": "Applications.Core/containers@2023-10-01-preview",
						"properties": map[string]any{
							"name": "demo",
							"properties": map[string]any{
								"application": "[reference('app').id]",
								"connections": map[string]any{
									"sql": map[string]any{
										"source": "[reference('db').id]",
									},
									"redis": map[string]any{
										"source": "[reference('cache').id]",
									},
								},
								"container": map[string]any{
									"image": "ghcr.io/demo:latest",
								},
							},
						},
						"dependsOn": []any{
							"app",
							"cache",
							"db",
						},
					},
					"db": map[string]any{
						"type": "Applications.Datastores/sqlDatabases@2023-10-01-preview",
						"properties": map[string]any{
							"name": "mydb",
							"properties": map[string]any{
								"environment": "[parameters('environment')]",
							},
						},
					},
					"cache": map[string]any{
						"type": "Applications.Datastores/redisCaches@2023-10-01-preview",
						"properties": map[string]any{
							"name": "mycache",
							"properties": map[string]any{
								"environment": "[parameters('environment')]",
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result *bicep.ARMTemplate) {
				require.Len(t, result.Resources, 4)

				resourcesByName := make(map[string]bicep.ARMResource)
				for _, r := range result.Resources {
					resourcesByName[r.Name] = r
				}

				// Verify double-nested properties are flattened
				demo := resourcesByName["demo"]
				assert.Equal(t, "demo", demo.SymbolicName)
				assert.Equal(t, "Applications.Core/containers", demo.Type)
				// Properties should contain the INNER properties (connections, container)
				_, hasConnections := demo.Properties["connections"]
				assert.True(t, hasConnections, "flattened properties should contain 'connections'")
				_, hasContainer := demo.Properties["container"]
				assert.True(t, hasContainer, "flattened properties should contain 'container'")

				// Verify symbolic names differ from resource names
				db := resourcesByName["mydb"]
				assert.Equal(t, "db", db.SymbolicName)
				assert.Equal(t, "mydb", db.Name)

				cache := resourcesByName["mycache"]
				assert.Equal(t, "cache", cache.SymbolicName)
				assert.Equal(t, "mycache", cache.Name)
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

func TestExtractConnections_V2MultipleConnections(t *testing.T) {
	// Test that v2 ARM JSON with double-nested properties and symbolic names
	// correctly detects ALL connections (not just the ones where symbolic name == resource name).
	template := map[string]any{
		"$schema":         "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
		"contentVersion":  "1.0.0.0",
		"languageVersion": "2.0",
		"resources": map[string]any{
			"app": map[string]any{
				"type": "Applications.Core/applications@2023-10-01-preview",
				"properties": map[string]any{
					"name": "testapp",
					"properties": map[string]any{
						"environment": "[parameters('environment')]",
					},
				},
			},
			"demo": map[string]any{
				"type": "Applications.Core/containers@2023-10-01-preview",
				"properties": map[string]any{
					"name": "demo",
					"properties": map[string]any{
						"application": "[reference('app').id]",
						"connections": map[string]any{
							"sql": map[string]any{
								"source": "[reference('db').id]",
							},
							"redis": map[string]any{
								"source": "[reference('cache').id]",
							},
						},
						"container": map[string]any{
							"image": "ghcr.io/demo:latest",
						},
					},
				},
				"dependsOn": []any{
					"app",
					"cache",
					"db",
				},
			},
			"db": map[string]any{
				"type": "Applications.Datastores/sqlDatabases@2023-10-01-preview",
				"properties": map[string]any{
					"name": "mydb",
					"properties": map[string]any{
						"environment": "[parameters('environment')]",
					},
				},
				"dependsOn": []any{
					"app",
				},
			},
			"cache": map[string]any{
				"type": "Applications.Datastores/redisCaches@2023-10-01-preview",
				"properties": map[string]any{
					"name": "mycache",
					"properties": map[string]any{
						"environment": "[parameters('environment')]",
					},
				},
				"dependsOn": []any{
					"app",
				},
			},
		},
	}

	parsed, err := bicep.ParseARMTemplate(template)
	require.NoError(t, err)

	connections := bicep.ExtractConnections(parsed)

	// Build a set of source→target pairs for easy checking
	connPairs := make(map[string]bool)
	for _, c := range connections {
		connPairs[c.SourceResourceID+"->"+c.TargetResourceID] = true
	}

	// Explicit connections via reference(): demo→mydb (via symbolic "db") and demo→mycache (via symbolic "cache")
	assert.True(t, connPairs["demo->mydb"], "should detect connection from demo to mydb via reference('db')")
	assert.True(t, connPairs["demo->mycache"], "should detect connection from demo to mycache via reference('cache')")

	// dependsOn connections: demo→testapp (via symbolic "app"), demo→mycache (via "cache"), demo→mydb (via "db")
	assert.True(t, connPairs["demo->testapp"], "should detect dependsOn from demo to testapp via symbolic 'app'")
}
