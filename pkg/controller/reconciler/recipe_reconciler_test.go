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
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
	"time"

	radappv1alpha3 "github.com/radius-project/radius/pkg/controller/api/rad_app/v1alpha3"
	"github.com/radius-project/radius/pkg/ucp/ucplog"
	"github.com/radius-project/radius/test/testcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	runtimelog "sigs.k8s.io/controller-runtime/pkg/log"
)

func Test_RecipeReconciler(t *testing.T) {
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
		err = (&RecipeReconciler{
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

	t.Run("WithoutSecret", func(t *testing.T) {
		runTest(RecipeReconciler_WithoutSecret)
	})

	t.Run("WithSecret", func(t *testing.T) {
		runTest(RecipeReconciler_WithSecret)
	})
}

func waitForRecipeStateUpdating(t *testing.T, client client.Client, name types.NamespacedName) *radappv1alpha3.RecipeStatus {
	ctx := testcontext.New(t)

	logger := t
	status := &radappv1alpha3.RecipeStatus{}
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		logger.Logf("Fetching Recipe: %+v", name)
		current := &radappv1alpha3.Recipe{}
		err := client.Get(ctx, name, current)
		require.NoError(t, err)

		status = &current.Status
		logger.Logf("Recipe.Status: %+v", current.Status)

		if assert.Equal(t, radappv1alpha3.PhraseUpdating, current.Status.Phrase) {
			assert.Equal(t, "/planes/radius/local/resourceGroups/default", current.Status.Scope)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/environments/default", current.Status.Environment)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/applications/default", current.Status.Application)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/"+current.Spec.Type+"/"+name.Name, current.Status.Resource)
			assert.NotEmpty(t, current.Status.Operation)
		}
	}, waitDuration, waitInterval)

	return status
}

func waitForRecipeStateReady(t *testing.T, client client.Client, name types.NamespacedName) *radappv1alpha3.RecipeStatus {
	ctx := testcontext.New(t)

	logger := t
	status := &radappv1alpha3.RecipeStatus{}
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		logger.Logf("Fetching Recipe: %+v", name)
		current := &radappv1alpha3.Recipe{}
		err := client.Get(ctx, name, current)
		require.NoError(t, err)

		status = &current.Status
		logger.Logf("Recipe.Status: %+v", current.Status)

		if assert.Equal(t, radappv1alpha3.PhraseReady, current.Status.Phrase) {
			assert.Equal(t, "/planes/radius/local/resourceGroups/default", current.Status.Scope)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/environments/default", current.Status.Environment)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/applications/default", current.Status.Application)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/"+current.Spec.Type+"/"+name.Name, current.Status.Resource)
			assert.Empty(t, current.Status.Operation)
		}
	}, waitDuration, waitInterval)

	return status
}

func waitForRecipeStateDeleting(t *testing.T, client client.Client, name types.NamespacedName) *radappv1alpha3.RecipeStatus {
	ctx := testcontext.New(t)

	logger := t
	status := &radappv1alpha3.RecipeStatus{}
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		logger.Logf("Fetching Recipe: %+v", name)
		current := &radappv1alpha3.Recipe{}
		err := client.Get(ctx, name, current)
		require.NoError(t, err)

		status = &current.Status
		logger.Logf("Recipe.Status: %+v", current.Status)

		if assert.Equal(t, radappv1alpha3.PhraseDeleting, current.Status.Phrase) {
			assert.Equal(t, "/planes/radius/local/resourceGroups/default", current.Status.Scope)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/environments/default", current.Status.Environment)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/Applications.Core/applications/default", current.Status.Application)
			assert.Equal(t, "/planes/radius/local/resourceGroups/default/providers/"+current.Spec.Type+"/"+name.Name, current.Status.Resource)
			assert.NotEmpty(t, current.Status.Operation)
		}
	}, waitDuration, waitInterval)

	return status
}

func waitForRecipeDeleted(t *testing.T, client client.Client, name types.NamespacedName) {
	ctx := testcontext.New(t)

	logger := t
	require.Eventually(t, func() bool {
		logger.Logf("Fetching Recipe: %+v", name)
		err := client.Get(ctx, name, &radappv1alpha3.Recipe{})
		return apierrors.IsNotFound(err)
	}, waitDuration, waitInterval)
}

func RecipeReconciler_WithoutSecret(t *testing.T, radius *mockRadiusClient, client client.Client) {
	ctx := testcontext.New(t)

	name := types.NamespacedName{Namespace: "default", Name: "test-recipe-withoutsecret"}
	recipe := NewRecipe(name, "Applications.Core/extenders")
	err := client.Create(ctx, recipe)
	require.NoError(t, err)

	// Recipe will be waiting for extender to complete provisioning.
	status := waitForRecipeStateUpdating(t, client, name)

	radius.CompleteOperation(status.Operation)

	// Recipe will update after operation completes
	status = waitForRecipeStateReady(t, client, name)

	extender, err := radius.Resources(status.Scope, "Applications.Core/extenders").Get(ctx, name.Name)
	require.NoError(t, err)
	require.Equal(t, "recipe", extender.Properties["resourceProvisioning"])

	err = client.Delete(ctx, recipe)
	require.NoError(t, err)

	// Deletion of the recipe is in progress.
	status = waitForRecipeStateDeleting(t, client, name)
	radius.CompleteOperation(status.Operation)

	// Now deleting of the deployment object can complete.
	waitForRecipeDeleted(t, client, name)
}

func RecipeReconciler_WithSecret(t *testing.T, radius *mockRadiusClient, client client.Client) {
	ctx := testcontext.New(t)

	name := types.NamespacedName{Namespace: "default", Name: "test-recipe-withsecret"}
	recipe := NewRecipe(name, "Applications.Core/extenders")
	recipe.Spec.SecretName = name.Name

	err := client.Create(ctx, recipe)
	require.NoError(t, err)

	// Recipe will be waiting for extender to complete provisioning.
	status := waitForRecipeStateUpdating(t, client, name)

	radius.Update(func() {
		resource := radius.resources[status.Resource]
		resource.Properties["a-value"] = "a"
		resource.Properties["secrets"] = map[string]string{
			"b-secret": "b",
		}
	})

	radius.CompleteOperation(status.Operation)

	// Recipe will update after operation completes
	status = waitForRecipeStateReady(t, client, name)

	secret := corev1.Secret{}
	err = client.Get(ctx, name, &secret)
	require.NoError(t, err)

	expectedData := map[string][]byte{
		"a-value":  []byte(base64.RawStdEncoding.EncodeToString([]byte("a"))),
		"b-secret": []byte(base64.RawStdEncoding.EncodeToString([]byte("b"))),
	}

	require.Equal(t, expectedData, secret.Data)

	extender, err := radius.Resources(status.Scope, "Applications.Core/extenders").Get(ctx, name.Name)
	require.NoError(t, err)
	require.Equal(t, "recipe", extender.Properties["resourceProvisioning"])

	err = client.Delete(ctx, recipe)
	require.NoError(t, err)

	// Deletion of the recipe is in progress.
	status = waitForRecipeStateDeleting(t, client, name)
	radius.CompleteOperation(status.Operation)

	// Now deleting of the deployment object can complete.
	waitForRecipeDeleted(t, client, name)

	err = client.Get(ctx, name, &secret)
	require.Error(t, err)
	require.True(t, apierrors.IsNotFound(err))
}
