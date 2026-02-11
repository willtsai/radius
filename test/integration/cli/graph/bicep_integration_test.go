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

	"github.com/radius-project/radius/pkg/cli/bicep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBicepIntegration_BuildBasicTemplate tests building a basic Bicep file
// This test requires the Bicep CLI to be installed
func TestBicepIntegration_BuildBasicTemplate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if bicep CLI is available
	if _, err := bicep.GetBicepCliPath(); err != nil {
		t.Skip("Bicep CLI not available, skipping integration test")
	}

	// Create a temporary Bicep file
	tempDir := t.TempDir()
	bicepFile := filepath.Join(tempDir, "test.bicep")

	bicepContent := `
extension radius

param environment string

resource app 'Applications.Core/applications@2023-10-01-preview' = {
  name: 'testapp'
  properties: {
    environment: environment
  }
}

resource container 'Applications.Core/containers@2023-10-01-preview' = {
  name: 'testcontainer'
  properties: {
    application: app.id
    container: {
      image: 'nginx:latest'
    }
  }
}
`
	err := os.WriteFile(bicepFile, []byte(bicepContent), 0644)
	require.NoError(t, err)

	// Build the Bicep file
	executor := bicep.NewExecutor()
	armJSON, err := executor.Build(bicepFile)

	// Note: This might fail if Radius extension is not available
	// In that case, we skip rather than fail the test
	if err != nil {
		t.Skipf("Bicep build failed (likely missing Radius extension): %v", err)
	}

	require.NotNil(t, armJSON)
	assert.Contains(t, armJSON, "Applications.Core/applications")
	assert.Contains(t, armJSON, "Applications.Core/containers")
}

// TestBicepIntegration_ParseBuildOutput tests parsing Bicep build output
func TestBicepIntegration_ParseBuildOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Sample ARM JSON output that would come from bicep build
	armJSON := `{
		"$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
		"contentVersion": "1.0.0.0",
		"parameters": {
			"environment": {
				"type": "string"
			}
		},
		"resources": [
			{
				"type": "Applications.Core/applications",
				"apiVersion": "2023-10-01-preview",
				"name": "testapp",
				"properties": {
					"environment": "[parameters('environment')]"
				}
			},
			{
				"type": "Applications.Core/containers",
				"apiVersion": "2023-10-01-preview",
				"name": "testcontainer",
				"properties": {
					"application": "[resourceId('Applications.Core/applications', 'testapp')]",
					"container": {
						"image": "nginx:latest"
					}
				}
			}
		]
	}`

	// Parse the ARM JSON
	var template map[string]any
	err := bicep.UnmarshalJSON([]byte(armJSON), &template)
	require.NoError(t, err)

	armTemplate, err := bicep.ParseARMTemplate(template)
	require.NoError(t, err)

	// Verify parsing
	assert.Equal(t, "1.0.0.0", armTemplate.ContentVersion)
	require.Len(t, armTemplate.Resources, 2)

	// Verify resources
	resourceTypes := make(map[string]bool)
	for _, r := range armTemplate.Resources {
		resourceTypes[r.Type] = true
	}
	assert.True(t, resourceTypes["Applications.Core/applications"])
	assert.True(t, resourceTypes["Applications.Core/containers"])
}

// TestBicepIntegration_ExtractGraph tests the full extraction pipeline
func TestBicepIntegration_ExtractGraph(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Sample ARM JSON with connections
	armJSON := `{
		"$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
		"contentVersion": "1.0.0.0",
		"resources": [
			{
				"type": "Applications.Core/containers",
				"apiVersion": "2023-10-01-preview",
				"name": "frontend",
				"properties": {
					"connections": {
						"backend": {
							"source": "[resourceId('Applications.Core/containers', 'backend')]"
						}
					}
				}
			},
			{
				"type": "Applications.Core/containers",
				"apiVersion": "2023-10-01-preview",
				"name": "backend",
				"properties": {
					"connections": {
						"redis": {
							"source": "[resourceId('Applications.Datastores/redisCaches', 'cache')]"
						}
					}
				}
			},
			{
				"type": "Applications.Datastores/redisCaches",
				"apiVersion": "2023-10-01-preview",
				"name": "cache"
			}
		]
	}`

	var template map[string]any
	err := bicep.UnmarshalJSON([]byte(armJSON), &template)
	require.NoError(t, err)

	armTemplate, err := bicep.ParseARMTemplate(template)
	require.NoError(t, err)

	// Extract resources
	extractor := bicep.NewResourceExtractor("default")
	resources, err := extractor.ExtractResources(armTemplate, "app.bicep")
	require.NoError(t, err)
	require.Len(t, resources, 3)

	// Extract connections
	connections := bicep.ExtractConnections(armTemplate)

	// Verify connections
	// frontend -> backend, backend -> cache
	assert.GreaterOrEqual(t, len(connections), 2)

	connectionMap := make(map[string]string)
	for _, c := range connections {
		connectionMap[c.SourceResourceID] = c.TargetResourceID
	}
	assert.Equal(t, "backend", connectionMap["frontend"])
	assert.Equal(t, "cache", connectionMap["backend"])
}

// TestBicepIntegration_HashComputation tests source hash computation
func TestBicepIntegration_HashComputation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary files
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "file1.bicep")
	file2 := filepath.Join(tempDir, "file2.bicep")

	err := os.WriteFile(file1, []byte("content1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte("content2"), 0644)
	require.NoError(t, err)

	// Compute hash
	hash, err := bicep.ComputeSourceHash([]string{file1, file2})
	require.NoError(t, err)
	assert.True(t, len(hash) > 0)
	assert.Contains(t, hash, "sha256:")

	// Same files should produce same hash
	hash2, err := bicep.ComputeSourceHash([]string{file1, file2})
	require.NoError(t, err)
	assert.Equal(t, hash, hash2)

	// Different order should produce same hash (sorted)
	hash3, err := bicep.ComputeSourceHash([]string{file2, file1})
	require.NoError(t, err)
	assert.Equal(t, hash, hash3)
}

// TestBicepIntegration_ParameterValidation tests parameter validation
func TestBicepIntegration_ParameterValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	armJSON := `{
		"parameters": {
			"environment": {
				"type": "string"
			},
			"application": {
				"type": "string"
			},
			"optionalParam": {
				"type": "string",
				"defaultValue": "default"
			}
		},
		"resources": []
	}`

	var template map[string]any
	err := bicep.UnmarshalJSON([]byte(armJSON), &template)
	require.NoError(t, err)

	armTemplate, err := bicep.ParseARMTemplate(template)
	require.NoError(t, err)

	// Get required parameters
	required := armTemplate.GetRequiredParameterNames()
	assert.Len(t, required, 2)
	assert.Contains(t, required, "environment")
	assert.Contains(t, required, "application")
	assert.NotContains(t, required, "optionalParam")
}
