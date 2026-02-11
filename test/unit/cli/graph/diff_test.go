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
	"github.com/radius-project/radius/pkg/corerp/api/v20231001preview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeDiff_NoChanges(t *testing.T) {
	t.Parallel()

	before := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/applications/app1", Name: "app1", Type: "Applications.Core/applications"},
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container1", Name: "container1", Type: "Applications.Core/containers"},
		},
		Connections: []v20231001preview.StaticAppGraphConnection{
			{SourceID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container1", TargetID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/applications/app1", Type: v20231001preview.StaticConnectionTypeDependsOn},
		},
	}

	after := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/applications/app1", Name: "app1", Type: "Applications.Core/applications"},
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container1", Name: "container1", Type: "Applications.Core/containers"},
		},
		Connections: []v20231001preview.StaticAppGraphConnection{
			{SourceID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container1", TargetID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/applications/app1", Type: v20231001preview.StaticConnectionTypeDependsOn},
		},
	}

	diff := graph.ComputeDiff(before, after)

	assert.Empty(t, diff.AddedResources)
	assert.Empty(t, diff.RemovedResources)
	assert.Empty(t, diff.ModifiedResources)
	assert.Empty(t, diff.AddedConnections)
	assert.Empty(t, diff.RemovedConnections)
	assert.True(t, diff.IsEmpty())
}

func TestComputeDiff_AddedResource(t *testing.T) {
	t.Parallel()

	before := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/applications/app1", Name: "app1", Type: "Applications.Core/applications"},
		},
	}

	after := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/applications/app1", Name: "app1", Type: "Applications.Core/applications"},
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container1", Name: "container1", Type: "Applications.Core/containers"},
		},
	}

	diff := graph.ComputeDiff(before, after)

	require.Len(t, diff.AddedResources, 1)
	assert.Equal(t, "container1", diff.AddedResources[0].Name)
	assert.Equal(t, "Applications.Core/containers", diff.AddedResources[0].Type)
	assert.Empty(t, diff.RemovedResources)
	assert.Empty(t, diff.ModifiedResources)
	assert.False(t, diff.IsEmpty())
	assert.Equal(t, 1, diff.Summary.ResourcesAdded)
}

func TestComputeDiff_RemovedResource(t *testing.T) {
	t.Parallel()

	before := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/applications/app1", Name: "app1", Type: "Applications.Core/applications"},
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container1", Name: "container1", Type: "Applications.Core/containers"},
		},
	}

	after := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/applications/app1", Name: "app1", Type: "Applications.Core/applications"},
		},
	}

	diff := graph.ComputeDiff(before, after)

	assert.Empty(t, diff.AddedResources)
	require.Len(t, diff.RemovedResources, 1)
	assert.Equal(t, "container1", diff.RemovedResources[0].Name)
	assert.Empty(t, diff.ModifiedResources)
	assert.False(t, diff.IsEmpty())
	assert.Equal(t, 1, diff.Summary.ResourcesRemoved)
}

func TestComputeDiff_ModifiedResource(t *testing.T) {
	t.Parallel()

	before := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{
				ID:   "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container1",
				Name: "container1",
				Type: "Applications.Core/containers",
				Properties: map[string]any{
					"image":    "nginx:1.19",
					"replicas": 2,
				},
			},
		},
	}

	after := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{
				ID:   "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container1",
				Name: "container1",
				Type: "Applications.Core/containers",
				Properties: map[string]any{
					"image":    "nginx:1.21",
					"replicas": 3,
				},
			},
		},
	}

	diff := graph.ComputeDiff(before, after)

	assert.Empty(t, diff.AddedResources)
	assert.Empty(t, diff.RemovedResources)
	require.Len(t, diff.ModifiedResources, 1)
	assert.Equal(t, "container1", diff.ModifiedResources[0].Name)
	assert.False(t, diff.IsEmpty())
	assert.Equal(t, 1, diff.Summary.ResourcesModified)

	// Check property changes
	require.Len(t, diff.ModifiedResources[0].ChangedProperties, 2)
	changeMap := make(map[string]v20231001preview.PropertyChange)
	for _, c := range diff.ModifiedResources[0].ChangedProperties {
		changeMap[c.Path] = c
	}
	assert.Equal(t, "nginx:1.19", changeMap["image"].OldValue)
	assert.Equal(t, "nginx:1.21", changeMap["image"].NewValue)
}

func TestComputeDiff_AddedConnection(t *testing.T) {
	t.Parallel()

	before := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container1", Name: "container1", Type: "Applications.Core/containers"},
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Datastores/redisCaches/redis1", Name: "redis1", Type: "Applications.Datastores/redisCaches"},
		},
		Connections: []v20231001preview.StaticAppGraphConnection{},
	}

	after := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container1", Name: "container1", Type: "Applications.Core/containers"},
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Datastores/redisCaches/redis1", Name: "redis1", Type: "Applications.Datastores/redisCaches"},
		},
		Connections: []v20231001preview.StaticAppGraphConnection{
			{
				SourceID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container1",
				TargetID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Datastores/redisCaches/redis1",
				Type:     v20231001preview.StaticConnectionTypeConnection,
			},
		},
	}

	diff := graph.ComputeDiff(before, after)

	require.Len(t, diff.AddedConnections, 1)
	assert.Equal(t, v20231001preview.StaticConnectionTypeConnection, diff.AddedConnections[0].Type)
	assert.Empty(t, diff.RemovedConnections)
}

func TestComputeDiff_RemovedConnection(t *testing.T) {
	t.Parallel()

	before := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container1", Name: "container1", Type: "Applications.Core/containers"},
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Datastores/redisCaches/redis1", Name: "redis1", Type: "Applications.Datastores/redisCaches"},
		},
		Connections: []v20231001preview.StaticAppGraphConnection{
			{
				SourceID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container1",
				TargetID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Datastores/redisCaches/redis1",
				Type:     v20231001preview.StaticConnectionTypeConnection,
			},
		},
	}

	after := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container1", Name: "container1", Type: "Applications.Core/containers"},
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Datastores/redisCaches/redis1", Name: "redis1", Type: "Applications.Datastores/redisCaches"},
		},
		Connections: []v20231001preview.StaticAppGraphConnection{},
	}

	diff := graph.ComputeDiff(before, after)

	assert.Empty(t, diff.AddedConnections)
	require.Len(t, diff.RemovedConnections, 1)
}

func TestComputeDiff_ComplexScenario(t *testing.T) {
	t.Parallel()

	before := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/applications/app1", Name: "app1", Type: "Applications.Core/applications"},
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container1", Name: "container1", Type: "Applications.Core/containers", Properties: map[string]any{"image": "nginx:1.19"}},
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container2", Name: "container2", Type: "Applications.Core/containers"},
		},
		Connections: []v20231001preview.StaticAppGraphConnection{
			{SourceID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container1", TargetID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/applications/app1", Type: v20231001preview.StaticConnectionTypeDependsOn},
		},
	}

	after := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/applications/app1", Name: "app1", Type: "Applications.Core/applications"},
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container1", Name: "container1", Type: "Applications.Core/containers", Properties: map[string]any{"image": "nginx:1.21"}},
			{ID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container3", Name: "container3", Type: "Applications.Core/containers"},
		},
		Connections: []v20231001preview.StaticAppGraphConnection{
			{SourceID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container1", TargetID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/applications/app1", Type: v20231001preview.StaticConnectionTypeDependsOn},
			{SourceID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/container3", TargetID: "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/applications/app1", Type: v20231001preview.StaticConnectionTypeDependsOn},
		},
	}

	diff := graph.ComputeDiff(before, after)

	// container2 removed, container3 added, container1 modified
	assert.Equal(t, 1, len(diff.AddedResources))
	assert.Equal(t, "container3", diff.AddedResources[0].Name)

	assert.Equal(t, 1, len(diff.RemovedResources))
	assert.Equal(t, "container2", diff.RemovedResources[0].Name)

	assert.Equal(t, 1, len(diff.ModifiedResources))
	assert.Equal(t, "container1", diff.ModifiedResources[0].Name)

	// One new connection for container3
	assert.Equal(t, 1, len(diff.AddedConnections))

	// Summary
	assert.False(t, diff.IsEmpty())
	assert.Equal(t, 1, diff.Summary.ResourcesAdded)
	assert.Equal(t, 1, diff.Summary.ResourcesRemoved)
	assert.Equal(t, 1, diff.Summary.ResourcesModified)
}

func TestComputeDiff_NilGraphs(t *testing.T) {
	t.Parallel()

	// Both nil
	diff := graph.ComputeDiff(nil, nil)
	assert.True(t, diff.IsEmpty())

	// Before nil
	after := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "id1", Name: "resource1", Type: "type1"},
		},
	}
	diff = graph.ComputeDiff(nil, after)
	assert.Equal(t, 1, len(diff.AddedResources))

	// After nil
	before := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "id1", Name: "resource1", Type: "type1"},
		},
	}
	diff = graph.ComputeDiff(before, nil)
	assert.Equal(t, 1, len(diff.RemovedResources))
}

func TestComputeDiff_ConnectionKey(t *testing.T) {
	t.Parallel()

	// Test that connections are compared by source+target+type tuple
	before := &v20231001preview.StaticAppGraph{
		Connections: []v20231001preview.StaticAppGraphConnection{
			{SourceID: "a", TargetID: "b", Type: v20231001preview.StaticConnectionTypeConnection},
		},
	}

	after := &v20231001preview.StaticAppGraph{
		Connections: []v20231001preview.StaticAppGraphConnection{
			{SourceID: "a", TargetID: "b", Type: v20231001preview.StaticConnectionTypeDependsOn}, // Same source/target, different type
		},
	}

	diff := graph.ComputeDiff(before, after)

	// Should detect as different connections
	assert.Equal(t, 1, len(diff.AddedConnections))
	assert.Equal(t, 1, len(diff.RemovedConnections))
}
