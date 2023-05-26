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

package workspaceswitch // switch is a reserved word in go, so we can't use it as a package name.

import (
	"context"
	"strings"

	"github.com/project-radius/radius/pkg/cli"
	"github.com/project-radius/radius/pkg/cli/cmd/commonflags"
	"github.com/project-radius/radius/pkg/cli/framework"
	"github.com/project-radius/radius/pkg/cli/output"
	"github.com/project-radius/radius/pkg/cli/workspaces"
	"github.com/spf13/cobra"
)

// NewCommand creates an instance of the command and runner for the `rad workspace switch` command.
func NewCommand(factory framework.Factory) (*cobra.Command, framework.Runner) {
	runner := NewRunner(factory)

	cmd := &cobra.Command{
		Use:   "switch",
		Short: "Switch current workspace",
		Long:  `Switch current workspace`,
		Example: `# Switch current workspace
rad workspace switch my-workspace`,
		Args: cobra.RangeArgs(0, 1),
		RunE: framework.RunCommand(runner),
	}

	commonflags.AddWorkspaceNameFlagVar(cmd, &runner.WorkspaceName)

	return cmd, runner
}

// Runner is the runner implementation for the `rad workspace switch` command.
type Runner struct {
	ConfigHolder        *framework.ConfigHolder
	ConfigFileInterface framework.ConfigFileInterface
	Output              output.Interface
	WorkspaceName       string
}

// NewRunner creates a new instance of the `rad workspace switch` runner.
func NewRunner(factory framework.Factory) *Runner {
	return &Runner{
		ConfigHolder:        factory.GetConfigHolder(),
		ConfigFileInterface: factory.GetConfigFileInterface(),
		Output:              factory.GetOutput(),
	}
}

// Validate runs validation for the `rad workspace switch` command.
func (r *Runner) Validate(cmd *cobra.Command, args []string) error {
	// We read the name explicitly rather than calling LoadWorkspace
	// because we require a workspace to be specified. LoadWorkspace would
	// apply our defaulting logic and miss some error cases.
	err := commonflags.AcceptWorkspaceNamePositionalArg(cmd, args, &r.WorkspaceName)
	if err != nil {
		return err
	}

	if r.WorkspaceName == "" {
		return workspaces.ErrNamedWorkspaceRequired
	}

	// We don't actually need the workspace, but we want to make sure it exists.
	//
	// So this is being called for the side-effect of running the validation.
	_, err = cli.GetWorkspace(r.ConfigHolder.Config, r.WorkspaceName)
	if err != nil {
		return err
	}

	return nil
}

// Run runs the `rad workspace switch` command.
func (r *Runner) Run(ctx context.Context) error {
	section, err := cli.ReadWorkspaceSection(r.ConfigHolder.Config)
	if err != nil {
		return err
	}

	if strings.EqualFold(section.Default, r.WorkspaceName) {
		r.Output.LogInfo("Default environment is already set to %v", r.WorkspaceName)
		return nil
	}

	if section.Default == "" {
		r.Output.LogInfo("Switching default workspace to %v", r.WorkspaceName)
	} else {
		r.Output.LogInfo("Switching default workspace from %v to %v", section.Default, r.WorkspaceName)
	}

	err = r.ConfigFileInterface.SetDefaultWorkspace(ctx, r.ConfigHolder.Config, r.WorkspaceName)
	if err != nil {
		return err
	}

	return nil
}
