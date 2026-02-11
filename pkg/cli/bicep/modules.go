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
)

// ModuleResolver resolves Bicep module references and collects all source files.
type ModuleResolver struct {
	// BaseDir is the directory containing the main Bicep file
	BaseDir string

	// ResolvedFiles is the list of all resolved Bicep files (including the main file)
	ResolvedFiles []string

	// visited tracks files we've already processed to avoid cycles
	visited map[string]bool
}

// NewModuleResolver creates a new ModuleResolver for resolving Bicep modules.
func NewModuleResolver(mainFile string) *ModuleResolver {
	return &ModuleResolver{
		BaseDir:       filepath.Dir(mainFile),
		ResolvedFiles: []string{mainFile},
		visited:       map[string]bool{filepath.Clean(mainFile): true},
	}
}

// ResolveModulesFromARM extracts module references from an ARM template
// and resolves them to file paths.
//
// In ARM JSON, Bicep modules are represented as Microsoft.Resources/deployments
// resources. This function identifies these and extracts the original Bicep
// module paths when possible.
func (r *ModuleResolver) ResolveModulesFromARM(template *ARMTemplate) ([]string, error) {
	var moduleFiles []string

	for _, resource := range template.Resources {
		if !IsModuleDeployment(resource.Type) {
			continue
		}

		// Try to extract module path from metadata or comments
		// Note: The actual Bicep source path is often lost during compilation
		// We can only detect that modules exist, not their original paths
		if modulePath := r.extractModulePath(resource); modulePath != "" {
			resolved := r.resolvePath(modulePath)
			if !r.visited[resolved] {
				r.visited[resolved] = true
				r.ResolvedFiles = append(r.ResolvedFiles, resolved)
				moduleFiles = append(moduleFiles, resolved)
			}
		}
	}

	return moduleFiles, nil
}

// extractModulePath attempts to extract the original Bicep module path from
// a deployment resource. This is best-effort as the path is often not preserved.
func (r *ModuleResolver) extractModulePath(resource ARMResource) string {
	// Check for module path in comments (some Bicep versions preserve this)
	if resource.Comments != "" {
		if strings.Contains(resource.Comments, ".bicep") {
			// Try to extract the path
			// Format might be: "Module: ./modules/foo.bicep"
			parts := strings.Split(resource.Comments, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[len(parts)-1])
			}
		}
	}

	// Check for module metadata in properties
	if props := resource.Properties; props != nil {
		if meta, ok := props["metadata"].(map[string]any); ok {
			if modulePath, ok := meta["_generator"].(map[string]any); ok {
				if name, ok := modulePath["templateHash"].(string); ok {
					_ = name // Template hash doesn't give us the path
				}
			}
		}
	}

	return ""
}

// resolvePath resolves a module path relative to the base directory.
func (r *ModuleResolver) resolvePath(modulePath string) string {
	if filepath.IsAbs(modulePath) {
		return filepath.Clean(modulePath)
	}
	return filepath.Clean(filepath.Join(r.BaseDir, modulePath))
}

// GetAllSourceFiles returns all resolved source files including transitively
// imported modules.
func (r *ModuleResolver) GetAllSourceFiles() []string {
	return r.ResolvedFiles
}

// DiscoverLocalModules scans the ARM template and discovers any locally
// referenced module files. This is useful for computing source hashes.
//
// Note: This only works for modules that exist as separate files on disk.
// Inline modules or registry modules cannot be discovered this way.
func DiscoverLocalModules(mainFile string, template *ARMTemplate) ([]string, error) {
	resolver := NewModuleResolver(mainFile)
	_, err := resolver.ResolveModulesFromARM(template)
	if err != nil {
		return nil, err
	}
	return resolver.GetAllSourceFiles(), nil
}

// LoadBicepConfig loads the Bicep configuration from bicepconfig.json
// if it exists in the same directory as the Bicep file or any parent.
func LoadBicepConfig(bicepFile string) (*BicepConfig, error) {
	dir := filepath.Dir(bicepFile)

	for {
		configPath := filepath.Join(dir, "bicepconfig.json")
		if _, err := os.Stat(configPath); err == nil {
			return loadBicepConfigFile(configPath)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent
	}

	// No config found, return default
	return &BicepConfig{}, nil
}

// BicepConfig represents the bicepconfig.json configuration.
type BicepConfig struct {
	// ModuleAliases defines shortcuts for module paths
	ModuleAliases map[string]ModuleAlias `json:"moduleAliases,omitempty"`

	// Analyzers configures linting rules
	Analyzers map[string]any `json:"analyzers,omitempty"`

	// CacheRootDirectory specifies where to cache external modules
	CacheRootDirectory string `json:"cacheRootDirectory,omitempty"`
}

// ModuleAlias defines a module path alias.
type ModuleAlias struct {
	Registry   string `json:"registry,omitempty"`
	ModulePath string `json:"modulePath,omitempty"`
}

// loadBicepConfigFile loads and parses a bicepconfig.json file.
func loadBicepConfigFile(path string) (*BicepConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read bicepconfig.json: %w", err)
	}

	var config BicepConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse bicepconfig.json: %w", err)
	}

	return &config, nil
}
