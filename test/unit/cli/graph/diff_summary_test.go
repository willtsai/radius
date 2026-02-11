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
)

func TestDiffSummary_Empty(t *testing.T) {
	t.Parallel()

	diff := &v20231001preview.GraphDiff{}
	diff.CalculateSummary()

	assert.True(t, diff.IsEmpty())
	assert.Equal(t, 0, diff.Summary.ResourcesAdded)
	assert.Equal(t, 0, diff.Summary.ResourcesRemoved)
	assert.Equal(t, 0, diff.Summary.ResourcesModified)
}

func TestDiffSummary_WithChanges(t *testing.T) {
	t.Parallel()

	diff := &v20231001preview.GraphDiff{
		AddedResources: []v20231001preview.StaticAppGraphResource{
			{ID: "id1", Name: "added1", Type: "type1"},
			{ID: "id2", Name: "added2", Type: "type2"},
		},
		RemovedResources: []v20231001preview.StaticAppGraphResource{
			{ID: "id3", Name: "removed1", Type: "type1"},
		},
		ModifiedResources: []v20231001preview.ResourceDiff{
			{ID: "id4", Name: "modified1", Type: "type1"},
		},
	}

	summary := graph.ComputeDiffSummary(diff)

	assert.False(t, diff.IsEmpty())
	assert.Equal(t, 2, summary.ResourcesAdded)
	assert.Equal(t, 1, summary.ResourcesRemoved)
	assert.Equal(t, 1, summary.ResourcesModified)
}

func TestDiffSummary_OnlyAdded(t *testing.T) {
	t.Parallel()

	diff := &v20231001preview.GraphDiff{
		AddedResources: []v20231001preview.StaticAppGraphResource{
			{ID: "id1", Name: "added1", Type: "type1"},
		},
	}

	summary := graph.ComputeDiffSummary(diff)

	assert.False(t, diff.IsEmpty())
	assert.Equal(t, 1, summary.ResourcesAdded)
	assert.Equal(t, 0, summary.ResourcesRemoved)
	assert.Equal(t, 0, summary.ResourcesModified)
}

func TestDiffSummary_OnlyRemoved(t *testing.T) {
	t.Parallel()

	diff := &v20231001preview.GraphDiff{
		RemovedResources: []v20231001preview.StaticAppGraphResource{
			{ID: "id1", Name: "removed1", Type: "type1"},
		},
	}

	summary := graph.ComputeDiffSummary(diff)

	assert.False(t, diff.IsEmpty())
	assert.Equal(t, 0, summary.ResourcesAdded)
	assert.Equal(t, 1, summary.ResourcesRemoved)
	assert.Equal(t, 0, summary.ResourcesModified)
}

func TestDiffSummary_OnlyModified(t *testing.T) {
	t.Parallel()

	diff := &v20231001preview.GraphDiff{
		ModifiedResources: []v20231001preview.ResourceDiff{
			{ID: "id1", Name: "modified1", Type: "type1", ChangedProperties: []v20231001preview.PropertyChange{{Path: "prop", OldValue: "a", NewValue: "b"}}},
		},
	}

	summary := graph.ComputeDiffSummary(diff)

	assert.False(t, diff.IsEmpty())
	assert.Equal(t, 0, summary.ResourcesAdded)
	assert.Equal(t, 0, summary.ResourcesRemoved)
	assert.Equal(t, 1, summary.ResourcesModified)
}

func TestResourceDiff_Fields(t *testing.T) {
	t.Parallel()

	rd := v20231001preview.ResourceDiff{
		ID:   "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/containers/test",
		Name: "test",
		Type: "Applications.Core/containers",
		ChangedProperties: []v20231001preview.PropertyChange{
			{Path: "image", OldValue: "nginx:1.19", NewValue: "nginx:1.21"},
			{Path: "replicas", OldValue: 2, NewValue: 3},
		},
	}

	assert.Equal(t, "test", rd.Name)
	assert.Equal(t, "Applications.Core/containers", rd.Type)
	assert.Len(t, rd.ChangedProperties, 2)
	assert.Equal(t, "image", rd.ChangedProperties[0].Path)
}

func TestPropertyChange_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		old      any
		new      any
		expected bool
	}{
		{
			name:     "string change",
			old:      "old",
			new:      "new",
			expected: true,
		},
		{
			name:     "int change",
			old:      1,
			new:      2,
			expected: true,
		},
		{
			name:     "nil to value",
			old:      nil,
			new:      "value",
			expected: true,
		},
		{
			name:     "value to nil",
			old:      "value",
			new:      nil,
			expected: true,
		},
		{
			name:     "same value",
			old:      "same",
			new:      "same",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pc := v20231001preview.PropertyChange{
				Path:     "test",
				OldValue: tc.old,
				NewValue: tc.new,
			}

			hasChange := pc.OldValue != pc.NewValue
			assert.Equal(t, tc.expected, hasChange)
		})
	}
}
