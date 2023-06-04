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

	"github.com/go-logr/logr"
	"github.com/project-radius/radius/pkg/recipes"
	"github.com/project-radius/radius/pkg/sdk"
)

func Deploy(ctx context.Context, ucpConn *sdk.Connection, tfDir string, configuration *recipes.Configuration, recipe *recipes.Metadata, definition *recipes.Definition) (*recipes.RecipeOutput, error) {
	logger := logr.FromContextOrDiscard(ctx)

	// Install Terraform
	execPath, err := Install(ctx, tfDir)
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("Terraform installation path: %q", execPath))

	// TODO Create Working Directory

	// TODO Generate Terraform json config in the working directory

	// TODO Run TF Init and Apply

	// TODO Retun recipe output

	return nil, nil
}
