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

package git

import (
	"github.com/radius-project/radius/pkg/corerp/api/v20231001preview"
)

// MetadataEnricher enriches resources with git metadata.
type MetadataEnricher struct {
	// Repository is the git repository to query
	Repository *Repository

	// CacheBames caches blame results by file for efficiency
	blameCache map[string]*BlameResult
}

// NewMetadataEnricher creates a new MetadataEnricher for the given repository.
func NewMetadataEnricher(repo *Repository) *MetadataEnricher {
	return &MetadataEnricher{
		Repository: repo,
		blameCache: make(map[string]*BlameResult),
	}
}

// EnrichResource adds git metadata to a resource based on its source location.
// Returns nil GitInfo if git metadata cannot be retrieved.
func (e *MetadataEnricher) EnrichResource(resource *v20231001preview.StaticAppGraphResource) error {
	if resource.SourceLocation.File == "" || resource.SourceLocation.Line == 0 {
		return nil
	}

	// Check for uncommitted changes first
	hasUncommitted, err := e.Repository.HasUncommittedChanges(resource.SourceLocation.File)
	if err != nil {
		// Non-fatal - continue without uncommitted check
		hasUncommitted = false
	}

	// Get blame info for the resource line
	blameInfo, err := e.getBlameInfo(resource.SourceLocation.File, resource.SourceLocation.Line)
	if err != nil {
		// Non-fatal - mark as uncommitted if blame fails
		if hasUncommitted {
			resource.GitInfo = &v20231001preview.GitInfo{
				Uncommitted: true,
			}
		}
		return nil
	}

	if blameInfo == nil {
		return nil
	}

	// Build GitInfo from blame
	resource.GitInfo = &v20231001preview.GitInfo{
		CommitSHA:   blameInfo.SHA,
		CommitShort: blameInfo.SHAShort,
		Author:      formatAuthor(blameInfo.Author, blameInfo.AuthorEmail),
		Date:        blameInfo.AuthorTime,
		Message:     blameInfo.Summary,
		Uncommitted: hasUncommitted,
	}

	return nil
}

// EnrichResources enriches multiple resources with git metadata.
func (e *MetadataEnricher) EnrichResources(resources []v20231001preview.StaticAppGraphResource) error {
	for i := range resources {
		if err := e.EnrichResource(&resources[i]); err != nil {
			return err
		}
	}
	return nil
}

// EnrichGraph adds git metadata to all resources in a graph.
func (e *MetadataEnricher) EnrichGraph(graph *v20231001preview.StaticAppGraph) error {
	// Add current commit to metadata
	commitSHA, err := e.Repository.GetCurrentCommit()
	if err == nil {
		graph.Metadata.GitCommit = commitSHA
	}

	// Enrich each resource
	return e.EnrichResources(graph.Resources)
}

// getBlameInfo retrieves blame information, using cache when available.
func (e *MetadataEnricher) getBlameInfo(filePath string, line int) (*BlameInfo, error) {
	// Check cache first
	cached, ok := e.blameCache[filePath]
	if ok {
		if cached.HistoryUnavailable {
			return nil, nil
		}
		if info, ok := cached.Lines[line]; ok {
			return info, nil
		}
	}

	// Fetch blame for this line
	result, err := e.Repository.Blame(filePath, line, line)
	if err != nil {
		return nil, err
	}

	if result.HistoryUnavailable {
		// Cache the unavailable status
		e.blameCache[filePath] = result
		return nil, nil
	}

	// Merge with existing cache
	if cached == nil {
		e.blameCache[filePath] = result
	} else {
		for l, info := range result.Lines {
			cached.Lines[l] = info
		}
	}

	return result.Lines[line], nil
}

// formatAuthor formats author information for display.
func formatAuthor(name, email string) string {
	if name == "" {
		return email
	}
	return name
}

// EnrichGraphFromPath creates a MetadataEnricher and enriches a graph from a file path.
// This is a convenience function for single-use enrichment.
func EnrichGraphFromPath(graph *v20231001preview.StaticAppGraph, filePath string) error {
	repo, err := FindRepository(filePath)
	if err != nil {
		return err
	}
	if repo == nil {
		// Not in a git repository - skip enrichment
		return nil
	}

	enricher := NewMetadataEnricher(repo)
	return enricher.EnrichGraph(graph)
}

// GetResourceGitInfo retrieves git information for a single resource location.
// This is a convenience function for getting git info without full enrichment.
func GetResourceGitInfo(filePath string, line int) (*v20231001preview.GitInfo, error) {
	repo, err := FindRepository(filePath)
	if err != nil {
		return nil, err
	}
	if repo == nil {
		return nil, nil
	}

	// Check for uncommitted changes
	hasUncommitted, _ := repo.HasUncommittedChanges(filePath)

	// Get blame info
	blameInfo, err := repo.GetBlameInfoForLine(filePath, line)
	if err != nil {
		if hasUncommitted {
			return &v20231001preview.GitInfo{Uncommitted: true}, nil
		}
		return nil, err
	}

	if blameInfo == nil {
		return nil, nil
	}

	return &v20231001preview.GitInfo{
		CommitSHA:   blameInfo.SHA,
		CommitShort: blameInfo.SHAShort,
		Author:      formatAuthor(blameInfo.Author, blameInfo.AuthorEmail),
		Date:        blameInfo.AuthorTime,
		Message:     blameInfo.Summary,
		Uncommitted: hasUncommitted,
	}, nil
}
