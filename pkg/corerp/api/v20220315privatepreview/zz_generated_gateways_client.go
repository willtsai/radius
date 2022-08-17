//go:build go1.16
// +build go1.16

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

package v20220315privatepreview

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"net/http"
	"net/url"
	"strings"
)

// GatewaysClient contains the methods for the Gateways group.
// Don't use this type directly, use NewGatewaysClient() instead.
type GatewaysClient struct {
	ep string
	pl runtime.Pipeline
	rootScope string
}

// NewGatewaysClient creates a new instance of GatewaysClient with the specified values.
func NewGatewaysClient(con *arm.Connection, rootScope string) *GatewaysClient {
	return &GatewaysClient{ep: con.Endpoint(), pl: con.NewPipeline(module, version), rootScope: rootScope}
}

// CreateOrUpdate - Create or update a Gateway.
// If the operation fails it returns the *ErrorResponse error type.
func (client *GatewaysClient) CreateOrUpdate(ctx context.Context, gatewayName string, gatewayResource GatewayResource, options *GatewaysCreateOrUpdateOptions) (GatewaysCreateOrUpdateResponse, error) {
	req, err := client.createOrUpdateCreateRequest(ctx, gatewayName, gatewayResource, options)
	if err != nil {
		return GatewaysCreateOrUpdateResponse{}, err
	}
	resp, err := 	client.pl.Do(req)
	if err != nil {
		return GatewaysCreateOrUpdateResponse{}, err
	}
	if !runtime.HasStatusCode(resp, http.StatusOK, http.StatusCreated, http.StatusNoContent) {
		return GatewaysCreateOrUpdateResponse{}, client.createOrUpdateHandleError(resp)
	}
	return client.createOrUpdateHandleResponse(resp)
}

// createOrUpdateCreateRequest creates the CreateOrUpdate request.
func (client *GatewaysClient) createOrUpdateCreateRequest(ctx context.Context, gatewayName string, gatewayResource GatewayResource, options *GatewaysCreateOrUpdateOptions) (*policy.Request, error) {
	urlPath := "/{rootScope}/providers/Applications.Core/gateways/{gatewayName}"
	if client.rootScope == "" {
		return nil, errors.New("parameter client.rootScope cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{rootScope}", client.rootScope)
	if gatewayName == "" {
		return nil, errors.New("parameter gatewayName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{gatewayName}", url.PathEscape(gatewayName))
	req, err := runtime.NewRequest(ctx, http.MethodPut, runtime.JoinPaths(	client.ep, urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2022-03-15-privatepreview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")
	return req, runtime.MarshalAsJSON(req, gatewayResource)
}

// createOrUpdateHandleResponse handles the CreateOrUpdate response.
func (client *GatewaysClient) createOrUpdateHandleResponse(resp *http.Response) (GatewaysCreateOrUpdateResponse, error) {
	result := GatewaysCreateOrUpdateResponse{RawResponse: resp}
	if err := runtime.UnmarshalAsJSON(resp, &result.GatewayResource); err != nil {
		return GatewaysCreateOrUpdateResponse{}, err
	}
	return result, nil
}

// createOrUpdateHandleError handles the CreateOrUpdate error response.
func (client *GatewaysClient) createOrUpdateHandleError(resp *http.Response) error {
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

// Delete - Delete a Gateway.
// If the operation fails it returns the *ErrorResponse error type.
func (client *GatewaysClient) Delete(ctx context.Context, gatewayName string, options *GatewaysDeleteOptions) (GatewaysDeleteResponse, error) {
	req, err := client.deleteCreateRequest(ctx, gatewayName, options)
	if err != nil {
		return GatewaysDeleteResponse{}, err
	}
	resp, err := 	client.pl.Do(req)
	if err != nil {
		return GatewaysDeleteResponse{}, err
	}
	if !runtime.HasStatusCode(resp, http.StatusOK, http.StatusAccepted, http.StatusNoContent) {
		return GatewaysDeleteResponse{}, client.deleteHandleError(resp)
	}
	return GatewaysDeleteResponse{RawResponse: resp}, nil
}

// deleteCreateRequest creates the Delete request.
func (client *GatewaysClient) deleteCreateRequest(ctx context.Context, gatewayName string, options *GatewaysDeleteOptions) (*policy.Request, error) {
	urlPath := "/{rootScope}/providers/Applications.Core/gateways/{gatewayName}"
	if client.rootScope == "" {
		return nil, errors.New("parameter client.rootScope cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{rootScope}", client.rootScope)
	if gatewayName == "" {
		return nil, errors.New("parameter gatewayName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{gatewayName}", url.PathEscape(gatewayName))
	req, err := runtime.NewRequest(ctx, http.MethodDelete, runtime.JoinPaths(	client.ep, urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2022-03-15-privatepreview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")
	return req, nil
}

// deleteHandleError handles the Delete error response.
func (client *GatewaysClient) deleteHandleError(resp *http.Response) error {
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

// Get - Gets the properties of a Gateway.
// If the operation fails it returns the *ErrorResponse error type.
func (client *GatewaysClient) Get(ctx context.Context, gatewayName string, options *GatewaysGetOptions) (GatewaysGetResponse, error) {
	req, err := client.getCreateRequest(ctx, gatewayName, options)
	if err != nil {
		return GatewaysGetResponse{}, err
	}
	resp, err := 	client.pl.Do(req)
	if err != nil {
		return GatewaysGetResponse{}, err
	}
	if !runtime.HasStatusCode(resp, http.StatusOK) {
		return GatewaysGetResponse{}, client.getHandleError(resp)
	}
	return client.getHandleResponse(resp)
}

// getCreateRequest creates the Get request.
func (client *GatewaysClient) getCreateRequest(ctx context.Context, gatewayName string, options *GatewaysGetOptions) (*policy.Request, error) {
	urlPath := "/{rootScope}/providers/Applications.Core/gateways/{gatewayName}"
	if client.rootScope == "" {
		return nil, errors.New("parameter client.rootScope cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{rootScope}", client.rootScope)
	if gatewayName == "" {
		return nil, errors.New("parameter gatewayName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{gatewayName}", url.PathEscape(gatewayName))
	req, err := runtime.NewRequest(ctx, http.MethodGet, runtime.JoinPaths(	client.ep, urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2022-03-15-privatepreview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")
	return req, nil
}

// getHandleResponse handles the Get response.
func (client *GatewaysClient) getHandleResponse(resp *http.Response) (GatewaysGetResponse, error) {
	result := GatewaysGetResponse{RawResponse: resp}
	if err := runtime.UnmarshalAsJSON(resp, &result.GatewayResource); err != nil {
		return GatewaysGetResponse{}, err
	}
	return result, nil
}

// getHandleError handles the Get error response.
func (client *GatewaysClient) getHandleError(resp *http.Response) error {
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

// ListByScope - List all Gateways in the given scope.
// If the operation fails it returns the *ErrorResponse error type.
func (client *GatewaysClient) ListByScope(options *GatewaysListByScopeOptions) (*GatewaysListByScopePager) {
	return &GatewaysListByScopePager{
		client: client,
		requester: func(ctx context.Context) (*policy.Request, error) {
			return client.listByScopeCreateRequest(ctx, options)
		},
		advancer: func(ctx context.Context, resp GatewaysListByScopeResponse) (*policy.Request, error) {
			return runtime.NewRequest(ctx, http.MethodGet, *resp.GatewayResourceList.NextLink)
		},
	}
}

// listByScopeCreateRequest creates the ListByScope request.
func (client *GatewaysClient) listByScopeCreateRequest(ctx context.Context, options *GatewaysListByScopeOptions) (*policy.Request, error) {
	urlPath := "/{rootScope}/providers/Applications.Core/gateways"
	if client.rootScope == "" {
		return nil, errors.New("parameter client.rootScope cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{rootScope}", client.rootScope)
	req, err := runtime.NewRequest(ctx, http.MethodGet, runtime.JoinPaths(	client.ep, urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2022-03-15-privatepreview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")
	return req, nil
}

// listByScopeHandleResponse handles the ListByScope response.
func (client *GatewaysClient) listByScopeHandleResponse(resp *http.Response) (GatewaysListByScopeResponse, error) {
	result := GatewaysListByScopeResponse{RawResponse: resp}
	if err := runtime.UnmarshalAsJSON(resp, &result.GatewayResourceList); err != nil {
		return GatewaysListByScopeResponse{}, err
	}
	return result, nil
}

// listByScopeHandleError handles the ListByScope error response.
func (client *GatewaysClient) listByScopeHandleError(resp *http.Response) error {
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

// Update - Update the properties of an existing Gateway.
// If the operation fails it returns the *ErrorResponse error type.
func (client *GatewaysClient) Update(ctx context.Context, gatewayName string, gatewayResource GatewayResource, options *GatewaysUpdateOptions) (GatewaysUpdateResponse, error) {
	req, err := client.updateCreateRequest(ctx, gatewayName, gatewayResource, options)
	if err != nil {
		return GatewaysUpdateResponse{}, err
	}
	resp, err := 	client.pl.Do(req)
	if err != nil {
		return GatewaysUpdateResponse{}, err
	}
	if !runtime.HasStatusCode(resp, http.StatusOK, http.StatusCreated, http.StatusNoContent) {
		return GatewaysUpdateResponse{}, client.updateHandleError(resp)
	}
	return client.updateHandleResponse(resp)
}

// updateCreateRequest creates the Update request.
func (client *GatewaysClient) updateCreateRequest(ctx context.Context, gatewayName string, gatewayResource GatewayResource, options *GatewaysUpdateOptions) (*policy.Request, error) {
	urlPath := "/{rootScope}/providers/Applications.Core/gateways/{gatewayName}"
	if client.rootScope == "" {
		return nil, errors.New("parameter client.rootScope cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{rootScope}", client.rootScope)
	if gatewayName == "" {
		return nil, errors.New("parameter gatewayName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{gatewayName}", url.PathEscape(gatewayName))
	req, err := runtime.NewRequest(ctx, http.MethodPatch, runtime.JoinPaths(	client.ep, urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2022-03-15-privatepreview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")
	return req, runtime.MarshalAsJSON(req, gatewayResource)
}

// updateHandleResponse handles the Update response.
func (client *GatewaysClient) updateHandleResponse(resp *http.Response) (GatewaysUpdateResponse, error) {
	result := GatewaysUpdateResponse{RawResponse: resp}
	if err := runtime.UnmarshalAsJSON(resp, &result.GatewayResource); err != nil {
		return GatewaysUpdateResponse{}, err
	}
	return result, nil
}

// updateHandleError handles the Update error response.
func (client *GatewaysClient) updateHandleError(resp *http.Response) error {
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
