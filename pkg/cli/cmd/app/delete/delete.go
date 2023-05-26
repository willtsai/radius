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

package list

import (
	"context"
	"errors"
	"fmt"

	"github.com/project-radius/radius/pkg/cli"
	"github.com/project-radius/radius/pkg/cli/cmd/commonflags"
	"github.com/project-radius/radius/pkg/cli/connections"
	"github.com/project-radius/radius/pkg/cli/framework"
	"github.com/project-radius/radius/pkg/cli/output"
	"github.com/project-radius/radius/pkg/cli/prompt"
	"github.com/project-radius/radius/pkg/cli/workspaces"
	"github.com/spf13/cobra"
)

const (
	deleteConfirmation = "Are you sure you want to delete application '%v' from '%v'?"
)

// NewCommand creates an instance of the `rad app delete` command and runner.
func NewCommand(factory framework.Factory) (*cobra.Command, framework.Runner) {
	runner := NewRunner(factory)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete Radius application",
		Long:  "Delete the specified Radius application deployed in the default environment",
		Example: `
# Delete current application
rad app delete

# Delete current application and bypass confirmation prompt
rad app delete --yes

# Delete specified application
rad app delete my-app

# Delete specified application in a specified resource group
rad app delete my-app --group my-group
`,
		Args: cobra.MaximumNArgs(1),
		RunE: framework.RunCommand(runner),
	}

	commonflags.AddApplicationScopedOptionsVar(cmd, &runner.WorkspaceOptions)
	commonflags.AddConfirmationFlagVar(cmd, &runner.Confirm)

	return cmd, runner
}

// Runner is the Runner implementation for the `rad app delete` command.
type Runner struct {
	ConfigHolder      *framework.ConfigHolder
	ConnectionFactory connections.Factory
	InputPrompter     prompt.Interface
	Output            output.Interface

	Confirm          bool
	WorkspaceOptions commonflags.WorkspaceOptions
	Workspace        *workspaces.Workspace
}

// NewRunner creates an instance of the runner for the `rad app delete` command.
func NewRunner(factory framework.Factory) *Runner {
	return &Runner{
		ConfigHolder:      factory.GetConfigHolder(),
		ConnectionFactory: factory.GetConnectionFactory(),
		InputPrompter:     factory.GetPrompter(),
		Output:            factory.GetOutput(),
	}
}

// Validate runs validation for the `rad app delete` command.
func (r *Runner) Validate(cmd *cobra.Command, args []string) error {
	err := commonflags.AcceptApplicationNamePositionalArg(cmd, args, &r.WorkspaceOptions.Application)
	if err != nil {
		return err
	}

	r.Workspace, err = cli.LoadWorkspace(r.ConfigHolder.Config, r.ConfigHolder.DirectoryConfig, r.WorkspaceOptions, cli.RequiresApplication)
	if err != nil {
		return err
	}

	return nil
}

// Run runs the `rad app delete` command.
func (r *Runner) Run(ctx context.Context) error {
	// Prompt user to confirm deletion
	if !r.Confirm {
		confirmed, err := prompt.YesOrNoPrompt(fmt.Sprintf(deleteConfirmation, r.WorkspaceOptions.Application, r.Workspace.Name), prompt.ConfirmNo, r.InputPrompter)
		if err != nil {
			if errors.Is(err, &prompt.ErrExitConsole{}) {
				return &cli.FriendlyError{Message: err.Error()}
			}
			return err
		}
		if !confirmed {
			return nil
		}
	}

	client, err := r.ConnectionFactory.CreateApplicationsManagementClient(ctx, *r.Workspace)
	if err != nil {
		return err
	}

	deleted, err := client.DeleteApplication(ctx, r.WorkspaceOptions.Application)
	if err != nil {
		return err
	}

	if deleted {
		r.Output.LogInfo("Application deleted")
	} else {
		r.Output.LogInfo("Application '%s' does not exist or has already been deleted.", r.WorkspaceOptions.Application)
	}

	return nil
}
