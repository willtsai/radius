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
	"github.com/radius-project/radius/pkg/corerp/api/v20231001preview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticGraphGeneration_SimpleContainer(t *testing.T) {
	// Test that a simple container resource is correctly extracted
	template := map[string]any{
		"$schema":        "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
		"contentVersion": "1.0.0.0",
		"resources": []any{
			map[string]any{
				"type":       "Applications.Core/containers",
				"apiVersion": "2023-10-01-preview",
				"name":       "mycontainer",
				"properties": map[string]any{
					"application": "[parameters('application')]",
					"container": map[string]any{
						"image": "myregistry/myapp:v1",
					},
				},
			},
		},
	}

	armTemplate, err := bicep.ParseARMTemplate(template)
	require.NoError(t, err)

	extractor := bicep.NewResourceExtractor("default")
	resources, err := extractor.ExtractResources(armTemplate, "app.bicep")
	require.NoError(t, err)

	require.Len(t, resources, 1)
	assert.Equal(t, "mycontainer", resources[0].Name)
	assert.Equal(t, "Applications.Core/containers", resources[0].Type)
	assert.Equal(t, "app.bicep", resources[0].SourceLocation.File)
}

func TestStaticGraphGeneration_MultipleResources(t *testing.T) {
	// Test graph generation with multiple resources
	template := map[string]any{
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
			map[string]any{
				"type":       "Applications.Datastores/redisCaches",
				"apiVersion": "2023-10-01-preview",
				"name":       "cache",
			},
		},
	}

	armTemplate, err := bicep.ParseARMTemplate(template)
	require.NoError(t, err)

	extractor := bicep.NewResourceExtractor("default")
	resources, err := extractor.ExtractResources(armTemplate, "app.bicep")
	require.NoError(t, err)

	require.Len(t, resources, 3)

	// Verify each resource
	resourceNames := make(map[string]bool)
	for _, r := range resources {
		resourceNames[r.Name] = true
	}
	assert.True(t, resourceNames["frontend"])
	assert.True(t, resourceNames["backend"])
	assert.True(t, resourceNames["cache"])
}

func TestStaticGraphGeneration_WithConnections(t *testing.T) {
	// Test connection extraction
	template := map[string]any{
		"resources": []any{
			map[string]any{
				"type":       "Applications.Core/containers",
				"apiVersion": "2023-10-01-preview",
				"name":       "frontend",
				"properties": map[string]any{
					"connections": map[string]any{
						"backend": map[string]any{
							"source": "[resourceId('Applications.Core/containers', 'backend')]",
						},
					},
				},
			},
			map[string]any{
				"type":       "Applications.Core/containers",
				"apiVersion": "2023-10-01-preview",
				"name":       "backend",
			},
		},
	}

	armTemplate, err := bicep.ParseARMTemplate(template)
	require.NoError(t, err)

	// Extract connections
	connections := bicep.ExtractConnections(armTemplate)
	require.Len(t, connections, 1)
	assert.Equal(t, "frontend", connections[0].SourceResourceID)
	assert.Equal(t, "backend", connections[0].TargetResourceID)
}

func TestStaticGraphGeneration_ResourceTypes(t *testing.T) {
	// Test that different Radius resource types are recognized
	tests := []struct {
		resourceType string
		isRadius     bool
	}{
		{"Applications.Core/containers", true},
		{"Applications.Core/gateways", true},
		{"Applications.Core/environments", true},
		{"Applications.Datastores/redisCaches", true},
		{"Applications.Datastores/sqlDatabases", true},
		{"Applications.Messaging/rabbitMQQueues", true},
		{"Applications.Dapr/stateStores", true},
		{"Microsoft.Storage/storageAccounts", false},
		{"Microsoft.Compute/virtualMachines", false},
	}

	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			result := bicep.IsRadiusResource(tt.resourceType)
			assert.Equal(t, tt.isRadius, result)
		})
	}
}

func TestStaticGraphGeneration_ResourceProviderExtraction(t *testing.T) {
	// Test provider extraction from resource type
	tests := []struct {
		resourceType     string
		expectedProvider string
		expectedTypeName string
	}{
		{"Applications.Core/containers", "Applications.Core", "containers"},
		{"Applications.Datastores/redisCaches", "Applications.Datastores", "redisCaches"},
		{"Microsoft.Storage/storageAccounts", "Microsoft.Storage", "storageAccounts"},
	}

	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			provider := bicep.GetResourceProvider(tt.resourceType)
			typeName := bicep.GetResourceTypeName(tt.resourceType)
			assert.Equal(t, tt.expectedProvider, provider)
			assert.Equal(t, tt.expectedTypeName, typeName)
		})
	}
}

func TestStaticGraphGeneration_MetadataPopulated(t *testing.T) {
	// Test that graph metadata is properly populated
	graph := v20231001preview.StaticAppGraph{
		Metadata: v20231001preview.StaticAppGraphMetadata{
			RadiusCliVersion: "0.30.0",
			SourceFiles:      []string{"app.bicep"},
			SourceHash:       "sha256:abc123",
		},
		Resources:   []v20231001preview.StaticAppGraphResource{},
		Connections: []v20231001preview.StaticAppGraphConnection{},
	}

	assert.Equal(t, "0.30.0", graph.Metadata.RadiusCliVersion)
	assert.Len(t, graph.Metadata.SourceFiles, 1)
	assert.NotEmpty(t, graph.Metadata.SourceHash)
}

func TestStaticGraphGeneration_DependsOnRelationships(t *testing.T) {
	// Test explicit dependsOn relationships
	template := map[string]any{
		"resources": []any{
			map[string]any{
				"type":       "Applications.Core/containers",
				"apiVersion": "2023-10-01-preview",
				"name":       "frontend",
				"dependsOn": []any{
					"backend",
				},
			},
			map[string]any{
				"type":       "Applications.Core/containers",
				"apiVersion": "2023-10-01-preview",
				"name":       "backend",
			},
		},
	}

	armTemplate, err := bicep.ParseARMTemplate(template)
	require.NoError(t, err)

	// Verify dependsOn is captured
	resources := armTemplate.GetResources()
	require.Len(t, resources, 2)

	// Find frontend resource
	var frontend *bicep.ARMResource
	for i := range resources {
		if resources[i].Name == "frontend" {
			frontend = &resources[i]
			break
		}
	}
	require.NotNil(t, frontend)
	assert.Contains(t, frontend.DependsOn, "backend")
}

func TestStaticGraphGeneration_ConnectionIDsMatchResourceIDs(t *testing.T) {
	// Verify that connection source/target IDs are remapped from names
	// to full resource IDs matching the extracted resources.
	template := map[string]any{
		"resources": []any{
			map[string]any{
				"type":       "Applications.Core/containers",
				"apiVersion": "2023-10-01-preview",
				"name":       "demo",
				"properties": map[string]any{
					"connections": map[string]any{
						"redis": map[string]any{
							"source": "db",
						},
					},
				},
			},
			map[string]any{
				"type":       "Applications.Datastores/redisCaches",
				"apiVersion": "2023-10-01-preview",
				"name":       "db",
			},
		},
	}

	armTemplate, err := bicep.ParseARMTemplate(template)
	require.NoError(t, err)

	// Extract resources (full IDs)
	extractor := bicep.NewResourceExtractor("default")
	resources, err := extractor.ExtractResources(armTemplate, "app.bicep")
	require.NoError(t, err)
	require.Len(t, resources, 2)

	// Extract connections (name-based IDs)
	connections := bicep.ExtractConnections(armTemplate)
	require.Len(t, connections, 1)
	// Before remapping, connection IDs are just names
	assert.Equal(t, "demo", connections[0].SourceResourceID)
	assert.Equal(t, "db", connections[0].TargetResourceID)

	// Simulate the remapping done in static.go's Generate()
	nameToID := make(map[string]string)
	for _, r := range resources {
		nameToID[r.Name] = r.ID
	}

	srcID := connections[0].SourceResourceID
	if fullID, ok := nameToID[srcID]; ok {
		srcID = fullID
	}
	tgtID := connections[0].TargetResourceID
	if fullID, ok := nameToID[tgtID]; ok {
		tgtID = fullID
	}

	// After remapping, connection IDs must match resource IDs
	assert.Equal(t, resources[0].ID, srcID, "source ID should match resource ID for 'demo'")
	assert.Equal(t, resources[1].ID, tgtID, "target ID should match resource ID for 'db'")

	// Both should be full /planes/... paths
	assert.Contains(t, srcID, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/demo")
	assert.Contains(t, tgtID, "/planes/radius/local/resourceGroups/default/providers/Applications.Datastores/redisCaches/db")
}
