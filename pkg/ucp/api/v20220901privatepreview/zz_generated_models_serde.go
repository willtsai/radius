//go:build go1.18
// +build go1.18

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.
// DO NOT EDIT.

package v20220901privatepreview

import (
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"reflect"
)

// MarshalJSON implements the json.Marshaller interface for type AWSAccessKeyCredentialProperties.
func (a AWSAccessKeyCredentialProperties) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "accessKeyId", a.AccessKeyID)
	objectMap["kind"] = "AccessKey"
	populate(objectMap, "provisioningState", a.ProvisioningState)
	populate(objectMap, "secretAccessKey", a.SecretAccessKey)
	populate(objectMap, "storage", a.Storage)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type AWSAccessKeyCredentialProperties.
func (a *AWSAccessKeyCredentialProperties) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", a, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "accessKeyId":
				err = unpopulate(val, "AccessKeyID", &a.AccessKeyID)
				delete(rawMsg, key)
		case "kind":
				err = unpopulate(val, "Kind", &a.Kind)
				delete(rawMsg, key)
		case "provisioningState":
				err = unpopulate(val, "ProvisioningState", &a.ProvisioningState)
				delete(rawMsg, key)
		case "secretAccessKey":
				err = unpopulate(val, "SecretAccessKey", &a.SecretAccessKey)
				delete(rawMsg, key)
		case "storage":
				a.Storage, err = unmarshalCredentialStoragePropertiesClassification(val)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", a, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type AWSCredentialProperties.
func (a AWSCredentialProperties) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	objectMap["kind"] = a.Kind
	populate(objectMap, "provisioningState", a.ProvisioningState)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type AWSCredentialProperties.
func (a *AWSCredentialProperties) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", a, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "kind":
				err = unpopulate(val, "Kind", &a.Kind)
				delete(rawMsg, key)
		case "provisioningState":
				err = unpopulate(val, "ProvisioningState", &a.ProvisioningState)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", a, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type AWSCredentialResource.
func (a AWSCredentialResource) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "id", a.ID)
	populate(objectMap, "location", a.Location)
	populate(objectMap, "name", a.Name)
	populate(objectMap, "properties", a.Properties)
	populate(objectMap, "systemData", a.SystemData)
	populate(objectMap, "tags", a.Tags)
	populate(objectMap, "type", a.Type)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type AWSCredentialResource.
func (a *AWSCredentialResource) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", a, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "id":
				err = unpopulate(val, "ID", &a.ID)
				delete(rawMsg, key)
		case "location":
				err = unpopulate(val, "Location", &a.Location)
				delete(rawMsg, key)
		case "name":
				err = unpopulate(val, "Name", &a.Name)
				delete(rawMsg, key)
		case "properties":
				a.Properties, err = unmarshalAWSCredentialPropertiesClassification(val)
				delete(rawMsg, key)
		case "systemData":
				err = unpopulate(val, "SystemData", &a.SystemData)
				delete(rawMsg, key)
		case "tags":
				err = unpopulate(val, "Tags", &a.Tags)
				delete(rawMsg, key)
		case "type":
				err = unpopulate(val, "Type", &a.Type)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", a, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type AWSCredentialResourceListResult.
func (a AWSCredentialResourceListResult) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "nextLink", a.NextLink)
	populate(objectMap, "value", a.Value)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type AWSCredentialResourceListResult.
func (a *AWSCredentialResourceListResult) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", a, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "nextLink":
				err = unpopulate(val, "NextLink", &a.NextLink)
				delete(rawMsg, key)
		case "value":
				err = unpopulate(val, "Value", &a.Value)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", a, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type AzureCredentialProperties.
func (a AzureCredentialProperties) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	objectMap["kind"] = a.Kind
	populate(objectMap, "provisioningState", a.ProvisioningState)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type AzureCredentialProperties.
func (a *AzureCredentialProperties) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", a, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "kind":
				err = unpopulate(val, "Kind", &a.Kind)
				delete(rawMsg, key)
		case "provisioningState":
				err = unpopulate(val, "ProvisioningState", &a.ProvisioningState)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", a, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type AzureCredentialResource.
func (a AzureCredentialResource) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "id", a.ID)
	populate(objectMap, "location", a.Location)
	populate(objectMap, "name", a.Name)
	populate(objectMap, "properties", a.Properties)
	populate(objectMap, "systemData", a.SystemData)
	populate(objectMap, "tags", a.Tags)
	populate(objectMap, "type", a.Type)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type AzureCredentialResource.
func (a *AzureCredentialResource) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", a, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "id":
				err = unpopulate(val, "ID", &a.ID)
				delete(rawMsg, key)
		case "location":
				err = unpopulate(val, "Location", &a.Location)
				delete(rawMsg, key)
		case "name":
				err = unpopulate(val, "Name", &a.Name)
				delete(rawMsg, key)
		case "properties":
				a.Properties, err = unmarshalAzureCredentialPropertiesClassification(val)
				delete(rawMsg, key)
		case "systemData":
				err = unpopulate(val, "SystemData", &a.SystemData)
				delete(rawMsg, key)
		case "tags":
				err = unpopulate(val, "Tags", &a.Tags)
				delete(rawMsg, key)
		case "type":
				err = unpopulate(val, "Type", &a.Type)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", a, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type AzureCredentialResourceListResult.
func (a AzureCredentialResourceListResult) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "nextLink", a.NextLink)
	populate(objectMap, "value", a.Value)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type AzureCredentialResourceListResult.
func (a *AzureCredentialResourceListResult) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", a, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "nextLink":
				err = unpopulate(val, "NextLink", &a.NextLink)
				delete(rawMsg, key)
		case "value":
				err = unpopulate(val, "Value", &a.Value)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", a, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type AzureServicePrincipalProperties.
func (a AzureServicePrincipalProperties) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "clientId", a.ClientID)
	populate(objectMap, "clientSecret", a.ClientSecret)
	objectMap["kind"] = "ServicePrincipal"
	populate(objectMap, "provisioningState", a.ProvisioningState)
	populate(objectMap, "storage", a.Storage)
	populate(objectMap, "tenantId", a.TenantID)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type AzureServicePrincipalProperties.
func (a *AzureServicePrincipalProperties) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", a, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "clientId":
				err = unpopulate(val, "ClientID", &a.ClientID)
				delete(rawMsg, key)
		case "clientSecret":
				err = unpopulate(val, "ClientSecret", &a.ClientSecret)
				delete(rawMsg, key)
		case "kind":
				err = unpopulate(val, "Kind", &a.Kind)
				delete(rawMsg, key)
		case "provisioningState":
				err = unpopulate(val, "ProvisioningState", &a.ProvisioningState)
				delete(rawMsg, key)
		case "storage":
				a.Storage, err = unmarshalCredentialStoragePropertiesClassification(val)
				delete(rawMsg, key)
		case "tenantId":
				err = unpopulate(val, "TenantID", &a.TenantID)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", a, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type BasicResourceProperties.
func (b BasicResourceProperties) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "provisioningState", b.ProvisioningState)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type BasicResourceProperties.
func (b *BasicResourceProperties) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", b, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "provisioningState":
				err = unpopulate(val, "ProvisioningState", &b.ProvisioningState)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", b, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type CredentialStorageProperties.
func (c CredentialStorageProperties) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	objectMap["kind"] = c.Kind
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type CredentialStorageProperties.
func (c *CredentialStorageProperties) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", c, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "kind":
				err = unpopulate(val, "Kind", &c.Kind)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", c, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type ErrorAdditionalInfo.
func (e ErrorAdditionalInfo) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "info", e.Info)
	populate(objectMap, "type", e.Type)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type ErrorAdditionalInfo.
func (e *ErrorAdditionalInfo) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", e, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "info":
				err = unpopulate(val, "Info", &e.Info)
				delete(rawMsg, key)
		case "type":
				err = unpopulate(val, "Type", &e.Type)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", e, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type ErrorDetail.
func (e ErrorDetail) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "additionalInfo", e.AdditionalInfo)
	populate(objectMap, "code", e.Code)
	populate(objectMap, "details", e.Details)
	populate(objectMap, "message", e.Message)
	populate(objectMap, "target", e.Target)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type ErrorDetail.
func (e *ErrorDetail) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", e, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "additionalInfo":
				err = unpopulate(val, "AdditionalInfo", &e.AdditionalInfo)
				delete(rawMsg, key)
		case "code":
				err = unpopulate(val, "Code", &e.Code)
				delete(rawMsg, key)
		case "details":
				err = unpopulate(val, "Details", &e.Details)
				delete(rawMsg, key)
		case "message":
				err = unpopulate(val, "Message", &e.Message)
				delete(rawMsg, key)
		case "target":
				err = unpopulate(val, "Target", &e.Target)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", e, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type ErrorResponse.
func (e ErrorResponse) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "error", e.Error)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type ErrorResponse.
func (e *ErrorResponse) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", e, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "error":
				err = unpopulate(val, "Error", &e.Error)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", e, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type InternalCredentialStorageProperties.
func (i InternalCredentialStorageProperties) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	objectMap["kind"] = "Internal"
	populate(objectMap, "secretName", i.SecretName)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type InternalCredentialStorageProperties.
func (i *InternalCredentialStorageProperties) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", i, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "kind":
				err = unpopulate(val, "Kind", &i.Kind)
				delete(rawMsg, key)
		case "secretName":
				err = unpopulate(val, "SecretName", &i.SecretName)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", i, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type PlaneResource.
func (p PlaneResource) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "id", p.ID)
	populate(objectMap, "location", p.Location)
	populate(objectMap, "name", p.Name)
	populate(objectMap, "properties", p.Properties)
	populate(objectMap, "systemData", p.SystemData)
	populate(objectMap, "tags", p.Tags)
	populate(objectMap, "type", p.Type)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type PlaneResource.
func (p *PlaneResource) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", p, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "id":
				err = unpopulate(val, "ID", &p.ID)
				delete(rawMsg, key)
		case "location":
				err = unpopulate(val, "Location", &p.Location)
				delete(rawMsg, key)
		case "name":
				err = unpopulate(val, "Name", &p.Name)
				delete(rawMsg, key)
		case "properties":
				err = unpopulate(val, "Properties", &p.Properties)
				delete(rawMsg, key)
		case "systemData":
				err = unpopulate(val, "SystemData", &p.SystemData)
				delete(rawMsg, key)
		case "tags":
				err = unpopulate(val, "Tags", &p.Tags)
				delete(rawMsg, key)
		case "type":
				err = unpopulate(val, "Type", &p.Type)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", p, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type PlaneResourceListResult.
func (p PlaneResourceListResult) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "nextLink", p.NextLink)
	populate(objectMap, "value", p.Value)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type PlaneResourceListResult.
func (p *PlaneResourceListResult) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", p, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "nextLink":
				err = unpopulate(val, "NextLink", &p.NextLink)
				delete(rawMsg, key)
		case "value":
				err = unpopulate(val, "Value", &p.Value)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", p, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type PlaneResourceProperties.
func (p PlaneResourceProperties) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "kind", p.Kind)
	populate(objectMap, "provisioningState", p.ProvisioningState)
	populate(objectMap, "resourceProviders", p.ResourceProviders)
	populate(objectMap, "url", p.URL)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type PlaneResourceProperties.
func (p *PlaneResourceProperties) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", p, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "kind":
				err = unpopulate(val, "Kind", &p.Kind)
				delete(rawMsg, key)
		case "provisioningState":
				err = unpopulate(val, "ProvisioningState", &p.ProvisioningState)
				delete(rawMsg, key)
		case "resourceProviders":
				err = unpopulate(val, "ResourceProviders", &p.ResourceProviders)
				delete(rawMsg, key)
		case "url":
				err = unpopulate(val, "URL", &p.URL)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", p, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type Resource.
func (r Resource) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "id", r.ID)
	populate(objectMap, "name", r.Name)
	populate(objectMap, "systemData", r.SystemData)
	populate(objectMap, "type", r.Type)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type Resource.
func (r *Resource) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", r, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "id":
				err = unpopulate(val, "ID", &r.ID)
				delete(rawMsg, key)
		case "name":
				err = unpopulate(val, "Name", &r.Name)
				delete(rawMsg, key)
		case "systemData":
				err = unpopulate(val, "SystemData", &r.SystemData)
				delete(rawMsg, key)
		case "type":
				err = unpopulate(val, "Type", &r.Type)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", r, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type ResourceGroupResource.
func (r ResourceGroupResource) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "id", r.ID)
	populate(objectMap, "location", r.Location)
	populate(objectMap, "name", r.Name)
	populate(objectMap, "properties", r.Properties)
	populate(objectMap, "systemData", r.SystemData)
	populate(objectMap, "tags", r.Tags)
	populate(objectMap, "type", r.Type)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type ResourceGroupResource.
func (r *ResourceGroupResource) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", r, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "id":
				err = unpopulate(val, "ID", &r.ID)
				delete(rawMsg, key)
		case "location":
				err = unpopulate(val, "Location", &r.Location)
				delete(rawMsg, key)
		case "name":
				err = unpopulate(val, "Name", &r.Name)
				delete(rawMsg, key)
		case "properties":
				err = unpopulate(val, "Properties", &r.Properties)
				delete(rawMsg, key)
		case "systemData":
				err = unpopulate(val, "SystemData", &r.SystemData)
				delete(rawMsg, key)
		case "tags":
				err = unpopulate(val, "Tags", &r.Tags)
				delete(rawMsg, key)
		case "type":
				err = unpopulate(val, "Type", &r.Type)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", r, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type ResourceGroupResourceListResult.
func (r ResourceGroupResourceListResult) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "nextLink", r.NextLink)
	populate(objectMap, "value", r.Value)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type ResourceGroupResourceListResult.
func (r *ResourceGroupResourceListResult) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", r, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "nextLink":
				err = unpopulate(val, "NextLink", &r.NextLink)
				delete(rawMsg, key)
		case "value":
				err = unpopulate(val, "Value", &r.Value)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", r, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type SystemData.
func (s SystemData) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populateTimeRFC3339(objectMap, "createdAt", s.CreatedAt)
	populate(objectMap, "createdBy", s.CreatedBy)
	populate(objectMap, "createdByType", s.CreatedByType)
	populateTimeRFC3339(objectMap, "lastModifiedAt", s.LastModifiedAt)
	populate(objectMap, "lastModifiedBy", s.LastModifiedBy)
	populate(objectMap, "lastModifiedByType", s.LastModifiedByType)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type SystemData.
func (s *SystemData) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", s, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "createdAt":
				err = unpopulateTimeRFC3339(val, "CreatedAt", &s.CreatedAt)
				delete(rawMsg, key)
		case "createdBy":
				err = unpopulate(val, "CreatedBy", &s.CreatedBy)
				delete(rawMsg, key)
		case "createdByType":
				err = unpopulate(val, "CreatedByType", &s.CreatedByType)
				delete(rawMsg, key)
		case "lastModifiedAt":
				err = unpopulateTimeRFC3339(val, "LastModifiedAt", &s.LastModifiedAt)
				delete(rawMsg, key)
		case "lastModifiedBy":
				err = unpopulate(val, "LastModifiedBy", &s.LastModifiedBy)
				delete(rawMsg, key)
		case "lastModifiedByType":
				err = unpopulate(val, "LastModifiedByType", &s.LastModifiedByType)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", s, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type TrackedResource.
func (t TrackedResource) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "id", t.ID)
	populate(objectMap, "location", t.Location)
	populate(objectMap, "name", t.Name)
	populate(objectMap, "systemData", t.SystemData)
	populate(objectMap, "tags", t.Tags)
	populate(objectMap, "type", t.Type)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type TrackedResource.
func (t *TrackedResource) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", t, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "id":
				err = unpopulate(val, "ID", &t.ID)
				delete(rawMsg, key)
		case "location":
				err = unpopulate(val, "Location", &t.Location)
				delete(rawMsg, key)
		case "name":
				err = unpopulate(val, "Name", &t.Name)
				delete(rawMsg, key)
		case "systemData":
				err = unpopulate(val, "SystemData", &t.SystemData)
				delete(rawMsg, key)
		case "tags":
				err = unpopulate(val, "Tags", &t.Tags)
				delete(rawMsg, key)
		case "type":
				err = unpopulate(val, "Type", &t.Type)
				delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", t, err)
		}
	}
	return nil
}

func populate(m map[string]interface{}, k string, v interface{}) {
	if v == nil {
		return
	} else if azcore.IsNullValue(v) {
		m[k] = nil
	} else if !reflect.ValueOf(v).IsNil() {
		m[k] = v
	}
}

func unpopulate(data json.RawMessage, fn string, v interface{}) error {
	if data == nil {
		return nil
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("struct field %s: %v", fn, err)
	}
	return nil
}

