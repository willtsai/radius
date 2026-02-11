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
	"bufio"
	"os/exec"
	"strings"
)

// FileStatus represents the git status of a file.
type FileStatus struct {
	// Path is the file path relative to the repository root
	Path string

	// IndexStatus is the status in the staging area (first character)
	IndexStatus rune

	// WorktreeStatus is the status in the working tree (second character)
	WorktreeStatus rune
}

// StatusResult contains the result of a git status operation.
type StatusResult struct {
	// Files maps file paths to their status
	Files map[string]*FileStatus

	// HasUncommitted indicates there are uncommitted changes
	HasUncommitted bool

	// HasStaged indicates there are staged but uncommitted changes
	HasStaged bool

	// HasUnstaged indicates there are unstaged changes
	HasUnstaged bool

	// HasUntracked indicates there are untracked files
	HasUntracked bool
}

// Status runs git status and returns information about the working tree.
func (r *Repository) Status() (*StatusResult, error) {
	cmd := exec.Command("git", "status", "--porcelain", "-uall")
	cmd.Dir = r.RootPath

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseStatusOutput(string(output))
}

// parseStatusOutput parses git status --porcelain output.
func parseStatusOutput(output string) (*StatusResult, error) {
	result := &StatusResult{
		Files: make(map[string]*FileStatus),
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 3 {
			continue
		}

		// Format: XY filename
		// X = index status, Y = worktree status
		indexStatus := rune(line[0])
		worktreeStatus := rune(line[1])
		path := strings.TrimSpace(line[3:])

		// Handle renamed files (format: "R  old -> new")
		if strings.Contains(path, " -> ") {
			parts := strings.Split(path, " -> ")
			path = parts[len(parts)-1]
		}

		status := &FileStatus{
			Path:           path,
			IndexStatus:    indexStatus,
			WorktreeStatus: worktreeStatus,
		}
		result.Files[path] = status

		// Track overall status
		if indexStatus != ' ' && indexStatus != '?' {
			result.HasStaged = true
		}
		if worktreeStatus != ' ' && worktreeStatus != '?' {
			result.HasUnstaged = true
		}
		if indexStatus == '?' && worktreeStatus == '?' {
			result.HasUntracked = true
		}
		result.HasUncommitted = true
	}

	return result, scanner.Err()
}

// HasUncommittedChanges checks if a specific file has uncommitted changes.
func (r *Repository) HasUncommittedChanges(filePath string) (bool, error) {
	status, err := r.Status()
	if err != nil {
		return false, err
	}

	// Check for the file in the status
	for path := range status.Files {
		if path == filePath || strings.HasSuffix(filePath, path) {
			return true, nil
		}
	}

	return false, nil
}

// IsFileTracked checks if a file is tracked by git.
func (r *Repository) IsFileTracked(filePath string) (bool, error) {
	cmd := exec.Command("git", "ls-files", "--error-unmatch", filePath)
	cmd.Dir = r.RootPath

	err := cmd.Run()
	if err != nil {
		// Exit code 1 means file is not tracked
		return false, nil
	}
	return true, nil
}

// IsFileIgnored checks if a file is ignored by .gitignore.
func (r *Repository) IsFileIgnored(filePath string) (bool, error) {
	cmd := exec.Command("git", "check-ignore", "-q", filePath)
	cmd.Dir = r.RootPath

	err := cmd.Run()
	if err != nil {
		// Exit code 1 means file is not ignored
		return false, nil
	}
	return true, nil
}

// GetModifiedFiles returns a list of files with uncommitted changes.
func (r *Repository) GetModifiedFiles() ([]string, error) {
	status, err := r.Status()
	if err != nil {
		return nil, err
	}

	var modified []string
	for path, fileStatus := range status.Files {
		// Include modified, added, deleted, renamed files (not untracked)
		if fileStatus.IndexStatus != '?' || fileStatus.WorktreeStatus != '?' {
			modified = append(modified, path)
		}
	}

	return modified, nil
}

// GetUntrackedFiles returns a list of untracked files.
func (r *Repository) GetUntrackedFiles() ([]string, error) {
	status, err := r.Status()
	if err != nil {
		return nil, err
	}

	var untracked []string
	for path, fileStatus := range status.Files {
		if fileStatus.IndexStatus == '?' && fileStatus.WorktreeStatus == '?' {
			untracked = append(untracked, path)
		}
	}

	return untracked, nil
}

// IsDirty checks if the repository has any uncommitted changes.
func (r *Repository) IsDirty() (bool, error) {
	status, err := r.Status()
	if err != nil {
		return false, err
	}
	return status.HasUncommitted, nil
}
