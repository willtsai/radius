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

package driver

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/project-radius/radius/pkg/recipes"
	"github.com/project-radius/radius/pkg/recipes/terraform"
	"github.com/project-radius/radius/pkg/sdk"
	"github.com/project-radius/radius/pkg/ucp/util"
)

const (
	// terraformDirRoot = "/terraform" // TODO make it configurable
	terraformDirRoot = "/tmp/terraform" // Temporary for local testing
)

var _ Driver = (*terraformDriver)(nil)

// NewTerraformDriver creates a new instance of driver to execute a Terraform recipe.
func NewTerraformDriver(ucpConn sdk.Connection) Driver {
	return &terraformDriver{UcpConn: ucpConn}
}

type terraformDriver struct {
	UcpConn sdk.Connection
}

// Execute executes a Terraform recipe by using the Terraform CLI through terraform-exec
func (d *terraformDriver) Execute(ctx context.Context, configuration recipes.Configuration, recipe recipes.ResourceMetadata, definition recipes.EnvironmentDefinition) (*recipes.RecipeOutput, error) {
	logger := logr.FromContextOrDiscard(ctx)

	logger.Info(fmt.Sprintf("Deploying recipe: %q, template: %q", recipe.Name, definition.TemplatePath))
	terraformDir := terraformDirRoot + "/" + util.NormalizeStringToLower(recipe.ResourceID) + "-" + uuid.NewString()
	recipeOutputs, err := terraform.Deploy(ctx, &d.UcpConn, terraformDir, &configuration, &recipe, &definition)
	if err != nil {
		cleanup(ctx, terraformDir)
		return nil, err
	}

	// Cleanup Terraform directories
	cleanup(ctx, terraformDir)

	return recipeOutputs, nil
}

func cleanup(ctx context.Context, tfDir string) {
	logger := logr.FromContextOrDiscard(ctx)
	// TODO Delete Kubernetes Secrets
	// k8s.DeleteSecrets(ctx, tfDir)
	err := os.RemoveAll(tfDir)
	if err != nil {
		logger.Error(err, "Failed to remove Terraform installation directory")
	}
}
