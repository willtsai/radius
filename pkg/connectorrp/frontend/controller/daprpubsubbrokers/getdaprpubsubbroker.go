// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package daprpubsubbrokers

import (
	"context"
	"errors"
	"net/http"

	manager "github.com/project-radius/radius/pkg/armrpc/asyncoperation/statusmanager"
	ctrl "github.com/project-radius/radius/pkg/armrpc/frontend/controller"
	"github.com/project-radius/radius/pkg/armrpc/servicecontext"
	"github.com/project-radius/radius/pkg/connectorrp/datamodel"
	"github.com/project-radius/radius/pkg/connectorrp/datamodel/converter"
	"github.com/project-radius/radius/pkg/connectorrp/frontend/deployment"
	"github.com/project-radius/radius/pkg/radrp/rest"
	"github.com/project-radius/radius/pkg/ucp/store"
)

var _ ctrl.Controller = (*GetDaprPubSubBroker)(nil)

// GetDaprPubSubBroker is the controller implementation to get the daprPubSubBroker conenctor resource.
type GetDaprPubSubBroker struct {
	ctrl.BaseController
}

// NewGetDaprPubSubBroker creates a new instance of GetDaprPubSubBroker.
func NewGetDaprPubSubBroker(ds store.StorageClient, sm manager.StatusManager, dp deployment.DeploymentProcessor) (ctrl.Controller, error) {
	return &GetDaprPubSubBroker{ctrl.NewBaseController(ds, sm, dp)}, nil
}

func (daprPubSub *GetDaprPubSubBroker) Run(ctx context.Context, req *http.Request) (rest.Response, error) {
	serviceCtx := servicecontext.ARMRequestContextFromContext(ctx)

	existingResource := &datamodel.DaprPubSubBroker{}
	_, err := daprPubSub.GetResource(ctx, serviceCtx.ResourceID.String(), existingResource)
	if err != nil {
		if errors.Is(&store.ErrNotFound{}, err) {
			return rest.NewNotFoundResponse(serviceCtx.ResourceID), nil
		}
		return nil, err
	}

	versioned, _ := converter.DaprPubSubBrokerDataModelToVersioned(existingResource, serviceCtx.APIVersion)
	return rest.NewOKResponse(versioned), nil
}
