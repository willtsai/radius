// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package setup

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/cobra"

	radAWS "github.com/project-radius/radius/pkg/cli/aws"
	"github.com/project-radius/radius/pkg/cli/prompt"
)

const (
	AWSProviderFlagName                = "provider-aws"
	AWSProviderAccessKeyIdFlagName     = "provider-aws-access-key-id"
	AWSProviderSecretAccessKeyFlagName = "provider-aws-secret-access-key"
	AWSProviderRegionFlagName          = "provider-aws-region"
)

func RegisterPersistentAWSProviderArgs(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolP(
		AWSProviderFlagName,
		"",
		false,
		"Add AWS provider for cloud resources",
	)
	cmd.PersistentFlags().String(
		AWSProviderAccessKeyIdFlagName,
		"",
		"Specifies an AWS access key associated with an IAM user or role",
	)
	cmd.PersistentFlags().String(
		AWSProviderSecretAccessKeyFlagName,
		"",
		"Specifies the secret key associated with the access key. This is essentially the \"password\" for the access key",
	)
	cmd.PersistentFlags().String(
		AWSProviderRegionFlagName,
		"",
		"Specifies the region to be used for resources deployed by this provider",
	)
}

func ParseAWSProviderFromArgs(cmd *cobra.Command, interactive bool) (*radAWS.Provider, error) {
	if interactive {
		return parseAWSProviderInteractive(cmd)
	}
	return parseAWSProviderNonInteractive(cmd)

}

func parseAWSProviderInteractive(cmd *cobra.Command) (*radAWS.Provider, error) {
	ctx := cmd.Context()

	addAWSCred, err := prompt.ConfirmWithDefault("Add AWS provider for cloud resources [y/N]?", prompt.No)
	if err != nil {
		return nil, err
	}
	if !addAWSCred {
		return nil, nil
	}

	region, err := prompt.Text("Enter the region you would like to use to deploy AWS resources:", prompt.EmptyValidator)
	if err != nil {
		return nil, err
	}

	keyID, err := prompt.Text("Enter the IAM Access Key ID:", prompt.EmptyValidator)
	if err != nil {
		return nil, err
	}

	secretAccessKey, err := prompt.Text("Enter your IAM Secret Access Keys:", prompt.EmptyValidator)
	if err != nil {
		return nil, err
	}

	return verifyAWSCredentials(ctx, keyID, secretAccessKey, region)
}

func parseAWSProviderNonInteractive(cmd *cobra.Command) (*radAWS.Provider, error) {
	ctx := cmd.Context()

	addAWSProvider, err := cmd.Flags().GetBool(AWSProviderFlagName)
	if err != nil {
		return nil, err
	}
	if !addAWSProvider {
		return nil, nil
	}

	keyID, err := cmd.Flags().GetString(AWSProviderAccessKeyIdFlagName)
	if err != nil {
		return nil, err
	}

	secretAccessKey, err := cmd.Flags().GetString(AWSProviderSecretAccessKeyFlagName)
	if err != nil {
		return nil, err
	}

	region, err := cmd.Flags().GetString(AWSProviderRegionFlagName)
	if err != nil {
		return nil, err
	}

	return verifyAWSCredentials(ctx, keyID, secretAccessKey, region)
}

func verifyAWSCredentials(ctx context.Context, keyID string, secretAccessKey string, region string) (*radAWS.Provider, error) {
	credentialsProvider := credentials.NewStaticCredentialsProvider(keyID, secretAccessKey, "")
	stsClient := sts.New(sts.Options{
		Region:      region,
		Credentials: credentialsProvider,
	})
	result, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("AWS credential verification failed: %s", err.Error())
	}

	return &radAWS.Provider{
		AccessKeyId:     keyID,
		SecretAccessKey: secretAccessKey,
		TargetRegion:    region,
		AccountId:       *result.Account,
	}, nil
}
