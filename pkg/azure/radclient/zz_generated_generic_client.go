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

// GenericClient contains the methods for the Generic group.
// Don't use this type directly, use NewGenericClient() instead.
type GenericClient struct {
	ep string
	pl runtime.Pipeline
	subscriptionID string
}

// NewGenericClient creates a new instance of GenericClient with the specified values.
func NewGenericClient(con *arm.Connection, subscriptionID string) *GenericClient {
	return &GenericClient{ep: con.Endpoint(), pl: con.NewPipeline(module, version), subscriptionID: subscriptionID}
}

// BeginCreateOrUpdate - Creates or updates a Generic resource.
// If the operation fails it returns the *ErrorResponse error type.
func (client *GenericClient) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, applicationName string, genericName string, parameters GenericResource, options *GenericBeginCreateOrUpdateOptions) (GenericCreateOrUpdatePollerResponse, error) {
	resp, err := client.createOrUpdate(ctx, resourceGroupName, applicationName, genericName, parameters, options)
	if err != nil {
		return GenericCreateOrUpdatePollerResponse{}, err
	}
	result := GenericCreateOrUpdatePollerResponse{
		RawResponse: resp,
	}
	pt, err := armruntime.NewPoller("GenericClient.CreateOrUpdate", "location", resp, 	client.pl, client.createOrUpdateHandleError)
	if err != nil {
		return GenericCreateOrUpdatePollerResponse{}, err
	}
	result.Poller = &GenericCreateOrUpdatePoller {
		pt: pt,
	}
	return result, nil
}

// CreateOrUpdate - Creates or updates a Generic resource.
// If the operation fails it returns the *ErrorResponse error type.
func (client *GenericClient) createOrUpdate(ctx context.Context, resourceGroupName string, applicationName string, genericName string, parameters GenericResource, options *GenericBeginCreateOrUpdateOptions) (*http.Response, error) {
	req, err := client.createOrUpdateCreateRequest(ctx, resourceGroupName, applicationName, genericName, parameters, options)
	if err != nil {
		return nil, err
	}
	resp, err := 	client.pl.Do(req)
	if err != nil {
		return nil, err
	}
	if !runtime.HasStatusCode(resp, http.StatusOK, http.StatusCreated, http.StatusAccepted) {
		return nil, client.createOrUpdateHandleError(resp)
	}
	 return resp, nil
}

// createOrUpdateCreateRequest creates the CreateOrUpdate request.
func (client *GenericClient) createOrUpdateCreateRequest(ctx context.Context, resourceGroupName string, applicationName string, genericName string, parameters GenericResource, options *GenericBeginCreateOrUpdateOptions) (*policy.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.CustomProviders/resourceProviders/radiusv3/Application/{applicationName}/Generic/{genericName}"
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
	if genericName == "" {
		return nil, errors.New("parameter genericName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{genericName}", url.PathEscape(genericName))
	req, err := runtime.NewRequest(ctx, http.MethodPut, runtime.JoinPaths(	client.ep, urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2018-09-01-preview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")
	return req, runtime.MarshalAsJSON(req, parameters)
}

// createOrUpdateHandleError handles the CreateOrUpdate error response.
func (client *GenericClient) createOrUpdateHandleError(resp *http.Response) error {
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

// BeginDelete - Deletes a Generic resource.
// If the operation fails it returns the *ErrorResponse error type.
func (client *GenericClient) BeginDelete(ctx context.Context, resourceGroupName string, applicationName string, genericName string, options *GenericBeginDeleteOptions) (GenericDeletePollerResponse, error) {
	resp, err := client.deleteOperation(ctx, resourceGroupName, applicationName, genericName, options)
	if err != nil {
		return GenericDeletePollerResponse{}, err
	}
	result := GenericDeletePollerResponse{
		RawResponse: resp,
	}
	pt, err := armruntime.NewPoller("GenericClient.Delete", "location", resp, 	client.pl, client.deleteHandleError)
	if err != nil {
		return GenericDeletePollerResponse{}, err
	}
	result.Poller = &GenericDeletePoller {
		pt: pt,
	}
	return result, nil
}

// Delete - Deletes a Generic resource.
// If the operation fails it returns the *ErrorResponse error type.
func (client *GenericClient) deleteOperation(ctx context.Context, resourceGroupName string, applicationName string, genericName string, options *GenericBeginDeleteOptions) (*http.Response, error) {
	req, err := client.deleteCreateRequest(ctx, resourceGroupName, applicationName, genericName, options)
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
func (client *GenericClient) deleteCreateRequest(ctx context.Context, resourceGroupName string, applicationName string, genericName string, options *GenericBeginDeleteOptions) (*policy.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.CustomProviders/resourceProviders/radiusv3/Application/{applicationName}/Generic/{genericName}"
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
	if genericName == "" {
		return nil, errors.New("parameter genericName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{genericName}", url.PathEscape(genericName))
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
func (client *GenericClient) deleteHandleError(resp *http.Response) error {
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

// Get - Gets a Generic resource by name.
// If the operation fails it returns the *ErrorResponse error type.
func (client *GenericClient) Get(ctx context.Context, resourceGroupName string, applicationName string, genericName string, options *GenericGetOptions) (GenericGetResponse, error) {
	req, err := client.getCreateRequest(ctx, resourceGroupName, applicationName, genericName, options)
	if err != nil {
		return GenericGetResponse{}, err
	}
	resp, err := 	client.pl.Do(req)
	if err != nil {
		return GenericGetResponse{}, err
	}
	if !runtime.HasStatusCode(resp, http.StatusOK) {
		return GenericGetResponse{}, client.getHandleError(resp)
	}
	return client.getHandleResponse(resp)
}

// getCreateRequest creates the Get request.
func (client *GenericClient) getCreateRequest(ctx context.Context, resourceGroupName string, applicationName string, genericName string, options *GenericGetOptions) (*policy.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.CustomProviders/resourceProviders/radiusv3/Application/{applicationName}/Generic/{genericName}"
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
	if genericName == "" {
		return nil, errors.New("parameter genericName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{genericName}", url.PathEscape(genericName))
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

// getHandleResponse handles the Get response.
func (client *GenericClient) getHandleResponse(resp *http.Response) (GenericGetResponse, error) {
	result := GenericGetResponse{RawResponse: resp}
	if err := runtime.UnmarshalAsJSON(resp, &result.GenericResource); err != nil {
		return GenericGetResponse{}, err
	}
	return result, nil
}

// getHandleError handles the Get error response.
func (client *GenericClient) getHandleError(resp *http.Response) error {
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

// List - List the Generic resources deployed in the application.
// If the operation fails it returns the *ErrorResponse error type.
func (client *GenericClient) List(ctx context.Context, resourceGroupName string, applicationName string, options *GenericListOptions) (GenericListResponse, error) {
	req, err := client.listCreateRequest(ctx, resourceGroupName, applicationName, options)
	if err != nil {
		return GenericListResponse{}, err
	}
	resp, err := 	client.pl.Do(req)
	if err != nil {
		return GenericListResponse{}, err
	}
	if !runtime.HasStatusCode(resp, http.StatusOK) {
		return GenericListResponse{}, client.listHandleError(resp)
	}
	return client.listHandleResponse(resp)
}

// listCreateRequest creates the List request.
func (client *GenericClient) listCreateRequest(ctx context.Context, resourceGroupName string, applicationName string, options *GenericListOptions) (*policy.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.CustomProviders/resourceProviders/radiusv3/Application/{applicationName}/Generic"
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
func (client *GenericClient) listHandleResponse(resp *http.Response) (GenericListResponse, error) {
	result := GenericListResponse{RawResponse: resp}
	if err := runtime.UnmarshalAsJSON(resp, &result.GenericList); err != nil {
		return GenericListResponse{}, err
	}
	return result, nil
}

// listHandleError handles the List error response.
func (client *GenericClient) listHandleError(resp *http.Response) error {
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

