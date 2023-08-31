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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/radius-project/radius/pkg/cli/clients_new/generated"
	radappv1alpha3 "github.com/radius-project/radius/pkg/controller/api/rad_app/v1alpha3"
	"github.com/radius-project/radius/pkg/corerp/api/v20220315privatepreview"
	"github.com/radius-project/radius/pkg/to"
	"github.com/radius-project/radius/pkg/ucp/ucplog"
	"github.com/radius-project/radius/test/testcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	runtimelog "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	waitDuration = 10 * time.Second
	waitInterval = 250 * time.Millisecond
)

func Test_DeploymentReconciler(t *testing.T) {
	if os.Getenv("KUBEBUILDER_ASSETS") == "" {
		t.Skip("Skipping test because KUBEBUILDER_ASSETS. Running `make test` will run tests with this environment variable populated.")
		return
	}

	ctx := testcontext.New(t)
	runtimelog.SetLogger(ucplog.FromContextOrDiscard(ctx))

	env := &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "deploy", "Chart", "crds", "radius")},
		ErrorIfCRDPathMissing: true,
	}

	t.Log("Starting test environment")
	cfg, err := env.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		t.Log("Stopping test environment")
		err = env.Stop()
		require.NoError(t, err)
	})

	s := runtime.NewScheme()
	err = clientgoscheme.AddToScheme(s)
	require.NoError(t, err)
	err = radappv1alpha3.AddToScheme(s)
	require.NoError(t, err)

	runTest := func(exec func(t *testing.T, radius *mockRadiusClient, client client.Client)) {
		ctx, cancel := testcontext.NewWithCancel(t)
		defer cancel()
		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: s,
		})
		require.NoError(t, err)

		radius := NewMockRadiusClient()
		err = (&DeploymentReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
			Radius: radius,
			Delay:  time.Millisecond * 100,
		}).SetupWithManager(mgr)
		require.NoError(t, err)

		go func() {
			err := mgr.Start(ctx)
			require.NoError(t, err)
		}()

		exec(t, radius, mgr.GetClient())
	}

	t.Run("Enabled", func(t *testing.T) {
		runTest(DeploymentReconciler_Enabled)
	})

	t.Run("Connections", func(t *testing.T) {
		runTest(DeploymentReconciler_Connections)
	})
}

func waitForStateWaiting(t *testing.T, client client.Client, name types.NamespacedName) *deploymentAnnotations {
	ctx := testcontext.New(t)

	logger := t
	var annotations *deploymentAnnotations
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		logger.Logf("Fetching Deployment: %+v", name)
		current := &appsv1.Deployment{}
		err := client.Get(ctx, name, current)
		require.NoError(t, err)

		annotations, err = readAnnotations(current)
		require.NoError(t, err)
		assert.NotNil(t, annotations)
		logger.Logf("Annotations.Status: %+v", annotations.Status)

		if assert.NotNil(t, annotations.Status) && assert.Equal(t, DeploymentStateWaiting, annotations.Status.State) {
			assert.Equal(t, "/planes/radius/local/resourceGroups/default", annotations.Status.Scope)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/environments/default", annotations.Status.Environment)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/applications/"+name.Namespace, annotations.Status.Application)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/"+name.Name, annotations.Status.Container)
			assert.Empty(t, annotations.Status.Operation)
		}
	}, waitDuration, waitInterval)

	return annotations
}

func waitForStateUpdating(t *testing.T, client client.Client, name types.NamespacedName) *deploymentAnnotations {
	ctx := testcontext.New(t)

	logger := t
	var annotations *deploymentAnnotations
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		logger.Logf("Fetching Deployment: %+v", name)
		current := &appsv1.Deployment{}
		err := client.Get(ctx, name, current)
		require.NoError(t, err)

		annotations, err = readAnnotations(current)
		require.NoError(t, err)
		assert.NotNil(t, annotations)
		logger.Logf("Annotations.Status: %+v", annotations.Status)

		if assert.NotNil(t, annotations.Status) && assert.Equal(t, DeploymentStateUpdating, annotations.Status.State) {

			assert.Equal(t, "/planes/radius/local/resourceGroups/default", annotations.Status.Scope)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/environments/default", annotations.Status.Environment)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/applications/default", annotations.Status.Application)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/"+name.Name, annotations.Status.Container)
			assert.NotEmpty(t, annotations.Status.Operation)
		}
	}, waitDuration, waitInterval)

	return annotations
}

func waitForStateReady(t *testing.T, client client.Client, name types.NamespacedName) *deploymentAnnotations {
	ctx := testcontext.New(t)

	logger := t
	var annotations *deploymentAnnotations
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		logger.Logf("Fetching Deployment: %+v", name)
		current := &appsv1.Deployment{}
		err := client.Get(ctx, name, current)
		require.NoError(t, err)

		annotations, err = readAnnotations(current)
		require.NoError(t, err)
		assert.NotNil(t, annotations)
		logger.Logf("Annotations.Status: %+v", annotations.Status)

		if assert.NotNil(t, annotations.Status) && assert.Equal(t, DeploymentStateReady, annotations.Status.State) {
			assert.Equal(t, "/planes/radius/local/resourceGroups/default", annotations.Status.Scope)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/environments/default", annotations.Status.Environment)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/applications/"+name.Namespace, annotations.Status.Application)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/"+name.Name, annotations.Status.Container)
			assert.Empty(t, annotations.Status.Operation)
		}
	}, waitDuration, waitInterval)

	return annotations
}

func waitForStateDeleting(t *testing.T, client client.Client, name types.NamespacedName) *deploymentAnnotations {
	ctx := testcontext.New(t)

	logger := t
	var annotations *deploymentAnnotations
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		logger.Logf("Fetching Deployment: %+v", name)
		current := &appsv1.Deployment{}
		err := client.Get(ctx, name, current)
		require.NoError(t, err)

		annotations, err = readAnnotations(current)
		require.NoError(t, err)
		assert.NotNil(t, annotations)
		logger.Logf("Annotations.Status: %+v", annotations.Status)

		if assert.NotNil(t, annotations.Status) && assert.Equal(t, DeploymentStateDeleting, annotations.Status.State) {
			assert.Equal(t, "/planes/radius/local/resourceGroups/default", annotations.Status.Scope)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/environments/default", annotations.Status.Environment)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/applications/"+name.Namespace, annotations.Status.Application)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/"+name.Name, annotations.Status.Container)
			assert.NotEmpty(t, annotations.Status.Operation)
		}
	}, waitDuration, waitInterval)

	return annotations
}

func waitForDeploymentDeleted(t *testing.T, client client.Client, name types.NamespacedName) {
	ctx := testcontext.New(t)

	logger := t
	require.Eventually(t, func() bool {
		logger.Logf("Fetching Deployment: %+v", name)
		err := client.Get(ctx, name, &appsv1.Deployment{})
		return apierrors.IsNotFound(err)
	}, waitDuration, waitInterval)
}

func DeploymentReconciler_Enabled(t *testing.T, radius *mockRadiusClient, client client.Client) {
	ctx := testcontext.New(t)

	name := types.NamespacedName{Namespace: "default", Name: "test-deployment-enabled"}
	deployment := NewEmptyDeployment(name)
	deployment.Annotations[AnnotationRadiusEnabled] = "true"
	err := client.Create(ctx, deployment)
	require.NoError(t, err)

	// Deployment will be waiting for container to complete deployment.
	annotations := waitForStateUpdating(t, client, name)

	radius.CompleteOperation(annotations.Status.Operation)

	// Deployment will update after operation completes
	annotations = waitForStateReady(t, client, name)

	container, err := radius.Containers(annotations.Status.Scope).Get(ctx, deployment.Name, nil)
	require.NoError(t, err)
	require.Equal(t, "manual", string(*container.Properties.ResourceProvisioning))
	require.Equal(t, []*v20220315privatepreview.ResourceReference{{ID: to.Ptr("/planes/kubernetes/local/namespaces/default/providers/apps/Deployment/" + deployment.Name)}}, container.Properties.Resources)

	err = client.Delete(ctx, deployment)
	require.NoError(t, err)

	// Deletion of the container is in progress.
	annotations = waitForStateDeleting(t, client, name)
	radius.CompleteOperation(annotations.Status.Operation)

	// Now deleting of the deployment object can complete.
	waitForDeploymentDeleted(t, client, name)
}

func DeploymentReconciler_Connections(t *testing.T, radius *mockRadiusClient, client client.Client) {
	ctx := testcontext.New(t)

	name := types.NamespacedName{Namespace: "default", Name: "test-deployment-connections"}
	deployment := NewEmptyDeployment(name)
	deployment.Annotations[AnnotationRadiusEnabled] = "true"
	deployment.Annotations[AnnotationRadiusConnectionPrefix+"a"] = "recipe-a"
	deployment.Annotations[AnnotationRadiusConnectionPrefix+"b"] = "recipe-b"

	err := client.Create(ctx, deployment)
	require.NoError(t, err)

	// Deployment will be waiting for recipe resources to be created
	_ = waitForStateWaiting(t, client, name)

	// Create the recipes, but don't mark them as provisioned yet.
	recipeA := NewRecipe(types.NamespacedName{Namespace: "default", Name: "recipe-a"}, "Applications.Core/extenders")
	recipeB := NewRecipe(types.NamespacedName{Namespace: "default", Name: "recipe-b"}, "Applications.Core/extenders")

	err = client.Create(ctx, recipeA)
	require.NoError(t, err)
	err = client.Create(ctx, recipeB)
	require.NoError(t, err)

	// Deployment will be waiting for recipe resources to be created.
	annotations := waitForStateWaiting(t, client, name)

	// Create the radius resources associated with the recipes
	extenderA := generated.GenericResource{
		Properties: map[string]any{
			"a-value": "a",
			"secrets": map[string]string{
				"a-secret": "a",
			},
		},
	}
	poller, err := radius.Resources(annotations.Status.Scope, "Applications.Core/extenders").BeginCreateOrUpdate(ctx, recipeA.Name, extenderA, nil)
	require.NoError(t, err)
	token, err := poller.ResumeToken()
	require.NoError(t, err)
	radius.CompleteOperation(token)

	extenderB := generated.GenericResource{
		Properties: map[string]any{
			"b-value": "b",
			"secrets": map[string]string{
				"b-secret": "b",
			},
		},
	}
	poller, err = radius.Resources(annotations.Status.Scope, "Applications.Core/extenders").BeginCreateOrUpdate(ctx, recipeB.Name, extenderB, nil)
	require.NoError(t, err)
	token, err = poller.ResumeToken()
	require.NoError(t, err)
	radius.CompleteOperation(token)

	recipeA.Status = radappv1alpha3.RecipeStatus{
		Resource: annotations.Status.Scope + "/providers/Applications.Core/extenders/" + recipeA.Name,
	}
	recipeB.Status = radappv1alpha3.RecipeStatus{
		Resource: annotations.Status.Scope + "/providers/Applications.Core/extenders/" + recipeB.Name,
	}

	// Mark the recipes as provisioned.
	err = client.Status().Update(ctx, recipeA)
	require.NoError(t, err)
	err = client.Status().Update(ctx, recipeB)
	require.NoError(t, err)

	// Now we can create the container
	annotations = waitForStateUpdating(t, client, name)

	radius.CompleteOperation(annotations.Status.Operation)

	// Deployment will update after operation completes
	annotations = waitForStateReady(t, client, name)

	container, err := radius.Containers(annotations.Status.Scope).Get(ctx, deployment.Name, nil)
	require.NoError(t, err)
	require.Equal(t, "manual", string(*container.Properties.ResourceProvisioning))
	require.Equal(t, map[string]*v20220315privatepreview.ConnectionProperties{
		"a": {
			Source: to.Ptr(annotations.Status.Scope + "/providers/Applications.Core/extenders/" + recipeA.Name),
		},
		"b": {
			Source: to.Ptr(annotations.Status.Scope + "/providers/Applications.Core/extenders/" + recipeB.Name),
		},
	}, container.Properties.Connections)
	require.Equal(t, []*v20220315privatepreview.ResourceReference{{ID: to.Ptr("/planes/kubernetes/local/namespaces/default/providers/apps/Deployment/" + deployment.Name)}}, container.Properties.Resources)

	err = client.Get(ctx, name, deployment)
	require.NoError(t, err)

	expectedEnvVars := []corev1.EnvVar{
		{
			Name:  "CONNECTION_A_A-SECRET",
			Value: "a",
		},
		{
			Name:  "CONNECTION_A_A-VALUE",
			Value: "a",
		},
		{
			Name:  "CONNECTION_B_B-SECRET",
			Value: "b",
		},
		{
			Name:  "CONNECTION_B_B-VALUE",
			Value: "b",
		},
	}

	require.Equal(t, expectedEnvVars, deployment.Spec.Template.Spec.Containers[0].Env)

	err = client.Delete(ctx, deployment)
	require.NoError(t, err)

	// Deletion of the container is in progress.
	annotations = waitForStateDeleting(t, client, name)
	radius.CompleteOperation(annotations.Status.Operation)

	// Now deleting of the deployment object can complete.
	waitForDeploymentDeleted(t, client, name)
}

func NewEmptyDeployment(name types.NamespacedName) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name.Name,
			Namespace:   name.Namespace,
			Annotations: map[string]string{},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name.Name,
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}
}

func NewRecipe(name types.NamespacedName, resourceType string) *radappv1alpha3.Recipe {
	return &radappv1alpha3.Recipe{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name.Name,
			Namespace:   name.Namespace,
			Annotations: map[string]string{},
			Labels:      map[string]string{},
		},
		Spec: radappv1alpha3.RecipeSpec{
			Type: resourceType,
		},
	}
}
