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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	aztoken "github.com/radius-project/radius/pkg/azure/tokencredentials"
	"github.com/radius-project/radius/pkg/cli/clients_new/generated"
	"github.com/radius-project/radius/pkg/corerp/api/v20220315privatepreview"
	"github.com/radius-project/radius/pkg/sdk"
)

type Poller[T any] interface {
	Done() bool
	Poll(ctx context.Context) (*http.Response, error)
	Result(ctx context.Context) (T, error)
	ResumeToken() (string, error)
}

var _ Poller[v20220315privatepreview.ContainersClientCreateOrUpdateResponse] = (*runtime.Poller[v20220315privatepreview.ContainersClientCreateOrUpdateResponse])(nil)

type RadiusClient interface {
	Applications(scope string) ApplicationsClient
	Containers(scope string) ContainersClient
	Environments(scope string) EnvironmentsClient
	Resources(scope string, resourceType string) ResourcesClient
}

type ApplicationsClient interface {
	CreateOrUpdate(ctx context.Context, applicationName string, resource v20220315privatepreview.ApplicationResource, options *v20220315privatepreview.ApplicationsClientCreateOrUpdateOptions) (v20220315privatepreview.ApplicationsClientCreateOrUpdateResponse, error)
	Delete(ctx context.Context, applicationName string, options *v20220315privatepreview.ApplicationsClientDeleteOptions) (v20220315privatepreview.ApplicationsClientDeleteResponse, error)
	Get(ctx context.Context, applicationName string, options *v20220315privatepreview.ApplicationsClientGetOptions) (v20220315privatepreview.ApplicationsClientGetResponse, error)
}

type ContainersClient interface {
	BeginCreateOrUpdate(ctx context.Context, containerName string, resource v20220315privatepreview.ContainerResource, options *v20220315privatepreview.ContainersClientBeginCreateOrUpdateOptions) (Poller[v20220315privatepreview.ContainersClientCreateOrUpdateResponse], error)
	BeginDelete(ctx context.Context, containerName string, options *v20220315privatepreview.ContainersClientBeginDeleteOptions) (Poller[v20220315privatepreview.ContainersClientDeleteResponse], error)
	ContinueCreateOperation(ctx context.Context, resumeToken string) (Poller[v20220315privatepreview.ContainersClientCreateOrUpdateResponse], error)
	ContinueDeleteOperation(ctx context.Context, resumeToken string) (Poller[v20220315privatepreview.ContainersClientDeleteResponse], error)
	Get(ctx context.Context, containerName string, options *v20220315privatepreview.ContainersClientGetOptions) (v20220315privatepreview.ContainersClientGetResponse, error)
}

type EnvironmentsClient interface {
}

type ResourcesClient interface {
	BeginCreateOrUpdate(ctx context.Context, resourceName string, resource generated.GenericResource, options *generated.GenericResourcesClientBeginCreateOrUpdateOptions) (Poller[generated.GenericResourcesClientCreateOrUpdateResponse], error)
	BeginDelete(ctx context.Context, resourceName string, options *generated.GenericResourcesClientBeginDeleteOptions) (Poller[generated.GenericResourcesClientDeleteResponse], error)
	ContinueCreateOperation(ctx context.Context, resumeToken string) (Poller[generated.GenericResourcesClientCreateOrUpdateResponse], error)
	ContinueDeleteOperation(ctx context.Context, resumeToken string) (Poller[generated.GenericResourcesClientDeleteResponse], error)
	Get(ctx context.Context, resourceName string) (generated.GenericResourcesClientGetResponse, error)
	ListSecrets(ctx context.Context, resourceName string) (generated.GenericResourcesClientListSecretsResponse, error)
}

type Client struct {
	connection sdk.Connection
}

func NewClient(connection sdk.Connection) *Client {
	return &Client{connection: connection}
}

var _ RadiusClient = (*Client)(nil)

func (c *Client) Applications(scope string) ApplicationsClient {
	ac, err := v20220315privatepreview.NewApplicationsClient(scope, &aztoken.AnonymousCredential{}, sdk.NewClientOptions(c.connection))
	if err != nil {
		panic("failed to create client: " + err.Error())
	}

	return &ApplicationsClientImpl{inner: ac}
}

func (c *Client) Containers(scope string) ContainersClient {
	cc, err := v20220315privatepreview.NewContainersClient(scope, &aztoken.AnonymousCredential{}, sdk.NewClientOptions(c.connection))
	if err != nil {
		panic("failed to create client: " + err.Error())
	}

	return &ContainersClientImpl{inner: cc}
}

func (c *Client) Environments(scope string) EnvironmentsClient {
	ec, err := v20220315privatepreview.NewEnvironmentsClient(scope, &aztoken.AnonymousCredential{}, sdk.NewClientOptions(c.connection))
	if err != nil {
		panic("failed to create client: " + err.Error())
	}

	return &EnvironmentsClientImpl{inner: ec}
}

func (c *Client) Resources(scope string, resourceType string) ResourcesClient {
	gc, err := generated.NewGenericResourcesClient(scope, resourceType, &aztoken.AnonymousCredential{}, sdk.NewClientOptions(c.connection))
	if err != nil {
		panic("failed to create client: " + err.Error())
	}

	return &ResourcesClientImpl{inner: gc}
}

var _ ApplicationsClient = (*ApplicationsClientImpl)(nil)

type ApplicationsClientImpl struct {
	inner *v20220315privatepreview.ApplicationsClient
}

func (ac *ApplicationsClientImpl) CreateOrUpdate(ctx context.Context, applicationName string, resource v20220315privatepreview.ApplicationResource, options *v20220315privatepreview.ApplicationsClientCreateOrUpdateOptions) (v20220315privatepreview.ApplicationsClientCreateOrUpdateResponse, error) {
	return ac.inner.CreateOrUpdate(ctx, applicationName, resource, options)
}

func (ac *ApplicationsClientImpl) Delete(ctx context.Context, applicationName string, options *v20220315privatepreview.ApplicationsClientDeleteOptions) (v20220315privatepreview.ApplicationsClientDeleteResponse, error) {
	return ac.inner.Delete(ctx, applicationName, options)
}

func (ac *ApplicationsClientImpl) Get(ctx context.Context, applicationName string, options *v20220315privatepreview.ApplicationsClientGetOptions) (v20220315privatepreview.ApplicationsClientGetResponse, error) {
	return ac.inner.Get(ctx, applicationName, options)
}

var _ ContainersClient = (*ContainersClientImpl)(nil)

type ContainersClientImpl struct {
	inner *v20220315privatepreview.ContainersClient
}

func (cc *ContainersClientImpl) BeginCreateOrUpdate(ctx context.Context, containerName string, resource v20220315privatepreview.ContainerResource, options *v20220315privatepreview.ContainersClientBeginCreateOrUpdateOptions) (Poller[v20220315privatepreview.ContainersClientCreateOrUpdateResponse], error) {
	return cc.inner.BeginCreateOrUpdate(ctx, containerName, resource, options)
}

func (cc *ContainersClientImpl) BeginDelete(ctx context.Context, containerName string, options *v20220315privatepreview.ContainersClientBeginDeleteOptions) (Poller[v20220315privatepreview.ContainersClientDeleteResponse], error) {
	return cc.inner.BeginDelete(ctx, containerName, options)
}

func (cc *ContainersClientImpl) ContinueCreateOperation(ctx context.Context, resumeToken string) (Poller[v20220315privatepreview.ContainersClientCreateOrUpdateResponse], error) {
	return cc.inner.BeginCreateOrUpdate(ctx, "", v20220315privatepreview.ContainerResource{}, &v20220315privatepreview.ContainersClientBeginCreateOrUpdateOptions{ResumeToken: resumeToken})
}

func (cc *ContainersClientImpl) ContinueDeleteOperation(ctx context.Context, resumeToken string) (Poller[v20220315privatepreview.ContainersClientDeleteResponse], error) {
	return cc.inner.BeginDelete(ctx, "", &v20220315privatepreview.ContainersClientBeginDeleteOptions{ResumeToken: resumeToken})
}

func (cc *ContainersClientImpl) Get(ctx context.Context, containerName string, options *v20220315privatepreview.ContainersClientGetOptions) (v20220315privatepreview.ContainersClientGetResponse, error) {
	return cc.inner.Get(ctx, containerName, options)
}

var _ EnvironmentsClient = (*EnvironmentsClientImpl)(nil)

type EnvironmentsClientImpl struct {
	inner *v20220315privatepreview.EnvironmentsClient
}

var _ ResourcesClient = (*ResourcesClientImpl)(nil)

type ResourcesClientImpl struct {
	inner *generated.GenericResourcesClient
}

func (rc *ResourcesClientImpl) BeginCreateOrUpdate(ctx context.Context, resourceName string, resource generated.GenericResource, options *generated.GenericResourcesClientBeginCreateOrUpdateOptions) (Poller[generated.GenericResourcesClientCreateOrUpdateResponse], error) {
	return rc.inner.BeginCreateOrUpdate(ctx, resourceName, resource, options)
}

func (rc *ResourcesClientImpl) BeginDelete(ctx context.Context, resourceName string, options *generated.GenericResourcesClientBeginDeleteOptions) (Poller[generated.GenericResourcesClientDeleteResponse], error) {
	return rc.inner.BeginDelete(ctx, resourceName, options)
}

func (rc *ResourcesClientImpl) ContinueCreateOperation(ctx context.Context, resumeToken string) (Poller[generated.GenericResourcesClientCreateOrUpdateResponse], error) {
	return rc.inner.BeginCreateOrUpdate(ctx, "", generated.GenericResource{}, &generated.GenericResourcesClientBeginCreateOrUpdateOptions{ResumeToken: resumeToken})
}

func (rc *ResourcesClientImpl) ContinueDeleteOperation(ctx context.Context, resumeToken string) (Poller[generated.GenericResourcesClientDeleteResponse], error) {
	return rc.inner.BeginDelete(ctx, "", &generated.GenericResourcesClientBeginDeleteOptions{ResumeToken: resumeToken})
}

func (rc *ResourcesClientImpl) Get(ctx context.Context, resourceName string) (generated.GenericResourcesClientGetResponse, error) {
	return rc.inner.Get(ctx, resourceName, nil)
}

func (rc *ResourcesClientImpl) ListSecrets(ctx context.Context, resourceName string) (generated.GenericResourcesClientListSecretsResponse, error) {
	return rc.inner.ListSecrets(ctx, resourceName, nil)
}
