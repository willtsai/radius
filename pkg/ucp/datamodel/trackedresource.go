// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package datamodel

import (
	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
)

// Resource represents a tracking entry inside UCP that corresponds to a tracked resource.
type TrackedResource struct {
	v1.TrackedResource

	// NOTE: we omit SystemData here because we're not the source of truth for the resource. We're just
	// keeping track of its existance. In the future we might want to make SystemData part of the data
	// we expose.

	// NOTE: we include InternalMetadata here, but it refers to the API version of the resources API
	// NOT the underlying resource.

	// InternalMetadata is the internal metadata which is used for conversion.
	v1.InternalMetadata
}

func (_ TrackedResource) ResourceTypeName() string {
	return "System.Resources/resources"
}
