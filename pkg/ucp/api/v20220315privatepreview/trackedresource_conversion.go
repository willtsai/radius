// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package v20220315privatepreview

import (
	"github.com/project-radius/radius/pkg/armrpc/api/conv"
	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
	"github.com/project-radius/radius/pkg/ucp/datamodel"

	"github.com/Azure/go-autorest/autorest/to"
)

// ConvertTo converts from the versioned TrackedResource resource to version-agnostic datamodel.
func (src *TrackedResource) ConvertTo() (conv.DataModelInterface, error) {
	converted := &datamodel.TrackedResource{
		TrackedResource: v1.TrackedResource{
			ID:       to.String(src.ID),
			Name:     to.String(src.Name),
			Type:     to.String(src.Type),
			Location: to.String(src.Location),
			Tags:     to.StringMap(src.Tags),
		},
		InternalMetadata: v1.InternalMetadata{
			UpdatedAPIVersion: Version,
		},
	}

	return converted, nil
}

// ConvertFrom converts from version-agnostic datamodel to the versioned Resource resource.
func (dst *TrackedResource) ConvertFrom(src conv.DataModelInterface) error {
	resource, ok := src.(*datamodel.TrackedResource)
	if !ok {
		return conv.ErrInvalidModelConversion
	}

	dst.ID = to.StringPtr(resource.ID)
	dst.Name = to.StringPtr(resource.Name)
	dst.Type = to.StringPtr(resource.Type)
	dst.Location = to.StringPtr(resource.Location)
	dst.Tags = *to.StringMapPtr(resource.Tags)

	return nil
}
