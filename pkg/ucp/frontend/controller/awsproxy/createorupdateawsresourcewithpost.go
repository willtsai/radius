// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------
package awsproxy

import (
	"context"
	"encoding/json"
	"fmt"
	http "net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/google/uuid"
	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
	armrpc_controller "github.com/project-radius/radius/pkg/armrpc/frontend/controller"
	armrpc_rest "github.com/project-radius/radius/pkg/armrpc/rest"
	awserror "github.com/project-radius/radius/pkg/ucp/aws"
	ctrl "github.com/project-radius/radius/pkg/ucp/frontend/controller"
	"github.com/project-radius/radius/pkg/ucp/ucplog"
	"github.com/wI2L/jsondiff"
)

var _ armrpc_controller.Controller = (*CreateOrUpdateAWSResourceWithPost)(nil)

// CreateOrUpdateAWSResourceWithPost is the controller implementation to create/update an AWS resource.
type CreateOrUpdateAWSResourceWithPost struct {
	ctrl.BaseController
}

// NewCreateOrUpdateAWSResourceWithPost creates a new CreateOrUpdateAWSResourceWithPost.
func NewCreateOrUpdateAWSResourceWithPost(opts ctrl.Options) (armrpc_controller.Controller, error) {
	return &CreateOrUpdateAWSResourceWithPost{ctrl.NewBaseController(opts)}, nil
}

func (p *CreateOrUpdateAWSResourceWithPost) Run(ctx context.Context, w http.ResponseWriter, req *http.Request) (armrpc_rest.Response, error) {
	logger := ucplog.GetLogger(ctx)
	// Lookup resource type schema
	cloudControlClient, cloudFormationClient, resourceType, id, err := ParseAWSRequest(ctx, p.Options, req)
	if err != nil {
		return nil, err
	}

	properties, err := readPropertiesFromBody(req)
	if err != nil {
		e := v1.ErrorResponse{
			Error: v1.ErrorDetails{
				Code:    v1.CodeInvalid,
				Message: "failed to read request body",
			},
		}
		return armrpc_rest.NewBadRequestARMResponse(e), nil
	}

	// TODO
	// 1. Split this method up into two calls
	// 2. If we can't create the multi-identifier resource
	// create the id afterwards and assume create

	primaryIdentifers, err := lookupPrimaryIdentifiersForResourceType(p.Options, resourceType)
	if err != nil {
		e := v1.ErrorResponse{
			Error: v1.ErrorDetails{
				Code:    v1.CodeInvalid,
				Message: err.Error(),
			},
		}
		return armrpc_rest.NewBadRequestARMResponse(e), nil
	}

	existing := true

	responseProperties := map[string]interface{}{}

	var operation uuid.UUID
	desiredState, err := json.Marshal(properties)
	if err != nil {
		return awserror.HandleAWSError(err)
	}

	awsResourceIdentifier, err := getResourceIDFromPrimaryIdentifiers(primaryIdentifers, properties)
	computedResourceID := ""

	var getResponse *cloudcontrol.GetResourceOutput = nil
	if err != nil {
		// assume that if we can't get the AWS resource identifier, we need to create the resource
		existing = false
	} else {
		computedResourceID = computeResourceID(id, awsResourceIdentifier)

		// Create and update work differently for AWS - we need to know if the resource
		// we're working on exists already.

		getResponse, err = client.GetResource(ctx, &cloudcontrol.GetResourceInput{
			TypeName:   &resourceType,
			Identifier: aws.String(awsResourceIdentifier),
		})
		if awserror.IsAWSResourceNotFound(err) {
			existing = false
		} else if err != nil {
			return awserror.HandleAWSError(err)
		} else {
			err = json.Unmarshal([]byte(*getResponse.ResourceDescription.Properties), &responseProperties)
			if err != nil {
				return awserror.HandleAWSError(err)
			}
		}
	}

	// Properties specified by users take precedence
	for k, v := range properties {
		responseProperties[k] = v
	}

	if existing {
		logger.Info("Updating resource", "resourceType", resourceType, "resourceID", awsResourceIdentifier)
		// For an existing resource we need to convert the desired state into a JSON-patch document
		patch, err := jsondiff.CompareJSON([]byte(*getResponse.ResourceDescription.Properties), desiredState)
		if err != nil {
			return awserror.HandleAWSError(err)
		}

		// We need to take out readonly properties. Those are usually not specified by the client, and so
		// our library will generate "remove" operations.
		//
		// Iterate backwards because we're removing items from the array
		for i := len(patch) - 1; i >= 0; i-- {
			if patch[i].Type == "remove" {
				patch = append(patch[:i], patch[i+1:]...)
			}
		}

		// Call update only if the patch is not empty
		if len(patch) > 0 {
			marshaled, err := json.Marshal(&patch)
			if err != nil {
				return awserror.HandleAWSError(err)
			}

			response, err := cloudControlClient.UpdateResource(ctx, &cloudcontrol.UpdateResourceInput{
				TypeName:      &resourceType,
				Identifier:    aws.String(awsResourceIdentifier),
				PatchDocument: aws.String(string(marshaled)),
			})
			if err != nil {
				return awserror.HandleAWSError(err)
			}

			operation, err = uuid.Parse(*response.ProgressEvent.RequestToken)
			if err != nil {
				return awserror.HandleAWSError(err)
			}
		} else {
			logger.Info("No changes detected, skipping update", "resourceType", resourceType, "resourceID", awsResourceIdentifier)
			// mark provisioning state as succeeded here
			// and return 200, telling the deployment engine that the resource has already been created
			responseProperties["provisioningState"] = v1.ProvisioningStateSucceeded
			responseBody := map[string]interface{}{
				"id":         computedResourceID,
				"name":       awsResourceIdentifier,
				"type":       id.Type(),
				"properties": responseProperties,
			}

			resp := armrpc_rest.NewOKResponse(responseBody)
			return resp, nil
		}
	} else {
		logger.Info("Creating resource", "resourceType", resourceType, "resourceID", awsResourceIdentifier)
		response, err := cloudControlClient.CreateResource(ctx, &cloudcontrol.CreateResourceInput{
			TypeName:     &resourceType,
			DesiredState: aws.String(string(desiredState)),
		})
		if err != nil {
			return awserror.HandleAWSError(err)
		}

		if response == nil || response.ProgressEvent == nil || response.ProgressEvent.Identifier == nil {
			return awserror.HandleAWSError(fmt.Errorf("empty response from AWS Cloud Control Create API for type %s", resourceType))
		}

		operation, err = uuid.Parse(*response.ProgressEvent.RequestToken)
		if err != nil {
			return awserror.HandleAWSError(err)
		}

		awsResourceIdentifier := *response.ProgressEvent.Identifier

		computedResourceID = computeResourceID(id, awsResourceIdentifier)
	}

	responseProperties["provisioningState"] = v1.ProvisioningStateProvisioning

	responseBody := map[string]interface{}{
		"id":         computedResourceID,
		"name":       awsResourceIdentifier,
		"type":       id.Type(),
		"properties": responseProperties,
	}

	resp := armrpc_rest.NewAsyncOperationResponse(responseBody, v1.LocationGlobal, 201, id, operation, "", id.RootScope(), p.Options.BasePath)
	return resp, nil
}
