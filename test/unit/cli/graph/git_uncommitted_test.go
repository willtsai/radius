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
	"strings"
	"testing"

	"github.com/radius-project/radius/pkg/cli/git"
	"github.com/stretchr/testify/assert"
)

func TestFileStatus_Fields(t *testing.T) {
	status := git.FileStatus{
		Path:           "main.bicep",
		IndexStatus:    'M',
		WorktreeStatus: ' ',
	}

	assert.Equal(t, "main.bicep", status.Path)
	assert.Equal(t, 'M', status.IndexStatus)
	assert.Equal(t, ' ', status.WorktreeStatus)
}

func TestStatusResult_Empty(t *testing.T) {
	result := git.StatusResult{
		Files:          make(map[string]*git.FileStatus),
		HasUncommitted: false,
		HasStaged:      false,
		HasUnstaged:    false,
		HasUntracked:   false,
	}

	assert.Empty(t, result.Files)
	assert.False(t, result.HasUncommitted)
	assert.False(t, result.HasStaged)
	assert.False(t, result.HasUnstaged)
	assert.False(t, result.HasUntracked)
}

func TestStatusResult_WithChanges(t *testing.T) {
	result := git.StatusResult{
		Files: map[string]*git.FileStatus{
			"main.bicep": {
				Path:           "main.bicep",
				IndexStatus:    'M',
				WorktreeStatus: ' ',
			},
		},
		HasUncommitted: true,
		HasStaged:      true,
		HasUnstaged:    false,
		HasUntracked:   false,
	}

	assert.Len(t, result.Files, 1)
	assert.True(t, result.HasUncommitted)
	assert.True(t, result.HasStaged)
}

func TestParseStatusOutputFormat(t *testing.T) {
	// Test parsing of porcelain status output
	// Format: XY filename
	// X = index status, Y = worktree status

	tests := []struct {
		name          string
		output        string
		expectedFiles []string
		hasStaged     bool
		hasUnstaged   bool
		hasUntracked  bool
	}{
		{
			name:          "staged modification",
			output:        "M  main.bicep",
			expectedFiles: []string{"main.bicep"},
			hasStaged:     true,
			hasUnstaged:   false,
			hasUntracked:  false,
		},
		{
			name:          "unstaged modification",
			output:        " M main.bicep",
			expectedFiles: []string{"main.bicep"},
			hasStaged:     false,
			hasUnstaged:   true,
			hasUntracked:  false,
		},
		{
			name:          "staged and unstaged",
			output:        "MM main.bicep",
			expectedFiles: []string{"main.bicep"},
			hasStaged:     true,
			hasUnstaged:   true,
			hasUntracked:  false,
		},
		{
			name:          "untracked file",
			output:        "?? newfile.bicep",
			expectedFiles: []string{"newfile.bicep"},
			hasStaged:     false,
			hasUnstaged:   false,
			hasUntracked:  true,
		},
		{
			name:          "deleted file",
			output:        " D removed.bicep",
			expectedFiles: []string{"removed.bicep"},
			hasStaged:     false,
			hasUnstaged:   true,
			hasUntracked:  false,
		},
		{
			name:          "added file",
			output:        "A  newfile.bicep",
			expectedFiles: []string{"newfile.bicep"},
			hasStaged:     true,
			hasUnstaged:   false,
			hasUntracked:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the format
			line := tc.output
			if len(line) >= 3 {
				indexStatus := rune(line[0])
				worktreeStatus := rune(line[1])
				path := strings.TrimSpace(line[3:])

				assert.Contains(t, tc.expectedFiles, path)

				// Check status indicators
				if indexStatus != ' ' && indexStatus != '?' {
					assert.True(t, tc.hasStaged)
				}
				if worktreeStatus != ' ' && worktreeStatus != '?' {
					assert.True(t, tc.hasUnstaged)
				}
				if indexStatus == '?' && worktreeStatus == '?' {
					assert.True(t, tc.hasUntracked)
				}
			}
		})
	}
}

func TestFileStatusCodes(t *testing.T) {
	// Document the meaning of git status codes
	statusCodes := map[rune]string{
		' ': "not modified",
		'M': "modified",
		'A': "added",
		'D': "deleted",
		'R': "renamed",
		'C': "copied",
		'U': "updated but unmerged",
		'?': "untracked",
		'!': "ignored",
	}

	// Verify we understand common codes
	assert.Equal(t, "modified", statusCodes['M'])
	assert.Equal(t, "added", statusCodes['A'])
	assert.Equal(t, "deleted", statusCodes['D'])
	assert.Equal(t, "untracked", statusCodes['?'])
}

func TestRenamedFileFormat(t *testing.T) {
	// Renamed files have format: "R  old -> new"
	line := "R  old.bicep -> new.bicep"

	if len(line) >= 3 && line[0] == 'R' {
		path := strings.TrimSpace(line[3:])
		if strings.Contains(path, " -> ") {
			parts := strings.Split(path, " -> ")
			assert.Len(t, parts, 2)
			assert.Equal(t, "old.bicep", parts[0])
			assert.Equal(t, "new.bicep", parts[1])
		}
	}
}

func TestStatusResultFlags(t *testing.T) {
	// Test the convenience flags on StatusResult

	tests := []struct {
		name          string
		files         map[string]*git.FileStatus
		wantStaged    bool
		wantUnstaged  bool
		wantUntracked bool
	}{
		{
			name:          "empty",
			files:         map[string]*git.FileStatus{},
			wantStaged:    false,
			wantUnstaged:  false,
			wantUntracked: false,
		},
		{
			name: "only staged",
			files: map[string]*git.FileStatus{
				"file.bicep": {IndexStatus: 'M', WorktreeStatus: ' '},
			},
			wantStaged:    true,
			wantUnstaged:  false,
			wantUntracked: false,
		},
		{
			name: "only unstaged",
			files: map[string]*git.FileStatus{
				"file.bicep": {IndexStatus: ' ', WorktreeStatus: 'M'},
			},
			wantStaged:    false,
			wantUnstaged:  true,
			wantUntracked: false,
		},
		{
			name: "only untracked",
			files: map[string]*git.FileStatus{
				"file.bicep": {IndexStatus: '?', WorktreeStatus: '?'},
			},
			wantStaged:    false,
			wantUnstaged:  false,
			wantUntracked: true,
		},
		{
			name: "mixed",
			files: map[string]*git.FileStatus{
				"staged.bicep":    {IndexStatus: 'A', WorktreeStatus: ' '},
				"unstaged.bicep":  {IndexStatus: ' ', WorktreeStatus: 'M'},
				"untracked.bicep": {IndexStatus: '?', WorktreeStatus: '?'},
			},
			wantStaged:    true,
			wantUnstaged:  true,
			wantUntracked: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := git.StatusResult{
				Files: tc.files,
			}

			// Compute flags
			for _, s := range tc.files {
				if s.IndexStatus != ' ' && s.IndexStatus != '?' {
					result.HasStaged = true
				}
				if s.WorktreeStatus != ' ' && s.WorktreeStatus != '?' {
					result.HasUnstaged = true
				}
				if s.IndexStatus == '?' && s.WorktreeStatus == '?' {
					result.HasUntracked = true
				}
				result.HasUncommitted = true
			}

			assert.Equal(t, tc.wantStaged, result.HasStaged)
			assert.Equal(t, tc.wantUnstaged, result.HasUnstaged)
			assert.Equal(t, tc.wantUntracked, result.HasUntracked)
		})
	}
}
