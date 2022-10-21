// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package controller

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/project-radius/radius/pkg/armrpc/api/conv"
	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
	armrpc_controller "github.com/project-radius/radius/pkg/armrpc/frontend/controller"
	armrpc_rest "github.com/project-radius/radius/pkg/armrpc/rest"
	"github.com/project-radius/radius/pkg/radlogger"
	"github.com/project-radius/radius/pkg/ucp/aws"
	"github.com/project-radius/radius/pkg/ucp/resources"
	"github.com/project-radius/radius/pkg/ucp/store"
)

// Options represents controller options.
type Options struct {
	BasePath                string
	DB                      store.StorageClient
	Address                 string
	AWSCloudControlClient   aws.AWSCloudControlClient
	AWSCloudFormationClient aws.AWSCloudFormationClient
}

type ControllerFunc func(Options) (armrpc_controller.Controller, error)

type HandlerOptions struct {
	ParentRouter   *mux.Router
	ResourceType   string
	Path           string
	Method         v1.OperationMethod
	HandlerFactory ControllerFunc
}

// BaseController is the base operation controller.
type BaseController struct {
	Options Options
}

// NewBaseController creates BaseController instance.
func NewBaseController(options Options) BaseController {
	return BaseController{
		options,
	}
}

func RegisterHandler(ctx context.Context, opts HandlerOptions, ctrlOpts Options) error {
	ctrl, err := opts.HandlerFactory(ctrlOpts)
	if err != nil {
		return err
	}

	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		response, err := ctrl.Run(ctx, w, req)
		if err != nil {
			HandleError(ctx, w, req, err)
			return
		}
		if response != nil {
			err = response.Apply(ctx, w, req)
			if err != nil {
				HandleError(ctx, w, req, err)
				return
			}
		}
	}

	ot := v1.OperationType{Type: opts.Path, Method: opts.Method}
	if opts.Method != "" {
		opts.ParentRouter.Methods(opts.Method.HTTPMethod()).HandlerFunc(fn).Name(ot.String())
	} else {
		// Path is used to proxy plane request irrespective of the http method
		opts.ParentRouter.PathPrefix(opts.Path).HandlerFunc(fn).Name(ot.String())
	}
	return nil
}

// StorageClient gets storage client for this controller.
func (b *BaseController) StorageClient() store.StorageClient {
	return b.Options.DB
}

// GetResource is the helper to get the resource via storage client.
func (c *BaseController) GetResource(ctx context.Context, id string, out interface{}) (etag string, err error) {
	etag = ""
	var res *store.Object
	if res, err = c.StorageClient().Get(ctx, id); err == nil && res != nil {
		if err = res.As(out); err == nil {
			etag = res.ETag
			return
		}
	}
	return
}

// SaveResource is the helper to save the resource via storage client.
func (c *BaseController) SaveResource(ctx context.Context, id string, in interface{}, etag string) (*store.Object, error) {
	nr := &store.Object{
		Metadata: store.Metadata{
			ID: id,
		},
		Data: in,
	}
	err := c.StorageClient().Save(ctx, nr, store.WithETag(etag))
	if err != nil {
		return nil, err
	}
	return nr, nil
}

// DeleteResource is the helper to delete the resource via storage client.
func (c *BaseController) DeleteResource(ctx context.Context, id string, etag string) error {
	err := c.StorageClient().Delete(ctx, id, store.WithETag(etag))
	if err != nil {
		return err
	}
	return nil
}

// Responds with an HTTP 500
func HandleError(ctx context.Context, w http.ResponseWriter, req *http.Request, err error) {
	logger := radlogger.GetLogger(ctx)

	var response armrpc_rest.Response
	// Try to use the ARM format to send back the error info
	// if the error is due to api conversion failure return bad request
	switch v := err.(type) {
	case *conv.ErrModelConversion:
		response = armrpc_rest.NewBadRequestARMResponse(v1.ErrorResponse{
			Error: v1.ErrorDetails{
				Code:    v1.CodeHTTPRequestPayloadAPISpecValidationFailed,
				Message: err.Error(),
			},
		})
	case *conv.ErrClientRP:
		response = armrpc_rest.NewBadRequestARMResponse(v1.ErrorResponse{
			Error: v1.ErrorDetails{
				Code:    v.Code,
				Message: v.Message,
			},
		})
	default:
		if err.Error() == conv.ErrInvalidModelConversion.Error() {
			response = armrpc_rest.NewBadRequestARMResponse(v1.ErrorResponse{
				Error: v1.ErrorDetails{
					Code:    v1.CodeHTTPRequestPayloadAPISpecValidationFailed,
					Message: err.Error(),
				},
			})
		} else {
			logger.V(radlogger.Debug).Error(err, "unhandled error")
			response = armrpc_rest.NewInternalServerErrorARMResponse(v1.ErrorResponse{
				Error: v1.ErrorDetails{
					Code:    v1.CodeInternal,
					Message: err.Error(),
				},
			})
		}
	}
	err = response.Apply(ctx, w, req)
	if err != nil {
		body := v1.ErrorResponse{
			Error: v1.ErrorDetails{
				Code:    v1.CodeInternal,
				Message: err.Error(),
			},
		}
		// There's no way to recover if we fail writing here, we likly partially wrote to the response stream.
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error(err, fmt.Sprintf("error writing marshaled %T bytes to output", body))
	}
}

func (b *BaseController) GetRelativePath(path string) string {
	trimmedPath := strings.TrimPrefix(path, b.Options.BasePath)
	return trimmedPath
}

func (b *BaseController) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	path := b.GetRelativePath(r.URL.Path)
	restResponse := armrpc_rest.NewNoResourceMatchResponse(path)
	err := restResponse.Apply(r.Context(), w, r)
	if err != nil {
		HandleError(r.Context(), w, r, err)
		return
	}
}

func (b *BaseController) MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	path := b.GetRelativePath(r.URL.Path)
	target := ""
	if rID, err := resources.Parse(path); err == nil {
		target = rID.Type() + "/" + rID.Name()
	}
	restResponse := armrpc_rest.NewMethodNotAllowedResponse(target, fmt.Sprintf("The request method '%s' is invalid.", r.Method))
	if err := restResponse.Apply(r.Context(), w, r); err != nil {
		HandleError(r.Context(), w, r, err)
	}
}

func ReadRequestBody(req *http.Request) ([]byte, error) {
	defer req.Body.Close()
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %w", err)
	}
	return data, nil
}

func ConfigureDefaultHandlers(router *mux.Router, opts Options) {
	b := NewBaseController(opts)
	router.NotFoundHandler = http.HandlerFunc(b.NotFoundHandler)
	router.MethodNotAllowedHandler = http.HandlerFunc(b.MethodNotAllowedHandler)
}