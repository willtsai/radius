// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------
package resourcegroups

import (
	"context"
	http "net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/golang/mock/gomock"
	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
	"github.com/project-radius/radius/pkg/ucp/api/v20220315privatepreview"
	"github.com/project-radius/radius/pkg/ucp/datamodel"
	ctrl "github.com/project-radius/radius/pkg/ucp/frontend/controller"
	"github.com/project-radius/radius/pkg/ucp/rest"
	"github.com/project-radius/radius/pkg/ucp/store"
	"github.com/project-radius/radius/pkg/ucp/util/testcontext"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func Test_ResourceGroup_ListResources(t *testing.T) {
	ctx, cancel := testcontext.New(t)
	defer cancel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("empty", func(t *testing.T) {
		storage := store.NewMockStorageClient(mockCtrl)

		rgCtrl, err := NewListResourceGroupResources(ctrl.Options{
			BasePath: "/apis/api.ucp.dev/v1alpha3",
			DB:       storage,
		})
		require.NoError(t, err)

		path := "/apis/api.ucp.dev/v1alpha3/planes/radius/local/resourceGroups/test-rg/resources"

		query := store.Query{
			RootScope:    "/planes/radius/local/resourceGroups/test-rg",
			ResourceType: "System.Resources/resources",
		}

		expected := v20220315privatepreview.TrackedResourceList{}
		expectedResponse := rest.NewOKResponse(expected)

		storage.EXPECT().
			Query(gomock.Any(), query).
			DoAndReturn(func(ctx context.Context, query store.Query, options ...store.QueryOptions) (*store.ObjectQueryResult, error) {
				return &store.ObjectQueryResult{
					// Empty result
					Items: []store.Object{},
				}, nil
			})
		request, err := http.NewRequest(http.MethodGet, path, nil)
		require.NoError(t, err)
		actualResponse, err := rgCtrl.Run(ctx, nil, request)
		require.NoError(t, err)
		assert.DeepEqual(t, expectedResponse, actualResponse)
	})

	t.Run("some", func(t *testing.T) {
		storage := store.NewMockStorageClient(mockCtrl)

		rgCtrl, err := NewListResourceGroupResources(ctrl.Options{
			BasePath: "/apis/api.ucp.dev/v1alpha3",
			DB:       storage,
		})
		require.NoError(t, err)

		path := "/apis/api.ucp.dev/v1alpha3/planes/radius/local/resourceGroups/test-rg/resources"

		query := store.Query{
			RootScope:    "/planes/radius/local/resourceGroups/test-rg",
			ResourceType: "System.Resources/resources",
		}

		expected := v20220315privatepreview.TrackedResourceList{
			Value: []*v20220315privatepreview.TrackedResource{
				{
					Location: to.Ptr("global"),
					Tags: map[string]*string{
						"test": to.Ptr("value"),
					},
					ID:   to.Ptr("/planes/radius/local/resourceGroups/test-rg/providers/Applications.Core/applications/test-app"),
					Name: to.Ptr("test-app"),
					Type: to.Ptr("Applications.Core/applications"),
				},
				{
					Location: to.Ptr("global"),
					Tags:     map[string]*string{},
					ID:       to.Ptr("/planes/radius/local/resourceGroups/test-rg/providers/Applications.Core/environments/test-env"),
					Name:     to.Ptr("test-env"),
					Type:     to.Ptr("Applications.Core/environments"),
				},
			},
		}
		expectedResponse := rest.NewOKResponse(expected)

		storage.EXPECT().
			Query(gomock.Any(), query).
			DoAndReturn(func(ctx context.Context, query store.Query, options ...store.QueryOptions) (*store.ObjectQueryResult, error) {
				return &store.ObjectQueryResult{
					Items: []store.Object{
						{
							Metadata: store.Metadata{
								ID: "/planes/radius/local/resourceGroups/test-rg/providers/Applications.Core/applications/test-app",
							},
							Data: datamodel.TrackedResource{
								TrackedResource: v1.TrackedResource{
									Location: "global",
									Tags: map[string]string{
										"test": "value",
									},
									ID:   "/planes/radius/local/resourceGroups/test-rg/providers/Applications.Core/applications/test-app",
									Name: "test-app",
									Type: "Applications.Core/applications",
								},
							},
						},
						{
							Metadata: store.Metadata{
								ID: "/planes/radius/local/resourceGroups/test-rg/providers/Applications.Core/applications/test-app",
							},
							Data: datamodel.TrackedResource{
								TrackedResource: v1.TrackedResource{
									Location: "global",
									Tags:     nil,
									ID:       "/planes/radius/local/resourceGroups/test-rg/providers/Applications.Core/environments/test-env",
									Name:     "test-env",
									Type:     "Applications.Core/environments",
								},
							},
						},
					},
				}, nil
			})

		request, err := http.NewRequest(http.MethodGet, path, nil)
		require.NoError(t, err)
		actualResponse, err := rgCtrl.Run(ctx, nil, request)
		require.NoError(t, err)
		assert.DeepEqual(t, expectedResponse, actualResponse)
	})
}
