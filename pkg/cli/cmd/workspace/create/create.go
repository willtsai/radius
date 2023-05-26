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

package create

import (
	"context"
	"fmt"
	"strings"

	"github.com/project-radius/radius/pkg/cli"
	"github.com/project-radius/radius/pkg/cli/clients"
	"github.com/project-radius/radius/pkg/cli/cmd/commonflags"
	"github.com/project-radius/radius/pkg/cli/connections"
	"github.com/project-radius/radius/pkg/cli/framework"
	"github.com/project-radius/radius/pkg/cli/helm"
	"github.com/project-radius/radius/pkg/cli/kubernetes"
	"github.com/project-radius/radius/pkg/cli/output"
	"github.com/project-radius/radius/pkg/cli/workspaces"
	"github.com/spf13/cobra"
)

// NewCommand creates an instance of the command and runner for the `rad workspace create` command.
func NewCommand(factory framework.Factory) (*cobra.Command, framework.Runner) {
	runner := NewRunner(factory)

	cmd := &cobra.Command{
		Use:   "create [workspaceType] [workspaceName]",
		Short: "Create a workspace",
		Long: `Create a workspace.
		
Available workspaceTypes: kubernetes

Workspaces allow you to manage multiple Radius platforms and environments using a local configuration file. 

You can easily define and switch between workspaces to deploy and manage applications across local, test, and production environments.`,
		Args: ValidateArgs(),
		Example: `
# Create a workspace with name 'myworkspace' and kubernetes context 'aks'
rad workspace create kubernetes myworkspace --context aks
# Create a workspace with name of current kubernetes context in current kubernetes context
rad workspace create kubernetes`,
		RunE: framework.RunCommand(runner),
	}

	commonflags.AddWorkspaceNameFlagVar(cmd, &runner.WorkspaceName)
	commonflags.AddResourceGroupFlagVar(cmd, &runner.ResourceGroup)
	commonflags.AddEnvironmentNameFlagVar(cmd, &runner.EnvironmentName)
	commonflags.AddContextFlagVar(cmd, &runner.Context)
	commonflags.AddForceFlagVar(cmd, &runner.Force)

	return cmd, runner
}

// Runner is the runner implementation for the `rad workspace create` command.
type Runner struct {
	ConfigFileInterface framework.ConfigFileInterface
	ConfigHolder        *framework.ConfigHolder
	ConnectionFactory   connections.Factory
	HelmInterface       helm.Interface
	KubernetesInterface kubernetes.Interface
	Output              output.Interface

	Context         string
	EnvironmentName string
	Force           bool
	ResourceGroup   string
	WorkspaceName   string
	Workspace       *workspaces.Workspace
}

// NewRunner creates a new instance of the `rad workspace create` runner.
func NewRunner(factory framework.Factory) *Runner {
	return &Runner{
		ConnectionFactory:   factory.GetConnectionFactory(),
		ConfigHolder:        factory.GetConfigHolder(),
		ConfigFileInterface: factory.GetConfigFileInterface(),
		Output:              factory.GetOutput(),
		HelmInterface:       factory.GetHelmInterface(),
		KubernetesInterface: factory.GetKubernetesInterface(),
	}
}

// Validate runs validation for the `rad workspace create` command.
func (r *Runner) Validate(cmd *cobra.Command, args []string) error {
	// Read the second arg, as 'kubernetes' is the first arg.
	err := commonflags.AcceptWorkspaceNamePositionalArg(cmd, args[1:], &r.WorkspaceName)
	if err != nil {
		return err
	}

	kubeContextList, err := r.KubernetesInterface.GetKubeContext()
	if err != nil {
		return &cli.FriendlyError{Message: "Failed to read kube config"}
	}

	if r.Context == "" {
		r.Context = kubeContextList.CurrentContext
	}

	context, err := cli.RequireKubeContext(cmd, r.Context)
	if err != nil {
		return err
	}

	_, ok := kubeContextList.Contexts[context]
	if !ok {
		return fmt.Errorf("the kubeconfig does not contain a context called %q", context)
	}

	if r.WorkspaceName == "" {
		r.WorkspaceName = context
	}

	state, err := r.HelmInterface.CheckRadiusInstall(context)
	if !state.Installed || err != nil {
		return fmt.Errorf("unable to create workspace %q. Radius control plane not installed on target platform. Run 'rad install' and try again", r.WorkspaceName)
	}

	workspaceExists, err := cli.HasWorkspace(r.ConfigHolder.Config, r.WorkspaceName)
	if err != nil {
		return err
	}

	if !r.Force && workspaceExists {
		return fmt.Errorf("workspace exists. please specify --force to overwrite")
	}

	if workspaceExists {
		workspace, err := cli.GetWorkspace(r.ConfigHolder.Config, r.WorkspaceName)
		if err != nil {
			return err
		}
		r.Workspace = workspace
	} else {
		r.Workspace = &workspaces.Workspace{Name: r.WorkspaceName}
	}
	r.Workspace.Connection = map[string]any{}
	r.Workspace.Connection["context"] = context
	r.Workspace.Connection["kind"] = args[0]

	var client clients.ApplicationsManagementClient
	if r.ResourceGroup != "" {
		r.Workspace.Scope = "/planes/radius/local/resourceGroups/" + r.ResourceGroup

		client, err = r.ConnectionFactory.CreateApplicationsManagementClient(cmd.Context(), *r.Workspace)
		if err != nil {
			return err
		}
		_, err := client.ShowUCPGroup(cmd.Context(), "radius", "local", r.ResourceGroup)
		if err != nil {
			return &cli.FriendlyError{Message: fmt.Sprintf("group %q does not exist. Run `rad env create` try again \n", r.Workspace.Scope)}
		}

		//we want to make sure we dont have a workspace which has environment in a different scope from workspace's scope
		if r.Workspace.Environment != "" && !strings.HasPrefix(r.Workspace.Environment, r.Workspace.Scope) && r.EnvironmentName == "" {
			return fmt.Errorf("workspace is currently using an environment which is in different scope. use -e to specify an environment which is in the scope of this workspace")
		}
	}

	if r.EnvironmentName != "" {
		if r.Workspace.Scope == "" {
			return fmt.Errorf("cannot set environment for workspace with empty scope. use -g to set a scope")
		}
		r.Workspace.Environment = r.Workspace.Scope + "/providers/applications.core/environments/" + r.EnvironmentName

		_, err = client.GetEnvDetails(cmd.Context(), r.EnvironmentName)
		if err != nil {
			return &cli.FriendlyError{Message: fmt.Sprintf("environment %q does not exist. Run `rad env create` try again \n", r.Workspace.Environment)}
		}
	}

	return nil
}

// Run runs the `rad workspace create` command.
func (r *Runner) Run(ctx context.Context) error {

	r.Output.LogInfo("creating workspace...")
	err := r.ConfigFileInterface.EditWorkspaces(ctx, r.ConfigHolder.Config, r.Workspace)
	if err != nil {
		return err
	}
	output.LogInfo("Set %q as current workspace", r.Workspace.Name)

	return nil
}
