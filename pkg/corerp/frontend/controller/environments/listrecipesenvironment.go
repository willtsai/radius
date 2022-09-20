// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package environments

import (
	"context"
	"errors"
	"net/http"

	ctrl "github.com/project-radius/radius/pkg/armrpc/frontend/controller"
	"github.com/project-radius/radius/pkg/armrpc/rest"
	"github.com/project-radius/radius/pkg/armrpc/servicecontext"
	"github.com/project-radius/radius/pkg/corerp/datamodel"
	"github.com/project-radius/radius/pkg/ucp/store"
)

var _ ctrl.Controller = (*ListRecipesEnvironment)(nil)

// ListRecipesEnvironment controller implementation to list recipes linked to the environment
type ListRecipesEnvironment struct {
	ctrl.BaseController
}

// NewListRecipesEnvironment creates a new instance of ListRecipesEnvironment.
func NewListRecipesEnvironment(opts ctrl.Options) (ctrl.Controller, error) {
	return &ListRecipesEnvironment{ctrl.NewBaseController(opts)}, nil
}

// Run returns recipes linked to the specified Environment resource
func (ctrl *ListRecipesEnvironment) Run(ctx context.Context, req *http.Request) (rest.Response, error) {
	sCtx := servicecontext.ARMRequestContextFromContext(ctx)

	resource := &datamodel.Environment{}

	// Request route for listrecipes has the action name as suffix which should be removed to get the resource id.
	// subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/radius-test-rg/providers/Applications.Core/environments/env0/listrecipes
	parsedResourceID := sCtx.ResourceID.Truncate()
	_, err := ctrl.GetResource(ctx, parsedResourceID.String(), resource)
	if err != nil {
		if errors.Is(&store.ErrNotFound{}, err) {
			return rest.NewNotFoundResponse(sCtx.ResourceID), nil
		}
		return nil, err
	}

	// versioned, _ := converter.EnvironmentRecipePropertiesDataModelToVersioned(resource.Properties.Recipes, sCtx.APIVersion)
	versioned := make(map[string]datamodel.EnvironmentRecipeProperties)
	return rest.NewOKResponse(versioned), nil
}
