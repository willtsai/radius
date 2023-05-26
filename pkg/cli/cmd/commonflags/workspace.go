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

package commonflags

import (
	"errors"

	"github.com/spf13/cobra"
)

type WorkspaceOptions struct {
	Workspace     string
	ResourceGroup string
	Environment   string
	Application   string
}

func AddResourceGroupScopedOptionsVar(cmd *cobra.Command, ref *WorkspaceOptions) {
	AddWorkspaceNameFlagVar(cmd, &ref.Workspace)
	AddResourceGroupFlagVar(cmd, &ref.ResourceGroup)
}

func AddEnvironmentScopedOptionsVar(cmd *cobra.Command, ref *WorkspaceOptions) {
	AddWorkspaceNameFlagVar(cmd, &ref.Workspace)
	AddResourceGroupFlagVar(cmd, &ref.ResourceGroup)
	AddEnvironmentNameFlagVar(cmd, &ref.Environment)
}

func AddApplicationScopedOptionsVar(cmd *cobra.Command, ref *WorkspaceOptions) {
	AddWorkspaceNameFlagVar(cmd, &ref.Workspace)
	AddResourceGroupFlagVar(cmd, &ref.ResourceGroup)
	AddEnvironmentNameFlagVar(cmd, &ref.Environment)
	AddApplicationNameFlagVar(cmd, &ref.Application)
}

// AddWorkspaceNameFlagVar defines the '--workspace' flag for a command and binds it to a variable.
func AddWorkspaceNameFlagVar(cmd *cobra.Command, ref *string) {
	cmd.Flags().StringP("workspace", "w", "", "The workspace name")
}

// AcceptWorkspaceNamePositionalArg accepts a workspace name as a positional argument or via the '--workspace' flag.
// This function returns an error if the user specifies both the workspace name via a positional argument and via the
// flag.
func AcceptWorkspaceNamePositionalArg(cmd *cobra.Command, args []string, ref *string) error {
	if len(args) > 0 && *ref != "" {
		return errors.New("cannot specify workspace name via both arguments and `--workspace`")
	} else if len(args) > 0 {
		*ref = args[0]
	}

	return nil
}

// AddResourceGroupFlagVar defines the '--group' flag for a command and binds it to a variable.
func AddResourceGroupFlagVar(cmd *cobra.Command, ref *string) {
	cmd.Flags().StringVarP(ref, "group", "g", "", "The resource group name")
}

// AcceptResourceGroupPositionalArg accepts a resource group name as a positional argument or via the '--group' flag.
// This function returns an error if the user specifies both the resource group name via a positional argument and via
// the flag.
func AcceptResourceGroupPositionalArg(cmd *cobra.Command, args []string, ref *string) error {
	if len(args) > 0 && *ref != "" {
		return errors.New("cannot specify resource group name via both arguments and `--group`")
	} else if len(args) > 0 {
		*ref = args[0]
	}

	return nil
}

// AddEnvironmentNameFlagVar defines the '--environment' flag for a command and binds it to a variable.
func AddEnvironmentNameFlagVar(cmd *cobra.Command, ref *string) {
	cmd.Flags().StringVarP(ref, "environment", "e", "", "The environment name")
}

// AcceptEnvironmentNamePositionalArg accepts an environment name as a positional argument or via the '--environment' flag.
// This function returns an error if the user specifies both the environment name via a positional argument and via
// the flag.
func AcceptEnvironmentNamePositionalArg(cmd *cobra.Command, args []string, ref *string) error {
	if len(args) > 0 && *ref != "" {
		return errors.New("cannot specify environment name via both arguments and `--environment`")
	} else if len(args) > 0 {
		*ref = args[0]
	}

	return nil
}

// AddApplicationNameFlagVar defines the '--application' flag for a command and binds it to a variable.
func AddApplicationNameFlagVar(cmd *cobra.Command, ref *string) {
	cmd.Flags().StringVarP(ref, "application", "a", "", "The application name")
}

// AcceptApplicationNamePositionalArg accepts an application name as a positional argument or via the '--application' flag.
// This function returns an error if the user specifies both the application name via a positional argument and via
// the flag.
func AcceptApplicationNamePositionalArg(cmd *cobra.Command, args []string, ref *string) error {
	if len(args) > 0 && *ref != "" {
		return errors.New("cannot specify application name via both arguments and `--application`")
	} else if len(args) > 0 {
		*ref = args[0]
	}

	return nil
}

// TODO: the APIs below this point are planned for removal once we migrate all of the usage.
func AddResourceGroupFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("group", "g", "", "The resource group name")
}

func AddEnvironmentNameFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("environment", "e", "", "The environment name")
}

func AddApplicationNameFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("application", "a", "", "The application name")
}
