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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repository represents a git repository and provides methods for
// extracting git metadata from files.
type Repository struct {
	// RootPath is the absolute path to the repository root
	RootPath string

	// IsShallow indicates whether the repository is a shallow clone
	IsShallow bool
}

// FindRepository discovers the git repository containing the given path.
// Returns nil if the path is not inside a git repository.
func FindRepository(path string) (*Repository, error) {
	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// Start from the given path and walk up to find .git
	dir := absPath
	if info, err := os.Stat(absPath); err == nil && !info.IsDir() {
		dir = filepath.Dir(absPath)
	}

	for {
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && (info.IsDir() || info.Mode().IsRegular()) {
			// Found .git directory or file (worktree)
			repo := &Repository{RootPath: dir}
			repo.detectShallowClone()
			return repo, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding .git
			return nil, nil
		}
		dir = parent
	}
}

// IsGitRepository checks if the given path is inside a git repository.
func IsGitRepository(path string) bool {
	repo, err := FindRepository(path)
	return err == nil && repo != nil
}

// detectShallowClone checks if the repository is a shallow clone.
func (r *Repository) detectShallowClone() {
	shallowFile := filepath.Join(r.RootPath, ".git", "shallow")
	if _, err := os.Stat(shallowFile); err == nil {
		r.IsShallow = true
	}
}

// GetCurrentCommit returns the current HEAD commit SHA.
// Returns empty string if not in a git repository or on error.
func (r *Repository) GetCurrentCommit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = r.RootPath

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCurrentCommitShort returns the abbreviated commit SHA (7 characters).
func (r *Repository) GetCurrentCommitShort() (string, error) {
	sha, err := r.GetCurrentCommit()
	if err != nil {
		return "", err
	}
	if len(sha) >= 7 {
		return sha[:7], nil
	}
	return sha, nil
}

// GetRelativePath converts an absolute path to a path relative to the repo root.
func (r *Repository) GetRelativePath(absPath string) (string, error) {
	return filepath.Rel(r.RootPath, absPath)
}

// GetAbsolutePath converts a relative path to an absolute path from repo root.
func (r *Repository) GetAbsolutePath(relPath string) string {
	return filepath.Join(r.RootPath, relPath)
}

// GetCurrentBranch returns the current branch name.
// Returns "HEAD" if in detached HEAD state.
func (r *Repository) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = r.RootPath

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// HasRemote checks if the repository has a configured remote.
func (r *Repository) HasRemote(name string) bool {
	cmd := exec.Command("git", "remote", "get-url", name)
	cmd.Dir = r.RootPath

	err := cmd.Run()
	return err == nil
}

// GetRemoteURL returns the URL of the specified remote.
func (r *Repository) GetRemoteURL(name string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", name)
	cmd.Dir = r.RootPath

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}
