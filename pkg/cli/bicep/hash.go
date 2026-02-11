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

package bicep

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sort"
)

// ComputeSourceHash computes a SHA256 hash of the given source files.
// The hash is deterministic - the same set of files with the same contents
// will always produce the same hash, regardless of file order.
//
// The returned hash is formatted as "sha256:<64 hex characters>".
func ComputeSourceHash(filePaths []string) (string, error) {
	if len(filePaths) == 0 {
		return "", fmt.Errorf("no source files provided")
	}

	// Sort file paths for deterministic ordering
	sortedPaths := make([]string, len(filePaths))
	copy(sortedPaths, filePaths)
	sort.Strings(sortedPaths)

	// Create a hasher
	h := sha256.New()

	// Hash each file's path and contents
	for _, path := range sortedPaths {
		// Include the file path in the hash for disambiguation
		h.Write([]byte(path))
		h.Write([]byte{0}) // Null separator

		// Read and hash file contents
		content, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read file %q: %w", path, err)
		}
		h.Write(content)
		h.Write([]byte{0}) // Null separator between files
	}

	// Format the hash
	hashBytes := h.Sum(nil)
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(hashBytes)), nil
}

// ComputeSourceHashFromReader computes a SHA256 hash from readers.
// This is useful for testing or when files are already open.
func ComputeSourceHashFromReader(readers map[string]io.Reader) (string, error) {
	if len(readers) == 0 {
		return "", fmt.Errorf("no source readers provided")
	}

	// Get sorted keys for deterministic ordering
	keys := make([]string, 0, len(readers))
	for k := range readers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create a hasher
	h := sha256.New()

	// Hash each reader's path and contents
	for _, path := range keys {
		reader := readers[path]

		// Include the path in the hash
		h.Write([]byte(path))
		h.Write([]byte{0})

		// Read and hash contents
		if _, err := io.Copy(h, reader); err != nil {
			return "", fmt.Errorf("failed to read from %q: %w", path, err)
		}
		h.Write([]byte{0})
	}

	hashBytes := h.Sum(nil)
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(hashBytes)), nil
}

// ValidateSourceHash checks if a hash string is valid.
// A valid hash is in the format "sha256:<64 hex characters>".
func ValidateSourceHash(hash string) bool {
	if len(hash) != len("sha256:")+64 {
		return false
	}
	if hash[:7] != "sha256:" {
		return false
	}

	// Check that the rest is valid hex
	_, err := hex.DecodeString(hash[7:])
	return err == nil
}

// CompareSourceHash compares two hashes and returns true if they match.
func CompareSourceHash(hash1, hash2 string) bool {
	return hash1 == hash2
}
