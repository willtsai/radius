//go:build go1.18
// +build go1.18

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.
// DO NOT EDIT.

package v20220315privatepreview

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	armruntime "github.com/Azure/azure-sdk-for-go/sdk/azcore/arm/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"net/http"
	"net/url"
	"strings"
)

// CredentialClient contains the methods for the Credential group.
// Don't use this type directly, use NewCredentialClient() instead.
type CredentialClient struct {
	host string
	pl runtime.Pipeline
}

// NewCredentialClient creates a new instance of CredentialClient with the specified values.
// credential - used to authorize requests. Usually a credential from azidentity.
// options - pass nil to accept the default values.
func NewCredentialClient(credential azcore.TokenCredential, options *arm.ClientOptions) (*CredentialClient, error) {
	if options == nil {
		options = &arm.ClientOptions{}
	}
	ep := cloud.AzurePublic.Services[cloud.ResourceManager].Endpoint
	if c, ok := options.Cloud.Services[cloud.ResourceManager]; ok {
		ep = c.Endpoint
	}
	pl, err := armruntime.NewPipeline(moduleName, moduleVersion, credential, runtime.PipelineOptions{}, options)
	if err != nil {
		return nil, err
	}
	client := &CredentialClient{
		host: ep,
pl: pl,
	}
	return client, nil
}

// CreateOrUpdate - Create or update a Credential.
// If the operation fails it returns an *azcore.ResponseError type.
// Generated from API version 2022-03-15-privatepreview
// planeType - The type of the plane
// planeName - The name of the plane
// credentialName - The name of the credential
// credential - Credential details
// options - CredentialClientCreateOrUpdateOptions contains the optional parameters for the CredentialClient.CreateOrUpdate
// method.
func (client *CredentialClient) CreateOrUpdate(ctx context.Context, planeType string, planeName string, credentialName string, credential CredentialResource, options *CredentialClientCreateOrUpdateOptions) (CredentialClientCreateOrUpdateResponse, error) {
	req, err := client.createOrUpdateCreateRequest(ctx, planeType, planeName, credentialName, credential, options)
	if err != nil {
		return CredentialClientCreateOrUpdateResponse{}, err
	}
	resp, err := client.pl.Do(req)
	if err != nil {
		return CredentialClientCreateOrUpdateResponse{}, err
	}
	if !runtime.HasStatusCode(resp, http.StatusOK, http.StatusCreated) {
		return CredentialClientCreateOrUpdateResponse{}, runtime.NewResponseError(resp)
	}
	return client.createOrUpdateHandleResponse(resp)
}

// createOrUpdateCreateRequest creates the CreateOrUpdate request.
func (client *CredentialClient) createOrUpdateCreateRequest(ctx context.Context, planeType string, planeName string, credentialName string, credential CredentialResource, options *CredentialClientCreateOrUpdateOptions) (*policy.Request, error) {
	urlPath := "/planes/{planeType}/{planeName}/providers/System.Azure/credentials/{credentialName}"
	if planeType == "" {
		return nil, errors.New("parameter planeType cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{planeType}", url.PathEscape(planeType))
	if planeName == "" {
		return nil, errors.New("parameter planeName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{planeName}", url.PathEscape(planeName))
	if credentialName == "" {
		return nil, errors.New("parameter credentialName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{credentialName}", url.PathEscape(credentialName))
	req, err := runtime.NewRequest(ctx, http.MethodPut, runtime.JoinPaths(client.host, urlPath))
	if err != nil {
		return nil, err
	}
	req.Raw().Header["Accept"] = []string{"application/json"}
	return req, runtime.MarshalAsJSON(req, credential)
}

// createOrUpdateHandleResponse handles the CreateOrUpdate response.
func (client *CredentialClient) createOrUpdateHandleResponse(resp *http.Response) (CredentialClientCreateOrUpdateResponse, error) {
	result := CredentialClientCreateOrUpdateResponse{}
	if err := runtime.UnmarshalAsJSON(resp, &result.CredentialResource); err != nil {
		return CredentialClientCreateOrUpdateResponse{}, err
	}
	return result, nil
}
