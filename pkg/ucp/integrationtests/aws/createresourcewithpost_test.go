// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package aws

// Tests that test with Mock RP functionality and UCP Server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol/types"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CreateAWSResourceWithPost(t *testing.T) {
	ucp, ucpClient, cloudcontrolClient, cloudformationClient := initializeTest(t)

	primaryIdentifiers := map[string]interface{}{
		"primaryIdentifier": []interface{}{
			"/properties/Name",
		},
	}
	serialized, err := json.Marshal(primaryIdentifiers)
	require.NoError(t, err)
	output := cloudformation.DescribeTypeOutput{
		TypeName: aws.String("AWS::Kinesis::Stream"),
		Schema:   to.Ptr(string(serialized)),
	}

	cloudformationClient.EXPECT().DescribeType(gomock.Any(), gomock.Any()).Return(&output, nil)

	cloudcontrolClient.EXPECT().GetResource(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, params *cloudcontrol.GetResourceInput, optFns ...func(*cloudcontrol.Options)) (*cloudcontrol.GetResourceOutput, error) {
		notfound := types.ResourceNotFoundException{
			Message: to.Ptr("Resource not found"),
		}
		return nil, &notfound
	})

	cloudcontrolClient.EXPECT().CreateResource(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, params *cloudcontrol.CreateResourceInput, optFns ...func(*cloudcontrol.Options)) (*cloudcontrol.CreateResourceOutput, error) {
		output := cloudcontrol.CreateResourceOutput{
			ProgressEvent: &types.ProgressEvent{
				OperationStatus: types.OperationStatusSuccess,
				RequestToken:    to.Ptr(testAWSRequestToken),
				Identifier:      to.Ptr("testStream"),
			},
		}
		return &output, nil
	})

	requestBody := map[string]interface{}{
		"properties": map[string]interface{}{
			"Name":                 "testStream",
			"RetentionPeriodHours": 178,
			"ShardCount":           3,
		},
	}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	createRequest, err := http.NewRequest(http.MethodPost, ucp.URL+basePath+testProxyRequestAWSCollectionPath+"/:put", bytes.NewBuffer(body))
	require.NoError(t, err)
	createResponse, err := ucpClient.httpClient.Do(createRequest)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, createResponse.StatusCode)
}
