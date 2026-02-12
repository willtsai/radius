// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package graph

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/radius-project/radius/pkg/cli"
	"github.com/radius-project/radius/pkg/cli/clients"
	"github.com/radius-project/radius/pkg/cli/clierrors"
	"github.com/radius-project/radius/pkg/cli/cmd/commonflags"
	"github.com/radius-project/radius/pkg/cli/connections"
	"github.com/radius-project/radius/pkg/cli/filesystem"
	"github.com/radius-project/radius/pkg/cli/framework"
	"github.com/radius-project/radius/pkg/cli/output"
	"github.com/radius-project/radius/pkg/cli/workspaces"
	"github.com/spf13/cobra"
)

// NewCommand creates an instance of the command and runner for the `rad app graph` command.
func NewCommand(factory framework.Factory) (*cobra.Command, framework.Runner) {
	runner := NewRunner(factory)
	cmd := &cobra.Command{
		Use:   "graph [application-name | file.bicep]",
		Short: "Shows the application graph for an application or Bicep file.",
		Long: `Shows the application graph for an application or Bicep file.

When given an application name, fetches the graph from a running Radius environment.
When given a Bicep file (.bicep), generates a static graph without deployment.

Static graph generation extracts resources, connections, and metadata from Bicep files
and outputs to .radius/app-graph.json by default. Use --stdout to print to stdout,
or --output to specify a custom path.`,
		Args: cobra.MaximumNArgs(1),
		Example: `
# Show graph for current application (requires Radius environment)
rad app graph

# Show graph for specified application (requires Radius environment)
rad app graph my-application

# Generate static graph from Bicep file (no deployment required)
rad app graph app.bicep

# Generate static graph with parameters file
rad app graph app.bicep --parameters app.bicepparam

# Output static graph to stdout instead of file
rad app graph app.bicep --stdout

# Generate static graph with Markdown output (includes Mermaid diagrams)
rad app graph app.bicep --format markdown`,
		RunE: framework.RunCommand(runner),
	}

	commonflags.AddWorkspaceFlag(cmd)
	commonflags.AddResourceGroupFlag(cmd)
	commonflags.AddApplicationNameFlag(cmd)

	// Add static graph generation flags
	cmd.Flags().StringP("parameters", "p", "", "Path to Bicep parameters file (.bicepparam or .json)")
	cmd.Flags().StringP("output", "o", "", "Output path for generated graph (default: .radius/app-graph.json)")
	cmd.Flags().Bool("stdout", false, "Output graph to stdout instead of file")
	cmd.Flags().Bool("no-git", false, "Skip git metadata enrichment (faster)")
	cmd.Flags().String("format", "json", "Output format: json, markdown (markdown also generates JSON)")

	return cmd, runner
}

// Runner is the runner implementation for the `rad app graph` command.
type Runner struct {
	ConfigHolder      *framework.ConfigHolder
	ConnectionFactory connections.Factory
	Output            output.Interface
	FileSystem        filesystem.FileSystem

	// Input detection
	InputType InputType
	BicepFile string

	// Existing app graph fields
	ApplicationName string
	Workspace       *workspaces.Workspace

	// Static graph options
	ParametersFile     string
	OutputPath         string
	UseStdout          bool
	IncludeGitMetadata bool
	OutputFormat       string
}

// NewRunner creates a new instance of the `rad app graph` runner.
func NewRunner(factory framework.Factory) *Runner {
	return &Runner{
		ConfigHolder:       factory.GetConfigHolder(),
		Output:             factory.GetOutput(),
		ConnectionFactory:  factory.GetConnectionFactory(),
		FileSystem:         filesystem.NewOSFS(),
		IncludeGitMetadata: true,
	}
}

// Validate runs validation for the `rad app graph` command.
func (r *Runner) Validate(cmd *cobra.Command, args []string) error {
	// Parse flags for static generation
	r.ParametersFile, _ = cmd.Flags().GetString("parameters")
	r.OutputPath, _ = cmd.Flags().GetString("output")
	r.UseStdout, _ = cmd.Flags().GetBool("stdout")
	noGit, _ := cmd.Flags().GetBool("no-git")
	r.IncludeGitMetadata = !noGit
	r.OutputFormat, _ = cmd.Flags().GetString("format")

	// Determine input type
	var input string
	if len(args) > 0 {
		input = args[0]
	}

	r.InputType = DetectInputType(input)

	// Handle based on input type
	switch r.InputType {
	case InputTypeBicepFile:
		return r.validateBicepInput(cmd, input)
	case InputTypeAppName:
		return r.validateAppInput(cmd, args)
	case InputTypeUnknown:
		// No argument - try to get from workspace/flags (existing behavior)
		return r.validateAppInput(cmd, args)
	default:
		return fmt.Errorf("could not determine input type. Provide an application name or Bicep file path")
	}
}

// validateBicepInput validates input for static graph generation from Bicep files.
func (r *Runner) validateBicepInput(cmd *cobra.Command, bicepFile string) error {
	// Check file exists
	if _, err := r.FileSystem.Stat(bicepFile); err != nil {
		return clierrors.Message("Bicep file %q not found: %v", bicepFile, err)
	}

	// Resolve to absolute path so output paths are predictable
	absBicep, err := filepath.Abs(bicepFile)
	if err == nil {
		bicepFile = absBicep
	}
	r.BicepFile = bicepFile

	// Validate parameters file if specified
	if r.ParametersFile != "" {
		if _, err := r.FileSystem.Stat(r.ParametersFile); err != nil {
			return clierrors.Message("Parameters file %q not found: %v", r.ParametersFile, err)
		}
	}

	// Set default output path if not specified and not using stdout
	if r.OutputPath == "" && !r.UseStdout {
		r.OutputPath = DefaultOutputPath(bicepFile)
	}

	// Resolve to absolute path so the output location is unambiguous
	if r.OutputPath != "" {
		absPath, err := filepath.Abs(r.OutputPath)
		if err == nil {
			r.OutputPath = absPath
		}
	}

	// Validate output format
	if r.OutputFormat != "json" && r.OutputFormat != "markdown" {
		return clierrors.Message("Invalid output format %q. Use 'json' or 'markdown'", r.OutputFormat)
	}

	return nil
}

// validateAppInput validates input for fetching graph from running Radius environment.
func (r *Runner) validateAppInput(cmd *cobra.Command, args []string) error {
	workspace, err := cli.RequireWorkspace(cmd, r.ConfigHolder.Config, r.ConfigHolder.DirectoryConfig)
	if err != nil {
		return err
	}
	r.Workspace = workspace

	r.Workspace.Scope, err = cli.RequireScope(cmd, *r.Workspace)
	if err != nil {
		return err
	}

	r.ApplicationName, err = cli.RequireApplicationArgs(cmd, args, *r.Workspace)
	if err != nil {
		return err
	}

	client, err := r.ConnectionFactory.CreateApplicationsManagementClient(cmd.Context(), *r.Workspace)
	if err != nil {
		return err
	}

	// Validate that the application exists
	_, err = client.GetApplication(cmd.Context(), r.ApplicationName)
	if clients.Is404Error(err) {
		return clierrors.Message("Application %q does not exist or has been deleted.", r.ApplicationName)
	} else if err != nil {
		return err
	}

	return nil
}

// Run runs the `rad app graph` command.
func (r *Runner) Run(ctx context.Context) error {
	switch r.InputType {
	case InputTypeBicepFile:
		return r.runStaticGraphGeneration(ctx)
	default:
		return r.runLiveGraphFetch(ctx)
	}
}

// runStaticGraphGeneration generates a graph from Bicep files without deployment.
func (r *Runner) runStaticGraphGeneration(ctx context.Context) error {
	generator := NewStaticGraphGenerator(r.FileSystem, r.Output)
	generator.IncludeGitMetadata = r.IncludeGitMetadata

	opts := GenerateOptions{
		ParametersFile:     r.ParametersFile,
		IncludeGitMetadata: r.IncludeGitMetadata,
	}

	graph, err := generator.Generate([]string{r.BicepFile}, opts)
	if err != nil {
		return clierrors.Message("Failed to generate graph: %v", err)
	}

	// Output the graph
	if r.UseStdout {
		return r.outputToStdout(graph)
	}

	return r.outputToFile(graph)
}

// outputToStdout writes the graph to stdout.
func (r *Runner) outputToStdout(graph interface{}) error {
	jsonBytes, err := output.DeterministicJSON(graph)
	if err != nil {
		return fmt.Errorf("failed to serialize graph: %w", err)
	}

	fmt.Println(string(jsonBytes))
	return nil
}

// outputToFile writes the graph to the configured output file.
func (r *Runner) outputToFile(graph interface{}) error {
	// Ensure output directory exists
	outputDir := filepath.Dir(r.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Serialize to JSON
	jsonBytes, err := output.DeterministicJSON(graph)
	if err != nil {
		return fmt.Errorf("failed to serialize graph: %w", err)
	}

	// Write JSON file
	if err := os.WriteFile(r.OutputPath, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	r.Output.LogInfo("Graph written to %s", r.OutputPath)

	// Also generate Markdown if requested
	if r.OutputFormat == "markdown" {
		mdPath := DefaultMarkdownOutputPath(r.BicepFile)
		r.Output.LogInfo("Markdown output would be written to %s (not yet implemented)", mdPath)
		// TODO: Implement Markdown generation in US2
	}

	return nil
}

// runLiveGraphFetch fetches the graph from a running Radius environment.
func (r *Runner) runLiveGraphFetch(ctx context.Context) error {
	client, err := r.ConnectionFactory.CreateApplicationsManagementClient(ctx, *r.Workspace)
	if err != nil {
		return err
	}

	applicationGraphResponse, err := client.GetApplicationGraph(ctx, r.ApplicationName)
	if err != nil {
		return err
	}
	graphResources := applicationGraphResponse.Resources
	displayOutput := display(graphResources, r.ApplicationName)
	r.Output.LogInfo(displayOutput)

	return nil
}
