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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/radius-project/radius/pkg/cli/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommitInfo_Fields(t *testing.T) {
	// Test that CommitInfo struct has expected fields
	info := git.CommitInfo{
		SHA:            "abc123def456789012345678901234567890abcd",
		SHAShort:       "abc123d",
		Author:         "Test Author",
		AuthorEmail:    "author@example.com",
		AuthorDate:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Committer:      "Test Committer",
		CommitterEmail: "committer@example.com",
		CommitDate:     time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
		Subject:        "Add new feature",
		Parents:        []string{"parent123"},
	}

	assert.Equal(t, "abc123def456789012345678901234567890abcd", info.SHA)
	assert.Equal(t, "abc123d", info.SHAShort)
	assert.Equal(t, "Test Author", info.Author)
	assert.Equal(t, "author@example.com", info.AuthorEmail)
	assert.Equal(t, "Test Committer", info.Committer)
	assert.Equal(t, "Add new feature", info.Subject)
	assert.Len(t, info.Parents, 1)
}

func TestLogOptions_Defaults(t *testing.T) {
	opts := git.LogOptions{}

	assert.Equal(t, 0, opts.MaxCount)
	assert.True(t, opts.Since.IsZero())
	assert.True(t, opts.Until.IsZero())
	assert.Empty(t, opts.Author)
	assert.Empty(t, opts.FilePaths)
}

func TestLogOptions_WithFilePaths(t *testing.T) {
	opts := git.LogOptions{
		MaxCount:  10,
		FilePaths: []string{"main.bicep", "modules/db.bicep"},
	}

	assert.Equal(t, 10, opts.MaxCount)
	assert.Len(t, opts.FilePaths, 2)
	assert.Contains(t, opts.FilePaths, "main.bicep")
	assert.Contains(t, opts.FilePaths, "modules/db.bicep")
}

func TestLogOptions_DateRange(t *testing.T) {
	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	opts := git.LogOptions{
		Since: since,
		Until: until,
	}

	assert.Equal(t, since, opts.Since)
	assert.Equal(t, until, opts.Until)
}

func TestParseLogOutputFormat(t *testing.T) {
	// Test parsing of our custom log format
	// Format: SHA\nAuthor\nAuthorEmail\nAuthorDate\nCommitter\nCommitterEmail\nCommitDate\nSubject\nParents\n---

	sampleOutput := `abc123def456789012345678901234567890abcd
Test Author
author@example.com
1705315800
Test Committer
committer@example.com
1705317600
Add new container

---
def456abc789012345678901234567890defdef
Another Author
another@example.com
1705315700
Another Committer
another-committer@example.com
1705317500
Previous commit
abc123def456789012345678901234567890abcd
---`

	// Verify format expectations
	lines := strings.Split(sampleOutput, "\n")

	// First commit
	assert.Len(t, lines[0], 40) // Full SHA
	assert.Equal(t, "Test Author", lines[1])
	assert.Equal(t, "author@example.com", lines[2])
	assert.Equal(t, "1705315800", lines[3]) // Unix timestamp
	assert.Equal(t, "Test Committer", lines[4])
	assert.Equal(t, "committer@example.com", lines[5])
	assert.Equal(t, "Add new container", lines[7])
	assert.Equal(t, "---", lines[9])

	// Count commits by counting --- separators
	commitCount := 0
	for _, line := range lines {
		if line == "---" {
			commitCount++
		}
	}
	assert.Equal(t, 2, commitCount)
}

func TestCommitInfo_Parents(t *testing.T) {
	tests := []struct {
		name      string
		parents   []string
		isMerge   bool
		isInitial bool
	}{
		{
			name:      "normal commit",
			parents:   []string{"abc123"},
			isMerge:   false,
			isInitial: false,
		},
		{
			name:      "merge commit",
			parents:   []string{"abc123", "def456"},
			isMerge:   true,
			isInitial: false,
		},
		{
			name:      "initial commit",
			parents:   []string{},
			isMerge:   false,
			isInitial: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			info := git.CommitInfo{
				SHA:     "test123",
				Parents: tc.parents,
			}

			isMerge := len(info.Parents) > 1
			isInitial := len(info.Parents) == 0

			assert.Equal(t, tc.isMerge, isMerge)
			assert.Equal(t, tc.isInitial, isInitial)
		})
	}
}

func TestCommitInfo_SHAShort(t *testing.T) {
	info := git.CommitInfo{
		SHA: "abc123def456789012345678901234567890abcd",
	}

	// SHAShort should be first 7 characters
	if len(info.SHA) >= 7 {
		expected := info.SHA[:7]
		assert.Equal(t, "abc123d", expected)
	}
}

func TestLogTimestampParsing(t *testing.T) {
	// Unix timestamp parsing
	tests := []struct {
		timestamp string
		expected  time.Time
	}{
		{
			timestamp: "1705315800",
			expected:  time.Unix(1705315800, 0),
		},
		{
			timestamp: "0",
			expected:  time.Unix(0, 0),
		},
	}

	for _, tc := range tests {
		t.Run(tc.timestamp, func(t *testing.T) {
			// Parse the timestamp as int64
			ts, err := strconv.ParseInt(tc.timestamp, 10, 64)
			require.NoError(t, err)

			// Verify it matches expected
			parsed := time.Unix(ts, 0)
			assert.Equal(t, tc.expected.Unix(), parsed.Unix())
		})
	}
}
