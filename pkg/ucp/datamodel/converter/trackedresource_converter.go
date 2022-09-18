// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package converter

import (
	"encoding/json"

	"github.com/project-radius/radius/pkg/armrpc/api/conv"
	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
	"github.com/project-radius/radius/pkg/ucp/api/v20220315privatepreview"
	"github.com/project-radius/radius/pkg/ucp/datamodel"
)

// TrackedResourceDataModelToVersioned converts version agnostic TrackedResource datamodel to versioned model.
func TrackedResourceDataModelToVersioned(model *datamodel.TrackedResource, version string) (conv.VersionedModelInterface, error) {
	switch version {
	case v20220315privatepreview.Version:
		versioned := &v20220315privatepreview.TrackedResource{}
		err := versioned.ConvertFrom(model)
		return versioned, err

	default:
		return nil, v1.ErrUnsupportedAPIVersion
	}
}

// TrackedResourceDataModelFromVersioned converts versioned TrackedResource model to datamodel.
func TrackedResourceDataModelFromVersioned(content []byte, version string) (*datamodel.TrackedResource, error) {
	switch version {
	case v20220315privatepreview.Version:
		am := &v20220315privatepreview.TrackedResource{}
		if err := json.Unmarshal(content, am); err != nil {
			return nil, err
		}
		dm, err := am.ConvertTo()
		return dm.(*datamodel.TrackedResource), err

	default:
		return nil, v1.ErrUnsupportedAPIVersion
	}
}
