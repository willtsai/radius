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

// BlameInfo contains git blame information for a specific line.
type BlameInfo struct {
	// SHA is the full 40-character commit hash
	SHA string

	// SHAShort is the abbreviated 7-character commit hash
	SHAShort string

	// Author is the author's name
	Author string

	// AuthorEmail is the author's email address
	AuthorEmail string

	// AuthorTime is the original authoring timestamp
	AuthorTime time.Time

	// Summary is the commit message summary (first line)
	Summary string

	// LineNumber is the 1-based line number in the file
	LineNumber int

	// Original indicates if this is the original commit (not copied from another file)
	Original bool
}

// BlameResult contains the result of a git blame operation.
type BlameResult struct {
	// FilePath is the file that was blamed
	FilePath string

	// Lines maps line numbers to BlameInfo
	Lines map[int]*BlameInfo

	// HistoryUnavailable indicates the file's git history is not available
	// (e.g., due to shallow clone)
	HistoryUnavailable bool
}

// Blame executes git blame on a file and returns blame information.
// If lineRange is specified (start, end), only blames those lines.
func (r *Repository) Blame(filePath string, lineRange ...int) (*BlameResult, error) {
	args := []string{"blame", "--porcelain"}

	// Add line range if specified
	if len(lineRange) >= 2 {
		args = append(args, fmt.Sprintf("-L%d,%d", lineRange[0], lineRange[1]))
	} else if len(lineRange) == 1 {
		args = append(args, fmt.Sprintf("-L%d,%d", lineRange[0], lineRange[0]))
	}

	args = append(args, "--", filePath)

	cmd := exec.Command("git", args...)
	cmd.Dir = r.RootPath

	output, err := cmd.Output()
	if err != nil {
		// Check if it's a shallow clone issue
		if r.IsShallow {
			return &BlameResult{
				FilePath:           filePath,
				Lines:              make(map[int]*BlameInfo),
				HistoryUnavailable: true,
			}, nil
		}
		return nil, fmt.Errorf("git blame failed: %w", err)
	}

	return parseBlameOutput(filePath, string(output))
}

// BlameLines executes git blame for specific line numbers.
func (r *Repository) BlameLines(filePath string, lines []int) (*BlameResult, error) {
	if len(lines) == 0 {
		return &BlameResult{
			FilePath: filePath,
			Lines:    make(map[int]*BlameInfo),
		}, nil
	}

	result := &BlameResult{
		FilePath: filePath,
		Lines:    make(map[int]*BlameInfo),
	}

	// Group consecutive lines into ranges for efficiency
	ranges := groupIntoRanges(lines)

	for _, r2 := range ranges {
		blameResult, err := r.Blame(filePath, r2[0], r2[1])
		if err != nil {
			return nil, err
		}
		if blameResult.HistoryUnavailable {
			result.HistoryUnavailable = true
			return result, nil
		}
		for lineNum, info := range blameResult.Lines {
			result.Lines[lineNum] = info
		}
	}

	return result, nil
}

// parseBlameOutput parses git blame --porcelain output.
func parseBlameOutput(filePath string, output string) (*BlameResult, error) {
	result := &BlameResult{
		FilePath: filePath,
		Lines:    make(map[int]*BlameInfo),
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	var currentInfo *BlameInfo
	var currentLine int

	for scanner.Scan() {
		line := scanner.Text()

		// SHA line: <sha> <original-line> <final-line> [group-count]
		if len(line) >= 40 && isHexString(line[:40]) {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				currentInfo = &BlameInfo{
					SHA:      parts[0],
					SHAShort: parts[0][:7],
				}
				currentLine, _ = strconv.Atoi(parts[2])
				currentInfo.LineNumber = currentLine
			}
			continue
		}

		if currentInfo == nil {
			continue
		}

		// Parse header fields
		switch {
		case strings.HasPrefix(line, "author "):
			currentInfo.Author = strings.TrimPrefix(line, "author ")
		case strings.HasPrefix(line, "author-mail "):
			email := strings.TrimPrefix(line, "author-mail ")
			currentInfo.AuthorEmail = strings.Trim(email, "<>")
		case strings.HasPrefix(line, "author-time "):
			ts := strings.TrimPrefix(line, "author-time ")
			if unixTime, err := strconv.ParseInt(ts, 10, 64); err == nil {
				currentInfo.AuthorTime = time.Unix(unixTime, 0)
			}
		case strings.HasPrefix(line, "summary "):
			currentInfo.Summary = strings.TrimPrefix(line, "summary ")
		case strings.HasPrefix(line, "boundary"):
			// Indicates the commit is a boundary (root commit or not reachable)
		case strings.HasPrefix(line, "\t"):
			// This is the actual line content, marks end of header
			if currentInfo != nil && currentLine > 0 {
				result.Lines[currentLine] = currentInfo
			}
			currentInfo = nil
		}
	}

	return result, scanner.Err()
}

// groupIntoRanges groups line numbers into consecutive ranges.
// Returns a slice of [start, end] pairs.
func groupIntoRanges(lines []int) [][2]int {
	if len(lines) == 0 {
		return nil
	}

	// Sort lines
	sorted := make([]int, len(lines))
	copy(sorted, lines)
	sortInts(sorted)

	var ranges [][2]int
	start := sorted[0]
	end := sorted[0]

	for i := 1; i < len(sorted); i++ {
		if sorted[i] == end+1 {
			end = sorted[i]
		} else {
			ranges = append(ranges, [2]int{start, end})
			start = sorted[i]
			end = sorted[i]
		}
	}
	ranges = append(ranges, [2]int{start, end})

	return ranges
}

// sortInts sorts a slice of ints in ascending order (simple bubble sort for small slices).
func sortInts(a []int) {
	for i := 0; i < len(a); i++ {
		for j := i + 1; j < len(a); j++ {
			if a[j] < a[i] {
				a[i], a[j] = a[j], a[i]
			}
		}
	}
}

// isHexString checks if a string contains only hex characters.
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// GetBlameInfoForLine gets blame information for a specific line in a file.
func (r *Repository) GetBlameInfoForLine(filePath string, line int) (*BlameInfo, error) {
	result, err := r.Blame(filePath, line, line)
	if err != nil {
		return nil, err
	}
	if result.HistoryUnavailable {
		return nil, fmt.Errorf("git history unavailable (shallow clone)")
	}
	return result.Lines[line], nil
}
