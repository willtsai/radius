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
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// CommitInfo contains information about a git commit.
type CommitInfo struct {
	// SHA is the full 40-character commit hash
	SHA string

	// SHAShort is the abbreviated 7-character commit hash
	SHAShort string

	// Author is the author's name
	Author string

	// AuthorEmail is the author's email address
	AuthorEmail string

	// AuthorDate is the original authoring timestamp
	AuthorDate time.Time

	// Committer is the committer's name
	Committer string

	// CommitterEmail is the committer's email address
	CommitterEmail string

	// CommitDate is the commit timestamp
	CommitDate time.Time

	// Subject is the commit message first line
	Subject string

	// Body is the full commit message
	Body string

	// Parents are the parent commit SHAs
	Parents []string
}

// LogOptions configures the git log query.
type LogOptions struct {
	// MaxCount limits the number of commits returned
	MaxCount int

	// Since filters commits after this date
	Since time.Time

	// Until filters commits before this date
	Until time.Time

	// Author filters by author name/email
	Author string

	// FilePaths filters commits that touched these files
	FilePaths []string
}

// Log retrieves commit history with optional filtering.
func (r *Repository) Log(opts LogOptions) ([]CommitInfo, error) {
	// Use a custom format for parsing
	// Format: SHA\nAuthor\nAuthorEmail\nAuthorDate\nCommitter\nCommitterEmail\nCommitDate\nSubject\nParents\n---
	format := "%H%n%an%n%ae%n%at%n%cn%n%ce%n%ct%n%s%n%P%n---"

	args := []string{"log", fmt.Sprintf("--format=%s", format)}

	if opts.MaxCount > 0 {
		args = append(args, fmt.Sprintf("-n%d", opts.MaxCount))
	}
	if !opts.Since.IsZero() {
		args = append(args, fmt.Sprintf("--since=%s", opts.Since.Format(time.RFC3339)))
	}
	if !opts.Until.IsZero() {
		args = append(args, fmt.Sprintf("--until=%s", opts.Until.Format(time.RFC3339)))
	}
	if opts.Author != "" {
		args = append(args, fmt.Sprintf("--author=%s", opts.Author))
	}

	// Add -- separator before file paths
	if len(opts.FilePaths) > 0 {
		args = append(args, "--")
		args = append(args, opts.FilePaths...)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = r.RootPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	return parseLogOutput(string(output))
}

// parseLogOutput parses the custom format git log output.
func parseLogOutput(output string) ([]CommitInfo, error) {
	var commits []CommitInfo

	scanner := bufio.NewScanner(strings.NewReader(output))
	var lines []string

	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			if len(lines) >= 9 {
				commit := parseCommitLines(lines)
				commits = append(commits, commit)
			}
			lines = nil
		} else {
			lines = append(lines, line)
		}
	}

	// Handle last commit if no trailing ---
	if len(lines) >= 9 {
		commit := parseCommitLines(lines)
		commits = append(commits, commit)
	}

	return commits, scanner.Err()
}

// parseCommitLines parses commit lines in our custom format.
func parseCommitLines(lines []string) CommitInfo {
	commit := CommitInfo{
		SHA:            lines[0],
		Author:         lines[1],
		AuthorEmail:    lines[2],
		Committer:      lines[4],
		CommitterEmail: lines[5],
		Subject:        lines[7],
	}

	if len(commit.SHA) >= 7 {
		commit.SHAShort = commit.SHA[:7]
	}

	// Parse author timestamp
	if ts, err := strconv.ParseInt(lines[3], 10, 64); err == nil {
		commit.AuthorDate = time.Unix(ts, 0)
	}

	// Parse commit timestamp
	if ts, err := strconv.ParseInt(lines[6], 10, 64); err == nil {
		commit.CommitDate = time.Unix(ts, 0)
	}

	// Parse parents
	if lines[8] != "" {
		commit.Parents = strings.Fields(lines[8])
	}

	return commit
}

// GetCommit retrieves information about a specific commit.
func (r *Repository) GetCommit(sha string) (*CommitInfo, error) {
	format := "%H%n%an%n%ae%n%at%n%cn%n%ce%n%ct%n%s%n%P"

	cmd := exec.Command("git", "log", "-1", fmt.Sprintf("--format=%s", format), sha)
	cmd.Dir = r.RootPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git show failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 9 {
		return nil, fmt.Errorf("unexpected git log output")
	}

	commit := parseCommitLines(lines)
	return &commit, nil
}

// GetRecentCommits returns the most recent N commits affecting the given files.
func (r *Repository) GetRecentCommits(n int, filePaths ...string) ([]CommitInfo, error) {
	return r.Log(LogOptions{
		MaxCount:  n,
		FilePaths: filePaths,
	})
}

// GetLastCommitForFile returns the most recent commit that modified a file.
func (r *Repository) GetLastCommitForFile(filePath string) (*CommitInfo, error) {
	commits, err := r.GetRecentCommits(1, filePath)
	if err != nil {
		return nil, err
	}
	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits found for file: %s", filePath)
	}
	return &commits[0], nil
}

// GetCommitsBetween returns commits between two references (exclusive of base).
func (r *Repository) GetCommitsBetween(baseSHA, headSHA string) ([]CommitInfo, error) {
	format := "%H%n%an%n%ae%n%at%n%cn%n%ce%n%ct%n%s%n%P%n---"

	args := []string{"log", fmt.Sprintf("--format=%s", format), fmt.Sprintf("%s..%s", baseSHA, headSHA)}

	cmd := exec.Command("git", args...)
	cmd.Dir = r.RootPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	return parseLogOutput(string(output))
}

// GetFilesChangedInCommit returns the list of files modified in a commit.
func (r *Repository) GetFilesChangedInCommit(sha string) ([]string, error) {
	cmd := exec.Command("git", "diff-tree", "--no-commit-id", "--name-only", "-r", sha)
	cmd.Dir = r.RootPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff-tree failed: %w", err)
	}

	var files []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			files = append(files, line)
		}
	}

	return files, scanner.Err()
}
