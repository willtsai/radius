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
	"fmt"
	"strings"
	"testing"

	"github.com/radius-project/radius/pkg/cli/output"
	"github.com/radius-project/radius/pkg/corerp/api/v20231001preview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffMarkdownRenderer_RenderPRComment_NoChanges(t *testing.T) {
	t.Parallel()

	renderer := output.NewDiffMarkdownRenderer()
	diff := &v20231001preview.GraphDiff{
		Summary: v20231001preview.DiffSummary{TotalChanges: 0},
	}

	result := renderer.RenderPRComment(diff, nil, nil)

	assert.Contains(t, result, "No changes detected")
}

func TestDiffMarkdownRenderer_RenderPRComment_WithChanges(t *testing.T) {
	t.Parallel()

	renderer := output.NewDiffMarkdownRenderer()

	before := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "id1", Name: "container1", Type: "Applications.Core/containers"},
		},
	}

	after := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "id1", Name: "container1", Type: "Applications.Core/containers"},
			{ID: "id2", Name: "container2", Type: "Applications.Core/containers"},
		},
	}

	diff := &v20231001preview.GraphDiff{
		AddedResources: []v20231001preview.StaticAppGraphResource{
			{ID: "id2", Name: "container2", Type: "Applications.Core/containers"},
		},
		Summary: v20231001preview.DiffSummary{
			ResourcesAdded: 1,
			TotalChanges:   1,
		},
	}

	result := renderer.RenderPRComment(diff, before, after)

	assert.Contains(t, result, "Application Graph Changes")
	assert.Contains(t, result, "➕")
	assert.Contains(t, result, "container2")
}

func TestDiffMarkdownRenderer_RenderSummaryTable(t *testing.T) {
	t.Parallel()

	renderer := output.NewDiffMarkdownRenderer()

	diff := &v20231001preview.GraphDiff{
		AddedResources: []v20231001preview.StaticAppGraphResource{
			{ID: "id1", Name: "added1", Type: "type1"},
		},
		RemovedResources: []v20231001preview.StaticAppGraphResource{
			{ID: "id2", Name: "removed1", Type: "type1"},
		},
		ModifiedResources: []v20231001preview.ResourceDiff{
			{ID: "id3", Name: "modified1", Type: "type1"},
		},
		Summary: v20231001preview.DiffSummary{
			ResourcesAdded:    1,
			ResourcesRemoved:  1,
			ResourcesModified: 1,
			TotalChanges:      3,
		},
	}

	result := renderer.RenderPRComment(diff, nil, nil)

	// Summary badges should show counts
	assert.Contains(t, result, "➕ 1 added")
	assert.Contains(t, result, "➖ 1 removed")
	assert.Contains(t, result, "🔄 1 modified")
}

func TestDiffMarkdownRenderer_RenderStalenessWarning(t *testing.T) {
	t.Parallel()

	renderer := output.NewDiffMarkdownRenderer()
	result := renderer.RenderStalenessWarning("sha256:abc123", "sha256:def456")

	assert.Contains(t, result, "⚠️")
	assert.Contains(t, result, "Stale")
	assert.Contains(t, result, "sha256:abc123")
	assert.Contains(t, result, "sha256:def456")
}

func TestDiffMarkdownRenderer_RenderPropertyChanges(t *testing.T) {
	t.Parallel()

	renderer := output.NewDiffMarkdownRenderer()

	diff := &v20231001preview.GraphDiff{
		ModifiedResources: []v20231001preview.ResourceDiff{
			{
				ID:   "id1",
				Name: "container1",
				Type: "Applications.Core/containers",
				ChangedProperties: []v20231001preview.PropertyChange{
					{Path: "image", OldValue: "nginx:1.19", NewValue: "nginx:1.21"},
					{Path: "replicas", OldValue: 2, NewValue: 3},
				},
			},
		},
		Summary: v20231001preview.DiffSummary{
			ResourcesModified: 1,
			TotalChanges:      1,
		},
	}

	result := renderer.RenderPRComment(diff, nil, nil)

	assert.Contains(t, result, "container1")
	assert.Contains(t, result, "image")
	assert.Contains(t, result, "nginx:1.19")
	assert.Contains(t, result, "nginx:1.21")
}

func TestDiffMarkdownRenderer_RenderMermaidDiff(t *testing.T) {
	t.Parallel()

	renderer := output.NewDiffMarkdownRenderer()
	renderer.IncludeMermaid = true

	before := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "id1", Name: "container1", Type: "Applications.Core/containers"},
		},
	}

	after := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "id1", Name: "container1", Type: "Applications.Core/containers"},
			{ID: "id2", Name: "container2", Type: "Applications.Core/containers"},
		},
	}

	diff := &v20231001preview.GraphDiff{
		AddedResources: []v20231001preview.StaticAppGraphResource{
			{ID: "id2", Name: "container2", Type: "Applications.Core/containers"},
		},
		Summary: v20231001preview.DiffSummary{
			ResourcesAdded: 1,
			TotalChanges:   1,
		},
	}

	result := renderer.RenderPRComment(diff, before, after)

	// Should contain Mermaid code block
	assert.Contains(t, result, "```mermaid")
	assert.Contains(t, result, "```")
}

func TestDiffMarkdownRenderer_EscapesSpecialCharacters(t *testing.T) {
	t.Parallel()

	renderer := output.NewDiffMarkdownRenderer()

	diff := &v20231001preview.GraphDiff{
		AddedResources: []v20231001preview.StaticAppGraphResource{
			{ID: "id1", Name: "my|pipe|name", Type: "type1"},
		},
		Summary: v20231001preview.DiffSummary{
			ResourcesAdded: 1,
			TotalChanges:   1,
		},
	}

	result := renderer.RenderPRComment(diff, nil, nil)

	// Pipes should be escaped in table cells
	assert.Contains(t, result, `my\|pipe\|name`)
}

func TestDiffMarkdownRenderer_CollapsibleSections(t *testing.T) {
	t.Parallel()

	renderer := output.NewDiffMarkdownRenderer()

	before := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "id1", Name: "container1", Type: "Applications.Core/containers"},
		},
	}

	after := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "id1", Name: "container1", Type: "Applications.Core/containers"},
			{ID: "id2", Name: "container2", Type: "Applications.Core/containers"},
		},
	}

	diff := &v20231001preview.GraphDiff{
		AddedResources: []v20231001preview.StaticAppGraphResource{
			{ID: "id2", Name: "container2", Type: "Applications.Core/containers"},
		},
		Summary: v20231001preview.DiffSummary{
			ResourcesAdded: 1,
			TotalChanges:   1,
		},
	}

	result := renderer.RenderPRComment(diff, before, after)

	// Implementation uses HTML details/summary for collapsible sections by default
	assert.Contains(t, result, "<details>")
	assert.Contains(t, result, "</details>")
}

func TestDiffMarkdownRenderer_IdentifiesResourceType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		resourceType string
		expectedIcon string
	}{
		{"Applications.Core/containers", "📦"},
		{"Applications.Core/gateways", "🚪"},
		{"Applications.Datastores/redisCaches", "🗄️"},
		{"Applications.Datastores/sqlDatabases", "🗄️"},
		{"Applications.Core/applications", "📱"},
		{"Unknown/type", "📄"},
	}

	for _, tc := range tests {
		t.Run(tc.resourceType, func(t *testing.T) {
			t.Parallel()

			icon := output.GetResourceTypeIcon(tc.resourceType)
			assert.Equal(t, tc.expectedIcon, icon)
		})
	}
}

func TestDiffMarkdownRenderer_EmptyGraphs(t *testing.T) {
	t.Parallel()

	renderer := output.NewDiffMarkdownRenderer()

	// Empty graphs
	diff := &v20231001preview.GraphDiff{
		Summary: v20231001preview.DiffSummary{TotalChanges: 0},
	}

	result := renderer.RenderPRComment(diff, nil, nil)

	// Should handle nil graphs gracefully
	require.NotEmpty(t, result)
	assert.Contains(t, result, "No changes detected")
}

func TestDiffMarkdownRenderer_LargeChangeset(t *testing.T) {
	t.Parallel()

	renderer := output.NewDiffMarkdownRenderer()

	// Create a large changeset with many resources
	var added []v20231001preview.StaticAppGraphResource
	for i := 0; i < 50; i++ {
		added = append(added, v20231001preview.StaticAppGraphResource{
			ID:   fmt.Sprintf("id%d", i),
			Name: fmt.Sprintf("resource%d", i),
			Type: "type",
		})
	}

	diff := &v20231001preview.GraphDiff{
		AddedResources: added,
		Summary: v20231001preview.DiffSummary{
			ResourcesAdded: 50,
			TotalChanges:   50,
		},
	}

	result := renderer.RenderPRComment(diff, nil, nil)

	// Should truncate or use collapsible for large changesets
	// The result should be reasonable size for PR comments
	lines := strings.Split(result, "\n")
	assert.Less(t, len(lines), 500, "output should be reasonably sized for PR comments")
}
