// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package daprpubsubbrokers

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

var _ ctrl.Controller = (*DeleteDaprPubSubBroker)(nil)

// DeleteDaprPubSubBroker is the controller implementation to delete daprPubSubBroker connector resource.
type DeleteDaprPubSubBroker struct {
	ctrl.BaseController
}

// NewDeleteDaprPubSubBroker creates a new instance DeleteDaprPubSubBroker.
func NewDeleteDaprPubSubBroker(opts ctrl.Options) (ctrl.Controller, error) {
	return &DeleteDaprPubSubBroker{ctrl.NewBaseController(opts)}, nil
}

func (daprPubSub *DeleteDaprPubSubBroker) Run(ctx context.Context, req *http.Request) (rest.Response, error) {
	serviceCtx := v1.ARMRequestContextFromContext(ctx)

	// Read resource metadata from the storage
	existingResource := &datamodel.DaprPubSubBroker{}
	etag, err := daprPubSub.GetResource(ctx, serviceCtx.ResourceID.String(), existingResource)
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

	err = daprPubSub.DeploymentProcessor().Delete(ctx, deployment.ResourceData{ID: serviceCtx.ResourceID, Resource: existingResource, OutputResources: existingResource.Properties.Status.OutputResources, ComputedValues: existingResource.ComputedValues, SecretValues: existingResource.SecretValues, RecipeData: existingResource.RecipeData})
	if err != nil {
		return nil, err
	}
	err = daprPubSub.StorageClient().Delete(ctx, serviceCtx.ResourceID.String())
	if err != nil {
		if errors.Is(&store.ErrNotFound{}, err) {
			return rest.NewNoContentResponse(), nil
		}
		return nil, err
	}

	return rest.NewOKResponse(nil), nil
}
