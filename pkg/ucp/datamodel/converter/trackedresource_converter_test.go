// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package converter

import (
	"encoding/json"
	"testing"

	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
	radiustesting "github.com/project-radius/radius/pkg/corerp/testing"
	"github.com/project-radius/radius/pkg/ucp/api/v20220315privatepreview"
	"github.com/project-radius/radius/pkg/ucp/datamodel"
	"github.com/stretchr/testify/require"
)

// Validates type conversion between versioned client side data model and RP data model.
func TestTrackedResourceDataModelToVersioned(t *testing.T) {
	testset := []struct {
		directory     string
		dataModelFile string
		apiVersion    string
		apiModelType  interface{}
		err           error
	}{
		{
			"../../api/v20220315privatepreview",
			"trackedresourcedatamodel.json",
			"2022-03-15-privatepreview",
			&v20220315privatepreview.TrackedResource{},
			nil,
		},
		{
			"../../api/v20220315privatepreview",
			"trackedresource.json",
			"unsupported",
			nil,
			v1.ErrUnsupportedAPIVersion,
		},
	}

	for _, tc := range testset {
		t.Run(tc.apiVersion, func(t *testing.T) {
			c := radiustesting.ReadPackageFixture(tc.directory, tc.dataModelFile)
			dm := &datamodel.TrackedResource{}
			_ = json.Unmarshal(c, dm)
			am, err := TrackedResourceDataModelToVersioned(dm, tc.apiVersion)
			if tc.err != nil {
				require.ErrorAs(t, tc.err, &err)
			} else {
				require.NoError(t, err)
				require.IsType(t, tc.apiModelType, am)
			}
		})
	}
}

func TestTrackedResourceDataModelFromVersioned(t *testing.T) {
	testset := []struct {
		directory          string
		versionedModelFile string
		apiVersion         string
		err                error
	}{
		{

			"../../api/v20220315privatepreview",
			"trackedresource.json",
			"2022-03-15-privatepreview",
			nil,
		},
		{
			"../../api/v20220315privatepreview",
			"trackedresource.json",
			"unsupported",
			v1.ErrUnsupportedAPIVersion,
		},
	}

	for _, tc := range testset {
		t.Run(tc.apiVersion, func(t *testing.T) {
			c := radiustesting.ReadPackageFixture(tc.directory, tc.versionedModelFile)
			dm, err := TrackedResourceDataModelFromVersioned(c, tc.apiVersion)
			if tc.err != nil {
				require.ErrorAs(t, tc.err, &err)
			} else {
				require.NoError(t, err)
				require.IsType(t, tc.apiVersion, dm.InternalMetadata.UpdatedAPIVersion)
			}
		})
	}
}
