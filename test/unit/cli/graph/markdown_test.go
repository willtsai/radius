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
)

func TestMarkdownGenerator_GenerateStaticGraphMarkdown_Basic(t *testing.T) {
	gen := output.NewMarkdownGenerator()
	gen.IncludeMermaid = true

	fixedTime := time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC)

	graph := &v20231001preview.StaticAppGraph{
		Metadata: v20231001preview.StaticAppGraphMetadata{
			GeneratedAt:      fixedTime,
			RadiusCliVersion: "1.0.0",
			SourceFiles:      []string{"main.bicep"},
			SourceHash:       "abc123",
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
		},
	}

	result := gen.GenerateStaticGraphMarkdown(graph)

	// Check title
	assert.Contains(t, result, "# Application Graph")

	// Check metadata table
	assert.Contains(t, result, "| Property | Value |")
	assert.Contains(t, result, "Generated At")
	assert.Contains(t, result, "Radius CLI Version")
	assert.Contains(t, result, "1.0.0")
	assert.Contains(t, result, "Source Files")

	// Check resources section
	assert.Contains(t, result, "## Resources")
	assert.Contains(t, result, "| Name | Type | Source | Git Info |")
	assert.Contains(t, result, "api")
	assert.Contains(t, result, "containers")

	// Check Mermaid diagram
	assert.Contains(t, result, "## Graph Visualization")
	assert.Contains(t, result, "```mermaid")
}

func TestMarkdownGenerator_ResourcesTable(t *testing.T) {
	gen := output.NewMarkdownGenerator()
	gen.IncludeMermaid = false

	graph := &v20231001preview.StaticAppGraph{
		Metadata: v20231001preview.StaticAppGraphMetadata{
			GeneratedAt:      time.Now(),
			RadiusCliVersion: "1.0.0",
		},
		Resources: []v20231001preview.StaticAppGraphResource{
			{
				ID:   "/test/api",
				Name: "api",
				Type: "Applications.Core/containers",
				SourceLocation: v20231001preview.SourceLocation{
					File: "app.bicep",
					Line: 15,
				},
			},
			{
				ID:   "/test/cache",
				Name: "cache",
				Type: "Applications.Datastores/redisCaches",
				SourceLocation: v20231001preview.SourceLocation{
					File: "app.bicep",
					Line: 30,
				},
			},
		},
	}

	result := gen.GenerateStaticGraphMarkdown(graph)

	// Verify table headers
	assert.Contains(t, result, "| Name | Type | Source | Git Info |")
	assert.Contains(t, result, "|------|------|--------|----------|")

	// Verify resources are listed
	assert.Contains(t, result, "api")
	assert.Contains(t, result, "cache")
	assert.Contains(t, result, "app.bicep:15")
	assert.Contains(t, result, "app.bicep:30")
}

func TestMarkdownGenerator_ConnectionsTable(t *testing.T) {
	gen := output.NewMarkdownGenerator()
	gen.IncludeMermaid = false

	graph := &v20231001preview.StaticAppGraph{
		Metadata: v20231001preview.StaticAppGraphMetadata{
			GeneratedAt:      time.Now(),
			RadiusCliVersion: "1.0.0",
		},
		Resources: []v20231001preview.StaticAppGraphResource{
			{
				ID:   "/test/frontend",
				Name: "frontend",
				Type: "Applications.Core/containers",
			},
			{
				ID:   "/test/api",
				Name: "api",
				Type: "Applications.Core/containers",
			},
		},
		Connections: []v20231001preview.StaticAppGraphConnection{
			{
				SourceID: "/test/frontend",
				TargetID: "/test/api",
				Type:     v20231001preview.StaticConnectionTypeConnection,
			},
		},
	}

	result := gen.GenerateStaticGraphMarkdown(graph)

	// Verify connections section exists
	assert.Contains(t, result, "## Connections")
	assert.Contains(t, result, "| Source | Target | Type |")
	assert.Contains(t, result, "frontend")
	assert.Contains(t, result, "api")
	assert.Contains(t, result, "connection")
}

func TestMarkdownGenerator_NoConnectionsSection(t *testing.T) {
	gen := output.NewMarkdownGenerator()
	gen.IncludeMermaid = false

	graph := &v20231001preview.StaticAppGraph{
		Metadata: v20231001preview.StaticAppGraphMetadata{
			GeneratedAt:      time.Now(),
			RadiusCliVersion: "1.0.0",
		},
		Resources: []v20231001preview.StaticAppGraphResource{
			{
				ID:   "/test/api",
				Name: "api",
				Type: "Applications.Core/containers",
			},
		},
		Connections: []v20231001preview.StaticAppGraphConnection{},
	}

	result := gen.GenerateStaticGraphMarkdown(graph)

	// Connections section should not be present
	assert.NotContains(t, result, "## Connections")
}

func TestMarkdownGenerator_GitInfoDisplay(t *testing.T) {
	gen := output.NewMarkdownGenerator()
	gen.IncludeMermaid = false
	gen.IncludeGitLinks = true
	gen.GitHubBaseURL = "https://github.com/test/repo"

	graph := &v20231001preview.StaticAppGraph{
		Metadata: v20231001preview.StaticAppGraphMetadata{
			GeneratedAt:      time.Now(),
			RadiusCliVersion: "1.0.0",
			GitCommit:        "abc123def456",
		},
		Resources: []v20231001preview.StaticAppGraphResource{
			{
				ID:   "/test/api",
				Name: "api",
				Type: "Applications.Core/containers",
				GitInfo: &v20231001preview.GitInfo{
					CommitSHA:   "abc123def456",
					CommitShort: "abc123d",
					Author:      "developer",
				},
			},
		},
	}

	result := gen.GenerateStaticGraphMarkdown(graph)

	// Verify git link in metadata
	assert.Contains(t, result, "[abc123d](https://github.com/test/repo/commit/abc123def456)")

	// Verify git info in resources table
	assert.Contains(t, result, "developer")
}

func TestMarkdownGenerator_GitInfoUncommitted(t *testing.T) {
	gen := output.NewMarkdownGenerator()
	gen.IncludeMermaid = false

	graph := &v20231001preview.StaticAppGraph{
		Metadata: v20231001preview.StaticAppGraphMetadata{
			GeneratedAt:      time.Now(),
			RadiusCliVersion: "1.0.0",
		},
		Resources: []v20231001preview.StaticAppGraphResource{
			{
				ID:   "/test/api",
				Name: "api",
				Type: "Applications.Core/containers",
				GitInfo: &v20231001preview.GitInfo{
					Uncommitted: true,
				},
			},
		},
	}

	result := gen.GenerateStaticGraphMarkdown(graph)

	// Verify uncommitted indicator
	assert.Contains(t, result, "⚠️ uncommitted")
}

func TestMarkdownGenerator_DiffMarkdown(t *testing.T) {
	gen := output.NewMarkdownGenerator()
	gen.IncludeMermaid = true

	before := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
			{ID: "/test/removed", Name: "removed", Type: "Applications.Core/containers"},
		},
	}

	after := &v20231001preview.StaticAppGraph{
		Resources: []v20231001preview.StaticAppGraphResource{
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

	result := gen.GenerateDiffMarkdown(diff, before, after)

	// Check title
	assert.Contains(t, result, "# Application Graph Changes")

	// Check summary table
	assert.Contains(t, result, "## Summary")
	assert.Contains(t, result, "| Change Type | Count |")
	assert.Contains(t, result, "➕ Added")
	assert.Contains(t, result, "➖ Removed")
	assert.Contains(t, result, "🔄 Modified")

	// Check added resources section
	assert.Contains(t, result, "## Added Resources")

	// Check removed resources section
	assert.Contains(t, result, "## Removed Resources")
}

func TestMarkdownGenerator_DiffMarkdownModified(t *testing.T) {
	gen := output.NewMarkdownGenerator()
	gen.IncludeMermaid = false

	diff := &v20231001preview.GraphDiff{
		AddedResources:   []v20231001preview.StaticAppGraphResource{},
		RemovedResources: []v20231001preview.StaticAppGraphResource{},
		ModifiedResources: []v20231001preview.ResourceDiff{
			{
				ID:   "/test/api",
				Name: "api",
				Type: "Applications.Core/containers",
				ChangedProperties: []v20231001preview.PropertyChange{
					{
						Path:     "properties.container.image",
						OldValue: "image:v1",
						NewValue: "image:v2",
					},
				},
			},
		},
	}

	result := gen.GenerateDiffMarkdown(diff, nil, nil)

	// Check modified resources section
	assert.Contains(t, result, "## Modified Resources")
	assert.Contains(t, result, "| Property | Before | After |")
	assert.Contains(t, result, "image:v1")
	assert.Contains(t, result, "image:v2")
}

func TestMarkdownGenerator_EscapesTableCells(t *testing.T) {
	gen := output.NewMarkdownGenerator()
	gen.IncludeMermaid = false

	graph := &v20231001preview.StaticAppGraph{
		Metadata: v20231001preview.StaticAppGraphMetadata{
			GeneratedAt:      time.Now(),
			RadiusCliVersion: "1.0.0",
		},
		Resources: []v20231001preview.StaticAppGraphResource{
			{
				ID:   "/test/api",
				Name: "api|with|pipes", // Contains pipe characters
				Type: "Applications.Core/containers",
				SourceLocation: v20231001preview.SourceLocation{
					File: "app.bicep",
					Line: 10,
				},
			},
		},
	}

	result := gen.GenerateStaticGraphMarkdown(graph)

	// Pipes should be escaped
	assert.Contains(t, result, `api\|with\|pipes`)
}

func TestMarkdownGenerator_TruncatesLongHash(t *testing.T) {
	gen := output.NewMarkdownGenerator()
	gen.IncludeMermaid = false

	longHash := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	graph := &v20231001preview.StaticAppGraph{
		Metadata: v20231001preview.StaticAppGraphMetadata{
			GeneratedAt:      time.Now(),
			RadiusCliVersion: "1.0.0",
			SourceHash:       longHash,
		},
		Resources: []v20231001preview.StaticAppGraphResource{},
	}

	result := gen.GenerateStaticGraphMarkdown(graph)

	// Full hash should not appear - it should be truncated
	assert.NotContains(t, result, longHash)
	// Truncated version should appear
	assert.True(t, strings.Contains(result, "...") || strings.Contains(result, "0123456789"))
}

func TestMarkdownGenerator_NoMermaid(t *testing.T) {
	gen := output.NewMarkdownGenerator()
	gen.IncludeMermaid = false

	graph := &v20231001preview.StaticAppGraph{
		Metadata: v20231001preview.StaticAppGraphMetadata{
			GeneratedAt:      time.Now(),
			RadiusCliVersion: "1.0.0",
		},
		Resources: []v20231001preview.StaticAppGraphResource{
			{
				ID:   "/test/api",
				Name: "api",
				Type: "Applications.Core/containers",
			},
		},
	}

	result := gen.GenerateStaticGraphMarkdown(graph)

	// Should not contain Mermaid section
	assert.NotContains(t, result, "## Graph Visualization")
	assert.NotContains(t, result, "```mermaid")
}
