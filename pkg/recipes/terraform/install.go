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
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
)

const (
	installSubDir = "install"
)

// Installs Terraform to the specified directory
// Returns the path to the installed Terraform binary
func Install(ctx context.Context, tfDir string) (string, error) {
	logger := logr.FromContextOrDiscard(ctx)

	// Create Terraform installation directory
	installDir := filepath.Join(tfDir, installSubDir)
	logger.Info(fmt.Sprintf("Installing Terraform in directory: %q", installDir))

	if err := os.MkdirAll(installDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory for terraform installation for resource: %w", err)
	}

	// Install Terraform
	// Using latest version, revisit this if we want to use a specific version.
	installer := &releases.LatestVersion{
		Product:    product.Terraform,
		InstallDir: installDir,
	}

	// We should look into checking if an exsiting installation of Terraform is available.
	// For initial iteration we will always install Terraform. Optimizations can be made later.
	execPath, err := installer.Install(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to install terraform: %w", err)
	}

	return execPath, nil
}
