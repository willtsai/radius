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

package v20231001preview

// GraphDiff represents the differences between two static app graphs.
// This is the result of comparing a base graph with a head graph.
type GraphDiff struct {
	// BaseCommit is the commit SHA of the base graph (optional)
	BaseCommit string `json:"baseCommit,omitempty"`

	// HeadCommit is the commit SHA of the head graph (optional)
	HeadCommit string `json:"headCommit,omitempty"`

	// AddedResources are resources present in head but not in base
	AddedResources []StaticAppGraphResource `json:"addedResources"`

	// RemovedResources are resources present in base but not in head
	RemovedResources []StaticAppGraphResource `json:"removedResources"`

	// ModifiedResources are resources with changed properties
	ModifiedResources []ResourceDiff `json:"modifiedResources"`

	// AddedConnections are connections present in head but not in base
	AddedConnections []StaticAppGraphConnection `json:"addedConnections"`

	// RemovedConnections are connections present in base but not in head
	RemovedConnections []StaticAppGraphConnection `json:"removedConnections"`

	// Summary provides a human-readable overview of the diff
	Summary DiffSummary `json:"summary"`
}

// ResourceDiff captures changes to a single resource between base and head graphs.
type ResourceDiff struct {
	// ID is the resource ID (same in both base and head)
	ID string `json:"id"`

	// Name is the resource name
	Name string `json:"name"`

	// Type is the resource type
	Type string `json:"type"`

	// ChangedProperties lists the property paths that changed
	ChangedProperties []PropertyChange `json:"changedProperties"`
}

// PropertyChange describes a single property modification between two graphs.
type PropertyChange struct {
	// Path is the JSON path to the changed property (e.g., "properties.container.image")
	Path string `json:"path"`

	// OldValue is the value in the base graph
	OldValue any `json:"oldValue,omitempty"`

	// NewValue is the value in the head graph
	NewValue any `json:"newValue,omitempty"`
}

// DiffSummary provides aggregate statistics for the graph diff.
type DiffSummary struct {
	// TotalChanges is the total number of changes (resources + connections)
	TotalChanges int `json:"totalChanges"`

	// ResourcesAdded is the count of resources added
	ResourcesAdded int `json:"resourcesAdded"`

	// ResourcesRemoved is the count of resources removed
	ResourcesRemoved int `json:"resourcesRemoved"`

	// ResourcesModified is the count of resources with property changes
	ResourcesModified int `json:"resourcesModified"`

	// ConnectionsAdded is the count of connections added
	ConnectionsAdded int `json:"connectionsAdded"`

	// ConnectionsRemoved is the count of connections removed
	ConnectionsRemoved int `json:"connectionsRemoved"`
}

// IsEmpty returns true if the diff contains no changes.
func (d *GraphDiff) IsEmpty() bool {
	return len(d.AddedResources) == 0 &&
		len(d.RemovedResources) == 0 &&
		len(d.ModifiedResources) == 0 &&
		len(d.AddedConnections) == 0 &&
		len(d.RemovedConnections) == 0
}

// CalculateSummary computes the DiffSummary from the diff contents.
// This should be called after populating the diff to ensure the summary is accurate.
func (d *GraphDiff) CalculateSummary() {
	d.Summary = DiffSummary{
		ResourcesAdded:     len(d.AddedResources),
		ResourcesRemoved:   len(d.RemovedResources),
		ResourcesModified:  len(d.ModifiedResources),
		ConnectionsAdded:   len(d.AddedConnections),
		ConnectionsRemoved: len(d.RemovedConnections),
	}
	d.Summary.TotalChanges = d.Summary.ResourcesAdded +
		d.Summary.ResourcesRemoved +
		d.Summary.ResourcesModified +
		d.Summary.ConnectionsAdded +
		d.Summary.ConnectionsRemoved
}
