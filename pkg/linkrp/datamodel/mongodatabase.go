// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package datamodel

import (
	v1 "github.com/project-radius/radius/pkg/armrpc/api/v1"
	"github.com/project-radius/radius/pkg/rp"
)

// MongoDatabase represents MongoDatabase link resource.
type MongoDatabase struct {
	v1.BaseResource

	// LinkMetadata represents internal DataModel properties common to all link types.
	LinkMetadata

	// Properties is the properties of the resource.
	Properties MongoDatabaseProperties `json:"properties"`
}

// MongoDatabaseProperties represents the properties of MongoDatabase resource.
type MongoDatabaseProperties struct {
	rp.BasicResourceProperties
	MongoDatabaseResourceProperties
	MongoDatabaseRecipeProperties
	MongoDatabaseValuesProperties
	Secrets           MongoDatabaseSecrets `json:"secrets,omitempty"`
	ProvisioningState v1.ProvisioningState `json:"provisioningState,omitempty"`
	Mode              LinkMode             `json:"mode"`
}

// Secrets values consisting of secrets provided for the resource
type MongoDatabaseSecrets struct {
	Username         string `json:"username"`
	Password         string `json:"password"`
	ConnectionString string `json:"connectionString"`
}

func (mongoSecrets MongoDatabaseSecrets) IsEmpty() bool {
	return mongoSecrets == MongoDatabaseSecrets{}
}

func (mongoSecrets MongoDatabaseSecrets) ResourceTypeName() string {
	return "Applications.Link/mongoDatabases"
}

func (mongo MongoDatabase) ResourceTypeName() string {
	return "Applications.Link/mongoDatabases"
}

type MongoDatabaseValuesProperties struct {
	Host     string `json:"host,omitempty"`
	Port     int32  `json:"port,omitempty"`
	Database string `json:"database,omitempty"`
}

type MongoDatabaseResourceProperties struct {
	Resource string `json:"resource,omitempty"`
}

type MongoDatabaseRecipeProperties struct {
	Recipe LinkRecipe `json:"recipe,omitempty"`
}
