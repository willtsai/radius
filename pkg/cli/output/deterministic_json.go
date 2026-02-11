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

package output

import (
	"bytes"
	"encoding/json"
	"sort"
)

// DeterministicJSON is an alias for MarshalDeterministicJSON.
// This provides a shorter function name for common usage.
func DeterministicJSON(v any) ([]byte, error) {
	return MarshalDeterministicJSON(v)
}

// MarshalDeterministicJSON serializes the given value to JSON with deterministic key ordering.
// This ensures that identical inputs produce byte-identical outputs, which is essential
// for version control diffs and staleness detection.
//
// The function sorts all map keys alphabetically at every level of the JSON structure.
func MarshalDeterministicJSON(v any) ([]byte, error) {
	// First, marshal to get the standard JSON
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	// Unmarshal into a generic structure that we can sort
	var generic any
	if err := json.Unmarshal(data, &generic); err != nil {
		return nil, err
	}

	// Sort and re-marshal
	sorted := sortValue(generic)

	// Marshal with indentation for readability
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(sorted); err != nil {
		return nil, err
	}

	// Remove trailing newline added by Encode
	result := buf.Bytes()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	return result, nil
}

// sortValue recursively sorts maps by their keys and processes arrays.
func sortValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		return sortMap(val)
	case []any:
		return sortArray(val)
	default:
		return v
	}
}

// sortMap creates a new map with sorted keys and recursively sorted values.
// We use a custom type to ensure consistent ordering during JSON marshaling.
func sortMap(m map[string]any) *orderedMap {
	if m == nil {
		return nil
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	om := &orderedMap{
		keys:   keys,
		values: make(map[string]any, len(m)),
	}

	for _, k := range keys {
		om.values[k] = sortValue(m[k])
	}

	return om
}

// sortArray processes each element of the array recursively.
func sortArray(arr []any) []any {
	if arr == nil {
		return nil
	}

	result := make([]any, len(arr))
	for i, v := range arr {
		result[i] = sortValue(v)
	}
	return result
}

// orderedMap is a map that maintains key ordering for deterministic JSON output.
type orderedMap struct {
	keys   []string
	values map[string]any
}

// MarshalJSON implements json.Marshaler for orderedMap.
func (om *orderedMap) MarshalJSON() ([]byte, error) {
	if om == nil {
		return []byte("null"), nil
	}

	var buf bytes.Buffer
	buf.WriteByte('{')

	for i, k := range om.keys {
		if i > 0 {
			buf.WriteByte(',')
		}

		// Write the key
		keyBytes, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')

		// Write the value
		valBytes, err := json.Marshal(om.values[k])
		if err != nil {
			return nil, err
		}
		buf.Write(valBytes)
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}
