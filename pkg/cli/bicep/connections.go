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

package bicep

import (
	"strings"

	"github.com/radius-project/radius/pkg/corerp/api/v20231001preview"
)

// ConnectionDetector detects connections between resources in an ARM template.
type ConnectionDetector struct {
	// ResourceIDs maps resource names to their full resource IDs
	ResourceIDs map[string]string
}

// NewConnectionDetector creates a new ConnectionDetector.
func NewConnectionDetector() *ConnectionDetector {
	return &ConnectionDetector{
		ResourceIDs: make(map[string]string),
	}
}

// BuildResourceIndex builds a lookup index from resource name to resource ID.
func (d *ConnectionDetector) BuildResourceIndex(resources []v20231001preview.StaticAppGraphResource) {
	for _, r := range resources {
		d.ResourceIDs[r.Name] = r.ID
		// Also index by type/name combination
		d.ResourceIDs[r.Type+"/"+r.Name] = r.ID
	}
}

// BuildResourceIndexWithSymbolicNames extends the resource index with v2 symbolic names.
// In ARM JSON v2.0, dependsOn and reference() expressions use the Bicep symbolic name
// rather than the resource name, so we need both in the lookup.
func (d *ConnectionDetector) BuildResourceIndexWithSymbolicNames(armResources []ARMResource, graphResources []v20231001preview.StaticAppGraphResource) {
	for i, armRes := range armResources {
		if i >= len(graphResources) {
			continue
		}
		if armRes.SymbolicName != "" && armRes.SymbolicName != armRes.Name {
			d.ResourceIDs[armRes.SymbolicName] = graphResources[i].ID
		}
	}
}

// DetectConnections extracts connections from ARM resources.
// It looks for:
// - Explicit connections in the "connections" property
// - Gateway routes in the "routes" property
// - dependsOn relationships
func (d *ConnectionDetector) DetectConnections(armResources []ARMResource, graphResources []v20231001preview.StaticAppGraphResource) []v20231001preview.StaticAppGraphConnection {
	d.BuildResourceIndex(graphResources)

	var connections []v20231001preview.StaticAppGraphConnection

	for i, armRes := range armResources {
		if i >= len(graphResources) {
			continue
		}
		sourceID := graphResources[i].ID

		// Check for explicit connections property
		connConns := d.extractConnectionsProperty(armRes.Properties, sourceID)
		connections = append(connections, connConns...)

		// Check for gateway routes
		routeConns := d.extractRoutes(armRes.Properties, sourceID)
		connections = append(connections, routeConns...)

		// Check for dependsOn
		depConns := d.extractDependsOn(armRes.DependsOn, sourceID)
		connections = append(connections, depConns...)
	}

	return connections
}

// extractConnectionsProperty extracts connections from the "connections" property.
// Radius containers define connections as:
//
//	connections: {
//	  backend: { source: backend.id }
//	}
func (d *ConnectionDetector) extractConnectionsProperty(properties map[string]any, sourceID string) []v20231001preview.StaticAppGraphConnection {
	var connections []v20231001preview.StaticAppGraphConnection

	connsAny, ok := properties["connections"]
	if !ok {
		return connections
	}

	connsMap, ok := connsAny.(map[string]any)
	if !ok {
		return connections
	}

	for _, connDef := range connsMap {
		connDefMap, ok := connDef.(map[string]any)
		if !ok {
			continue
		}

		// Look for "source" property
		sourceAny, ok := connDefMap["source"]
		if !ok {
			continue
		}

		targetID := d.resolveResourceReference(sourceAny)
		if targetID != "" {
			connections = append(connections, v20231001preview.StaticAppGraphConnection{
				SourceID: sourceID,
				TargetID: targetID,
				Type:     v20231001preview.StaticConnectionTypeConnection,
			})
		}
	}

	return connections
}

// extractRoutes extracts connections from gateway "routes" property.
// Radius gateways define routes as:
//
//	routes: {
//	  '/api': { destination: 'http://backend:8080' }
//	}
func (d *ConnectionDetector) extractRoutes(properties map[string]any, sourceID string) []v20231001preview.StaticAppGraphConnection {
	var connections []v20231001preview.StaticAppGraphConnection

	routesAny, ok := properties["routes"]
	if !ok {
		return connections
	}

	routesMap, ok := routesAny.(map[string]any)
	if !ok {
		return connections
	}

	for _, routeDef := range routesMap {
		routeDefMap, ok := routeDef.(map[string]any)
		if !ok {
			continue
		}

		// Look for "destination" property
		destAny, ok := routeDefMap["destination"]
		if !ok {
			continue
		}

		targetID := d.resolveResourceReference(destAny)
		if targetID != "" {
			connections = append(connections, v20231001preview.StaticAppGraphConnection{
				SourceID: sourceID,
				TargetID: targetID,
				Type:     v20231001preview.StaticConnectionTypeRoute,
			})
		}
	}

	return connections
}

// extractDependsOn extracts connections from ARM dependsOn array.
func (d *ConnectionDetector) extractDependsOn(dependsOn []string, sourceID string) []v20231001preview.StaticAppGraphConnection {
	var connections []v20231001preview.StaticAppGraphConnection

	for _, dep := range dependsOn {
		targetID := d.resolveResourceReference(dep)
		if targetID != "" {
			connections = append(connections, v20231001preview.StaticAppGraphConnection{
				SourceID: sourceID,
				TargetID: targetID,
				Type:     v20231001preview.StaticConnectionTypeDependsOn,
			})
		}
	}

	return connections
}

// resolveResourceReference attempts to resolve a resource reference to a full resource ID.
// References can be:
// - A full resource ID: "/planes/radius/local/..."
// - An ARM resource reference: "[resourceId('Microsoft.App/containers', 'mycontainer')]"
// - A simple name reference: "mycontainer"
func (d *ConnectionDetector) resolveResourceReference(ref any) string {
	refStr, ok := ref.(string)
	if !ok {
		return ""
	}

	// If it's already a full resource ID
	if strings.HasPrefix(refStr, "/") {
		return refStr
	}

	// Try to look up by name
	if id, ok := d.ResourceIDs[refStr]; ok {
		return id
	}

	// Handle ARM resourceId() expressions
	// e.g., "[resourceId('Applications.Core/containers', 'backend')]"
	if strings.HasPrefix(refStr, "[resourceId(") {
		// Extract the resource name from the expression
		// This is a simplified parser - production code would need more robust parsing
		parts := strings.Split(refStr, "'")
		if len(parts) >= 4 {
			// parts[1] = type, parts[3] = name
			name := parts[3]
			if id, ok := d.ResourceIDs[name]; ok {
				return id
			}
		}
	}

	// Handle reference() expressions
	// e.g., "[reference('backend').id]"
	if strings.HasPrefix(refStr, "[reference(") {
		parts := strings.Split(refStr, "'")
		if len(parts) >= 2 {
			name := parts[1]
			if id, ok := d.ResourceIDs[name]; ok {
				return id
			}
		}
	}

	return ""
}

// DeduplicateConnections removes duplicate connections from the list.
func DeduplicateConnections(connections []v20231001preview.StaticAppGraphConnection) []v20231001preview.StaticAppGraphConnection {
	seen := make(map[string]bool)
	var result []v20231001preview.StaticAppGraphConnection

	for _, conn := range connections {
		key := conn.SourceID + "|" + conn.TargetID + "|" + string(conn.Type)
		if !seen[key] {
			seen[key] = true
			result = append(result, conn)
		}
	}

	return result
}

// Connection represents a connection extracted from an ARM template.
type Connection struct {
	SourceResourceID string
	TargetResourceID string
	Type             string
}

// ExtractConnections extracts connections from an ARM template.
// This is a convenience function that creates a ConnectionDetector,
// extracts resources, and detects connections.
func ExtractConnections(template *ARMTemplate) []Connection {
	if template == nil {
		return nil
	}

	// Build a list of graph resources from ARM resources
	// Use name as ID for detection purposes since we don't have full resource IDs
	var graphResources []v20231001preview.StaticAppGraphResource
	for _, armRes := range template.Resources {
		graphResources = append(graphResources, v20231001preview.StaticAppGraphResource{
			ID:   armRes.Name, // Use name as ID for detection
			Name: armRes.Name,
			Type: armRes.Type,
		})
	}

	// Detect connections
	detector := NewConnectionDetector()
	// Also index symbolic names for v2 ARM JSON where dependsOn and
	// reference() expressions use the Bicep variable name.
	detector.BuildResourceIndexWithSymbolicNames(template.Resources, graphResources)
	connections := detector.DetectConnections(template.Resources, graphResources)

	// Convert to Connection type
	var result []Connection
	for _, conn := range connections {
		result = append(result, Connection{
			SourceResourceID: conn.SourceID,
			TargetResourceID: conn.TargetID,
			Type:             string(conn.Type),
		})
	}

	return result
}
