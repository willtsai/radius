//go:build go1.16
// +build go1.16

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

package v20220315privatepreview

const (
	module = "v20220315privatepreview"
	version = "v0.0.1"
)

// CreatedByType - The type of identity that created the resource.
type CreatedByType string

const (
	CreatedByTypeApplication CreatedByType = "Application"
	CreatedByTypeKey CreatedByType = "Key"
	CreatedByTypeManagedIdentity CreatedByType = "ManagedIdentity"
	CreatedByTypeUser CreatedByType = "User"
)

// PossibleCreatedByTypeValues returns the possible values for the CreatedByType const type.
func PossibleCreatedByTypeValues() []CreatedByType {
	return []CreatedByType{	
		CreatedByTypeApplication,
		CreatedByTypeKey,
		CreatedByTypeManagedIdentity,
		CreatedByTypeUser,
	}
}

// ToPtr returns a *CreatedByType pointing to the current value.
func (c CreatedByType) ToPtr() *CreatedByType {
	return &c
}

// DaprPubSubBrokerPropertiesKind - The DaprPubSubProperties kind
type DaprPubSubBrokerPropertiesKind string

const (
	DaprPubSubBrokerPropertiesKindGeneric DaprPubSubBrokerPropertiesKind = "generic"
	DaprPubSubBrokerPropertiesKindPubsubAzureServicebus DaprPubSubBrokerPropertiesKind = "pubsub.azure.servicebus"
)

// PossibleDaprPubSubBrokerPropertiesKindValues returns the possible values for the DaprPubSubBrokerPropertiesKind const type.
func PossibleDaprPubSubBrokerPropertiesKindValues() []DaprPubSubBrokerPropertiesKind {
	return []DaprPubSubBrokerPropertiesKind{	
		DaprPubSubBrokerPropertiesKindGeneric,
		DaprPubSubBrokerPropertiesKindPubsubAzureServicebus,
	}
}

// ToPtr returns a *DaprPubSubBrokerPropertiesKind pointing to the current value.
func (c DaprPubSubBrokerPropertiesKind) ToPtr() *DaprPubSubBrokerPropertiesKind {
	return &c
}

// DaprSecretStorePropertiesKind - Radius kind for Dapr Secret Store
type DaprSecretStorePropertiesKind string

const (
	DaprSecretStorePropertiesKindGeneric DaprSecretStorePropertiesKind = "generic"
)

// PossibleDaprSecretStorePropertiesKindValues returns the possible values for the DaprSecretStorePropertiesKind const type.
func PossibleDaprSecretStorePropertiesKindValues() []DaprSecretStorePropertiesKind {
	return []DaprSecretStorePropertiesKind{	
		DaprSecretStorePropertiesKindGeneric,
	}
}

// ToPtr returns a *DaprSecretStorePropertiesKind pointing to the current value.
func (c DaprSecretStorePropertiesKind) ToPtr() *DaprSecretStorePropertiesKind {
	return &c
}

// DaprStateStorePropertiesKind - The Dapr StateStore kind
type DaprStateStorePropertiesKind string

const (
	DaprStateStorePropertiesKindGeneric DaprStateStorePropertiesKind = "generic"
	DaprStateStorePropertiesKindStateAzureTablestorage DaprStateStorePropertiesKind = "state.azure.tablestorage"
	DaprStateStorePropertiesKindStateSqlserver DaprStateStorePropertiesKind = "state.sqlserver"
)

// PossibleDaprStateStorePropertiesKindValues returns the possible values for the DaprStateStorePropertiesKind const type.
func PossibleDaprStateStorePropertiesKindValues() []DaprStateStorePropertiesKind {
	return []DaprStateStorePropertiesKind{	
		DaprStateStorePropertiesKindGeneric,
		DaprStateStorePropertiesKindStateAzureTablestorage,
		DaprStateStorePropertiesKindStateSqlserver,
	}
}

// ToPtr returns a *DaprStateStorePropertiesKind pointing to the current value.
func (c DaprStateStorePropertiesKind) ToPtr() *DaprStateStorePropertiesKind {
	return &c
}

// ProvisioningState - Provisioning state of the connector at the time the operation was called
type ProvisioningState string

const (
	ProvisioningStateAccepted ProvisioningState = "Accepted"
	ProvisioningStateCanceled ProvisioningState = "Canceled"
	ProvisioningStateDeleting ProvisioningState = "Deleting"
	ProvisioningStateFailed ProvisioningState = "Failed"
	ProvisioningStateProvisioning ProvisioningState = "Provisioning"
	ProvisioningStateSucceeded ProvisioningState = "Succeeded"
	ProvisioningStateUpdating ProvisioningState = "Updating"
)

// PossibleProvisioningStateValues returns the possible values for the ProvisioningState const type.
func PossibleProvisioningStateValues() []ProvisioningState {
	return []ProvisioningState{	
		ProvisioningStateAccepted,
		ProvisioningStateCanceled,
		ProvisioningStateDeleting,
		ProvisioningStateFailed,
		ProvisioningStateProvisioning,
		ProvisioningStateSucceeded,
		ProvisioningStateUpdating,
	}
}

// ToPtr returns a *ProvisioningState pointing to the current value.
func (c ProvisioningState) ToPtr() *ProvisioningState {
	return &c
}
