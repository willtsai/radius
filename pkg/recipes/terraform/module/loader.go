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

package module

import (
	"fmt"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

func Load(workingDir, recipeName string) (map[string]*tfconfig.ProviderRequirement, map[string]*tfconfig.Variable, error) {
	// Load Terraform module downloaded in the working directory
	mod, diags := tfconfig.LoadModule(workingDir + "/.terraform/modules/" + recipeName)
	if diags.HasErrors() {
		return nil, nil, fmt.Errorf("error loading the module: %w", diags.Err())
	}

	return mod.RequiredProviders, mod.Variables, nil
}
