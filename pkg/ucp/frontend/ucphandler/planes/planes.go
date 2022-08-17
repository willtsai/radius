// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------
package planes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	planesdb "github.com/project-radius/radius/pkg/ucp/db/planes"
	resourcegroupsdb "github.com/project-radius/radius/pkg/ucp/db/resourcegroups"
	"github.com/project-radius/radius/pkg/ucp/planes"
	"github.com/project-radius/radius/pkg/ucp/proxy"
	"github.com/project-radius/radius/pkg/ucp/resources"
	"github.com/project-radius/radius/pkg/ucp/rest"
	"github.com/project-radius/radius/pkg/ucp/store"
	"github.com/project-radius/radius/pkg/ucp/ucplog"
)

const (
	PlanesPath = "/planes"
)

//go:generate mockgen -destination=./mock_planes_ucphandler.go -package=planes -self_package github.com/project-radius/radius/pkg/ucp/frontend/ucphandler/planes github.com/project-radius/radius/pkg/ucp/frontend/ucphandler/planes PlanesUCPHandler
type PlanesUCPHandler interface {
	CreateOrUpdate(ctx context.Context, db store.StorageClient, body []byte, path string) (rest.Response, error)
	List(ctx context.Context, db store.StorageClient, path string) (rest.Response, error)
	GetByID(ctx context.Context, db store.StorageClient, path string) (rest.Response, error)
	DeleteByID(ctx context.Context, db store.StorageClient, path string) (rest.Response, error)
	ProxyRequest(ctx context.Context, db store.StorageClient, w http.ResponseWriter, r *http.Request, incomingURL *url.URL) (rest.Response, error)
}

type Options struct {
	Address  string
	BasePath string
}

// NewPlanesUCPHandler creates a new Planes UCP handler
func NewPlanesUCPHandler(options Options) PlanesUCPHandler {
	return &ucpHandler{
		options: options,
	}
}

type ucpHandler struct {
	options Options
}

func (ucp *ucpHandler) CreateOrUpdate(ctx context.Context, db store.StorageClient, body []byte, path string) (rest.Response, error) {
	ctx = ucplog.WrapLogContext(ctx, ucplog.LogFieldRequestPath, path)
	var plane rest.Plane
	err := json.Unmarshal(body, &plane)
	if err != nil {
		return rest.NewBadRequestResponse(err.Error()), nil
	}
	plane.ID = path
	planeType, name, _, err := resources.ExtractPlanesPrefixFromURLPath(path)
	if err != nil {
		return rest.NewBadRequestResponse(err.Error()), nil
	}
	plane.Type = planes.PlaneTypePrefix + "/" + planeType
	plane.Name = name
	planeExists := true
	ID, err := resources.Parse(plane.ID)
	//cannot parse ID something wrong with request
	if err != nil {
		return rest.NewBadRequestResponse(err.Error()), nil
	}

	ctx = ucplog.WrapLogContext(ctx, ucplog.LogFieldPlaneKind, plane.Properties.Kind)
	logger := ucplog.GetLogger(ctx)
	// At least one provider needs to be configured
	if plane.Properties.Kind == rest.PlaneKindUCPNative {
		if plane.Properties.ResourceProviders == nil || len(plane.Properties.ResourceProviders) == 0 {
			err = fmt.Errorf("At least one resource provider must be configured for UCP native plane: %s", plane.Name)
			return rest.NewBadRequestResponse(err.Error()), nil
		}
	} else {
		if plane.Properties.URL == "" {
			err = fmt.Errorf("URL must be specified for plane: %s", plane.Name)
			return rest.NewBadRequestResponse(err.Error()), nil
		}
	}

	_, err = planesdb.GetByID(ctx, db, ID)
	if err != nil {
		if errors.Is(err, &store.ErrNotFound{}) {
			planeExists = false
			logger.Info(fmt.Sprintf("No existing plane %s found in db", ID))
		} else {
			return nil, err
		}
	}
	plane, err = planesdb.Save(ctx, db, plane)
	if err != nil {
		return nil, err
	}
	var restResp rest.Response
	if planeExists {
		restResp = rest.NewOKResponse(plane)
		logger.Info(fmt.Sprintf("Updated plane %s successfully", plane.Name))
	} else {
		restResp = rest.NewCreatedResponse(plane)
		logger.Info(fmt.Sprintf("Created plane %s successfully", plane.Name))
	}
	return restResp, nil
}

func (ucp *ucpHandler) List(ctx context.Context, db store.StorageClient, path string) (rest.Response, error) {
	ctx = ucplog.WrapLogContext(ctx, ucplog.LogFieldRequestPath, path)
	logger := ucplog.GetLogger(ctx)
	var query store.Query
	query.RootScope = path
	query.IsScopeQuery = true
	logger.Info(fmt.Sprintf("Listing planes in scope %s", query.RootScope))
	listOfPlanes, err := planesdb.GetScope(ctx, db, query)
	if err != nil {
		return nil, err
	}
	var ok = rest.NewOKResponse(listOfPlanes)
	return ok, nil
}

func (ucp *ucpHandler) GetByID(ctx context.Context, db store.StorageClient, path string) (rest.Response, error) {
	ctx = ucplog.WrapLogContext(ctx, ucplog.LogFieldRequestPath, path)
	logger := ucplog.GetLogger(ctx)
	id := strings.ToLower(path)
	resourceId, err := resources.Parse(id)
	if err != nil {
		if err != nil {
			return rest.NewBadRequestResponse(err.Error()), nil
		}
	}
	logger.Info(fmt.Sprintf("Getting plane %s from db", resourceId))
	plane, err := planesdb.GetByID(ctx, db, resourceId)
	if err != nil {
		if errors.Is(err, &store.ErrNotFound{}) {
			restResponse := rest.NewNotFoundResponse(path)
			logger.Info(fmt.Sprintf("Plane %s not found in db", resourceId))
			return restResponse, nil
		}
		return nil, err
	}
	restResponse := rest.NewOKResponse(plane)
	return restResponse, nil
}

func (ucp *ucpHandler) DeleteByID(ctx context.Context, db store.StorageClient, path string) (rest.Response, error) {
	ctx = ucplog.WrapLogContext(ctx, ucplog.LogFieldRequestPath, path)
	logger := ucplog.GetLogger(ctx)
	resourceId, err := resources.Parse(path)
	if err != nil {
		return rest.NewBadRequestResponse(err.Error()), nil
	}
	_, err = planesdb.GetByID(ctx, db, resourceId)
	if err != nil {
		if errors.Is(err, &store.ErrNotFound{}) {
			restResponse := rest.NewNoContentResponse()
			return restResponse, nil
		}
		return nil, err
	}
	err = planesdb.DeleteByID(ctx, db, resourceId)
	if err != nil {
		return nil, err
	}
	restResponse := rest.NewNoContentResponse()
	logger.Info(fmt.Sprintf("Successfully deleted plane %s", resourceId))
	return restResponse, nil
}

func (ucp *ucpHandler) ProxyRequest(ctx context.Context, db store.StorageClient, w http.ResponseWriter, r *http.Request, incomingURL *url.URL) (rest.Response, error) {
	ctx = ucplog.WrapLogContext(ctx, ucplog.LogFieldRequestPath, incomingURL)
	planeType, name, _, err := resources.ExtractPlanesPrefixFromURLPath(incomingURL.Path)
	if err != nil {
		return nil, err
	}

	// Lookup the plane
	planePath := PlanesPath + "/" + planeType + "/" + name
	planeID, err := resources.Parse(planePath)
	if err != nil {
		if err != nil {
			return nil, err
		}
	}
	plane, err := planesdb.GetByID(ctx, db, planeID)
	if err != nil {
		if errors.Is(err, &store.ErrNotFound{}) {
			return rest.NewNotFoundResponse(planePath), nil
		}
		return nil, err
	}

	ctx = ucplog.WrapLogContext(ctx,
		ucplog.LogFieldPlaneID, planeID,
		ucplog.LogFieldPlaneKind, plane.Properties.Kind)
	if plane.Properties.Kind == rest.PlaneKindUCPNative {
		// Check if the resource group exists
		id, err := resources.Parse(incomingURL.Path)
		if err != nil {
			return nil, err
		}
		rgPath := id.RootScope()
		rgID, err := resources.Parse(rgPath)
		if err != nil {
			return nil, err
		}
		_, err = resourcegroupsdb.GetByID(ctx, db, rgID)
		if err != nil {
			if errors.Is(err, &store.ErrNotFound{}) {
				return rest.NewNotFoundResponse(rgID.String()), nil
			}
			return nil, err
		}
	}

	// Get the resource provider
	resourceID, err := resources.Parse(incomingURL.Path)
	if err != nil {
		return nil, err
	}

	if resourceID.ProviderNamespace() == "" {
		err = fmt.Errorf("Invalid resourceID specified with no provider")
		return rest.NewBadRequestResponse(err.Error()), err
	}

	// Lookup the resource providers configured to determine the URL to proxy to
	// Not using map lookups to enable case insensitive comparisons
	// We need to preserve the case while storing data in DB and therefore iterating for case
	// insensitive comparisons

	var proxyURL string
	if plane.Properties.Kind == rest.PlaneKindUCPNative {
		proxyURL = plane.LookupResourceProvider(resourceID.ProviderNamespace())
		if proxyURL == "" {
			err = fmt.Errorf("Provider %s not configured", resourceID.ProviderNamespace())
			return nil, err
		}
		ctx = ucplog.WrapLogContext(ctx,
			ucplog.LogFieldPlaneURL, proxyURL,
			ucplog.LogFieldProvider, resourceID.ProviderNamespace())
	} else {
		// For a non UCP-native plane, the configuration should have a URL to which
		// all the requests will be forwarded
		proxyURL = plane.Properties.URL
		ctx = ucplog.WrapLogContext(ctx, ucplog.LogFieldPlaneURL, proxyURL)
	}

	downstream, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	options := proxy.ReverseProxyOptions{
		RoundTripper:     http.DefaultTransport,
		ProxyAddress:     ucp.options.Address,
		TrimPlanesPrefix: (plane.Properties.Kind != rest.PlaneKindUCPNative),
	}

	// // // As per https://github.com/golang/go/issues/28940#issuecomment-441749380, the way to check
	// // // for http vs https is check the TLS field
	httpScheme := "http"
	if r.TLS != nil {
		httpScheme = "https"
	}

	ctx = ucplog.WrapLogContext(ctx, ucplog.LogFieldHTTPScheme, httpScheme)

	requestInfo := proxy.UCPRequestInfo{
		PlaneURL:   proxyURL,
		PlaneKind:  plane.Properties.Kind,
		PlaneID:    planePath,
		HTTPScheme: httpScheme,
		// The Host field in the request that the client makes to UCP contains the UCP Host address
		// That address will be used to construct the URL for reverse proxying
		UCPHost: r.Host + ucp.options.BasePath,
	}

	url, err := url.Parse(incomingURL.Path)
	if err != nil {
		return nil, err
	}

	// Preserving the query strings on the incoming url on the newly constructed url
	url.RawQuery = incomingURL.Query().Encode()
	r.URL = url
	r.Header.Set("X-Forwarded-Proto", httpScheme)

	ctx = context.WithValue(ctx, proxy.UCPRequestInfoField, requestInfo)
	sender := proxy.NewARMProxy(options, downstream, nil)

	logger := ucplog.GetLogger(ctx)
	logger.Info(fmt.Sprintf("Proxying request to target %s", proxyURL))
	sender.ServeHTTP(w, r.WithContext(ctx))

	// The upstream response has already been sent at this point. Therefore, return nil response here
	return nil, nil
}