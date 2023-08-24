//go:build go1.18
// +build go1.18

// Licensed under the Apache License, Version 2.0 . See LICENSE in the repository root for license information.
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.
// DO NOT EDIT.

package v20220315privatepreview

const (
	moduleName = "v20220315privatepreview"
	moduleVersion = "v0.0.1"
)

// ActionType - Enum. Indicates the action type. "Internal" refers to actions that are for internal only APIs.
type ActionType string

const (
	ActionTypeInternal ActionType = "Internal"
)

// PossibleActionTypeValues returns the possible values for the ActionType const type.
func PossibleActionTypeValues() []ActionType {
	return []ActionType{	
		ActionTypeInternal,
	}
}

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

// IdentitySettingKind - IdentitySettingKind is the kind of supported external identity setting
type IdentitySettingKind string

const (
	// IdentitySettingKindAzureComWorkload - azure ad workload identity
	IdentitySettingKindAzureComWorkload IdentitySettingKind = "azure.com.workload"
	// IdentitySettingKindUndefined - undefined identity
	IdentitySettingKindUndefined IdentitySettingKind = "undefined"
)

// PossibleIdentitySettingKindValues returns the possible values for the IdentitySettingKind const type.
func PossibleIdentitySettingKindValues() []IdentitySettingKind {
	return []IdentitySettingKind{	
		IdentitySettingKindAzureComWorkload,
		IdentitySettingKindUndefined,
	}
}

// Origin - The intended executor of the operation; as in Resource Based Access Control (RBAC) and audit logs UX. Default
// value is "user,system"
type Origin string

const (
	OriginSystem Origin = "system"
	OriginUser Origin = "user"
	OriginUserSystem Origin = "user,system"
)

// PossibleOriginValues returns the possible values for the Origin const type.
func PossibleOriginValues() []Origin {
	return []Origin{	
		OriginSystem,
		OriginUser,
		OriginUserSystem,
	}
}

// ProvisioningState - Provisioning state of the portable resource at the time the operation was called
type ProvisioningState string

const (
	// ProvisioningStateAccepted - The resource create request has been accepted
	ProvisioningStateAccepted ProvisioningState = "Accepted"
	// ProvisioningStateCanceled - Resource creation was canceled.
	ProvisioningStateCanceled ProvisioningState = "Canceled"
	// ProvisioningStateDeleting - The resource is being deleted
	ProvisioningStateDeleting ProvisioningState = "Deleting"
	// ProvisioningStateFailed - Resource creation failed.
	ProvisioningStateFailed ProvisioningState = "Failed"
	// ProvisioningStateProvisioning - The resource is being provisioned
	ProvisioningStateProvisioning ProvisioningState = "Provisioning"
	// ProvisioningStateSucceeded - Resource has been created.
	ProvisioningStateSucceeded ProvisioningState = "Succeeded"
	// ProvisioningStateUpdating - The resource is updating
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

// ResourceProvisioning - Specifies how the underlying service/resource is provisioned and managed. Available values are 'recipe',
// where Radius manages the lifecycle of the resource through a Recipe, and 'manual', where a user
// manages the resource and provides the values.
type ResourceProvisioning string

const (
	// ResourceProvisioningManual - The resource lifecycle will be managed by the user
	ResourceProvisioningManual ResourceProvisioning = "manual"
	// ResourceProvisioningRecipe - The resource lifecycle will be managed by Radius
	ResourceProvisioningRecipe ResourceProvisioning = "recipe"
)

// PossibleResourceProvisioningValues returns the possible values for the ResourceProvisioning const type.
func PossibleResourceProvisioningValues() []ResourceProvisioning {
	return []ResourceProvisioning{	
		ResourceProvisioningManual,
		ResourceProvisioningRecipe,
	}
}

// Versions - Supported API versions for the Applications.Dapr resource provider.
type Versions string

const (
	// VersionsV20220315Privatepreview - 2022-03-15-privatepreview
	VersionsV20220315Privatepreview Versions = "2022-03-15-privatepreview"
)

// PossibleVersionsValues returns the possible values for the Versions const type.
func PossibleVersionsValues() []Versions {
	return []Versions{	
		VersionsV20220315Privatepreview,
	}
}

