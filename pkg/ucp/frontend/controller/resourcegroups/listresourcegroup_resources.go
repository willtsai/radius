// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------
package resourcegroups

import (
	"context"
	"net/http"

	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
	"github.com/project-radius/radius/pkg/middleware"
	"github.com/project-radius/radius/pkg/ucp/api/v20220315privatepreview"
	"github.com/project-radius/radius/pkg/ucp/datamodel"
	ctrl "github.com/project-radius/radius/pkg/ucp/frontend/controller"
	"github.com/project-radius/radius/pkg/ucp/resources"
	"github.com/project-radius/radius/pkg/ucp/rest"
	"github.com/project-radius/radius/pkg/ucp/store"
)

var _ ctrl.Controller = (*ListResourceGroupResources)(nil)

// ListResourceGroupResources is the controller implementation to get the list of tracked resources in a resource group.
type ListResourceGroupResources struct {
	ctrl.BaseController
}

// NewListResourceGroupResources creates a new ListResourceGroupResources.
func NewListResourceGroupResources(opts ctrl.Options) (ctrl.Controller, error) {
	return &ListResourceGroupResources{ctrl.NewBaseController(opts)}, nil
}

func (r *ListResourceGroupResources) Run(ctx context.Context, w http.ResponseWriter, req *http.Request) (rest.Response, error) {
	path := middleware.GetRelativePath(r.Options.BasePath, req.URL.Path)
	id, err := resources.ParseByMethod(path, req.Method)
	if err != nil {
		return rest.NewBadRequestARMResponse(rest.ErrorResponse{
			Error: rest.ErrorDetails{
				Code:    v1.CodeInvalidRequestContent,
				Message: err.Error(),
			},
		}), nil
	}

	// We store the tracked resources using the resource group as the scope and "System.Resources/resources" as the type.
	// This doesn't map to an actual resource type we expose through the API.
	result, err := r.Options.DB.Query(ctx, store.Query{
		RootScope:    id.Truncate().String(),
		ResourceType: "System.Resources/resources",
	})
	if err != nil {
		return nil, err
	}

	items := v20220315privatepreview.TrackedResourceList{}
	for _, obj := range result.Items {
		dm := datamodel.TrackedResource{}
		err := obj.As(&dm)
		if err != nil {
			return nil, err
		}

		converted := v20220315privatepreview.TrackedResource{}
		err = converted.ConvertFrom(&dm)
		if err != nil {
			return nil, err
		}
		items.Value = append(items.Value, &converted)
	}

	return rest.NewOKResponse(items), nil
}
