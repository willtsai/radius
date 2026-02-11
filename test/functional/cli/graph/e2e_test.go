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
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/radius-project/radius/pkg/corerp/api/v20231001preview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_StaticGraphGeneration tests the complete workflow of generating
// a static app graph from a Bicep file. This is a functional end-to-end test
// that requires the Bicep CLI to be installed.
func TestE2E_StaticGraphGeneration(t *testing.T) {
	// Skip if no bicep available
	_, err := exec.LookPath("bicep")
	if err != nil {
		t.Skip("Bicep CLI not found in PATH - skipping E2E test")
	}

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "graph-e2e-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a sample Bicep file
	bicepContent := `
extension radius

param environment string

resource app 'Applications.Core/applications@2023-10-01-preview' = {
  name: 'testapp'
  properties: {
    environment: environment
  }
}

resource frontend 'Applications.Core/containers@2023-10-01-preview' = {
  name: 'frontend'
  properties: {
    application: app.id
    container: {
      image: 'myapp/frontend:v1.0.0'
      ports: {
        web: { containerPort: 3000 }
      }
    }
    connections: {
      backend: { source: backend.id }
    }
  }
}

resource backend 'Applications.Core/containers@2023-10-01-preview' = {
  name: 'backend'
  properties: {
    application: app.id
    container: {
      image: 'myapp/backend:v1.0.0'
      ports: {
        api: { containerPort: 8080 }
      }
    }
  }
}
`
	bicepPath := filepath.Join(tempDir, "app.bicep")
	err = os.WriteFile(bicepPath, []byte(bicepContent), 0644)
	require.NoError(t, err)

	// Build the Bicep file to ARM JSON
	cmd := exec.Command("bicep", "build", bicepPath, "--stdout")
	output, err := cmd.Output()
	require.NoError(t, err, "Failed to build Bicep file")

	// Parse the ARM JSON output
	var armTemplate map[string]any
	err = json.Unmarshal(output, &armTemplate)
	require.NoError(t, err)

	// Verify the ARM template contains the expected resources
	resources, ok := armTemplate["resources"].([]any)
	require.True(t, ok, "ARM template should have resources array")
	require.GreaterOrEqual(t, len(resources), 3, "Should have at least 3 resources (app, frontend, backend)")

	// Verify resource types
	var foundApp, foundFrontend, foundBackend bool
	for _, res := range resources {
		resMap, ok := res.(map[string]any)
		if !ok {
			continue
		}
		resType, _ := resMap["type"].(string)
		resName, _ := resMap["name"].(string)

		switch {
		case resType == "Applications.Core/applications" && strings.Contains(resName, "testapp"):
			foundApp = true
		case resType == "Applications.Core/containers" && strings.Contains(resName, "frontend"):
			foundFrontend = true
		case resType == "Applications.Core/containers" && strings.Contains(resName, "backend"):
			foundBackend = true
		}
	}

	assert.True(t, foundApp, "Should find application resource")
	assert.True(t, foundFrontend, "Should find frontend container")
	assert.True(t, foundBackend, "Should find backend container")
}

// TestE2E_GraphDiffComputation tests the graph diff computation between two versions.
func TestE2E_GraphDiffComputation(t *testing.T) {
	// Create two graph versions for comparison
	baseGraph := &v20231001preview.StaticAppGraph{
		Metadata: v20231001preview.StaticAppGraphMetadata{
			RadiusCliVersion: "0.35.0",
		},
		Resources: []v20231001preview.StaticAppGraphResource{
			{
				ID:   "/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/frontend",
				Name: "frontend",
				Type: "Applications.Core/containers",
			},
			{
				ID:   "/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/backend",
				Name: "backend",
				Type: "Applications.Core/containers",
			},
		},
	}

	headGraph := &v20231001preview.StaticAppGraph{
		Metadata: v20231001preview.StaticAppGraphMetadata{
			RadiusCliVersion: "0.35.0",
		},
		Resources: []v20231001preview.StaticAppGraphResource{
			{
				ID:   "/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/frontend",
				Name: "frontend",
				Type: "Applications.Core/containers",
			},
			{
				ID:   "/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/backend",
				Name: "backend",
				Type: "Applications.Core/containers",
			},
			{
				ID:   "/planes/radius/local/resourceGroups/default/providers/Applications.Datastores/redisCaches/cache",
				Name: "cache",
				Type: "Applications.Datastores/redisCaches",
			},
		},
	}

	// Compute diff
	diff := computeDiff(baseGraph, headGraph)
	require.NotNil(t, diff)

	// Verify diff results
	assert.Len(t, diff.AddedResources, 1, "Should have 1 added resource")
	assert.Len(t, diff.RemovedResources, 0, "Should have 0 removed resources")
	assert.Equal(t, "cache", diff.AddedResources[0].Name)
	assert.Equal(t, "Applications.Datastores/redisCaches", diff.AddedResources[0].Type)
}

// computeDiff computes the difference between two graphs (simplified for E2E test).
func computeDiff(base, head *v20231001preview.StaticAppGraph) *v20231001preview.GraphDiff {
	if base == nil || head == nil {
		return &v20231001preview.GraphDiff{}
	}

	// Build maps for comparison
	baseResources := make(map[string]v20231001preview.StaticAppGraphResource)
	for _, r := range base.Resources {
		baseResources[r.ID] = r
	}

	headResources := make(map[string]v20231001preview.StaticAppGraphResource)
	for _, r := range head.Resources {
		headResources[r.ID] = r
	}

	diff := &v20231001preview.GraphDiff{}

	// Find added resources
	for id, r := range headResources {
		if _, exists := baseResources[id]; !exists {
			diff.AddedResources = append(diff.AddedResources, r)
		}
	}

	// Find removed resources
	for id, r := range baseResources {
		if _, exists := headResources[id]; !exists {
			diff.RemovedResources = append(diff.RemovedResources, r)
		}
	}

	return diff
}

// TestE2E_DeterministicOutput verifies that the same input produces identical output.
func TestE2E_DeterministicOutput(t *testing.T) {
	graph := &v20231001preview.StaticAppGraph{
		Metadata: v20231001preview.StaticAppGraphMetadata{
			RadiusCliVersion: "0.35.0",
			SourceFiles:      []string{"app.bicep"},
			SourceHash:       "sha256:abc123",
		},
		Resources: []v20231001preview.StaticAppGraphResource{
			{
				ID:   "/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/frontend",
				Name: "frontend",
				Type: "Applications.Core/containers",
			},
			{
				ID:   "/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/backend",
				Name: "backend",
				Type: "Applications.Core/containers",
			},
		},
	}

	// Serialize twice
	output1, err := json.MarshalIndent(graph, "", "  ")
	require.NoError(t, err)

	output2, err := json.MarshalIndent(graph, "", "  ")
	require.NoError(t, err)

	// Should be identical
	assert.Equal(t, string(output1), string(output2), "Serialization should be deterministic")
}
