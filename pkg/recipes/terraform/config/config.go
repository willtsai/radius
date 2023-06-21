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

package config

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/project-radius/radius/pkg/recipes"
	"github.com/project-radius/radius/pkg/sdk"
	"github.com/project-radius/radius/pkg/ucp/resources"
)

const (
	terraformVersion = "1.1.0" // TODO make it configurable
)

var (
	// TODO Outputs mapping between expected Radius resource outputs and module outputs.
	// This is hardcoded for now until integrated with the recipe definition (should be provided at recipe registration)
	// https://dev.azure.com/azure-octo/Incubations/_workitems/edit/7673
	outputMappingValues = map[string]string{
		"host": "host",
		"port": "port",
	}

	outputMappingSecrets = map[string]string{
		"connectionString": "connectionString",
	}
)

func GenerateConfigFiles(ctx context.Context, ucpConn *sdk.Connection, tfProviders []TerraformProviderMetadata, configuration *recipes.Configuration, recipe *recipes.ResourceMetadata, definition *recipes.EnvironmentDefinition, workingDir string) error {
	err := generateMainConfig(ctx, ucpConn, tfProviders, configuration, recipe, definition, workingDir)
	if err != nil {
		return err
	}

	err = generateOutputConfig(recipe.Name, workingDir)
	if err != nil {
		return err
	}

	return nil
}

// Generate Terraform configuration in JSON format for required providers and modules, and write it
// to a file in the specified working directory. This JSON configuration is needed to initialize
// and apply Terraform modules. See https://www.terraform.io/docs/language/syntax/json.html
// for more information on the JSON syntax for Terraform configuration.
// templatePath is the path to the Terraform module source, e.g. "Azure/cosmosdb/azurerm".
func generateMainConfig(ctx context.Context, ucpConn *sdk.Connection, tfProviders []TerraformProviderMetadata, configuration *recipes.Configuration, recipe *recipes.ResourceMetadata, definition *recipes.EnvironmentDefinition, workingDir string) error {
	providerConfigs, err := getProviderConfigs(ctx, ucpConn, tfProviders, &configuration.Providers)
	if err != nil {
		return err
	}

	moduleData, err := generateModuleData(ctx, definition.TemplatePath, recipe.Parameters, definition.Parameters)
	if err != nil {
		return err
	}

	backend, err := generateKubernetesBackendConfig(recipe.ResourceID, configuration.Runtime.Kubernetes.Namespace)
	if err != nil {
		return err
	}

	tfConfig := TerraformConfig{
		Terraform: TerraformDefinition{
			RequiredProviders: getRequiredProviders(tfProviders),
			Backend:           backend,
			RequiredVersion:   ">= " + terraformVersion,
		},
		Provider: providerConfigs,
		Module: map[string]interface{}{
			recipe.Name: moduleData,
		},
	}

	// Convert the Terraform config to JSON
	jsonData, err := json.MarshalIndent(tfConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	// Write the JSON data to a file in the working directory.
	// JSON configuration syntax for Terraform requires the file to be named with .tf.json suffix.
	// https://developer.hashicorp.com/terraform/language/syntax/json
	configFilePath := fmt.Sprintf("%s/main.tf.json", workingDir)
	file, err := os.Create(configFilePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

func generateModuleData(ctx context.Context, source string, resourceParams, envParams map[string]interface{}) (map[string]interface{}, error) {
	moduleConfig := map[string]interface{}{
		ModuleSourceKey:  source,
		ModuleVersionKey: "1.0.0", // This is module version and needs to come from the user - TODO: https://dev.azure.com/azure-octo/Incubations/_workitems/edit/8137
	}

	// Populate recipe parameters
	// Resource parameter overrides parameter set on the recipe definition in the environment,
	// if same parameter is defined in both environment and resource recipe metadata.
	for key, value := range envParams {
		moduleConfig[key] = value
	}

	for key, value := range resourceParams {
		moduleConfig[key] = value
	}

	// TODO add context parameter if present in the module
	// https://dev.azure.com/azure-octo/Incubations/_workitems/edit/8399

	return moduleConfig, nil
}

func generateKubernetesBackendConfig(resourceID, namespace string) (map[string]interface{}, error) {
	secretSuffix, err := generateSecretSuffix(resourceID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"kubernetes": map[string]interface{}{
			"config_path":   "~/.kube/config",
			"secret_suffix": secretSuffix,
			"namespace":     namespace,
		},
	}, nil
}

func generateSecretSuffix(resourceID string) (string, error) {
	parsedID, err := resources.Parse(resourceID)
	if err != nil {
		return "", err
	}

	name := parsedID.Name()
	maxResourceNameLen := 22
	if len(name) >= maxResourceNameLen {
		name = name[:maxResourceNameLen]
	}

	hasher := sha1.New()
	_, _ = hasher.Write([]byte(strings.ToLower(parsedID.String())))
	hash := hasher.Sum(nil)

	suffix := fmt.Sprintf("%s.%x", name, hash)

	return suffix, nil
}

// Generate output.tf file that references module outputs to populate expected Radius resource outputs.
// Outputs of modules are accessible through this format: module.<MODULE NAME>.<OUTPUT NAME>
func generateOutputConfig(recipeName, workingDir string) error {
	outputFilePath := fmt.Sprintf("%s/output.tf", workingDir)
	// Create a new output file object
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("error creating output.tf file: %w", err)
	}

	// Write the `output` blocks
	fmt.Fprintf(outputFile, "output \"values\" {\n")
	fmt.Fprintf(outputFile, "  value = {\n")
	for radiusResourceOutput, moduleOutput := range outputMappingValues {
		fmt.Fprintf(outputFile, "    %s = module.%s.%s,\n", radiusResourceOutput, recipeName, moduleOutput)
	}
	fmt.Fprintf(outputFile, "  }\n")
	fmt.Fprintf(outputFile, "}\n\n")

	fmt.Fprintf(outputFile, "output \"secrets\" {\n")
	fmt.Fprintf(outputFile, "  value = {\n")
	for radiusResourceOutput, moduleOutput := range outputMappingSecrets {
		fmt.Fprintf(outputFile, "    %s = module.%s.%s,\n", radiusResourceOutput, recipeName, moduleOutput)
	}
	fmt.Fprintf(outputFile, "  }\n")
	fmt.Fprintf(outputFile, "  sensitive = true\n")
	fmt.Fprintf(outputFile, "}\n")

	err = outputFile.Close()
	if err != nil {
		return fmt.Errorf("error closing the file: %w", err)
	}

	return nil
}
