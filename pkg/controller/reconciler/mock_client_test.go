/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package reconciler

import (
	"context"
	"net/http"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/google/uuid"
	v1 "github.com/radius-project/radius/pkg/armrpc/api/v1"
	"github.com/radius-project/radius/pkg/cli/clients_new/generated"
	"github.com/radius-project/radius/pkg/corerp/api/v20220315privatepreview"
	"github.com/radius-project/radius/pkg/to"
)

// This file contains mocks for the RadiusClient interface.

func NewMockRadiusClient() *mockRadiusClient {
	return &mockRadiusClient{
		applications: map[string]v20220315privatepreview.ApplicationResource{},
		containers:   map[string]v20220315privatepreview.ContainerResource{},
		environments: map[string]v20220315privatepreview.EnvironmentResource{},
		resources:    map[string]generated.GenericResource{},
		operations:   map[string]*operationState{},

		lock: &sync.Mutex{},
	}
}

var _ RadiusClient = (*mockRadiusClient)(nil)

type mockRadiusClient struct {
	applications map[string]v20220315privatepreview.ApplicationResource
	containers   map[string]v20220315privatepreview.ContainerResource
	environments map[string]v20220315privatepreview.EnvironmentResource
	resources    map[string]generated.GenericResource
	operations   map[string]*operationState

	lock *sync.Mutex
}

type operationState struct {
	complete bool
	value    any
	deleteID string
}

func (rc *mockRadiusClient) Update(exec func()) {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	exec()
}

func (rc *mockRadiusClient) Applications(scope string) ApplicationsClient {
	return &mockApplicationsClient{mock: rc, scope: scope}
}

func (rc *mockRadiusClient) Containers(scope string) ContainersClient {
	return &mockContainersClient{mock: rc, scope: scope}
}

func (rc *mockRadiusClient) Environments(scope string) EnvironmentsClient {
	return &mockEnvironmentsClient{mock: rc, scope: scope}
}

func (rc *mockRadiusClient) Resources(scope string, resourceType string) ResourcesClient {
	return &mockResourcesClient{mock: rc, scope: scope, resourceType: resourceType}
}

func (rc *mockRadiusClient) CompleteOperation(operationID string) {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	state, ok := rc.operations[operationID]
	if !ok {
		panic("operation not found: " + operationID)
	}

	state.complete = true

	if state.deleteID != "" {
		delete(rc.environments, state.deleteID)
		delete(rc.applications, state.deleteID)
		delete(rc.containers, state.deleteID)
		delete(rc.resources, state.deleteID)
	}
}

var _ ApplicationsClient = (*mockApplicationsClient)(nil)

type mockApplicationsClient struct {
	mock  *mockRadiusClient
	scope string
}

func (ac *mockApplicationsClient) id(applicationName string) string {
	return ac.scope + "/providers/Applications.Core/applications/" + applicationName
}

func (ac *mockApplicationsClient) CreateOrUpdate(ctx context.Context, applicationName string, resource v20220315privatepreview.ApplicationResource, options *v20220315privatepreview.ApplicationsClientCreateOrUpdateOptions) (v20220315privatepreview.ApplicationsClientCreateOrUpdateResponse, error) {
	id := ac.id(applicationName)

	ac.mock.lock.Lock()
	defer ac.mock.lock.Unlock()

	ac.mock.applications[id] = resource
	return v20220315privatepreview.ApplicationsClientCreateOrUpdateResponse{ApplicationResource: resource}, nil
}

func (ac *mockApplicationsClient) Delete(ctx context.Context, applicationName string, options *v20220315privatepreview.ApplicationsClientDeleteOptions) (v20220315privatepreview.ApplicationsClientDeleteResponse, error) {
	id := ac.id(applicationName)

	ac.mock.lock.Lock()
	defer ac.mock.lock.Unlock()

	delete(ac.mock.applications, id)
	return v20220315privatepreview.ApplicationsClientDeleteResponse{}, nil
}

func (ac *mockApplicationsClient) Get(ctx context.Context, applicationName string, options *v20220315privatepreview.ApplicationsClientGetOptions) (v20220315privatepreview.ApplicationsClientGetResponse, error) {
	id := ac.id(applicationName)

	ac.mock.lock.Lock()
	defer ac.mock.lock.Unlock()

	application, ok := ac.mock.applications[id]
	if !ok {
		err := &azcore.ResponseError{ErrorCode: v1.CodeNotFound, StatusCode: http.StatusNotFound}
		return v20220315privatepreview.ApplicationsClientGetResponse{}, err
	}

	return v20220315privatepreview.ApplicationsClientGetResponse{ApplicationResource: application}, nil
}

var _ ContainersClient = (*mockContainersClient)(nil)

type mockContainersClient struct {
	mock  *mockRadiusClient
	scope string
}

func (cc *mockContainersClient) id(containerName string) string {
	return cc.scope + "/providers/Applications.Core/containers/" + containerName
}

func (cc *mockContainersClient) BeginCreateOrUpdate(ctx context.Context, containerName string, resource v20220315privatepreview.ContainerResource, options *v20220315privatepreview.ContainersClientBeginCreateOrUpdateOptions) (Poller[v20220315privatepreview.ContainersClientCreateOrUpdateResponse], error) {
	id := cc.id(containerName)

	cc.mock.lock.Lock()
	defer cc.mock.lock.Unlock()

	value := v20220315privatepreview.ContainersClientCreateOrUpdateResponse{ContainerResource: resource}
	state := &operationState{complete: false, value: value}

	operationID := uuid.New().String()
	cc.mock.containers[id] = resource
	cc.mock.operations[operationID] = state

	return &mockPoller[v20220315privatepreview.ContainersClientCreateOrUpdateResponse]{mock: cc.mock, operationID: operationID, state: state}, nil
}

func (cc *mockContainersClient) BeginDelete(ctx context.Context, containerName string, options *v20220315privatepreview.ContainersClientBeginDeleteOptions) (Poller[v20220315privatepreview.ContainersClientDeleteResponse], error) {
	id := cc.id(containerName)

	cc.mock.lock.Lock()
	defer cc.mock.lock.Unlock()

	value := v20220315privatepreview.ContainersClientDeleteResponse{}
	state := &operationState{complete: false, value: value, deleteID: id}

	operationID := uuid.New().String()
	cc.mock.operations[operationID] = state

	return &mockPoller[v20220315privatepreview.ContainersClientDeleteResponse]{mock: cc.mock, operationID: operationID, state: state}, nil
}

func (cc *mockContainersClient) ContinueCreateOperation(ctx context.Context, resumeToken string) (Poller[v20220315privatepreview.ContainersClientCreateOrUpdateResponse], error) {
	cc.mock.lock.Lock()
	defer cc.mock.lock.Unlock()

	state, ok := cc.mock.operations[resumeToken]
	if !ok {
		panic("operation not found: " + resumeToken)
	}

	return &mockPoller[v20220315privatepreview.ContainersClientCreateOrUpdateResponse]{mock: cc.mock, operationID: resumeToken, state: state}, nil
}

func (cc *mockContainersClient) ContinueDeleteOperation(ctx context.Context, resumeToken string) (Poller[v20220315privatepreview.ContainersClientDeleteResponse], error) {
	cc.mock.lock.Lock()
	defer cc.mock.lock.Unlock()

	state, ok := cc.mock.operations[resumeToken]
	if !ok {
		panic("operation not found: " + resumeToken)
	}

	return &mockPoller[v20220315privatepreview.ContainersClientDeleteResponse]{mock: cc.mock, operationID: resumeToken, state: state}, nil
}

func (cc *mockContainersClient) Get(ctx context.Context, containerName string, options *v20220315privatepreview.ContainersClientGetOptions) (v20220315privatepreview.ContainersClientGetResponse, error) {
	id := cc.id(containerName)

	cc.mock.lock.Lock()
	defer cc.mock.lock.Unlock()

	container, ok := cc.mock.containers[id]
	if !ok {
		err := &azcore.ResponseError{ErrorCode: v1.CodeNotFound, StatusCode: http.StatusNotFound}
		return v20220315privatepreview.ContainersClientGetResponse{}, err
	}

	return v20220315privatepreview.ContainersClientGetResponse{ContainerResource: container}, nil
}

var _ EnvironmentsClient = (*mockEnvironmentsClient)(nil)

type mockEnvironmentsClient struct {
	mock  *mockRadiusClient
	scope string
}

var _ ResourcesClient = (*mockResourcesClient)(nil)

type mockResourcesClient struct {
	mock         *mockRadiusClient
	scope        string
	resourceType string
}

func (rc *mockResourcesClient) id(resourceName string) string {
	return rc.scope + "/providers/" + rc.resourceType + "/" + resourceName
}

func (rc *mockResourcesClient) BeginCreateOrUpdate(ctx context.Context, resourceName string, resource generated.GenericResource, options *generated.GenericResourcesClientBeginCreateOrUpdateOptions) (Poller[generated.GenericResourcesClientCreateOrUpdateResponse], error) {
	id := rc.id(resourceName)

	rc.mock.lock.Lock()
	defer rc.mock.lock.Unlock()

	value := generated.GenericResourcesClientCreateOrUpdateResponse{GenericResource: resource}
	state := &operationState{complete: false, value: value}

	operationID := uuid.New().String()
	rc.mock.resources[id] = resource
	rc.mock.operations[operationID] = state

	return &mockPoller[generated.GenericResourcesClientCreateOrUpdateResponse]{mock: rc.mock, operationID: operationID, state: state}, nil
}

func (rc *mockResourcesClient) BeginDelete(ctx context.Context, resourceName string, options *generated.GenericResourcesClientBeginDeleteOptions) (Poller[generated.GenericResourcesClientDeleteResponse], error) {
	id := rc.id(resourceName)

	rc.mock.lock.Lock()
	defer rc.mock.lock.Unlock()

	value := generated.GenericResourcesClientDeleteResponse{}
	state := &operationState{complete: false, value: value, deleteID: id}

	operationID := uuid.New().String()
	rc.mock.operations[operationID] = state

	return &mockPoller[generated.GenericResourcesClientDeleteResponse]{mock: rc.mock, operationID: operationID, state: state}, nil
}

func (rc *mockResourcesClient) ContinueCreateOperation(ctx context.Context, resumeToken string) (Poller[generated.GenericResourcesClientCreateOrUpdateResponse], error) {
	rc.mock.lock.Lock()
	defer rc.mock.lock.Unlock()

	state, ok := rc.mock.operations[resumeToken]
	if !ok {
		panic("operation not found: " + resumeToken)
	}

	return &mockPoller[generated.GenericResourcesClientCreateOrUpdateResponse]{mock: rc.mock, operationID: resumeToken, state: state}, nil
}

func (rc *mockResourcesClient) ContinueDeleteOperation(ctx context.Context, resumeToken string) (Poller[generated.GenericResourcesClientDeleteResponse], error) {
	rc.mock.lock.Lock()
	defer rc.mock.lock.Unlock()

	state, ok := rc.mock.operations[resumeToken]
	if !ok {
		panic("operation not found: " + resumeToken)
	}

	return &mockPoller[generated.GenericResourcesClientDeleteResponse]{mock: rc.mock, operationID: resumeToken, state: state}, nil
}

func (rc *mockResourcesClient) Get(ctx context.Context, resourceName string) (generated.GenericResourcesClientGetResponse, error) {
	id := rc.id(resourceName)

	rc.mock.lock.Lock()
	defer rc.mock.lock.Unlock()

	resource, ok := rc.mock.resources[id]
	if !ok {
		err := &azcore.ResponseError{ErrorCode: v1.CodeNotFound, StatusCode: http.StatusNotFound}
		return generated.GenericResourcesClientGetResponse{}, err
	}

	return generated.GenericResourcesClientGetResponse{GenericResource: resource}, nil
}

func (rc *mockResourcesClient) ListSecrets(ctx context.Context, resourceName string) (generated.GenericResourcesClientListSecretsResponse, error) {
	id := rc.id(resourceName)

	rc.mock.lock.Lock()
	defer rc.mock.lock.Unlock()

	resource, ok := rc.mock.resources[id]
	if !ok {
		err := &azcore.ResponseError{ErrorCode: v1.CodeNotFound, StatusCode: http.StatusNotFound}
		return generated.GenericResourcesClientListSecretsResponse{}, err
	}

	obj, ok := resource.Properties["secrets"]
	if !ok {
		err := &azcore.ResponseError{ErrorCode: v1.CodeNotFound, StatusCode: http.StatusNotFound}
		return generated.GenericResourcesClientListSecretsResponse{}, err
	}

	data := obj.(map[string]string)
	secrets := map[string]*string{}
	for k, v := range data {
		secrets[k] = to.Ptr(v)
	}

	return generated.GenericResourcesClientListSecretsResponse{Value: secrets}, nil
}

var _ Poller[v20220315privatepreview.ContainersClientCreateOrUpdateResponse] = (*mockPoller[v20220315privatepreview.ContainersClientCreateOrUpdateResponse])(nil)

type mockPoller[T any] struct {
	operationID string
	mock        *mockRadiusClient
	state       *operationState
}

func (mp *mockPoller[T]) Done() bool {
	mp.mock.lock.Lock()
	defer mp.mock.lock.Unlock()

	return mp.state.complete // Status updates are delivered via the Poll function.
}

func (mp *mockPoller[T]) Poll(ctx context.Context) (*http.Response, error) {
	mp.mock.lock.Lock()
	defer mp.mock.lock.Unlock()

	// NOTE: this is ok because our code ignores the actual result.
	mp.state = mp.mock.operations[mp.operationID]
	return nil, nil
}

func (mp *mockPoller[T]) Result(ctx context.Context) (T, error) {
	mp.mock.lock.Lock()
	defer mp.mock.lock.Unlock()

	if mp.state.complete {
		return mp.state.value.(T), nil
	}

	panic("operation not done")
}

func (mp *mockPoller[T]) ResumeToken() (string, error) {
	return mp.operationID, nil
}
