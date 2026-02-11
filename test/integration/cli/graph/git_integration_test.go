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
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/radius-project/radius/pkg/cli/git"
	"github.com/radius-project/radius/pkg/corerp/api/v20231001preview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGitIntegration_FindRepository tests finding git repos in real environments.
func TestGitIntegration_FindRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// The Radius repo itself should be a git repository
	repoRoot, err := findRepoRoot()
	require.NoError(t, err, "expected to find Radius repo root")

	repo, err := git.FindRepository(repoRoot)
	require.NoError(t, err)
	require.NotNil(t, repo)

	assert.Equal(t, repoRoot, repo.RootPath)
}

// TestGitIntegration_NonGitDirectory tests behavior outside git repos.
func TestGitIntegration_NonGitDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create a temp directory that's not a git repo
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	repo, err := git.FindRepository(tmpDir)
	require.NoError(t, err)
	assert.Nil(t, repo, "should return nil for non-git directory")

	isGit := git.IsGitRepository(tmpDir)
	assert.False(t, isGit)
}

// TestGitIntegration_GetCurrentCommit tests getting current HEAD commit.
func TestGitIntegration_GetCurrentCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo, err := findTestRepo()
	if err != nil {
		t.Skip("skipping - git repository not available")
	}

	sha, err := repo.GetCurrentCommit()
	require.NoError(t, err)
	assert.Len(t, sha, 40, "expected full 40-char SHA")
	assert.Regexp(t, `^[0-9a-f]{40}$`, sha, "should be valid hex SHA")

	shortSHA, err := repo.GetCurrentCommitShort()
	require.NoError(t, err)
	assert.Len(t, shortSHA, 7, "expected 7-char short SHA")
	assert.Equal(t, sha[:7], shortSHA)
}

// TestGitIntegration_Blame tests blame on actual files.
func TestGitIntegration_Blame(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo, err := findTestRepo()
	if err != nil {
		t.Skip("skipping - git repository not available")
	}

	// Find a file that definitely exists and has git history
	testFile := "go.mod"
	fullPath := filepath.Join(repo.RootPath, testFile)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Skip("skipping - test file not found")
	}

	result, err := repo.Blame(testFile, 1, 5)
	require.NoError(t, err)

	if result.HistoryUnavailable {
		t.Skip("skipping - git history unavailable (shallow clone)")
	}

	assert.NotEmpty(t, result.Lines, "expected some blame lines")

	// Verify blame info structure
	for lineNum, info := range result.Lines {
		assert.GreaterOrEqual(t, lineNum, 1)
		assert.LessOrEqual(t, lineNum, 5)
		assert.Len(t, info.SHA, 40, "expected 40-char SHA")
		assert.NotEmpty(t, info.Author, "expected author name")
		assert.False(t, info.AuthorTime.IsZero(), "expected valid timestamp")
	}
}

// TestGitIntegration_Status tests git status operations.
func TestGitIntegration_Status(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo, err := findTestRepo()
	if err != nil {
		t.Skip("skipping - git repository not available")
	}

	status, err := repo.Status()
	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.NotNil(t, status.Files)
	// Note: Can't assert specific status because it depends on workspace state
}

// TestGitIntegration_Log tests git log operations.
func TestGitIntegration_Log(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo, err := findTestRepo()
	if err != nil {
		t.Skip("skipping - git repository not available")
	}

	commits, err := repo.Log(git.LogOptions{MaxCount: 5})
	require.NoError(t, err)
	assert.NotEmpty(t, commits, "expected at least one commit")
	assert.LessOrEqual(t, len(commits), 5)

	// Verify commit structure
	for _, commit := range commits {
		assert.Len(t, commit.SHA, 40)
		assert.Len(t, commit.SHAShort, 7)
		assert.NotEmpty(t, commit.Author)
		assert.NotEmpty(t, commit.AuthorEmail)
		assert.False(t, commit.AuthorDate.IsZero())
		assert.NotEmpty(t, commit.Subject)
	}
}

// TestGitIntegration_MetadataEnricher tests full metadata enrichment flow.
func TestGitIntegration_MetadataEnricher(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo, err := findTestRepo()
	if err != nil {
		t.Skip("skipping - git repository not available")
	}

	enricher := git.NewMetadataEnricher(repo)

	// Create a test resource with source location pointing to go.mod
	resource := v20231001preview.StaticAppGraphResource{
		ID:   "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/applications/test",
		Name: "test",
		Type: "Applications.Core/applications",
		SourceLocation: v20231001preview.SourceLocation{
			File: "go.mod",
			Line: 1,
		},
	}

	err = enricher.EnrichResource(&resource)
	require.NoError(t, err)

	// Verify git info was populated (unless shallow clone)
	if resource.GitInfo != nil && !resource.GitInfo.Uncommitted {
		assert.NotEmpty(t, resource.GitInfo.CommitSHA)
		assert.NotEmpty(t, resource.GitInfo.CommitShort)
		assert.NotEmpty(t, resource.GitInfo.Author)
		assert.False(t, resource.GitInfo.Date.IsZero())
	}
}

// TestGitIntegration_EnrichGraph tests graph-level enrichment.
func TestGitIntegration_EnrichGraph(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo, err := findTestRepo()
	if err != nil {
		t.Skip("skipping - git repository not available")
	}

	graph := &v20231001preview.StaticAppGraph{
		Metadata: v20231001preview.StaticAppGraphMetadata{
			GeneratedAt:      time.Now().UTC(),
			RadiusCliVersion: "test",
			SourceFiles:      []string{"go.mod"},
			SourceHash:       "sha256:abc123",
		},
		Resources: []v20231001preview.StaticAppGraphResource{
			{
				ID:   "/planes/radius/local/resourceGroups/rg/providers/Applications.Core/applications/test",
				Name: "test",
				Type: "Applications.Core/applications",
				SourceLocation: v20231001preview.SourceLocation{
					File: "go.mod",
					Line: 1,
				},
			},
		},
	}

	enricher := git.NewMetadataEnricher(repo)
	err = enricher.EnrichGraph(graph)
	require.NoError(t, err)

	// Verify graph metadata includes git commit
	if graph.Metadata.GitCommit != "" {
		assert.Len(t, graph.Metadata.GitCommit, 40)
	}
}

// TestGitIntegration_ShallowClone tests behavior with shallow clone detection.
func TestGitIntegration_ShallowClone(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo, err := findTestRepo()
	if err != nil {
		t.Skip("skipping - git repository not available")
	}

	// IsShallow should be populated based on .git/shallow existence
	// We can't assert the value because it depends on how the repo was cloned
	t.Logf("IsShallow: %v", repo.IsShallow)

	// Even with shallow clone, basic operations should work
	sha, err := repo.GetCurrentCommit()
	require.NoError(t, err)
	assert.NotEmpty(t, sha)
}

// TestGitIntegration_FilePaths tests path operations.
func TestGitIntegration_FilePaths(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repo, err := findTestRepo()
	if err != nil {
		t.Skip("skipping - git repository not available")
	}

	// Test relative path conversion
	absPath := filepath.Join(repo.RootPath, "pkg", "cli", "git", "repo.go")
	relPath, err := repo.GetRelativePath(absPath)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("pkg", "cli", "git", "repo.go"), relPath)

	// Test absolute path conversion
	resultAbs := repo.GetAbsolutePath("pkg/cli/git/repo.go")
	expected := filepath.Join(repo.RootPath, "pkg/cli/git/repo.go")
	assert.Equal(t, expected, resultAbs)
}

// Helper functions

// findRepoRoot finds the Radius repository root directory.
func findRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return filepath.Clean(string(output[:len(output)-1])), nil
}

// findTestRepo returns a Repository for the current test environment.
func findTestRepo() (*git.Repository, error) {
	root, err := findRepoRoot()
	if err != nil {
		return nil, err
	}
	return git.FindRepository(root)
}
