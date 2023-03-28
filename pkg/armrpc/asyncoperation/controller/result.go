// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package controller

import (
	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
)

// Result is the response of async operation controller.
type Result struct {
	// Requeue tells the Controller to requeue the reconcile key. Defaults to false.
	Requeue bool

	// Error represents the error when status is Cancelled or Failed.
	Error *v1.ErrorDetails

	// state represents the provisioning status.
	state *v1.ProvisioningState
}

// NewCanceledResult creates the canceled asynchronous operation result.
func NewCanceledResult(message string) Result {
	r := Result{}
	r.SetCanceled(message)
	return r
}

// NewFailedResult creates the failed asynchronous operation result.
func NewFailedResult(err v1.ErrorDetails) Result {
	r := Result{}
	r.SetFailed(err, false)
	return r
}

// SetFailed sets the error response with Failed status.
func (r *Result) SetFailed(err v1.ErrorDetails, requeue bool) {
	if r == nil {
		r = &Result{}
	}
	r.Requeue = requeue
	r.SetProvisioningState(v1.ProvisioningStateFailed)
	r.Error = &v1.ErrorDetails{
		Code:    err.Code,
		Message: err.Message,
	}
}

// SetCanceled sets the response status to Canceled.
func (r *Result) SetCanceled(message string) {
	if r == nil {
		r = &Result{}
	}
	r.Requeue = false
	r.SetProvisioningState(v1.ProvisioningStateCanceled)
	r.Error = &v1.ErrorDetails{
		Code:    v1.CodeOperationCanceled,
		Message: message,
	}
}

// SetProvisioningState sets provisioning state.
func (r *Result) SetProvisioningState(s v1.ProvisioningState) {
	r.state = &s
}

// ProvisioningState gets the provisioning state of the result.
func (r *Result) ProvisioningState() v1.ProvisioningState {
	if r.state == nil {
		return v1.ProvisioningStateSucceeded
	}
	return *r.state
}