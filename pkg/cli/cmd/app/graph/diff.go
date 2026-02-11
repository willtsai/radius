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

package graph

import (
	"fmt"
	"reflect"

	"github.com/radius-project/radius/pkg/corerp/api/v20231001preview"
)

// ComputeDiff compares two application graphs and returns the differences.
// Either graph can be nil (treated as empty graph).
func ComputeDiff(before, after *v20231001preview.StaticAppGraph) *v20231001preview.GraphDiff {
	diff := &v20231001preview.GraphDiff{
		AddedResources:     []v20231001preview.StaticAppGraphResource{},
		RemovedResources:   []v20231001preview.StaticAppGraphResource{},
		ModifiedResources:  []v20231001preview.ResourceDiff{},
		AddedConnections:   []v20231001preview.StaticAppGraphConnection{},
		RemovedConnections: []v20231001preview.StaticAppGraphConnection{},
	}

	// Handle nil graphs
	beforeResources := getResources(before)
	afterResources := getResources(after)
	beforeConnections := getConnections(before)
	afterConnections := getConnections(after)

	// Build maps for efficient lookup
	beforeMap := buildResourceMap(beforeResources)
	afterMap := buildResourceMap(afterResources)

	// Find added and modified resources
	for id, afterRes := range afterMap {
		beforeRes, exists := beforeMap[id]
		if !exists {
			// Resource was added
			diff.AddedResources = append(diff.AddedResources, afterRes)
		} else {
			// Check if modified
			changes := compareResources(beforeRes, afterRes)
			if len(changes) > 0 {
				rd := v20231001preview.ResourceDiff{
					ID:                afterRes.ID,
					Name:              afterRes.Name,
					Type:              afterRes.Type,
					ChangedProperties: changes,
				}
				diff.ModifiedResources = append(diff.ModifiedResources, rd)
			}
		}
	}

	// Find removed resources
	for id, beforeRes := range beforeMap {
		if _, exists := afterMap[id]; !exists {
			diff.RemovedResources = append(diff.RemovedResources, beforeRes)
		}
	}

	// Compare connections
	beforeConnMap := buildConnectionMap(beforeConnections)
	afterConnMap := buildConnectionMap(afterConnections)

	for key, conn := range afterConnMap {
		if _, exists := beforeConnMap[key]; !exists {
			diff.AddedConnections = append(diff.AddedConnections, conn)
		}
	}

	for key, conn := range beforeConnMap {
		if _, exists := afterConnMap[key]; !exists {
			diff.RemovedConnections = append(diff.RemovedConnections, conn)
		}
	}

	// Compute summary
	diff.CalculateSummary()

	return diff
}

// ComputeDiffSummary generates summary statistics for a diff.
// Note: This is provided for convenience, but prefer using diff.CalculateSummary() method.
func ComputeDiffSummary(diff *v20231001preview.GraphDiff) v20231001preview.DiffSummary {
	diff.CalculateSummary()
	return diff.Summary
}

// getResources safely extracts resources from a graph (handles nil).
func getResources(graph *v20231001preview.StaticAppGraph) []v20231001preview.StaticAppGraphResource {
	if graph == nil {
		return nil
	}
	return graph.Resources
}

// getConnections safely extracts connections from a graph (handles nil).
func getConnections(graph *v20231001preview.StaticAppGraph) []v20231001preview.StaticAppGraphConnection {
	if graph == nil {
		return nil
	}
	return graph.Connections
}

// buildResourceMap creates a map of resource ID to resource.
func buildResourceMap(resources []v20231001preview.StaticAppGraphResource) map[string]v20231001preview.StaticAppGraphResource {
	m := make(map[string]v20231001preview.StaticAppGraphResource)
	for _, r := range resources {
		m[r.ID] = r
	}
	return m
}

// buildConnectionMap creates a map of connection key to connection.
// The key is "sourceID|targetID|type" to ensure uniqueness.
func buildConnectionMap(connections []v20231001preview.StaticAppGraphConnection) map[string]v20231001preview.StaticAppGraphConnection {
	m := make(map[string]v20231001preview.StaticAppGraphConnection)
	for _, c := range connections {
		key := fmt.Sprintf("%s|%s|%s", c.SourceID, c.TargetID, c.Type)
		m[key] = c
	}
	return m
}

// compareResources finds differences between two versions of the same resource.
func compareResources(before, after v20231001preview.StaticAppGraphResource) []v20231001preview.PropertyChange {
	var changes []v20231001preview.PropertyChange

	// Compare properties maps
	allKeys := make(map[string]bool)
	for k := range before.Properties {
		allKeys[k] = true
	}
	for k := range after.Properties {
		allKeys[k] = true
	}

	for key := range allKeys {
		beforeVal := before.Properties[key]
		afterVal := after.Properties[key]

		if !reflect.DeepEqual(beforeVal, afterVal) {
			changes = append(changes, v20231001preview.PropertyChange{
				Path:     key,
				OldValue: beforeVal,
				NewValue: afterVal,
			})
		}
	}

	// Compare source location changes
	if before.SourceLocation.File != after.SourceLocation.File {
		changes = append(changes, v20231001preview.PropertyChange{
			Path:     "sourceLocation.file",
			OldValue: before.SourceLocation.File,
			NewValue: after.SourceLocation.File,
		})
	}
	if before.SourceLocation.Line != after.SourceLocation.Line {
		changes = append(changes, v20231001preview.PropertyChange{
			Path:     "sourceLocation.line",
			OldValue: before.SourceLocation.Line,
			NewValue: after.SourceLocation.Line,
		})
	}

	return changes
}

// DiffFromJSON loads graphs from JSON and computes the diff.
// This is a convenience function for the GitHub Action.
func DiffFromJSON(beforeJSON, afterJSON []byte) (*v20231001preview.GraphDiff, error) {
	var before, after *v20231001preview.StaticAppGraph

	if len(beforeJSON) > 0 {
		before = &v20231001preview.StaticAppGraph{}
		if err := parseGraphJSON(beforeJSON, before); err != nil {
			return nil, fmt.Errorf("failed to parse before graph: %w", err)
		}
	}

	if len(afterJSON) > 0 {
		after = &v20231001preview.StaticAppGraph{}
		if err := parseGraphJSON(afterJSON, after); err != nil {
			return nil, fmt.Errorf("failed to parse after graph: %w", err)
		}
	}

	return ComputeDiff(before, after), nil
}

// parseGraphJSON parses JSON into a StaticAppGraph.
func parseGraphJSON(data []byte, graph *v20231001preview.StaticAppGraph) error {
	// Use standard encoding/json for simplicity
	// The DeterministicJSON package is for output only
	return nil // TODO: Implement proper JSON parsing
}
