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

package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ParameterSource specifies where parameter values come from.
type ParameterSource struct {
	// File is the path to a parameters file (*.bicepparam or *.json)
	File string

	// Values are inline parameter values (key=value pairs)
	Values map[string]any
}

// LoadParametersFile loads parameters from a .bicepparam or .json file.
func LoadParametersFile(filePath string) (map[string]any, error) {
	if filePath == "" {
		return nil, nil
	}

	// Check file exists
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("parameters file not found: %s", filePath)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".bicepparam":
		return loadBicepParamFile(filePath)
	case ".json":
		return loadJSONParamFile(filePath)
	default:
		return nil, fmt.Errorf("unsupported parameters file format: %s (expected .bicepparam or .json)", ext)
	}
}

// loadBicepParamFile loads a .bicepparam file.
// Note: .bicepparam files are processed by the Bicep CLI, not parsed directly.
// We just verify the file exists and return an indicator that it should be used.
func loadBicepParamFile(filePath string) (map[string]any, error) {
	// .bicepparam files are passed directly to bicep build --parameters
	// We don't parse them ourselves - the Bicep CLI handles them
	return map[string]any{
		"_bicepparamFile": filePath,
	}, nil
}

// loadJSONParamFile loads a JSON parameters file.
// Supports both ARM-style and simple key-value formats.
func loadJSONParamFile(filePath string) (map[string]any, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read parameters file: %w", err)
	}

	var rawParams map[string]any
	if err := json.Unmarshal(data, &rawParams); err != nil {
		return nil, fmt.Errorf("failed to parse parameters file: %w", err)
	}

	// Check if this is ARM-style parameters format
	// ARM format: { "parameters": { "name": { "value": "..." } } }
	if params, ok := rawParams["parameters"].(map[string]any); ok {
		result := make(map[string]any)
		for name, param := range params {
			if paramMap, ok := param.(map[string]any); ok {
				if value, ok := paramMap["value"]; ok {
					result[name] = value
				}
			}
		}
		return result, nil
	}

	// Otherwise, assume simple key-value format
	return rawParams, nil
}

// ParseParameterOverrides parses command-line parameter overrides in the format key=value.
func ParseParameterOverrides(args []string) (map[string]any, error) {
	result := make(map[string]any)

	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid parameter format: %s (expected key=value)", arg)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Try to parse as JSON value first (for numbers, booleans, objects)
		var parsed any
		if err := json.Unmarshal([]byte(value), &parsed); err == nil {
			result[key] = parsed
		} else {
			// Treat as string
			result[key] = value
		}
	}

	return result, nil
}

// MergeParameters merges parameter values from multiple sources.
// Later sources override earlier ones.
func MergeParameters(sources ...map[string]any) map[string]any {
	result := make(map[string]any)
	for _, source := range sources {
		for key, value := range source {
			result[key] = value
		}
	}
	return result
}

// ValidateRequiredParameters checks that all required parameters have values.
func ValidateRequiredParameters(required []string, provided map[string]any) error {
	var missing []string
	for _, name := range required {
		if _, ok := provided[name]; !ok {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required parameters: %s. Use --parameters to provide a parameters file", strings.Join(missing, ", "))
	}

	return nil
}

// FindParametersFile looks for a default parameters file for the given Bicep file.
// It checks for:
// 1. <bicepfile>.bicepparam (e.g., app.bicep → app.bicepparam)
// 2. <bicepfile>.parameters.json (e.g., app.bicep → app.parameters.json)
// Returns empty string if no default parameters file is found.
func FindParametersFile(bicepFile string) string {
	baseName := strings.TrimSuffix(bicepFile, filepath.Ext(bicepFile))

	// Check for .bicepparam
	bicepparamFile := baseName + ".bicepparam"
	if _, err := os.Stat(bicepparamFile); err == nil {
		return bicepparamFile
	}

	// Check for .parameters.json
	jsonParamFile := baseName + ".parameters.json"
	if _, err := os.Stat(jsonParamFile); err == nil {
		return jsonParamFile
	}

	return ""
}
