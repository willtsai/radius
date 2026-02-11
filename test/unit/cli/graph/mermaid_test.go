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
	"strings"
	"testing"
	"time"

	"github.com/radius-project/radius/pkg/cli/output"
	"github.com/radius-project/radius/pkg/corerp/api/v20231001preview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMermaidGenerator_GenerateStaticGraph_Basic(t *testing.T) {
	gen := output.NewMermaidGenerator()

	graph := &v20231001preview.StaticAppGraph{
		Metadata: v20231001preview.StaticAppGraphMetadata{
			GeneratedAt: time.Now(),
		},
		Resources: []v20231001preview.StaticAppGraphResource{
			{
				ID:   "/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/api",
				Name: "api",
				Type: "Applications.Core/containers",
			},
			{
				ID:   "/planes/radius/local/resourceGroups/default/providers/Applications.Datastores/redisCaches/cache",
				Name: "cache",
				Type: "Applications.Datastores/redisCaches",
			},
		},
		Connections: []v20231001preview.StaticAppGraphConnection{
			{
				SourceID: "/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/api",
				TargetID: "/planes/radius/local/resourceGroups/default/providers/Applications.Datastores/redisCaches/cache",
				Type:     v20231001preview.StaticConnectionTypeConnection,
			},
		},
	}

	result := gen.GenerateStaticGraph(graph)

	// Verify graph header
	assert.Contains(t, result, "graph TB")

	// Verify nodes are generated
	assert.Contains(t, result, "node0")
	assert.Contains(t, result, "node1")

	// Verify connection edge
	assert.Contains(t, result, "-->")
}

func TestMermaidGenerator_ShapeMapping(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		expectedText string
	}{
		{
			name:         "container gets rectangle",
			resourceType: "Applications.Core/containers",
			expectedText: "[", // Rectangle uses [text]
		},
		{
			name:         "gateway gets diamond",
			resourceType: "Applications.Core/gateways",
			expectedText: "{", // Diamond uses {text}
		},
		{
			name:         "redis gets cylinder",
			resourceType: "Applications.Datastores/redisCaches",
			expectedText: "[(", // Cylinder uses [(text)]
		},
		{
			name:         "environment gets hexagon",
			resourceType: "Applications.Core/environments",
			expectedText: "{{", // Hexagon uses {{text}}
		},
		{
			name:         "rabbitMQ gets parallelogram",
			resourceType: "Applications.Messaging/rabbitMQQueues",
			expectedText: "[/", // Parallelogram uses [/text/]
		},
		{
			name:         "dapr stateStore gets stadium",
			resourceType: "Applications.Dapr/stateStores",
			expectedText: "([", // Stadium uses ([text])
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gen := output.NewMermaidGenerator()

			graph := &v20231001preview.StaticAppGraph{
				Resources: []v20231001preview.StaticAppGraphResource{
					{
						ID:   "/test/resource",
						Name: "testresource",
						Type: tc.resourceType,
					},
				},
			}

			result := gen.GenerateStaticGraph(graph)
			assert.Contains(t, result, tc.expectedText, "Should contain shape marker for %s", tc.resourceType)
		})
	}
}

func TestMermaidGenerator_EdgeStyles(t *testing.T) {
	tests := []struct {
		name           string
		connectionType v20231001preview.StaticConnectionType
		expectedEdge   string
	}{
		{
			name:           "connection uses solid arrow",
			connectionType: v20231001preview.StaticConnectionTypeConnection,
			expectedEdge:   "-->",
		},
		{
			name:           "route uses dotted arrow",
			connectionType: v20231001preview.StaticConnectionTypeRoute,
			expectedEdge:   "-.->",
		},
		{
			name:           "dependsOn uses thick arrow",
			connectionType: v20231001preview.StaticConnectionTypeDependsOn,
			expectedEdge:   "==>",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gen := output.NewMermaidGenerator()

			graph := &v20231001preview.StaticAppGraph{
				Resources: []v20231001preview.StaticAppGraphResource{
					{ID: "/test/source", Name: "source", Type: "Applications.Core/containers"},
					{ID: "/test/target", Name: "target", Type: "Applications.Core/containers"},
				},
				Connections: []v20231001preview.StaticAppGraphConnection{
					{
						SourceID: "/test/source",
						TargetID: "/test/target",
						Type:     tc.connectionType,
					},
				},
			}

			result := gen.GenerateStaticGraph(graph)
			assert.Contains(t, result, tc.expectedEdge, "Should contain edge style for %s", tc.connectionType)
		})
	}
}

func TestMermaidGenerator_SpecialCharactersEscaped(t *testing.T) {
	gen := output.NewMermaidGenerator()

	graph := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{
				ID:   "/test/resource",
				Name: `my"container`, // Contains quote
				Type: "Applications.Core/containers",
			},
		},
	}

	result := gen.GenerateStaticGraph(graph)

	// Quotes should be escaped
	assert.NotContains(t, result, `"container`)
}

func TestMermaidGenerator_DiffDiagram(t *testing.T) {
	gen := output.NewMermaidGenerator()

	before := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "/test/removed", Name: "removed", Type: "Applications.Core/containers"},
			{ID: "/test/unchanged", Name: "unchanged", Type: "Applications.Core/containers"},
		},
	}

	after := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "/test/unchanged", Name: "unchanged", Type: "Applications.Core/containers"},
			{ID: "/test/added", Name: "added", Type: "Applications.Core/containers"},
		},
	}

	diff := &v20231001preview.GraphDiff{
		AddedResources: []v20231001preview.StaticAppGraphResource{
			{ID: "/test/added", Name: "added", Type: "Applications.Core/containers"},
		},
		RemovedResources: []v20231001preview.StaticAppGraphResource{
			{ID: "/test/removed", Name: "removed", Type: "Applications.Core/containers"},
		},
		ModifiedResources: []v20231001preview.ResourceDiff{},
	}

	result := gen.GenerateDiffDiagram(before, after, diff)

	// Verify graph header
	assert.Contains(t, result, "graph TB")

	// Verify style classes are defined
	assert.Contains(t, result, "classDef added")
	assert.Contains(t, result, "classDef removed")
	assert.Contains(t, result, "classDef modified")

	// Verify colors
	assert.Contains(t, result, "#90EE90") // Added - light green
	assert.Contains(t, result, "#FFB6C1") // Removed - light pink
}

func TestMermaidGenerator_EmptyGraph(t *testing.T) {
	gen := output.NewMermaidGenerator()

	graph := &v20231001preview.StaticAppGraph{
		Resources:   []v20231001preview.StaticAppGraphResource{},
		Connections: []v20231001preview.StaticAppGraphConnection{},
	}

	result := gen.GenerateStaticGraph(graph)

	// Should still have valid header
	assert.Contains(t, result, "graph TB")
}

func TestMermaidGenerator_Direction(t *testing.T) {
	tests := []struct {
		direction string
		expected  string
	}{
		{"TB", "graph TB"},
		{"LR", "graph LR"},
		{"BT", "graph BT"},
		{"RL", "graph RL"},
	}

	for _, tc := range tests {
		t.Run(tc.direction, func(t *testing.T) {
			gen := output.NewMermaidGenerator()
			gen.Direction = tc.direction

			graph := &v20231001preview.StaticAppGraph{
				Resources: []v20231001preview.StaticAppGraphResource{
					{ID: "/test/a", Name: "a", Type: "Applications.Core/containers"},
				},
			}

			result := gen.GenerateStaticGraph(graph)
			assert.Contains(t, result, tc.expected)
		})
	}
}

func TestMermaidGenerator_ResourceTypeAnnotations(t *testing.T) {
	gen := output.NewMermaidGenerator()
	gen.IncludeResourceTypes = true

	graph := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{
				ID:   "/test/api",
				Name: "api",
				Type: "Applications.Core/containers",
			},
		},
	}

	result := gen.GenerateStaticGraph(graph)

	// Should include type annotation
	assert.Contains(t, result, "containers") // Short type name
}

func TestWrapInCodeBlock(t *testing.T) {
	mermaid := "graph TB\n    A --> B"
	result := output.WrapInCodeBlock(mermaid)

	assert.True(t, strings.HasPrefix(result, "```mermaid"))
	assert.True(t, strings.HasSuffix(result, "```"))
	assert.Contains(t, result, "graph TB")
}

func TestValidateMermaidSyntax(t *testing.T) {
	tests := []struct {
		name    string
		mermaid string
		wantErr bool
	}{
		{
			name:    "valid graph TB",
			mermaid: "graph TB\n    A --> B",
			wantErr: false,
		},
		{
			name:    "valid graph LR",
			mermaid: "graph LR\n    A --> B",
			wantErr: false,
		},
		{
			name:    "valid flowchart",
			mermaid: "flowchart TB\n    A --> B",
			wantErr: false,
		},
		{
			name:    "invalid - missing direction",
			mermaid: "graph\n    A --> B",
			wantErr: true,
		},
		{
			name:    "invalid - empty",
			mermaid: "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := output.ValidateMermaidSyntax(tc.mermaid)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
