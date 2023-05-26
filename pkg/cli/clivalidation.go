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

package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/project-radius/radius/pkg/cli/cmd/commonflags"
	"github.com/project-radius/radius/pkg/cli/config"
	"github.com/project-radius/radius/pkg/cli/ucp"
	"github.com/project-radius/radius/pkg/cli/workspaces"
	"github.com/project-radius/radius/pkg/ucp/resources"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type RequiredScope int

const (
	RequiresWorkspace     RequiredScope = 0x00000000
	RequiresResourceGroup RequiredScope = 0x00000001
	RequiresEnvironment   RequiredScope = 0x00000011
	RequiresApplication   RequiredScope = 0x00000111
)

type AzureResource struct {
	Name           string
	ResourceType   string
	ResourceGroup  string
	SubscriptionID string
}

func RequireEnvironmentNameArgs(cmd *cobra.Command, args []string, workspace workspaces.Workspace) (string, error) {
	environmentName, err := ReadEnvironmentNameArgs(cmd, args)
	if err != nil {
		return "", err
	}

	// We store the environment id in config, but most commands work with the environment name.
	if environmentName == "" && workspace.Environment != "" {
		id, err := resources.ParseResource(workspace.Environment)
		if err != nil {
			return "", err
		}

		environmentName = id.Name()
	}

	if environmentName == "" {
		return "", fmt.Errorf("no environment name provided and no default environment set, " +
			"either pass in an environment name or set a default environment by using `rad env switch`")
	}

	return environmentName, err
}

func RequireEnvironmentName(cmd *cobra.Command, args []string, workspace workspaces.Workspace) (string, error) {
	environmentName, err := cmd.Flags().GetString("environment")
	if err != nil {
		return "", err
	}

	// We store the environment id in config, but most commands work with the environment name.
	if environmentName == "" && workspace.Environment != "" {
		id, err := resources.ParseResource(workspace.Environment)
		if err != nil {
			return "", err
		}

		environmentName = id.Name()
	}

	if environmentName == "" && workspace.IsEditableWorkspace() {
		// Setting a default environment only applies to editable workspaces
		return "", fmt.Errorf("no environment name provided and no default environment set, " +
			"either pass in an environment name or set a default environment by using `rad env switch`")
	} else if environmentName == "" {
		return "", fmt.Errorf("no environment name provided, pass in an environment name")
	}

	return environmentName, err
}

// RequireKubeContext is used by commands that need a kubernetes context name to be specified using -c flag or has a default kubecontext
func RequireKubeContext(cmd *cobra.Command, currentContext string) (string, error) {
	kubecontext, err := cmd.Flags().GetString("context")
	if err != nil {
		return "", err
	}

	if kubecontext == "" && currentContext == "" {
		return "", errors.New("the kubeconfig has no current context")
	} else if kubecontext == "" {
		kubecontext = currentContext
	}

	return kubecontext, nil
}

func ReadEnvironmentNameArgs(cmd *cobra.Command, args []string) (string, error) {
	name, err := cmd.Flags().GetString("environment")
	if err != nil {
		return "", err
	}

	if len(args) > 0 {
		if name != "" {
			return "", fmt.Errorf("cannot specify environment name via both arguments and `-e`")
		}
		name = args[0]
	}

	return name, err
}

// RequireApplicationArgs reads the application name from the following sources in priority order and returns
// an error if no application name is set.
//
// - '--application' flag
// - first positional arg
// - workspace default application
// - directory config application
func RequireApplicationArgs(cmd *cobra.Command, args []string, workspace workspaces.Workspace) (string, error) {
	applicationName, err := ReadApplicationNameArgs(cmd, args)
	if err != nil {
		return "", err
	}

	if applicationName == "" {
		applicationName = workspace.DefaultApplication
	}

	if applicationName == "" {
		applicationName = workspace.DirectoryConfig.Workspace.Application
	}

	if applicationName == "" {
		return "", fmt.Errorf("no application name provided and no default application set, " +
			"either pass in an application name or set a default application by using `rad application switch`")
	}

	return applicationName, nil
}

// ReadApplicationName reads the application name from the following sources in priority order and returns
// the empty string if no application is set.
//
// - '--application' flag
// - workspace default application
// - directory config application
func ReadApplicationName(cmd *cobra.Command, workspace workspaces.Workspace) (string, error) {
	applicationName, err := cmd.Flags().GetString("application")
	if err != nil {
		return "", err
	}

	if applicationName == "" {
		applicationName = workspace.DefaultApplication
	}

	if applicationName == "" {
		applicationName = workspace.DirectoryConfig.Workspace.Application
	}

	return applicationName, nil
}

// ReadApplicationName reads the application name from the following sources in priority order and returns
// the empty string if no application is set.
//
// - '--application' flag
// - first positional arg
func ReadApplicationNameArgs(cmd *cobra.Command, args []string) (string, error) {
	name, err := cmd.Flags().GetString("application")
	if err != nil {
		return "", err
	}

	if len(args) > 0 {
		if name != "" {
			return "", fmt.Errorf("cannot specify application name via both arguments and `-a`")
		}
		name = args[0]
	}

	return name, err
}

// RequireApplicationArgs reads the application name from the following sources in priority order and returns
// an error if no application name is set.
//
// - '--application' flag
// - workspace default application
// - directory config application
func RequireApplication(cmd *cobra.Command, workspace workspaces.Workspace) (string, error) {
	return RequireApplicationArgs(cmd, []string{}, workspace)
}

func RequireResource(cmd *cobra.Command, args []string) (resourceType string, resourceName string, err error) {
	results, err := requiredMultiple(cmd, args, "type", "resource")
	if err != nil {
		return "", "", err
	}
	return results[0], results[1], nil
}

func RequireResourceTypeAndName(args []string) (string, string, error) {
	if len(args) < 2 {
		return "", "", errors.New("no resource type or name provided")
	}
	resourceType, err := RequireResourceType(args)
	if err != nil {
		return "", "", err
	}
	resourceName := args[1]
	return resourceType, resourceName, nil
}

// example of resource Type: Applications.Core/httpRoutes, Applications.Link/redisCaches
func RequireResourceType(args []string) (string, error) {
	if len(args) < 1 {
		return "", errors.New("no resource type provided")
	}
	resourceTypeName := args[0]
	supportedTypes := []string{}
	for _, resourceType := range ucp.ResourceTypesList {
		supportedType := strings.Split(resourceType, "/")[1]
		supportedTypes = append(supportedTypes, supportedType)
		if strings.EqualFold(supportedType, resourceTypeName) {
			return resourceType, nil
		}
	}
	return "", fmt.Errorf("'%s' is not a valid resource type. Available Types are: \n\n%s\n",
		resourceTypeName, strings.Join(supportedTypes, "\n"))
}

// RequireWorkspace is used by commands that require an existing workspace either set as the default,
// or specified using the 'workspace' flag.
func RequireWorkspace(cmd *cobra.Command, config *viper.Viper, dc *config.DirectoryConfig) (*workspaces.Workspace, error) {
	name, err := cmd.Flags().GetString("workspace")
	if err != nil {
		return nil, err
	}

	section, err := ReadWorkspaceSection(config)
	if err != nil {
		return nil, err
	}

	ws, err := section.GetWorkspace(name)
	if err != nil {
		return nil, err
	}

	// If we get here and ws is nil then this means there's no default set (or no config).
	// Lets use the fallback configuration.
	if ws == nil {
		ws = workspaces.MakeFallbackWorkspace()
	}

	if dc != nil {
		ws.DirectoryConfig = *dc
	}

	return ws, nil
}

// RequireUCPResourceGroup is used by commands that require specifying a UCP resouce group name using flag or positional args
func RequireUCPResourceGroup(cmd *cobra.Command, args []string) (string, error) {
	group, err := ReadResourceGroupNameArgs(cmd, args)
	if err != nil {
		return "", err
	}
	if group == "" {
		return "", fmt.Errorf("resource group name is not provided or is empty ")
	}

	return group, nil
}

// ReadResourceGroupNameArgs is used to get the resource group name that is supplied as either the first argument for group commands or using a -g flag
func ReadResourceGroupNameArgs(cmd *cobra.Command, args []string) (string, error) {
	name, err := cmd.Flags().GetString("group")
	if err != nil {
		return "", err
	}

	if len(args) > 0 {
		if name != "" {
			return "", fmt.Errorf("cannot specify resource group name via both arguments and `-g`")
		}
		name = args[0]
	}

	return name, err
}

func requiredMultiple(cmd *cobra.Command, args []string, names ...string) ([]string, error) {
	results := make([]string, len(names))
	for i, name := range names {
		value, err := cmd.Flags().GetString(name)
		if err == nil {
			results[i] = value
		}
		if results[i] != "" {
			if len(args) > len(names)-i-1 {
				return nil, fmt.Errorf("cannot specify %v name via both arguments and switch", name)
			}
			continue
		}
		if len(args) == 0 {
			return nil, fmt.Errorf("no %v name provided", name)
		}
		results[i] = args[0]
		args = args[1:]
	}
	return results, nil
}

// RequireScope returns the scope the command should use to execute or an error if unset.
//
// This function considers the following sources:
//
// - --group flag
// - workspace scope
func RequireScope(cmd *cobra.Command, workspace workspaces.Workspace) (string, error) {
	resourceGroup, err := cmd.Flags().GetString("group")
	if err != nil {
		return "", err
	}

	if resourceGroup != "" {
		return fmt.Sprintf("/planes/radius/local/resourceGroups/%s", resourceGroup), nil
	} else if workspace.Scope != "" {
		return workspace.Scope, nil
	} else if workspace.IsEditableWorkspace() {
		return "", &FriendlyError{Message: "no resource group set, either use `rad group switch` to set, or use `--group` to pass in a resource group name"}
	} else {
		return "", &FriendlyError{Message: "no resource group set, use `--group` to pass in a resource group name"}
	}
}

func LoadWorkspace(config *viper.Viper, dc *config.DirectoryConfig, options commonflags.WorkspaceOptions, requires RequiredScope) (*workspaces.Workspace, error) {
	section, err := ReadWorkspaceSection(config)
	if err != nil {
		return nil, err
	}

	// Load workspace based on command line options.
	ws, err := section.GetWorkspace(options.Workspace)
	if err != nil {
		return nil, err
	}

	// If we get here and ws is nil then this means there's no default set (or no config).
	// Lets use the fallback configuration.
	if ws == nil {
		ws = workspaces.MakeFallbackWorkspace()
	}

	if dc != nil {
		ws.DirectoryConfig = *dc
	}

	if requires == RequiresWorkspace {
		return ws, nil
	}

	if options.ResourceGroup != "" {
		ws.Scope = fmt.Sprintf("/planes/radius/local/resourceGroups/%s", options.ResourceGroup)
	} else if ws.Scope == "" && ws.IsEditableWorkspace() {
		return nil, &FriendlyError{Message: "no resource group name provided and no default resource group name set, " +
			"either use `--group` to provide a resource group name or set a default resource group name using `rad group switch`"}
	} else if ws.Scope == "" {
		return nil, &FriendlyError{Message: "no resource group set, use `--group` to provide in a resource group name"}
	}

	if requires == RequiresResourceGroup {
		return ws, nil
	}

	if options.Environment != "" {
		ws.Environment = fmt.Sprintf("%s/providers/Applications.Core/environments/%s", ws.Scope, options.Environment)
	} else if ws.Environment == "" && ws.IsEditableWorkspace() {
		return nil, &FriendlyError{Message: "no environment name provided and no default environment set, " +
			"either use `--environment` to provide an environment name or set a default environment by using `rad env switch`"}
	} else if ws.Environment == "" {
		return nil, &FriendlyError{Message: "no environment set, use `--environment` to provide an environment name"}
	}

	if requires == RequiresEnvironment {
		return ws, nil
	}

	if options.Application != "" {
		ws.DefaultApplication = fmt.Sprintf("%s/providers/Applications.Core/applications/%s", ws.Scope, options.Application)
	} else if ws.DefaultApplication == "" {
		return nil, &FriendlyError{Message: "no application name provided, use `--application` to provide an application name"}
	}

	return ws, nil
}
