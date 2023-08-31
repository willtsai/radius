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

	v1 "github.com/radius-project/radius/pkg/armrpc/api/v1"
	"github.com/radius-project/radius/pkg/cli/clients"
	"github.com/radius-project/radius/pkg/cli/clients_new/generated"
	"github.com/radius-project/radius/pkg/corerp/api/v20220315privatepreview"
	"github.com/radius-project/radius/pkg/to"
	"github.com/radius-project/radius/pkg/ucp/resources"
	"github.com/radius-project/radius/pkg/ucp/ucplog"
)

func EnsureApplication(ctx context.Context, radius RadiusClient, environment string, application string) error {
	id, err := resources.Parse(application)
	if err != nil {
		return err
	}

	logger := ucplog.FromContextOrDiscard(ctx).WithValues("scope", id.RootScope(), "application", application, "environment", environment)
	logger.Info("Fetching application.", "application", application)

	_, err = radius.Applications(id.RootScope()).Get(context.Background(), id.Name(), nil)
	if clients.Is404Error(err) {
		// Need to create application. Keep going.
	} else if err != nil {
		return err
	} else {
		// Application already created.
		logger.Info("Application application.", "application", application)
		return nil
	}

	app := v20220315privatepreview.ApplicationResource{
		Location: to.Ptr(v1.LocationGlobal),
		Name:     to.Ptr(id.Name()),
		Properties: &v20220315privatepreview.ApplicationProperties{
			Environment: to.Ptr(environment),
			Extensions: []v20220315privatepreview.ExtensionClassification{
				&v20220315privatepreview.KubernetesNamespaceExtension{
					Kind:      to.Ptr("kubernetesNamespace"),
					Namespace: to.Ptr(id.Name()),
				},
			},
		},
	}
	_, err = radius.Applications(id.RootScope()).CreateOrUpdate(ctx, id.Name(), app, nil)
	if err != nil {
		return err
	}

	return nil
}

func DeleteContainer(ctx context.Context, radius RadiusClient, container string) (Poller[v20220315privatepreview.ContainersClientDeleteResponse], error) {
	id, err := resources.Parse(container)
	if err != nil {
		return nil, err
	}

	logger := ucplog.FromContextOrDiscard(ctx).WithValues("scope", id.RootScope(), "resourceType", id.Type())
	logger.Info("Deleting resource.")

	poller, err := radius.Containers(id.RootScope()).BeginDelete(ctx, id.Name(), nil)
	if err != nil {
		return nil, err
	}

	return poller, nil
}

func UpdateContainer(ctx context.Context, radius RadiusClient, container string, properties *v20220315privatepreview.ContainerProperties) (Poller[v20220315privatepreview.ContainersClientCreateOrUpdateResponse], error) {
	id, err := resources.Parse(container)
	if err != nil {
		return nil, err
	}

	logger := ucplog.FromContextOrDiscard(ctx).WithValues("scope", id.RootScope(), "resourceType", id.Type())
	logger.Info("Updating resource.")

	body := v20220315privatepreview.ContainerResource{
		Location:   to.Ptr(v1.LocationGlobal),
		Name:       to.Ptr(id.Name()),
		Properties: properties,
	}
	poller, err := radius.Containers(id.RootScope()).BeginCreateOrUpdate(ctx, id.Name(), body, nil)
	if err != nil {
		return nil, err
	}

	return poller, nil
}

func DeleteResource(ctx context.Context, radius RadiusClient, resource string) (Poller[generated.GenericResourcesClientDeleteResponse], error) {
	id, err := resources.Parse(resource)
	if err != nil {
		return nil, err
	}

	logger := ucplog.FromContextOrDiscard(ctx).WithValues("scope", id.RootScope(), "resourceType", id.Type())
	logger.Info("Deleting resource.")

	poller, err := radius.Resources(id.RootScope(), id.Type()).BeginDelete(ctx, id.Name(), nil)
	if err != nil {
		return nil, err
	}

	return poller, nil
}

func UpdateResource(ctx context.Context, radius RadiusClient, resource string, properties map[string]any) (Poller[generated.GenericResourcesClientCreateOrUpdateResponse], error) {
	id, err := resources.Parse(resource)
	if err != nil {
		return nil, err
	}

	logger := ucplog.FromContextOrDiscard(ctx).WithValues("scope", id.RootScope(), "resourceType", id.Type())
	logger.Info("Updating resource.")

	body := generated.GenericResource{
		Location:   to.Ptr(v1.LocationGlobal),
		Name:       to.Ptr(id.Name()),
		Properties: properties,
	}
	poller, err := radius.Resources(id.RootScope(), id.Type()).BeginCreateOrUpdate(ctx, id.Name(), body, nil)
	if err != nil {
		return nil, err
	}

	return poller, nil
}
