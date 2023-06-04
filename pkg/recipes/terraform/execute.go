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

package terraform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/project-radius/radius/pkg/recipes"
	"github.com/project-radius/radius/pkg/recipes/terraform/config"
	"github.com/project-radius/radius/pkg/sdk"
)

var (
	// TODO this is a placeholder for now - we need to get the providers from the user or automatically detect from the module
	tfProviders = []config.TerraformProviderMetadata{
		{
			Type: "azurerm",
		},
		{
			Type: "aws",
		},
		{
			Type: "kubernetes",
		},
	}
)

func Deploy(ctx context.Context, ucpConn *sdk.Connection, tfDir string, configuration *recipes.Configuration, recipe *recipes.Metadata, definition *recipes.Definition) (*recipes.RecipeOutput, error) {
	logger := logr.FromContextOrDiscard(ctx)

	// Install Terraform
	execPath, err := Install(ctx, tfDir)
	if err != nil {
		return nil, err
	}
	logger.Info(fmt.Sprintf("Terraform installation path: %q", execPath))

	// Create Working Directory
	workingDir, err := createWorkingDir(ctx, tfDir)
	if err != nil {
		return nil, err
	}

	// Generate Terraform json config in the working directory
	err = config.GenerateConfigFiles(ctx, ucpConn, tfProviders, configuration, recipe, definition, workingDir)
	if err != nil {
		return nil, err
	}

	// Run TF Init and Apply
	recipeOutputs, err := initAndApply(ctx, workingDir, execPath)
	if err != nil {
		return nil, err
	}

	return recipeOutputs, nil
}

func createWorkingDir(ctx context.Context, tfDir string) (string, error) {
	logger := logr.FromContextOrDiscard(ctx)

	workingDir := filepath.Join(tfDir, "exec")
	logger.Info(fmt.Sprintf("Creating Terraform working directory: %q", workingDir))
	if err := os.MkdirAll(workingDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create working directory for terraform execution: %w", err)
	}

	return workingDir, nil
}

// Runs Terraform init and apply in the provided working directory.
func initAndApply(ctx context.Context, workingDir, execPath string) (*recipes.RecipeOutput, error) {
	logger := logr.FromContextOrDiscard(ctx)

	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		return nil, err
	}

	// Initialize Terraform
	logger.Info("Initializing Terraform")
	if err := tf.Init(ctx); err != nil {
		return nil, fmt.Errorf("terraform init failure: %w", err)
	}

	// Apply Terraform configuration
	logger.Info("Running Terraform apply")
	if err := tf.Apply(ctx); err != nil {
		return nil, fmt.Errorf("terraform apply failure: %w", err)
	}

	logger.Info("Fetching module outputs")
	tfState, err := tf.Show(ctx)
	if err != nil {
		return nil, err
	}

	recipeOutput := recipes.RecipeOutput{
		Secrets: make(map[string]interface{}),
		Values:  make(map[string]interface{}),
	}
	outputs := tfState.Values.Outputs
	// Access the "secrets" key and assign it to the RecipeOutput struct
	if secretOutputs, ok := outputs["secrets"]; ok {
		recipeOutput.Secrets = secretOutputs.Value.(map[string]interface{})
	}
	// Access the "values" key and assign it to the RecipeOutput struct
	if valuesOutputs, ok := outputs["values"]; ok {
		recipeOutput.Values = valuesOutputs.Value.(map[string]interface{})
	}

	return &recipeOutput, nil
}
