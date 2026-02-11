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

package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/radius-project/radius/pkg/corerp/api/v20231001preview"
)

// MarkdownGenerator generates Markdown documentation from application graphs.
type MarkdownGenerator struct {
	// IncludeMermaid enables/disables Mermaid diagram generation
	IncludeMermaid bool

	// IncludeGitLinks enables linking commit SHAs to GitHub
	IncludeGitLinks bool

	// GitHubBaseURL is the base URL for GitHub links (e.g., "https://github.com/org/repo")
	GitHubBaseURL string

	// mermaid generator for diagram creation
	mermaid *MermaidGenerator
}

// NewMarkdownGenerator creates a new MarkdownGenerator with default settings.
func NewMarkdownGenerator() *MarkdownGenerator {
	return &MarkdownGenerator{
		IncludeMermaid:  true,
		IncludeGitLinks: true,
		mermaid:         NewMermaidGenerator(),
	}
}

// GenerateStaticGraphMarkdown generates Markdown documentation for a static app graph.
func (g *MarkdownGenerator) GenerateStaticGraphMarkdown(graph *v20231001preview.StaticAppGraph) string {
	var sb strings.Builder

	// Title and metadata
	sb.WriteString("# Application Graph\n\n")
	sb.WriteString(g.generateMetadataSection(graph.Metadata))

	// Diagram
	if g.IncludeMermaid {
		sb.WriteString("## Graph Visualization\n\n")
		mermaid := g.mermaid.GenerateStaticGraph(graph)
		sb.WriteString(WrapInCodeBlock(mermaid))
		sb.WriteString("\n\n")
	}

	// Resources table
	sb.WriteString("## Resources\n\n")
	sb.WriteString(g.generateResourcesTable(graph.Resources))

	// Connections table
	if len(graph.Connections) > 0 {
		sb.WriteString("\n## Connections\n\n")
		sb.WriteString(g.generateConnectionsTable(graph.Connections, graph.Resources))
	}

	return sb.String()
}

// generateMetadataSection creates the metadata section of the Markdown.
func (g *MarkdownGenerator) generateMetadataSection(metadata v20231001preview.StaticAppGraphMetadata) string {
	var sb strings.Builder

	sb.WriteString("| Property | Value |\n")
	sb.WriteString("|----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Generated At | %s |\n", metadata.GeneratedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("| Radius CLI Version | %s |\n", metadata.RadiusCliVersion))
	sb.WriteString(fmt.Sprintf("| Source Files | %s |\n", strings.Join(metadata.SourceFiles, ", ")))
	sb.WriteString(fmt.Sprintf("| Source Hash | `%s` |\n", truncateHash(metadata.SourceHash)))

	if metadata.GitCommit != "" {
		commitDisplay := metadata.GitCommit[:7]
		if g.IncludeGitLinks && g.GitHubBaseURL != "" {
			sb.WriteString(fmt.Sprintf("| Git Commit | [%s](%s/commit/%s) |\n",
				commitDisplay, g.GitHubBaseURL, metadata.GitCommit))
		} else {
			sb.WriteString(fmt.Sprintf("| Git Commit | `%s` |\n", commitDisplay))
		}
	}

	sb.WriteString("\n")
	return sb.String()
}

// generateResourcesTable creates the resources table.
func (g *MarkdownGenerator) generateResourcesTable(resources []v20231001preview.StaticAppGraphResource) string {
	var sb strings.Builder

	// Table header
	sb.WriteString("| Name | Type | Source | Git Info |\n")
	sb.WriteString("|------|------|--------|----------|\n")

	for _, r := range resources {
		name := escapeMarkdownTableCell(r.Name)
		resourceType := escapeMarkdownTableCell(getShortResourceType(r.Type))
		source := fmt.Sprintf("%s:%d", r.SourceLocation.File, r.SourceLocation.Line)
		gitInfo := g.formatGitInfo(r.GitInfo)

		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", name, resourceType, source, gitInfo))
	}

	return sb.String()
}

// generateConnectionsTable creates the connections table.
func (g *MarkdownGenerator) generateConnectionsTable(connections []v20231001preview.StaticAppGraphConnection, resources []v20231001preview.StaticAppGraphResource) string {
	var sb strings.Builder

	// Create ID to name map
	idToName := make(map[string]string)
	for _, r := range resources {
		idToName[r.ID] = r.Name
	}

	sb.WriteString("| Source | Target | Type |\n")
	sb.WriteString("|--------|--------|------|\n")

	for _, c := range connections {
		sourceName := idToName[c.SourceID]
		if sourceName == "" {
			sourceName = extractResourceName(c.SourceID)
		}
		targetName := idToName[c.TargetID]
		if targetName == "" {
			targetName = extractResourceName(c.TargetID)
		}

		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
			escapeMarkdownTableCell(sourceName),
			escapeMarkdownTableCell(targetName),
			string(c.Type)))
	}

	return sb.String()
}

// formatGitInfo formats git information for display.
func (g *MarkdownGenerator) formatGitInfo(gitInfo *v20231001preview.GitInfo) string {
	if gitInfo == nil {
		return "N/A"
	}

	if gitInfo.Uncommitted {
		return "⚠️ uncommitted"
	}

	var parts []string

	// Commit SHA (linked if possible)
	if gitInfo.CommitShort != "" {
		if g.IncludeGitLinks && g.GitHubBaseURL != "" {
			parts = append(parts, fmt.Sprintf("[%s](%s/commit/%s)",
				gitInfo.CommitShort, g.GitHubBaseURL, gitInfo.CommitSHA))
		} else {
			parts = append(parts, fmt.Sprintf("`%s`", gitInfo.CommitShort))
		}
	}

	// Author
	if gitInfo.Author != "" {
		parts = append(parts, gitInfo.Author)
	}

	return strings.Join(parts, " by ")
}

// GenerateDiffMarkdown generates Markdown documentation for a graph diff.
func (g *MarkdownGenerator) GenerateDiffMarkdown(diff *v20231001preview.GraphDiff, before, after *v20231001preview.StaticAppGraph) string {
	var sb strings.Builder

	sb.WriteString("# Application Graph Changes\n\n")

	// Summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString("| Change Type | Count |\n")
	sb.WriteString("|-------------|-------|\n")
	sb.WriteString(fmt.Sprintf("| ➕ Added | %d |\n", len(diff.AddedResources)))
	sb.WriteString(fmt.Sprintf("| ➖ Removed | %d |\n", len(diff.RemovedResources)))
	sb.WriteString(fmt.Sprintf("| 🔄 Modified | %d |\n", len(diff.ModifiedResources)))
	sb.WriteString("\n")

	// Diff diagram
	if g.IncludeMermaid && (len(diff.AddedResources) > 0 || len(diff.RemovedResources) > 0 || len(diff.ModifiedResources) > 0) {
		sb.WriteString("## Visual Diff\n\n")
		sb.WriteString("🟢 Added | 🔴 Removed | 🟡 Modified\n\n")
		mermaid := g.mermaid.GenerateDiffDiagram(before, after, diff)
		sb.WriteString(WrapInCodeBlock(mermaid))
		sb.WriteString("\n\n")
	}

	// Added resources
	if len(diff.AddedResources) > 0 {
		sb.WriteString("## Added Resources\n\n")
		sb.WriteString(g.generateAddedResourceTable(diff.AddedResources))
	}

	// Removed resources
	if len(diff.RemovedResources) > 0 {
		sb.WriteString("## Removed Resources\n\n")
		sb.WriteString(g.generateRemovedResourceTable(diff.RemovedResources))
	}

	// Modified resources
	if len(diff.ModifiedResources) > 0 {
		sb.WriteString("## Modified Resources\n\n")
		sb.WriteString(g.generateModifiedResourcesSection(diff.ModifiedResources))
	}

	return sb.String()
}

// generateAddedResourceTable creates a table for added resources.
func (g *MarkdownGenerator) generateAddedResourceTable(resources []v20231001preview.StaticAppGraphResource) string {
	var sb strings.Builder

	sb.WriteString("| Name | Type |\n")
	sb.WriteString("|------|------|\n")

	for _, r := range resources {
		name := r.Name
		if name == "" {
			name = extractResourceName(r.ID)
		}
		sb.WriteString(fmt.Sprintf("| %s | %s |\n", name, r.Type))
	}

	sb.WriteString("\n")
	return sb.String()
}

// generateRemovedResourceTable creates a table for removed resources.
func (g *MarkdownGenerator) generateRemovedResourceTable(resources []v20231001preview.StaticAppGraphResource) string {
	var sb strings.Builder

	sb.WriteString("| Name | Type |\n")
	sb.WriteString("|------|------|\n")

	for _, r := range resources {
		name := r.Name
		if name == "" {
			name = extractResourceName(r.ID)
		}
		sb.WriteString(fmt.Sprintf("| ~~%s~~ | %s |\n", name, r.Type))
	}

	sb.WriteString("\n")
	return sb.String()
}

// generateModifiedResourcesSection creates details for modified resources.
func (g *MarkdownGenerator) generateModifiedResourcesSection(diffs []v20231001preview.ResourceDiff) string {
	var sb strings.Builder

	for _, d := range diffs {
		name := d.Name
		if name == "" {
			name = extractResourceName(d.ID)
		}
		sb.WriteString(fmt.Sprintf("### %s\n\n", name))

		if len(d.ChangedProperties) > 0 {
			sb.WriteString("| Property | Before | After |\n")
			sb.WriteString("|----------|--------|-------|\n")

			for _, change := range d.ChangedProperties {
				before := formatPropertyValue(change.OldValue)
				after := formatPropertyValue(change.NewValue)
				sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
					escapeMarkdownTableCell(change.Path),
					escapeMarkdownTableCell(before),
					escapeMarkdownTableCell(after)))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// Helper functions

// escapeMarkdownTableCell escapes special characters for Markdown table cells.
func escapeMarkdownTableCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

// truncateHash truncates a hash for display.
func truncateHash(hash string) string {
	if len(hash) > 20 {
		return hash[:20] + "..."
	}
	return hash
}

// extractResourceName extracts the resource name from a resource ID.
func extractResourceName(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return resourceID
}

// formatPropertyValue formats a property value for display in tables.
func formatPropertyValue(v any) string {
	if v == nil {
		return "(none)"
	}

	s := fmt.Sprintf("%v", v)
	if len(s) > 50 {
		return s[:47] + "..."
	}
	return s
}

// GetResourceTypeIcon returns an emoji icon for the given resource type.
func GetResourceTypeIcon(resourceType string) string {
	// Normalize the type for comparison
	lowered := strings.ToLower(resourceType)

	// Order matters: check more specific patterns before generic ones
	switch {
	case strings.Contains(lowered, "containers"):
		return "📦"
	case strings.Contains(lowered, "gateways"):
		return "🚪"
	case strings.Contains(lowered, "datastores") || strings.Contains(lowered, "databases") || strings.Contains(lowered, "caches"):
		return "🗄️"
	case strings.Contains(lowered, "secretstores"):
		return "🔐"
	case strings.Contains(lowered, "extenders"):
		return "🔌"
	case strings.Contains(lowered, "/applications"):
		return "📱"
	default:
		return "📄"
	}
}
