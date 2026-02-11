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
	"regexp"
	"strings"

	"github.com/radius-project/radius/pkg/corerp/api/v20231001preview"
)

// MermaidGenerator generates Mermaid diagram markup from application graphs.
type MermaidGenerator struct {
	// Direction is the graph direction: TB (top-bottom), LR (left-right), BT, RL
	Direction string

	// Theme is the Mermaid theme (default, dark, forest, neutral)
	Theme string

	// IncludeResourceTypes adds resource type annotations
	IncludeResourceTypes bool
}

// NewMermaidGenerator creates a new MermaidGenerator with default settings.
func NewMermaidGenerator() *MermaidGenerator {
	return &MermaidGenerator{
		Direction:            "TB",
		Theme:                "default",
		IncludeResourceTypes: true,
	}
}

// GenerateStaticGraph generates Mermaid markup from a StaticAppGraph.
func (g *MermaidGenerator) GenerateStaticGraph(graph *v20231001preview.StaticAppGraph) string {
	var sb strings.Builder

	// Graph header
	sb.WriteString(fmt.Sprintf("graph %s\n", g.Direction))

	// Add theme if not default
	if g.Theme != "" && g.Theme != "default" {
		sb.WriteString(fmt.Sprintf("    %%%%{init: {'theme': '%s'}}%%%%\n", g.Theme))
	}

	// Generate nodes
	nodeIDs := make(map[string]string) // resource ID -> mermaid node ID
	for i, resource := range graph.Resources {
		nodeID := fmt.Sprintf("node%d", i)
		nodeIDs[resource.ID] = nodeID

		label := g.formatNodeLabel(resource)
		shape := g.getNodeShape(resource.Type)
		sb.WriteString(fmt.Sprintf("    %s%s\n", nodeID, formatMermaidNode(label, shape)))
	}

	// Generate edges
	for _, conn := range graph.Connections {
		sourceNode := nodeIDs[conn.SourceID]
		targetNode := nodeIDs[conn.TargetID]
		if sourceNode != "" && targetNode != "" {
			edgeStyle := g.getEdgeStyle(string(conn.Type))
			sb.WriteString(fmt.Sprintf("    %s %s %s\n", sourceNode, edgeStyle, targetNode))
		}
	}

	return sb.String()
}

// formatNodeLabel creates the label for a node.
func (g *MermaidGenerator) formatNodeLabel(resource v20231001preview.StaticAppGraphResource) string {
	label := sanitizeMermaidText(resource.Name)

	if g.IncludeResourceTypes {
		shortType := getShortResourceType(resource.Type)
		label = fmt.Sprintf("%s<br/><small>%s</small>", label, shortType)
	}

	return label
}

// getNodeShape returns the Mermaid shape based on resource type.
func (g *MermaidGenerator) getNodeShape(resourceType string) string {
	typeName := getResourceTypeName(resourceType)

	switch typeName {
	case "gateways":
		return "diamond"
	case "environments":
		return "hexagon"
	case "mongoDatabases", "redisCaches", "sqlDatabases":
		return "cylinder"
	case "rabbitMQQueues":
		return "parallelogram"
	case "pubSubBrokers", "secretStores", "stateStores", "configurationStores":
		return "stadium"
	case "containers":
		return "rectangle"
	case "applications":
		return "round"
	default:
		return "rectangle"
	}
}

// getEdgeStyle returns the Mermaid edge style based on connection type.
func (g *MermaidGenerator) getEdgeStyle(connType string) string {
	switch connType {
	case "connection":
		return "-->" // Solid arrow
	case "route":
		return "-.->." // Dotted arrow
	case "dependsOn":
		return "==>" // Thick arrow
	default:
		return "-->"
	}
}

// formatMermaidNode formats a node with the given label and shape.
func formatMermaidNode(label, shape string) string {
	switch shape {
	case "diamond":
		return fmt.Sprintf("{%s}", label)
	case "hexagon":
		return fmt.Sprintf("{{%s}}", label)
	case "cylinder":
		return fmt.Sprintf("[(%s)]", label)
	case "parallelogram":
		return fmt.Sprintf("[/%s/]", label)
	case "stadium":
		return fmt.Sprintf("([%s])", label)
	case "round":
		return fmt.Sprintf("(%s)", label)
	default:
		return fmt.Sprintf("[%s]", label)
	}
}

// sanitizeMermaidText escapes special characters for Mermaid.
func sanitizeMermaidText(text string) string {
	// Replace quotes with escaped quotes
	text = strings.ReplaceAll(text, `"`, `#quot;`)
	// Replace angle brackets
	text = strings.ReplaceAll(text, "<", "#lt;")
	text = strings.ReplaceAll(text, ">", "#gt;")
	return text
}

// getShortResourceType extracts a short type name from a full resource type.
func getShortResourceType(resourceType string) string {
	parts := strings.Split(resourceType, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return resourceType
}

// getResourceTypeName extracts just the type name from a resource type.
func getResourceTypeName(resourceType string) string {
	parts := strings.Split(resourceType, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return resourceType
}

// GenerateDiffDiagram generates a Mermaid diagram showing graph differences.
// Added resources are shown in green, removed in red, modified in yellow.
func (g *MermaidGenerator) GenerateDiffDiagram(before, after *v20231001preview.StaticAppGraph, diff *v20231001preview.GraphDiff) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("graph %s\n", g.Direction))

	// Create maps for quick lookup
	beforeResources := make(map[string]v20231001preview.StaticAppGraphResource)
	afterResources := make(map[string]v20231001preview.StaticAppGraphResource)

	if before != nil {
		for _, r := range before.Resources {
			beforeResources[r.ID] = r
		}
	}
	if after != nil {
		for _, r := range after.Resources {
			afterResources[r.ID] = r
		}
	}

	// Track node IDs and their status
	nodeIDs := make(map[string]string)
	nodeStatus := make(map[string]string) // "added", "removed", "modified", "unchanged"
	nodeCounter := 0

	// Process added resources
	for _, resource := range diff.AddedResources {
		nodeID := fmt.Sprintf("node%d", nodeCounter)
		nodeCounter++
		nodeIDs[resource.ID] = nodeID
		nodeStatus[resource.ID] = "added"

		label := g.formatNodeLabel(resource)
		shape := g.getNodeShape(resource.Type)
		sb.WriteString(fmt.Sprintf("    %s%s\n", nodeID, formatMermaidNode(label, shape)))
	}

	// Process removed resources
	for _, resource := range diff.RemovedResources {
		nodeID := fmt.Sprintf("node%d", nodeCounter)
		nodeCounter++
		nodeIDs[resource.ID] = nodeID
		nodeStatus[resource.ID] = "removed"

		label := g.formatNodeLabel(resource)
		shape := g.getNodeShape(resource.Type)
		sb.WriteString(fmt.Sprintf("    %s%s\n", nodeID, formatMermaidNode(label, shape)))
	}

	// Process modified resources
	for _, rd := range diff.ModifiedResources {
		nodeID := fmt.Sprintf("node%d", nodeCounter)
		nodeCounter++
		nodeIDs[rd.ID] = nodeID
		nodeStatus[rd.ID] = "modified"

		resource := afterResources[rd.ID]
		label := g.formatNodeLabel(resource)
		shape := g.getNodeShape(resource.Type)
		sb.WriteString(fmt.Sprintf("    %s%s\n", nodeID, formatMermaidNode(label, shape)))
	}

	// Add style classes
	sb.WriteString("\n")
	sb.WriteString("    classDef added fill:#90EE90,stroke:#228B22\n")
	sb.WriteString("    classDef removed fill:#FFB6C1,stroke:#DC143C\n")
	sb.WriteString("    classDef modified fill:#FFE4B5,stroke:#FF8C00\n")

	// Apply styles
	var addedNodes, removedNodes, modifiedNodes []string
	for id, status := range nodeStatus {
		nodeID := nodeIDs[id]
		switch status {
		case "added":
			addedNodes = append(addedNodes, nodeID)
		case "removed":
			removedNodes = append(removedNodes, nodeID)
		case "modified":
			modifiedNodes = append(modifiedNodes, nodeID)
		}
	}

	if len(addedNodes) > 0 {
		sb.WriteString(fmt.Sprintf("    class %s added\n", strings.Join(addedNodes, ",")))
	}
	if len(removedNodes) > 0 {
		sb.WriteString(fmt.Sprintf("    class %s removed\n", strings.Join(removedNodes, ",")))
	}
	if len(modifiedNodes) > 0 {
		sb.WriteString(fmt.Sprintf("    class %s modified\n", strings.Join(modifiedNodes, ",")))
	}

	return sb.String()
}

// WrapInCodeBlock wraps Mermaid markup in a Markdown code fence.
func WrapInCodeBlock(mermaid string) string {
	return fmt.Sprintf("```mermaid\n%s```", mermaid)
}

// ValidateMermaidSyntax performs basic validation of Mermaid syntax.
func ValidateMermaidSyntax(mermaid string) error {
	lines := strings.Split(mermaid, "\n")
	if len(lines) == 0 {
		return fmt.Errorf("empty Mermaid diagram")
	}

	// Check for valid graph declaration
	firstLine := strings.TrimSpace(lines[0])
	validStart := regexp.MustCompile(`^(graph|flowchart)\s+(TB|BT|LR|RL)`)
	if !validStart.MatchString(firstLine) {
		return fmt.Errorf("invalid graph declaration: %s", firstLine)
	}

	return nil
}
