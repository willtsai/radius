// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package rabbitmqmessagequeues

import (
	"context"
	"errors"
	"net/http"

	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
	ctrl "github.com/project-radius/radius/pkg/armrpc/frontend/controller"
	"github.com/project-radius/radius/pkg/armrpc/rest"
	"github.com/project-radius/radius/pkg/connectorrp/datamodel"
	"github.com/project-radius/radius/pkg/connectorrp/frontend/deployment"
	"github.com/project-radius/radius/pkg/ucp/store"
)

var _ ctrl.Controller = (*DeleteRabbitMQMessageQueue)(nil)

// DeleteRabbitMQMessageQueue is the controller implementation to delete rabbitmq connector resource.
type DeleteRabbitMQMessageQueue struct {
	ctrl.BaseController
}

// NewDeleteRabbitMQMessageQueue creates a new instance DeleteRabbitMQMessageQueue.
func NewDeleteRabbitMQMessageQueue(opts ctrl.Options) (ctrl.Controller, error) {
	return &DeleteRabbitMQMessageQueue{ctrl.NewBaseController(opts)}, nil
}

func (rabbitmq *DeleteRabbitMQMessageQueue) Run(ctx context.Context, req *http.Request) (rest.Response, error) {
	serviceCtx := v1.ARMRequestContextFromContext(ctx)

	// Read resource metadata from the storage
	existingResource := &datamodel.RabbitMQMessageQueue{}
	etag, err := rabbitmq.GetResource(ctx, serviceCtx.ResourceID.String(), existingResource)
	if err != nil {
		if errors.Is(&store.ErrNotFound{}, err) {
			return rest.NewNoContentResponse(), nil
		}
		return nil, err
	}

	if etag == "" {
		return rest.NewNoContentResponse(), nil
	}

	err = ctrl.ValidateETag(*serviceCtx, etag)
	if err != nil {
		return rest.NewPreconditionFailedResponse(serviceCtx.ResourceID.String(), err.Error()), nil
	}

	err = rabbitmq.DeploymentProcessor().Delete(ctx, deployment.ResourceData{ID: serviceCtx.ResourceID, Resource: existingResource, OutputResources: existingResource.Properties.Status.OutputResources, ComputedValues: existingResource.ComputedValues, SecretValues: existingResource.SecretValues, RecipeData: existingResource.RecipeData})
	if err != nil {
		return nil, err
	}

	err = rabbitmq.StorageClient().Delete(ctx, serviceCtx.ResourceID.String())
	if err != nil {
		if errors.Is(&store.ErrNotFound{}, err) {
			return rest.NewNoContentResponse(), nil
		}
		return nil, err
	}

	return rest.NewOKResponse(nil), nil
}
