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

	"github.com/go-logr/logr"
	"github.com/project-radius/radius/pkg/azure/tokencredentials"
	"github.com/project-radius/radius/pkg/sdk"
	"github.com/project-radius/radius/pkg/ucp/credentials"
	"github.com/project-radius/radius/pkg/ucp/resources"
	"github.com/project-radius/radius/pkg/ucp/secret/provider"
)

const (
	AWSProviderName    = "aws"
	AWSProviderSource  = "hashicorp/aws"
	AWSProviderVersion = "~> 4.0" // TODO make it configurable
)

// Returns the Terraform provider configuration for AWS provider.
// https://registry.terraform.io/providers/hashicorp/aws/latest/docs
func BuildAWSProviderConfig(ctx context.Context, ucpConn *sdk.Connection, scope string) (map[string]interface{}, error) {
	logger := logr.FromContextOrDiscard(ctx)
	if scope == "" {
		return map[string]interface{}{}, nil
	}

	account, region, err := parseAWSScope(scope)
	if err != nil {
		return nil, err
	}

	credentials, err := fetchAWSCredentials(ucpConn)
	if err != nil {
		return nil, err
	}
	logger.Info(fmt.Sprintf("Fetched AWS credentials for client id %q", credentials.AccessKeyID))

	awsConfig := map[string]interface{}{
		"region":              region,
		"allowed_account_ids": []string{account},
		"access_key":          credentials.AccessKeyID,
		"secret_key":          credentials.SecretAccessKey,
	}

	return awsConfig, nil
}

func parseAWSScope(scope string) (account string, region string, err error) {
	parsedScope, err := resources.Parse(scope)
	if err != nil {
		return "", "", fmt.Errorf("error parsing AWS scope: %w", err)
	}

	for _, segment := range parsedScope.ScopeSegments() {
		if segment.Type == resources.AccountsSegment {
			account = segment.Name
		}

		if segment.Type == resources.RegionsSegment {
			region = segment.Name
		}
	}
	return
}

func fetchAWSCredentials(ucpConn *sdk.Connection) (*credentials.AWSCredential, error) {
	// TODO add client.Get to validate that credentials exist

	secretProvider := provider.NewSecretProvider(provider.SecretProviderOptions{
		Provider: provider.TypeKubernetesSecret,
	})
	awsCredentialProvider, err := credentials.NewAWSCredentialProvider(secretProvider, *ucpConn, &tokencredentials.AnonymousCredential{})
	if err != nil {
		return nil, fmt.Errorf("error creating AWS credential provider: %w", err)
	}

	credentials, err := awsCredentialProvider.Fetch(context.Background(), credentials.AWSPublic, defaultCredential)
	if err != nil {
		return nil, fmt.Errorf("error fetching AWS credentials: %w", err)
	}

	if credentials.AccessKeyID == "" || credentials.SecretAccessKey == "" {
		return nil, errors.New("credentials are required to create AWS resources through Recipe. Use `rad credential register aws` to register AWS credentials")
	}

	return credentials, nil
}
