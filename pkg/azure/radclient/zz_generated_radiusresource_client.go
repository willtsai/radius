//go:build go1.16
// +build go1.16

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

package radclient

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	armruntime "github.com/Azure/azure-sdk-for-go/sdk/azcore/arm/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"net/http"
	"net/url"
	"strings"
)

// RadiusResourceClient contains the methods for the RadiusResource group.
// Don't use this type directly, use NewRadiusResourceClient() instead.
type RadiusResourceClient struct {
	ep string
	pl runtime.Pipeline
	subscriptionID string
}

// NewRadiusResourceClient creates a new instance of RadiusResourceClient with the specified values.
func NewRadiusResourceClient(con *arm.Connection, subscriptionID string) *RadiusResourceClient {
	return &RadiusResourceClient{ep: con.Endpoint(), pl: con.NewPipeline(module, version), subscriptionID: subscriptionID}
}

// BeginDelete - Deletes a RadiusResource resource.
// If the operation fails it returns the *ErrorResponse error type.
func (client *RadiusResourceClient) BeginDelete(ctx context.Context, resourceGroupName string, applicationName string, radiusResourceType string, radiusResourceName string, options *RadiusResourceBeginDeleteOptions) (RadiusResourceDeletePollerResponse, error) {
	resp, err := client.deleteOperation(ctx, resourceGroupName, applicationName, radiusResourceType, radiusResourceName, options)
	if err != nil {
		return RadiusResourceDeletePollerResponse{}, err
	}
	result := RadiusResourceDeletePollerResponse{
		RawResponse: resp,
	}
	pt, err := armruntime.NewPoller("RadiusResourceClient.Delete", "location", resp, 	client.pl, client.deleteHandleError)
	if err != nil {
		return RadiusResourceDeletePollerResponse{}, err
	}
	result.Poller = &RadiusResourceDeletePoller {
		pt: pt,
	}
	return result, nil
}

// Delete - Deletes a RadiusResource resource.
// If the operation fails it returns the *ErrorResponse error type.
func (client *RadiusResourceClient) deleteOperation(ctx context.Context, resourceGroupName string, applicationName string, radiusResourceType string, radiusResourceName string, options *RadiusResourceBeginDeleteOptions) (*http.Response, error) {
	req, err := client.deleteCreateRequest(ctx, resourceGroupName, applicationName, radiusResourceType, radiusResourceName, options)
	if err != nil {
		return nil, err
	}
	resp, err := 	client.pl.Do(req)
	if err != nil {
		return nil, err
	}
	if !runtime.HasStatusCode(resp, http.StatusAccepted, http.StatusNoContent) {
		return nil, client.deleteHandleError(resp)
	}
	 return resp, nil
}

// deleteCreateRequest creates the Delete request.
func (client *RadiusResourceClient) deleteCreateRequest(ctx context.Context, resourceGroupName string, applicationName string, radiusResourceType string, radiusResourceName string, options *RadiusResourceBeginDeleteOptions) (*policy.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.CustomProviders/resourceProviders/radiusv3/Application/{applicationName}/{radiusResourceType}/{radiusResourceName}"
	if client.subscriptionID == "" {
		return nil, errors.New("parameter client.subscriptionID cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionId}", url.PathEscape(client.subscriptionID))
	if resourceGroupName == "" {
		return nil, errors.New("parameter resourceGroupName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{resourceGroupName}", url.PathEscape(resourceGroupName))
	if applicationName == "" {
		return nil, errors.New("parameter applicationName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{applicationName}", url.PathEscape(applicationName))
	if radiusResourceType == "" {
		return nil, errors.New("parameter radiusResourceType cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{radiusResourceType}", url.PathEscape(radiusResourceType))
	if radiusResourceName == "" {
		return nil, errors.New("parameter radiusResourceName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{radiusResourceName}", url.PathEscape(radiusResourceName))
	req, err := runtime.NewRequest(ctx, http.MethodDelete, runtime.JoinPaths(	client.ep, urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2018-09-01-preview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")
	return req, nil
}

// deleteHandleError handles the Delete error response.
func (client *RadiusResourceClient) deleteHandleError(resp *http.Response) error {
	body, err := runtime.Payload(resp)
	if err != nil {
		return runtime.NewResponseError(err, resp)
	}
		errType := ErrorResponse{raw: string(body)}
	if err := runtime.UnmarshalAsJSON(resp, &errType); err != nil {
		return runtime.NewResponseError(fmt.Errorf("%s\n%s", string(body), err), resp)
	}
	return runtime.NewResponseError(&errType, resp)
}

// Get - Gets a RadiusResource resource by name.
// If the operation fails it returns the *ErrorResponse error type.
func (client *RadiusResourceClient) Get(ctx context.Context, resourceGroupName string, applicationName string, radiusResourceType string, radiusResourceName string, options *RadiusResourceGetOptions) (RadiusResourceGetResponse, error) {
	req, err := client.getCreateRequest(ctx, resourceGroupName, applicationName, radiusResourceType, radiusResourceName, options)
	if err != nil {
		return RadiusResourceGetResponse{}, err
	}
	resp, err := 	client.pl.Do(req)
	if err != nil {
		return RadiusResourceGetResponse{}, err
	}
	if !runtime.HasStatusCode(resp, http.StatusOK) {
		return RadiusResourceGetResponse{}, client.getHandleError(resp)
	}
	return client.getHandleResponse(resp)
}

// getCreateRequest creates the Get request.
func (client *RadiusResourceClient) getCreateRequest(ctx context.Context, resourceGroupName string, applicationName string, radiusResourceType string, radiusResourceName string, options *RadiusResourceGetOptions) (*policy.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.CustomProviders/resourceProviders/radiusv3/Application/{applicationName}/{radiusResourceType}/{radiusResourceName}"
	if client.subscriptionID == "" {
		return nil, errors.New("parameter client.subscriptionID cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionId}", url.PathEscape(client.subscriptionID))
	if resourceGroupName == "" {
		return nil, errors.New("parameter resourceGroupName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{resourceGroupName}", url.PathEscape(resourceGroupName))
	if applicationName == "" {
		return nil, errors.New("parameter applicationName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{applicationName}", url.PathEscape(applicationName))
	if radiusResourceType == "" {
		return nil, errors.New("parameter radiusResourceType cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{radiusResourceType}", url.PathEscape(radiusResourceType))
	if radiusResourceName == "" {
		return nil, errors.New("parameter radiusResourceName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{radiusResourceName}", url.PathEscape(radiusResourceName))
	req, err := runtime.NewRequest(ctx, http.MethodGet, runtime.JoinPaths(	client.ep, urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	if options != nil && options.ResourceSubscriptionID != nil {
		reqQP.Set("ResourceSubscriptionID", *options.ResourceSubscriptionID)
	}
	if options != nil && options.ResourceGroup != nil {
		reqQP.Set("ResourceGroup", *options.ResourceGroup)
	}
	if options != nil && options.ResourceType != nil {
		reqQP.Set("ResourceType", *options.ResourceType)
	}
	reqQP.Set("api-version", "2018-09-01-preview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")
	return req, nil
}

// getHandleResponse handles the Get response.
func (client *RadiusResourceClient) getHandleResponse(resp *http.Response) (RadiusResourceGetResponse, error) {
	result := RadiusResourceGetResponse{RawResponse: resp}
	if err := runtime.UnmarshalAsJSON(resp, &result.RadiusResource); err != nil {
		return RadiusResourceGetResponse{}, err
	}
	return result, nil
}

// getHandleError handles the Get error response.
func (client *RadiusResourceClient) getHandleError(resp *http.Response) error {
	body, err := runtime.Payload(resp)
	if err != nil {
		return runtime.NewResponseError(err, resp)
	}
		errType := ErrorResponse{raw: string(body)}
	if err := runtime.UnmarshalAsJSON(resp, &errType); err != nil {
		return runtime.NewResponseError(fmt.Errorf("%s\n%s", string(body), err), resp)
	}
	return runtime.NewResponseError(&errType, resp)
}

// List - List the RadiusResource resources deployed in the application.
// If the operation fails it returns the *ErrorResponse error type.
func (client *RadiusResourceClient) List(ctx context.Context, resourceGroupName string, applicationName string, options *RadiusResourceListOptions) (RadiusResourceListResponse, error) {
	req, err := client.listCreateRequest(ctx, resourceGroupName, applicationName, options)
	if err != nil {
		return RadiusResourceListResponse{}, err
	}
	resp, err := 	client.pl.Do(req)
	if err != nil {
		return RadiusResourceListResponse{}, err
	}
	if !runtime.HasStatusCode(resp, http.StatusOK) {
		return RadiusResourceListResponse{}, client.listHandleError(resp)
	}
	return client.listHandleResponse(resp)
}

// listCreateRequest creates the List request.
func (client *RadiusResourceClient) listCreateRequest(ctx context.Context, resourceGroupName string, applicationName string, options *RadiusResourceListOptions) (*policy.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.CustomProviders/resourceProviders/radiusv3/Application/{applicationName}/RadiusResource"
	if client.subscriptionID == "" {
		return nil, errors.New("parameter client.subscriptionID cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionId}", url.PathEscape(client.subscriptionID))
	if resourceGroupName == "" {
		return nil, errors.New("parameter resourceGroupName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{resourceGroupName}", url.PathEscape(resourceGroupName))
	if applicationName == "" {
		return nil, errors.New("parameter applicationName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{applicationName}", url.PathEscape(applicationName))
	req, err := runtime.NewRequest(ctx, http.MethodGet, runtime.JoinPaths(	client.ep, urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2018-09-01-preview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")
	return req, nil
}

// listHandleResponse handles the List response.
func (client *RadiusResourceClient) listHandleResponse(resp *http.Response) (RadiusResourceListResponse, error) {
	result := RadiusResourceListResponse{RawResponse: resp}
	if err := runtime.UnmarshalAsJSON(resp, &result.RadiusResourceList); err != nil {
		return RadiusResourceListResponse{}, err
	}
	return result, nil
}

// listHandleError handles the List error response.
func (client *RadiusResourceClient) listHandleError(resp *http.Response) error {
	body, err := runtime.Payload(resp)
	if err != nil {
		return runtime.NewResponseError(err, resp)
	}
		errType := ErrorResponse{raw: string(body)}
	if err := runtime.UnmarshalAsJSON(resp, &errType); err != nil {
		return runtime.NewResponseError(fmt.Errorf("%s\n%s", string(body), err), resp)
	}
	return runtime.NewResponseError(&errType, resp)
}
