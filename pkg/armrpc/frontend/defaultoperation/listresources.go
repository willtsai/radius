// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package defaultoperation

import (
	"context"
	"net/http"

	"github.com/project-radius/radius/pkg/armrpc/api/conv"
	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
	ctrl "github.com/project-radius/radius/pkg/armrpc/frontend/controller"
	"github.com/project-radius/radius/pkg/armrpc/rest"
	"github.com/project-radius/radius/pkg/ucp/store"
)

// ListResources is the controller implementation to get the list of resources in resource group.
type ListResources[P interface {
	*T
	conv.ResourceDataModel
}, T any] struct {
	ctrl.Operation[P, T]
}

// NewListResources creates a new ListResources instance.
func NewListResources[P interface {
	*T
	conv.ResourceDataModel
}, T any](opts ctrl.Options, ctrlOpts ctrl.ResourceOptions[T]) (ctrl.Controller, error) {
	return &ListResources[P, T]{
		ctrl.NewOperation[P](opts, ctrlOpts),
	}, nil
}

// Run fetches the list of all resources in resourcegroups.
func (e *ListResources[P, T]) Run(ctx context.Context, w http.ResponseWriter, req *http.Request) (rest.Response, error) {
	serviceCtx := v1.ARMRequestContextFromContext(ctx)

	query := store.Query{
		RootScope:    serviceCtx.ResourceID.RootScope(),
		ResourceType: serviceCtx.ResourceID.Type(),
	}

	result, err := e.StorageClient().Query(ctx, query, store.WithPaginationToken(serviceCtx.SkipToken), store.WithMaxQueryItemCount(serviceCtx.Top))
	if err != nil {
		return nil, err
	}

	pagination, err := e.createPaginationResponse(ctx, req, result)

	return rest.NewOKResponse(pagination), err
}

func (e *ListResources[P, T]) createPaginationResponse(ctx context.Context, req *http.Request, result *store.ObjectQueryResult) (*v1.PaginatedList, error) {
	serviceCtx := v1.ARMRequestContextFromContext(ctx)

	items := []interface{}{}
	for _, item := range result.Items {
		resource := new(T)
		if err := item.As(resource); err != nil {
			return nil, err
		}

		versioned, err := e.ResponseConverter()(resource, serviceCtx.APIVersion)
		if err != nil {
			return nil, err
		}

		items = append(items, versioned)
	}

	return &v1.PaginatedList{
		Value:    items,
		NextLink: ctrl.GetNextLinkURL(ctx, req, result.PaginationToken),
	}, nil
}