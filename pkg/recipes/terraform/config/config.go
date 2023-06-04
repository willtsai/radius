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
	"encoding/json"
	"errors"
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
	outputValues = map[string]string{
		"host": "host",
		"port": "port",
	}

	outputSecrets = map[string]string{
		"connectionString": "connectionString",
	}
)

func GenerateConfigFiles(ctx context.Context, ucpConn *sdk.Connection, tfProviders []TerraformProviderMetadata, configuration *recipes.Configuration, recipe *recipes.Metadata, definition *recipes.Definition, workingDir string) error {
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
func generateMainConfig(ctx context.Context, ucpConn *sdk.Connection, tfProviders []TerraformProviderMetadata, configuration *recipes.Configuration, recipe *recipes.Metadata, definition *recipes.Definition, workingDir string) error {
	providerConfigs, resourceGroup, err := getProviderConfigs(ctx, ucpConn, tfProviders, &configuration.Providers)
	if err != nil {
		return err
	}

	moduleData, err := generateModuleDataFromParams(ctx, definition.TemplatePath, resourceGroup, recipe.Parameters, definition.Parameters)
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

func generateOutputConfig(recipeName, workingDir string) error {
	outputBlock := ""
	outputBlock += `output "values" {
		value = {
	  `
	for key, value := range outputValues {
		outputBlock += fmt.Sprintf("    %s = module.%s.%s,\n", key, recipeName, value)
	}
	outputBlock += `  }
}
`
	// Generate secrets block
	outputBlock += `output "secrets" {
		value = {
	  `
	for key, value := range outputSecrets {
		outputBlock += fmt.Sprintf("    %s = module.%s.%s,\n", key, recipeName, value)
	}
	outputBlock += `  }
		sensitive = true
	  }`

	outputFilePath := fmt.Sprintf("%s/output.tf", workingDir)
	file, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	_, err = file.Write([]byte(outputBlock))
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

func generateModuleDataFromParams(ctx context.Context, source, resourceGroup string, devParams, operatorParams map[string]interface{}) (map[string]interface{}, error) {
	moduleConfig := map[string]interface{}{
		"source":  source,
		"version": "1.0.0",
	}

	for key, value := range operatorParams {
		moduleConfig[key] = value
	}

	for key, value := range devParams {
		moduleConfig[key] = value
	}

	// Workaround for now to pass resource group param name until the design is finalized
	if rgParam, ok := moduleConfig["resourceGroupParamName"]; ok {
		if resourceGroup == "" {
			return nil, errors.New("azure provider is not registered to the environment")
		}
		moduleConfig[rgParam.(string)] = resourceGroup
	}
	delete(moduleConfig, "resourceGroupParamName")

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

	var names []string
	for _, segment := range parsedID.ScopeSegments() {
		names = append(names, segment.Name)
	}
	suffix := strings.Join(names, "-")

	return suffix, nil
}
