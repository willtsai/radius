// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package environments

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
	ctrl "github.com/project-radius/radius/pkg/armrpc/frontend/controller"
	"github.com/project-radius/radius/pkg/armrpc/rest"
	"github.com/project-radius/radius/pkg/corerp/datamodel"
	"github.com/project-radius/radius/pkg/corerp/datamodel/converter"
)

var _ ctrl.Controller = (*RegisterRecipe)(nil)

type RegisterRecipe struct {
	ctrl.Operation[*datamodel.Environment, datamodel.Environment]
}

func NewRegisterRecipe(opts ctrl.Options) (ctrl.Controller, error) {
	return &GetRecipeMetadata{
		ctrl.NewOperation(opts,
			ctrl.ResourceOptions[datamodel.Environment]{
				RequestConverter:  converter.EnvironmentDataModelFromVersioned,
				ResponseConverter: converter.EnvironmentDataModelToVersioned,
			},
		),
	}, nil
}

func (r *RegisterRecipe) Run(ctx context.Context, w http.ResponseWriter, req *http.Request) (rest.Response, error) {
	serviceCtx := v1.ARMRequestContextFromContext(ctx)

	recipeSuffix := strings.Split(serviceCtx.OrignalURL.Path, "/environments/")[1]
	recipeName := strings.Split(recipeSuffix, "/")[1]
	resource, _, err := r.GetResource(ctx, serviceCtx.ResourceID)
	if err != nil {
		return nil, err
	}

	if resource == nil {
		return rest.NewNotFoundResponse(serviceCtx.ResourceID), nil
	}

	recipe, exists := resource.Properties.Recipes[recipeName]
	if !exists {
		return rest.NewNotFoundMessageResponse(fmt.Sprintf("Recipe with name %q not found on environment with id %q", recipeName, serviceCtx.ResourceID)), nil
	}

	recipeParams, err := getRecipeMetadataFromRegistry(ctx, recipe.TemplatePath, recipeName)
	if err != nil {
		return nil, err
	}

	ret := datamodel.EnvironmentRecipeProperties{
		LinkType:     recipe.LinkType,
		TemplatePath: recipe.TemplatePath,
		Parameters:   recipeParams,
	}

	versioned, err := converter.EnvironmentRecipePropertiesDataModelToVersioned(&ret, serviceCtx.APIVersion)
	if err != nil {
		return nil, err
	}
	return rest.NewOKResponse(versioned), nil
}
