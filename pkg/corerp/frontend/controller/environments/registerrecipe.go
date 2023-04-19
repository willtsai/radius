// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package environments

import (
	"context"
	"net/http"

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
	return &RegisterRecipe{
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
	content, err := ctrl.ReadJSONBody(req)
	if err != nil {
		return nil, err
	}
	recipe, err := converter.RecipeDatamodelFromVersioned(content, serviceCtx.APIVersion)
	if err != nil {
		return nil, err
	}
	resource, etag, err := r.GetResource(ctx, serviceCtx.ResourceID)
	if err != nil {
		return nil, err
	}

	if resource == nil {
		return rest.NewNotFoundResponse(serviceCtx.ResourceID), nil
	}

	recipeProperties := resource.Properties.Recipes
	if recipeProperties == nil {
		recipeProperties = map[string]datamodel.EnvironmentRecipeProperties{}
	}
	recipeProperties[recipe.Name] = datamodel.EnvironmentRecipeProperties{
		LinkType:     recipe.LinkType,
		TemplatePath: recipe.TemplatePath,
		Parameters:   recipe.Parameters,
	}
	_, err = r.SaveResource(ctx, serviceCtx.ResourceID.String(), resource, etag)
	if err != nil {
		return nil, err
	}
	versioned, err := converter.RecipeDatamodelToVersioned(recipe, serviceCtx.APIVersion)
	if err != nil {
		return nil, err
	}
	return rest.NewOKResponse(versioned), nil
}
