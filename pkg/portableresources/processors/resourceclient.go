/*
Copyright 2023 The Radius Authors.

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

package processors

import (
	context "context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/radius-project/radius/pkg/azure/armauth"
	"github.com/radius-project/radius/pkg/azure/clientv2"
	aztoken "github.com/radius-project/radius/pkg/azure/tokencredentials"
	"github.com/radius-project/radius/pkg/cli/clients"
	"github.com/radius-project/radius/pkg/cli/clients_new/generated"
	"github.com/radius-project/radius/pkg/components/kubernetesclient/kubernetesclientprovider"
	"github.com/radius-project/radius/pkg/components/trace"
	"github.com/radius-project/radius/pkg/sdk"
	"github.com/radius-project/radius/pkg/ucp/resources"
	resources_azure "github.com/radius-project/radius/pkg/ucp/resources/azure"
	resources_kubernetes "github.com/radius-project/radius/pkg/ucp/resources/kubernetes"
	"github.com/radius-project/radius/pkg/ucp/ucplog"
	"go.opentelemetry.io/otel/attribute"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtime_client "sigs.k8s.io/controller-runtime/pkg/client"
)

type resourceClient struct {
	arm *armauth.ArmConfig

	// armClientOptions is used to create ARM clients. Provide a Transport to override for testing.
	armClientOptions *arm.ClientOptions

	// connection is the connection to use for UCP resources. Override this for testing.
	connection sdk.Connection

	// kubernetesClient is the Kubernetes client provider used to create Kubernetes clients. Override this for testing.
	kubernetesClient *kubernetesclientprovider.KubernetesClientProvider
}

// NewResourceClient creates a new resourceClient instance with the given parameters.
func NewResourceClient(arm *armauth.ArmConfig, connection sdk.Connection, kubernetesClient *kubernetesclientprovider.KubernetesClientProvider) *resourceClient {
	return &resourceClient{arm: arm, connection: connection, kubernetesClient: kubernetesClient}
}

// Delete attempts to delete a resource, either through UCP, Azure, or Kubernetes, depending on the resource type.
func (c *resourceClient) Delete(ctx context.Context, id string) error {
	parsed, err := resources.ParseResource(id)
	if err != nil {
		return err
	}

	// Performing deletion is going to fire of potentially many requests due to discovery and polling. Creating
	// a span will help categorize the requests in traces.
	attributes := []attribute.KeyValue{{Key: attribute.Key(ucplog.LogFieldTargetResourceID), Value: attribute.StringValue(id)}}
	ctx, span := trace.StartCustomSpan(ctx, "resourceclient.Delete", trace.BackendTracerName, attributes)
	defer span.End()

	// Ideally we'd do all of our resource deletion through UCP. Unfortunately we have not yet integrated
	// Azure and Kubernetes resources yet, so those are handled as special cases here.
	ns := strings.ToLower(parsed.PlaneNamespace())

	if !parsed.IsUCPQualified() || strings.HasPrefix(ns, "azure/") {
		return c.wrapError(parsed, c.deleteAzureResource(ctx, parsed))
	} else if strings.HasPrefix(ns, "kubernetes/") {
		return c.wrapError(parsed, c.deleteKubernetesResource(ctx, parsed))
	} else {
		return c.wrapError(parsed, c.deleteUCPResource(ctx, parsed))
	}
}

func (c *resourceClient) wrapError(id resources.ID, err error) error {
	if err != nil {
		return &ResourceError{Inner: err, ID: id.String()}
	}

	return nil
}

func (c *resourceClient) deleteAzureResource(ctx context.Context, id resources.ID) error {
	var err error
	if id.IsUCPQualified() {
		id, err = resources.ParseResource(resources.MakeRelativeID(id.ScopeSegments()[1:], id.TypeSegments(), id.ExtensionSegments()))
		if err != nil {
			return err
		}
	}

	apiVersion, err := c.lookupARMAPIVersion(ctx, id)
	if err != nil {
		return err
	}

	client, err := clientv2.NewGenericResourceClient(id.FindScope(resources_azure.ScopeSubscriptions), &c.arm.ClientOptions, c.armClientOptions)
	if err != nil {
		return err
	}

	poller, err := client.BeginDeleteByID(ctx, id.String(), apiVersion, &armresources.ClientBeginDeleteByIDOptions{})
	if err != nil {
		if clients.Is404Error(err) {
			// If the resource that we want to delete doesn't exist, we don't need to delete it.
			return nil
		}

		return err
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		if clients.Is404Error(err) {
			// If the resource that we want to delete doesn't exist, we don't need to delete it.
			return nil
		}

		return err
	}

	return nil
}

func (c *resourceClient) lookupARMAPIVersion(ctx context.Context, id resources.ID) (string, error) {
	client, err := clientv2.NewProvidersClient(id.FindScope(resources_azure.ScopeSubscriptions), &c.arm.ClientOptions, c.armClientOptions)
	if err != nil {
		return "", err
	}

	resp, err := client.Get(ctx, id.ProviderNamespace(), nil)
	if err != nil {
		return "", err
	}

	// We need to match on the resource type name without the provider namespace.
	shortType := strings.TrimPrefix(id.TypeSegments()[0].Type, id.ProviderNamespace()+"/")
	for _, rt := range resp.ResourceTypes {
		if !strings.EqualFold(shortType, *rt.ResourceType) {
			continue
		}
		if rt.DefaultAPIVersion != nil {
			return *rt.DefaultAPIVersion, nil
		}

		if len(rt.APIVersions) > 0 {
			return *rt.APIVersions[0], nil
		}

		return "", fmt.Errorf("could not find API version for type %q, no supported API versions", id.Type())

	}

	return "", fmt.Errorf("could not find API version for type %q, type was not found", id.Type())
}

func (c *resourceClient) deleteUCPResource(ctx context.Context, id resources.ID) error {
	// NOTE: the API version passed in here is ignored.
	//
	// We're using a generated client that understands Radius' currently supported API version.
	//
	// For AWS resources, the server does not yet validate the API version.
	//
	// In the future we should change this to look up API versions dynamically like we do for ARM.
	client, err := generated.NewGenericResourcesClient(id.RootScope(), id.Type(), &aztoken.AnonymousCredential{}, sdk.NewClientOptions(c.connection))
	if err != nil {
		return err
	}

	poller, err := client.BeginDelete(ctx, id.Name(), nil)
	if err != nil {
		if clients.Is404Error(err) {
			// If the resource that we want to delete doesn't exist, we don't need to delete it.
			return nil
		}

		return err
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		if clients.Is404Error(err) {
			// If the resource that we want to delete doesn't exist, we don't need to delete it.
			return nil
		}

		return err
	}

	return nil
}

func (c *resourceClient) deleteKubernetesResource(ctx context.Context, id resources.ID) error {
	apiVersion, err := c.lookupKubernetesAPIVersion(id)
	if err != nil {
		return err
	}

	group, kind, namespace, name := resources_kubernetes.ToParts(id)

	metadata := map[string]any{
		"name": name,
	}
	if namespace != "" {
		metadata["namespace"] = namespace
	}

	if group != "" {
		apiVersion = fmt.Sprintf("%s/%s", group, apiVersion)
	}

	obj := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata":   metadata,
		},
	}

	runtimeClient, err := c.kubernetesClient.RuntimeClient()
	if err != nil {
		return err
	}

	err = runtime_client.IgnoreNotFound(runtimeClient.Delete(ctx, &obj))
	if err != nil {
		return err
	}

	return nil
}

func (c *resourceClient) lookupKubernetesAPIVersion(id resources.ID) (string, error) {
	group, kind, namespace, _ := resources_kubernetes.ToParts(id)
	var resourceLists []*v1.APIResourceList

	discoveryClient, err := c.kubernetesClient.DiscoveryClient()
	if err != nil {
		return "", err
	}

	if namespace == "" {
		resourceLists, err = discoveryClient.ServerPreferredResources()
		if err != nil {
			return "", fmt.Errorf("could not find API version for type %q: %w", id.Type(), err)
		}
	} else {
		resourceLists, err = discoveryClient.ServerPreferredNamespacedResources()
		if err != nil {
			return "", fmt.Errorf("could not find API version for type %q: %w", id.Type(), err)
		}
	}

	for _, resourceList := range resourceLists {
		// We know the group but not the version. This will give us the the list of resources and their preferred versions.
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			return "", err
		}

		if group != gv.Group {
			continue
		}

		for _, resource := range resourceList.APIResources {
			if resource.Kind == kind {
				return gv.Version, nil
			}
		}
	}

	return "", fmt.Errorf("could not find API version for type %q, type was not found", id.Type())
}
