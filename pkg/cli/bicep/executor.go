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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/radius-project/radius/pkg/cli/filesystem"
	"github.com/radius-project/radius/pkg/cli/output"
	"github.com/radius-project/radius/pkg/version"
)

// Executor provides a simple interface for Bicep CLI operations.
// Use NewExecutor() to create an instance.
type Executor struct{}

// NewExecutor creates a new Executor for Bicep CLI operations.
func NewExecutor() *Executor {
	return &Executor{}
}

// Build compiles a Bicep file and returns the ARM JSON as a string.
func (e *Executor) Build(filePath string) (string, error) {
	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(filePath), ".bicep") {
		return "", fmt.Errorf("the provided file %q must be a .bicep file", filePath)
	}

	// Ensure Bicep CLI is available
	ok, err := IsBicepInstalled()
	if err != nil {
		return "", fmt.Errorf("failed to check bicep installation: %w", err)
	}

	if !ok {
		if err := DownloadBicep(); err != nil {
			return "", fmt.Errorf("failed to download bicep: %w", err)
		}
	}

	// Check the file exists
	if _, err := os.Stat(filePath); err != nil {
		return "", fmt.Errorf("could not find file %q: %w", filePath, err)
	}

	// Compile to ARM JSON
	bytes, err := runBicepRaw("build", "--stdout", filePath)
	if err != nil {
		return "", fmt.Errorf("failed to compile bicep: %w", err)
	}

	return string(bytes), nil
}

// GetBicepCliPath returns the path to the Bicep CLI executable.
// If the CLI is not installed, it returns an error.
func GetBicepCliPath() (string, error) {
	ok, err := IsBicepInstalled()
	if err != nil {
		return "", fmt.Errorf("failed to check bicep installation: %w", err)
	}
	if !ok {
		return "", fmt.Errorf("bicep CLI is not installed")
	}

	// Get the home directory location where bicep is installed
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	bicepPath := filepath.Join(homeDir, ".rad", "bin", "rad-bicep")
	return bicepPath, nil
}

// StaticGraphExecutor provides methods for compiling Bicep files to ARM JSON
// for static graph generation. Unlike the standard Impl, this executor is
// optimized for extracting resource definitions without deployment.
type StaticGraphExecutor struct {
	FileSystem filesystem.FileSystem
	Output     output.Interface
}

// NewStaticGraphExecutor creates a new StaticGraphExecutor.
func NewStaticGraphExecutor(fs filesystem.FileSystem, out output.Interface) *StaticGraphExecutor {
	return &StaticGraphExecutor{
		FileSystem: fs,
		Output:     out,
	}
}

// CompileToARMJSON compiles a Bicep file to ARM JSON template.
// It returns the parsed ARM JSON as a map.
//
// The function handles:
// - Bicep CLI installation check and auto-download
// - File existence validation
// - Bicep compilation with --stdout
// - JSON parsing of the output
func (e *StaticGraphExecutor) CompileToARMJSON(filePath string) (map[string]any, error) {
	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(filePath), ".bicep") {
		return nil, fmt.Errorf("the provided file %q must be a .bicep file", filePath)
	}

	// Ensure Bicep CLI is available
	ok, err := IsBicepInstalled()
	if err != nil {
		return nil, fmt.Errorf("failed to check bicep installation: %w", err)
	}

	if !ok {
		e.Output.LogInfo("Downloading Bicep for channel %s...", version.Channel())
		if err := DownloadBicep(); err != nil {
			return nil, fmt.Errorf("failed to download bicep: %w", err)
		}
	}

	// Check the file exists
	if _, err := e.FileSystem.Stat(filePath); err != nil {
		return nil, fmt.Errorf("could not find file %q: %w", filePath, err)
	}

	// Compile to ARM JSON
	step := e.Output.BeginStep("Compiling %s to ARM JSON...", filePath)
	bytes, err := runBicepRaw("build", "--stdout", filePath)
	if err != nil {
		e.Output.CompleteStep(step)
		return nil, fmt.Errorf("failed to compile bicep: %w", err)
	}
	e.Output.CompleteStep(step)

	// Parse the JSON output
	template := make(map[string]any)
	if err := json.Unmarshal(bytes, &template); err != nil {
		return nil, fmt.Errorf("failed to parse ARM JSON output: %w", err)
	}

	return template, nil
}

// CompileToARMJSONWithParameters compiles a Bicep file with a parameters file.
// Parameters are merged into the compilation to resolve parameterized values.
func (e *StaticGraphExecutor) CompileToARMJSONWithParameters(filePath, paramsFilePath string) (map[string]any, error) {
	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(filePath), ".bicep") {
		return nil, fmt.Errorf("the provided file %q must be a .bicep file", filePath)
	}

	// Ensure Bicep CLI is available
	ok, err := IsBicepInstalled()
	if err != nil {
		return nil, fmt.Errorf("failed to check bicep installation: %w", err)
	}

	if !ok {
		e.Output.LogInfo("Downloading Bicep for channel %s...", version.Channel())
		if err := DownloadBicep(); err != nil {
			return nil, fmt.Errorf("failed to download bicep: %w", err)
		}
	}

	// Check the files exist
	if _, err := e.FileSystem.Stat(filePath); err != nil {
		return nil, fmt.Errorf("could not find bicep file %q: %w", filePath, err)
	}
	if _, err := e.FileSystem.Stat(paramsFilePath); err != nil {
		return nil, fmt.Errorf("could not find parameters file %q: %w", paramsFilePath, err)
	}

	// Compile with parameters
	step := e.Output.BeginStep("Compiling %s with parameters...", filePath)
	bytes, err := runBicepRaw("build", "--stdout", filePath, "--parameters", paramsFilePath)
	if err != nil {
		e.Output.CompleteStep(step)
		return nil, fmt.Errorf("failed to compile bicep with parameters: %w", err)
	}
	e.Output.CompleteStep(step)

	// Parse the JSON output
	template := make(map[string]any)
	if err := json.Unmarshal(bytes, &template); err != nil {
		return nil, fmt.Errorf("failed to parse ARM JSON output: %w", err)
	}

	return template, nil
}

// GetRequiredParameters analyzes a Bicep file and returns the list of parameters
// that are required (have no default value).
func (e *StaticGraphExecutor) GetRequiredParameters(filePath string) ([]string, error) {
	template, err := e.CompileToARMJSON(filePath)
	if err != nil {
		return nil, err
	}

	params, ok := template["parameters"].(map[string]any)
	if !ok {
		return nil, nil // No parameters
	}

	var required []string
	for name, paramDef := range params {
		paramMap, ok := paramDef.(map[string]any)
		if !ok {
			continue
		}

		// A parameter is required if it has no "defaultValue"
		if _, hasDefault := paramMap["defaultValue"]; !hasDefault {
			required = append(required, name)
		}
	}

	return required, nil
}
