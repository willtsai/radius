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
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"github.com/radius-project/radius/pkg/cli/clients"
	"github.com/radius-project/radius/pkg/cli/clients_new/generated"
	radappv1alpha3 "github.com/radius-project/radius/pkg/controller/api/rad_app/v1alpha3"
	"github.com/radius-project/radius/pkg/ucp/ucplog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RecipeReconciler reconciles a Recipe object.
type RecipeReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Radius RadiusClient
	Delay  time.Duration
}

// Reconcile is the main reconciliation loop for the Recipe resource.
func (r *RecipeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ucplog.FromContextOrDiscard(ctx).WithValues("kind", "Recipe", "name", req.Name, "namespace", req.Namespace)
	ctx = logr.NewContext(ctx, logger)

	recipe := radappv1alpha3.Recipe{}
	err := r.Client.Get(ctx, req.NamespacedName, &recipe)
	if apierrors.IsNotFound(err) {
		logger.Info("Recipe is being deleted.")
		return ctrl.Result{}, nil
	} else if err != nil {
		logger.Error(err, "Unable to fetch resource.")
		return ctrl.Result{}, err
	}

	// Check for deletion before checking whether we're up-to-date.
	if recipe.DeletionTimestamp != nil {
		return r.ReconcileDelete(ctx, &recipe)
	}

	// Right now we don't support updates to the spec after creation.
	if recipe.Status.Phrase == radappv1alpha3.PhraseReady {
		recipe.Status.ObservedGeneration = recipe.Generation
		err = r.Client.Status().Update(ctx, &recipe)
		if err != nil {
			logger.Error(err, "Unable to update resource.")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	return r.ReconcileUpdate(ctx, &recipe)
}

func (r *RecipeReconciler) HasCreatedResource(recipe *radappv1alpha3.Recipe) bool {
	return recipe.Status.Resource != ""
}

func (r *RecipeReconciler) ReconcileUpdate(ctx context.Context, recipe *radappv1alpha3.Recipe) (ctrl.Result, error) {
	logger := ucplog.FromContextOrDiscard(ctx)

	// Ensure that our finalizer is present before we start any operations.
	if controllerutil.AddFinalizer(recipe, RecipeFinalizer) {
		err := r.Client.Update(ctx, recipe)
		if err != nil {
			logger.Error(err, "Unable to update resource.")
			return ctrl.Result{}, err
		}
	}

	recipe.Status.ObservedGeneration = recipe.Generation
	recipe.SetDefaults()

	err := EnsureApplication(ctx, r.Radius, recipe.Status.Environment, recipe.Status.Application)
	if err != nil {
		logger.Error(err, "Unable to ensure application.")
		return ctrl.Result{}, err
	}

	poller, err := r.StartUpdateOperationIfNeeded(ctx, recipe)
	if err != nil {
		logger.Error(err, "Unable to update resource.")
		return ctrl.Result{}, err
	}

	_, err = poller.Poll(ctx)
	if err != nil {
		logger.Error(err, "Unable to query Radius resource status.")
		return ctrl.Result{}, err
	}

	if !poller.Done() {
		err = r.HandleUpdating(ctx, recipe, poller)
		if err != nil {
			logger.Error(err, "Unable to update resource.")
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true, RequeueAfter: r.GetRequeueDelay()}, nil
	}

	err = r.HandleReady(ctx, recipe, poller)
	if err != nil {
		logger.Error(err, "Unable to update resource.")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *RecipeReconciler) ReconcileDelete(ctx context.Context, recipe *radappv1alpha3.Recipe) (ctrl.Result, error) {
	logger := ucplog.FromContextOrDiscard(ctx)

	recipe.Status.ObservedGeneration = recipe.Generation

	poller, err := r.StartDeleteOperationIfNeeded(ctx, recipe)
	if err != nil {
		logger.Error(err, "Unable to delete container.")
		return ctrl.Result{}, err
	}

	_, err = poller.Poll(ctx)
	if err != nil {
		logger.Error(err, "Unable to check operation status.")
		return ctrl.Result{}, err
	}

	if !poller.Done() {
		err = r.HandleDeleting(ctx, recipe, poller)
		if err != nil {
			logger.Error(err, "Unable to update resource.")
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true, RequeueAfter: r.GetRequeueDelay()}, nil
	}

	err = r.HandleDeleted(ctx, recipe, poller)
	if err != nil {
		logger.Error(err, "Unable to update resource.")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *RecipeReconciler) StartUpdateOperationIfNeeded(ctx context.Context, recipe *radappv1alpha3.Recipe) (Poller[generated.GenericResourcesClientCreateOrUpdateResponse], error) {
	logger := ucplog.FromContextOrDiscard(ctx)
	if recipe.Status.Operation != "" {
		logger.Info("Resuming operation.")
		return r.Radius.Resources(recipe.Status.Scope, recipe.Spec.Type).ContinueCreateOperation(ctx, recipe.Status.Operation)

	}

	logger.Info("Starting operation.")
	resource := recipe.Status.Scope + "/providers/" + recipe.Spec.Type + "/" + recipe.Name
	properties := map[string]any{
		"application":          recipe.Status.Application,
		"environment":          recipe.Status.Environment,
		"resourceProvisioning": "recipe",
	}
	return UpdateResource(ctx, r.Radius, resource, properties)
}

func (r *RecipeReconciler) HandleUpdating(ctx context.Context, recipe *radappv1alpha3.Recipe, poller Poller[generated.GenericResourcesClientCreateOrUpdateResponse]) error {
	logger := ucplog.FromContextOrDiscard(ctx)
	logger.Info("Resource is still updating.")
	token, err := poller.ResumeToken()
	if err != nil {
		return fmt.Errorf("failed to get operation token: %w", err)
	}

	recipe.Status.Operation = token
	recipe.Status.Phrase = radappv1alpha3.PhraseUpdating
	return r.Client.Status().Update(ctx, recipe)
}

func (r *RecipeReconciler) HandleReady(ctx context.Context, recipe *radappv1alpha3.Recipe, poller Poller[generated.GenericResourcesClientCreateOrUpdateResponse]) error {
	logger := ucplog.FromContextOrDiscard(ctx)
	logger.Info("Resource is updated.")

	recipe.Status.Operation = ""
	recipe.Status.Resource = recipe.Status.Scope + "/providers/" + recipe.Spec.Type + "/" + recipe.Name

	if recipe.Spec.SecretName != "" {
		_, err := r.CreateSecret(ctx, recipe, poller)
		if err != nil {
			return err
		}
	}

	if recipe.Spec.SecretName != recipe.Status.Secret.Name && recipe.Status.Secret.Name != "" {
		err := r.Client.Delete(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      recipe.Spec.SecretName,
				Namespace: recipe.Namespace,
			},
		})
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to delete secret %s: %w", recipe.Spec.SecretName, err)
		}
	}

	recipe.Status.Phrase = radappv1alpha3.PhraseReady
	return r.Client.Status().Update(ctx, recipe)
}

func (r *RecipeReconciler) StartDeleteOperationIfNeeded(ctx context.Context, recipe *radappv1alpha3.Recipe) (Poller[generated.GenericResourcesClientDeleteResponse], error) {
	logger := ucplog.FromContextOrDiscard(ctx)
	if recipe.Status.Operation != "" && recipe.Status.Phrase == radappv1alpha3.PhraseDeleting {
		logger.Info("Resuming operation.")
		return r.Radius.Resources(recipe.Status.Scope, recipe.Spec.Type).ContinueDeleteOperation(ctx, recipe.Status.Operation)
	}

	logger.Info("Starting operation.")
	return DeleteResource(ctx, r.Radius, recipe.Status.Resource)
}

func (r *RecipeReconciler) HandleDeleting(ctx context.Context, recipe *radappv1alpha3.Recipe, poller Poller[generated.GenericResourcesClientDeleteResponse]) error {
	logger := ucplog.FromContextOrDiscard(ctx)
	logger.Info("Resource is still deleting.")

	token, err := poller.ResumeToken()
	if err != nil {
		return fmt.Errorf("unable to get resume token: %w", err)
	}

	recipe.Status.Operation = token
	recipe.Status.Phrase = radappv1alpha3.PhraseDeleting

	return r.Client.Status().Update(ctx, recipe)
}

func (r *RecipeReconciler) HandleDeleted(ctx context.Context, recipe *radappv1alpha3.Recipe, poller Poller[generated.GenericResourcesClientDeleteResponse]) error {
	logger := ucplog.FromContextOrDiscard(ctx)
	logger.Info("Resource is deleted.")

	recipe.Status.Operation = ""
	recipe.Status.Phrase = radappv1alpha3.PhraseDeleted

	// Update status to show that deletion is complete.
	err := r.Client.Status().Update(ctx, recipe)
	if err != nil {
		return err
	}

	if recipe.Status.Secret.Name != "" {
		err := r.Client.Delete(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      recipe.Status.Secret.Name,
				Namespace: recipe.Namespace,
			},
		})
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to delete secret %s: %w", recipe.Status.Secret.Name, err)
		}
	}

	_ = controllerutil.RemoveFinalizer(recipe, RecipeFinalizer)

	// Update the resource to remove the finalizer.
	return r.Client.Update(ctx, recipe)
}

func (r *RecipeReconciler) CreateSecret(ctx context.Context, recipe *radappv1alpha3.Recipe, poller Poller[generated.GenericResourcesClientCreateOrUpdateResponse]) (*corev1.Secret, error) {
	result, err := poller.Result(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	secret := corev1.Secret{}
	err = r.Client.Get(ctx, client.ObjectKey{Namespace: recipe.Namespace, Name: recipe.Spec.SecretName}, &secret)
	if apierrors.IsNotFound(err) {
		secret = corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      recipe.Spec.SecretName,
				Namespace: recipe.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(recipe, radappv1alpha3.GroupVersion.WithKind("Recipe")),
				},
			},
		}

		err = r.Client.Create(ctx, &secret)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	if secret.StringData == nil {
		secret.StringData = map[string]string{}
	}

	values, err := resourceToConnectionValues(recipe.Name, result.GenericResource)
	if err != nil {
		return nil, fmt.Errorf("failed to read connection values: %w", err)
	}

	for k, v := range values {
		secret.StringData[k] = v
	}

	secrets, err := r.Radius.Resources(recipe.Status.Scope, recipe.Spec.Type).ListSecrets(ctx, recipe.Name)
	if clients.Is404Error(err) {
		// Safe to ignore. Not everything implements this.
	} else if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	} else {
		for k, v := range secrets.Value {
			secret.StringData[k] = *v
		}
	}

	err = r.Client.Update(ctx, &secret)
	if err != nil {
		return nil, fmt.Errorf("failed to update secret %s: %w", secret.Name, err)
	}

	recipe.Status.Secret = corev1.ObjectReference{
		APIVersion: secret.APIVersion,
		Kind:       secret.Kind,
		Namespace:  secret.Namespace,
		Name:       secret.Name,
		UID:        secret.UID,
	}

	return &secret, nil
}

func (r *RecipeReconciler) GetRequeueDelay() time.Duration {
	delay := r.Delay
	if delay == 0 {
		delay = PollingDelay
	}

	return delay
}

// SetupWithManager sets up the controller with the Manager.
func (r *RecipeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&radappv1alpha3.Recipe{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
