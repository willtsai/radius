// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package renderers

import (
	"fmt"
	"strings"

	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
	coreDatamodel "github.com/project-radius/radius/pkg/corerp/datamodel"
	"github.com/project-radius/radius/pkg/ucp/resources"
)

func ValidateApplicationID(application string) (resources.ID, error) {
	app := &coreDatamodel.Application{}
	if application != "" {
		appId, err := resources.ParseResource(application)
		if err != nil {
			return resources.ID{}, v1.NewClientErrInvalidRequest(fmt.Sprintf("failed to parse application from the property: %s", err.Error()))
		}
		if !strings.EqualFold(appId.Type(), app.ResourceTypeName()) {
			return resources.ID{}, v1.NewClientErrInvalidRequest(fmt.Sprintf("provided application id type %q is not a valid type.", appId.Type()))
		}
		return appId, nil
	}
	return resources.ID{}, nil
}

func ValidateLinkType(resource v1.DataModelInterface, options RenderOptions) error {
	if options.RecipeProperties.LinkType != resource.ResourceTypeName() {
		return v1.NewClientErrInvalidRequest(fmt.Sprintf("link type %q of provided recipe %q is incompatible with %q resource type. Recipe link type must match link resource type.",
			options.RecipeProperties.LinkType, options.RecipeProperties.Name, resource.ResourceTypeName()))
	}
	return nil
}