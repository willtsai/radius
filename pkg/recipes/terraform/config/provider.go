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
	"fmt"

	"github.com/project-radius/radius/pkg/corerp/datamodel"
	tf_providers "github.com/project-radius/radius/pkg/recipes/terraform/providers"
	"github.com/project-radius/radius/pkg/sdk"
)

// Returns source and version for the required providers.
func getRequiredProviders(providers []TerraformProviderMetadata) map[string]ProviderDefinition {
	requiredProviders := make(map[string]ProviderDefinition)

	for _, provider := range providers {
		switch provider.Type {
		case tf_providers.AWSProviderName:
			requiredProviders[tf_providers.AWSProviderName] = ProviderDefinition{
				Source:  tf_providers.AWSProviderSource,
				Version: tf_providers.AWSProviderVersion,
			}
		case tf_providers.AzureProviderName:
			requiredProviders[tf_providers.AzureProviderName] = ProviderDefinition{
				Source:  tf_providers.AzureProviderSoruce,
				Version: tf_providers.AzureProviderVersion,
			}
		case tf_providers.KubernetesProviderName:
			requiredProviders[tf_providers.KubernetesProviderName] = ProviderDefinition{
				Source:  tf_providers.KubernetesProviderSource,
				Version: tf_providers.KubernetesProviderVersion,
			}
		}
	}

	return requiredProviders
}

// Returns Terraform provider configurations for the required providers and
// Azure resource group tied to Environment's scope if Azure provider is required.
func getProviderConfigs(ctx context.Context, ucpConn *sdk.Connection, providers []TerraformProviderMetadata, envProviders *datamodel.Providers) (map[string]interface{}, string, error) {
	providerConfigs := make(map[string]interface{})
	resourceGroup := ""

	for _, provider := range providers {
		switch provider.Type {
		case tf_providers.AWSProviderName:
			config, err := tf_providers.BuildAWSProviderConfig(ctx, ucpConn, envProviders.AWS.Scope)
			if err != nil {
				return nil, "", err
			}
			providerConfigs[tf_providers.AWSProviderName] = config
		case tf_providers.AzureProviderName:
			config, rg, err := tf_providers.BuildAzureProviderConfig(ctx, ucpConn, envProviders.Azure.Scope)
			if err != nil {
				return nil, "", err
			}
			resourceGroup = rg
			providerConfigs[tf_providers.AzureProviderName] = config
		case tf_providers.KubernetesProviderName:
			config, err := tf_providers.BuildKubernetesProviderConfig(ctx)
			if err != nil {
				return nil, "", err
			}
			providerConfigs[tf_providers.KubernetesProviderName] = config
		default:
			return nil, "", fmt.Errorf("unsupported provider type: %s", provider.Type)
		}
	}

	return providerConfigs, resourceGroup, nil
}
