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
	"fmt"
	"time"

	"github.com/radius-project/radius/pkg/cli/bicep"
	"github.com/radius-project/radius/pkg/cli/filesystem"
	"github.com/radius-project/radius/pkg/cli/output"
	"github.com/radius-project/radius/pkg/corerp/api/v20231001preview"
	"github.com/radius-project/radius/pkg/version"
)

// StaticGraphGenerator generates application graphs from Bicep files
// without requiring deployment to a Radius environment.
type StaticGraphGenerator struct {
	FileSystem    filesystem.FileSystem
	Output        output.Interface
	ResourceGroup string

	// Options
	IncludeGitMetadata bool
	ParametersFile     string
}

// NewStaticGraphGenerator creates a new StaticGraphGenerator.
func NewStaticGraphGenerator(fs filesystem.FileSystem, out output.Interface) *StaticGraphGenerator {
	return &StaticGraphGenerator{
		FileSystem:         fs,
		Output:             out,
		ResourceGroup:      "default",
		IncludeGitMetadata: true,
	}
}

// GenerateOptions configures the graph generation process.
type GenerateOptions struct {
	// ParametersFile is the path to a Bicep parameters file (.bicepparam or .json)
	ParametersFile string

	// ResourceGroup is the resource group name to use in resource IDs
	ResourceGroup string

	// IncludeGitMetadata enables git metadata enrichment (default: true)
	IncludeGitMetadata bool
}

// Generate creates a StaticAppGraph from one or more Bicep files.
//
// The generation process:
// 1. Compile Bicep file(s) to ARM JSON using bicep build --stdout
// 2. Parse the ARM JSON to extract resources and their properties
// 3. Detect connections from resource properties (connections, routes, dependsOn)
// 4. Build the graph structure with all resources and connections
// 5. Optionally enrich with git metadata (if in a git repo and --no-git not specified)
//
// Parameters:
//   - bicepFiles: One or more Bicep file paths to include in the graph
//   - opts: Generation options (parameters file, resource group, etc.)
//
// Returns the complete StaticAppGraph or an error if generation fails.
func (g *StaticGraphGenerator) Generate(bicepFiles []string, opts GenerateOptions) (*v20231001preview.StaticAppGraph, error) {
	if len(bicepFiles) == 0 {
		return nil, fmt.Errorf("at least one Bicep file is required")
	}

	// Apply options
	if opts.ResourceGroup != "" {
		g.ResourceGroup = opts.ResourceGroup
	}
	g.IncludeGitMetadata = opts.IncludeGitMetadata

	// Create executor for Bicep compilation
	executor := bicep.NewStaticGraphExecutor(g.FileSystem, g.Output)

	// Compile Bicep files to ARM JSON
	g.Output.LogInfo("Compiling Bicep files...")

	var allResources []v20231001preview.StaticAppGraphResource
	var allConnections []v20231001preview.StaticAppGraphConnection
	var compiledFiles []string

	for _, bicepFile := range bicepFiles {
		var template map[string]any
		var err error

		if opts.ParametersFile != "" {
			template, err = executor.CompileToARMJSONWithParameters(bicepFile, opts.ParametersFile)
		} else {
			template, err = executor.CompileToARMJSON(bicepFile)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to compile %s: %w", bicepFile, err)
		}

		// Parse the ARM template
		armTemplate, err := bicep.ParseARMTemplate(template)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ARM template from %s: %w", bicepFile, err)
		}

		// Validate required parameters if no parameters file provided
		if opts.ParametersFile == "" {
			required := armTemplate.GetRequiredParameterNames()
			if len(required) > 0 {
				return nil, fmt.Errorf("required parameters missing. Use --parameters to provide values for: %v", required)
			}
		}

		// Extract resources
		extractor := bicep.NewResourceExtractor(g.ResourceGroup)
		resources, err := extractor.ExtractResources(armTemplate, bicepFile)
		if err != nil {
			return nil, fmt.Errorf("failed to extract resources from %s: %w", bicepFile, err)
		}

		allResources = append(allResources, resources...)
		compiledFiles = append(compiledFiles, bicepFile)

		// Extract connections
		connections := bicep.ExtractConnections(armTemplate)
		for _, conn := range connections {
			allConnections = append(allConnections, v20231001preview.StaticAppGraphConnection{
				SourceID: conn.SourceResourceID,
				TargetID: conn.TargetResourceID,
				Type:     v20231001preview.StaticConnectionType(conn.Type),
			})
		}
	}

	// Compute source hash
	sourceHash, err := bicep.ComputeSourceHash(compiledFiles)
	if err != nil {
		g.Output.LogInfo("Warning: failed to compute source hash: %v", err)
		sourceHash = ""
	}

	// Build the graph
	graph := &v20231001preview.StaticAppGraph{
		Metadata: v20231001preview.StaticAppGraphMetadata{
			GeneratedAt:      time.Now().UTC(),
			RadiusCliVersion: version.Version(),
			SourceFiles:      compiledFiles,
			SourceHash:       sourceHash,
		},
		Resources:   allResources,
		Connections: allConnections,
	}

	// Enrich with git metadata if enabled
	if g.IncludeGitMetadata {
		if err := g.enrichWithGitMetadata(graph, compiledFiles); err != nil {
			// Git enrichment is optional - log warning but don't fail
			g.Output.LogInfo("Warning: git metadata enrichment skipped: %v", err)
		}
	}

	return graph, nil
}

// enrichWithGitMetadata adds git commit information to each resource.
// This is done by running git blame on the source files.
func (g *StaticGraphGenerator) enrichWithGitMetadata(graph *v20231001preview.StaticAppGraph, files []string) error {
	// Check if we're in a git repository
	// For now, we'll skip this as US3 implements the full git integration
	// This is a placeholder for the git enrichment logic

	// Try to get current commit SHA
	// gitCommit, err := git.GetCurrentCommit()
	// if err != nil {
	//     return err
	// }
	// graph.Metadata.GitCommit = gitCommit

	return nil
}

// GenerateFromSingleFile is a convenience method for generating a graph from a single Bicep file.
func (g *StaticGraphGenerator) GenerateFromSingleFile(bicepFile string, parametersFile string) (*v20231001preview.StaticAppGraph, error) {
	return g.Generate([]string{bicepFile}, GenerateOptions{
		ParametersFile:     parametersFile,
		IncludeGitMetadata: g.IncludeGitMetadata,
	})
}
