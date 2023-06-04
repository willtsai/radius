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

package providers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/project-radius/radius/pkg/azure/tokencredentials"
	"github.com/project-radius/radius/pkg/sdk"
	"github.com/project-radius/radius/pkg/ucp/credentials"
	"github.com/project-radius/radius/pkg/ucp/resources"
	"github.com/project-radius/radius/pkg/ucp/secret/provider"
)

const (
	AzureProviderName    = "azurerm"
	AzureProviderSoruce  = "hashicorp/azurerm"
	AzureProviderVersion = "~> 3.0.2" // TODO make it configurable
)

// Returns the Terraform provider configuration for Azure provider.
// https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs
func BuildAzureProviderConfig(ctx context.Context, ucpConn *sdk.Connection, scope string) (map[string]interface{}, string, error) {
	logger := logr.FromContextOrDiscard(ctx)
	if scope == "" {
		return map[string]interface{}{}, "", nil
	}

	subscriptionID, resourceGroup, err := parseAzureScope(scope)
	if err != nil {
		return nil, "", err
	}

	credentials, err := fetchAzureCredentials(ucpConn)
	if err != nil {
		return nil, "", err
	}
	logger.Info(fmt.Sprintf("Fetched Azure credentials for client id %q", credentials.ClientID))

	azureConfig := map[string]interface{}{
		"subscription_id": subscriptionID,
		"client_id":       credentials.ClientID,
		"client_secret":   credentials.ClientSecret,
		"tenant_id":       credentials.TenantID,
		"features":        map[string]interface{}{},
	}

	return azureConfig, resourceGroup, nil
}

func parseAzureScope(scope string) (subscriptionID string, resourceGroup string, err error) {
	parsedScope, err := resources.Parse(scope)
	if err != nil {
		return "", "", fmt.Errorf("error parsing Azure scope: %w", err)
	}

	for _, segment := range parsedScope.ScopeSegments() {
		if strings.EqualFold(segment.Type, resources.SubscriptionsSegment) {
			subscriptionID = segment.Name
		}

		if strings.EqualFold(segment.Type, resources.ResourceGroupsSegment) {
			resourceGroup = segment.Name
		}
	}

	return
}

func fetchAzureCredentials(ucpConn *sdk.Connection) (*credentials.AzureCredential, error) {
	// TODO add client.Get to validate that credentials exist

	secretProvider := provider.NewSecretProvider(provider.SecretProviderOptions{
		Provider: provider.TypeKubernetesSecret,
	})
	azureCredentialProvider, err := credentials.NewAzureCredentialProvider(secretProvider, *ucpConn, &tokencredentials.AnonymousCredential{})
	if err != nil {
		return nil, fmt.Errorf("error creating Azure credential provider: %w", err)
	}

	credentials, err := azureCredentialProvider.Fetch(context.Background(), credentials.AzureCloud, defaultCredential)
	if err != nil {
		return nil, fmt.Errorf("error fetching Azure credentials: %w", err)
	}

	if credentials.ClientID == "" || credentials.ClientSecret == "" || credentials.TenantID == "" {
		return nil, errors.New("credentials are required to create Azure resources through Recipe. Use `rad credential register azure` to register Azure credentials")
	}

	return credentials, nil
}
