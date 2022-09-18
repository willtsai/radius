// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package v20220315privatepreview

import (
	"encoding/json"
	"testing"

	"github.com/project-radius/radius/pkg/armrpc/api/conv"
	radiustesting "github.com/project-radius/radius/pkg/corerp/testing"
	"github.com/project-radius/radius/pkg/ucp/datamodel"
	"github.com/stretchr/testify/require"
)

func TestTrackedResource_ConvertVersionedToDataModel(t *testing.T) {
	testset := []string{"trackedresource.json"}
	for _, payload := range testset {
		rawPayload := radiustesting.ReadFixture(payload)
		versionedResource := &TrackedResource{}
		err := json.Unmarshal(rawPayload, versionedResource)
		require.NoError(t, err)

		dm, err := versionedResource.ConvertTo()
		require.NoError(t, err)

		convertedResource := dm.(*datamodel.TrackedResource)
		require.Equal(t, "/planes/radius/local/resourceGroups/radius-test-rg/providers/Applications.Core/applications/test-app", convertedResource.ID)
		require.Equal(t, "test-app", convertedResource.Name)
		require.Equal(t, "Applications.Core/applications", convertedResource.Type)
		require.Equal(t, "global", convertedResource.Location)
		require.Equal(t, map[string]string{
			"tag-name": "tag-value",
		}, convertedResource.Tags)
		require.Equal(t, "2022-03-15-privatepreview", convertedResource.InternalMetadata.UpdatedAPIVersion)
	}
}

func TestResource_ConvertDataModelToVersioned(t *testing.T) {
	testset := []string{"trackedresourcedatamodel.json"}

	for _, payload := range testset {
		rawPayload := radiustesting.ReadFixture(payload)
		resource := &datamodel.TrackedResource{}
		err := json.Unmarshal(rawPayload, resource)
		require.NoError(t, err)

		versionedResource := &TrackedResource{}
		err = versionedResource.ConvertFrom(resource)

		require.NoError(t, err)
		require.Equal(t, "/planes/radius/local/resourceGroups/radius-test-rg/providers/Applications.Core/applications/test-app", resource.ID)
		require.Equal(t, "test-app", resource.Name)
		require.Equal(t, "Applications.Core/applications", resource.Type)
		require.Equal(t, "global", resource.Location)
		require.Equal(t, map[string]string{
			"tag-name": "tag-value",
		}, resource.Tags)
	}
}

type fakeResource struct{}

func (f *fakeResource) ResourceTypeName() string {
	return "FakeResource"
}

func TestTrackedResource_ConvertFromValidation(t *testing.T) {
	validationTests := []struct {
		src conv.DataModelInterface
		err error
	}{
		{&fakeResource{}, conv.ErrInvalidModelConversion},
		{nil, conv.ErrInvalidModelConversion},
	}

	for _, tc := range validationTests {
		versioned := &TrackedResource{}
		err := versioned.ConvertFrom(tc.src)
		require.ErrorAs(t, tc.err, &err)
	}
}
