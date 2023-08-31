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
	"sort"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	"github.com/radius-project/radius/pkg/cli/clients"
	radappv1alpha3 "github.com/radius-project/radius/pkg/controller/api/rad_app/v1alpha3"
	"github.com/radius-project/radius/pkg/corerp/api/v20220315privatepreview"
	"github.com/radius-project/radius/pkg/to"
	"github.com/radius-project/radius/pkg/ucp/resources"
	"github.com/radius-project/radius/pkg/ucp/ucplog"
)

// DeploymentReconciler reconciles a Deployment object.
type DeploymentReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Radius RadiusClient
	Delay  time.Duration
}

// Reconcile is the main reconciliation loop for the Deployment resource.
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ucplog.FromContextOrDiscard(ctx).WithValues("kind", "Deployment", "name", req.Name, "namespace", req.Namespace)
	ctx = logr.NewContext(ctx, logger)

	deployment := appsv1.Deployment{}
	err := r.Client.Get(ctx, req.NamespacedName, &deployment)
	if apierrors.IsNotFound(err) {
		logger.Info("Deployment is being deleted.")
		return ctrl.Result{}, nil
	} else if err != nil {
		logger.Error(err, "Unable to fetch resource.")
		return ctrl.Result{}, err
	}

	annotations, err := readAnnotations(&deployment)
	if err != nil {
		logger.Error(err, "Failed to read deployment status.")
		deployment.Annotations[AnnotationRadiusStatus] = ""
		// continue processing
	} else if annotations == nil {
		logger.Info("Radius is not enabled for this deployment.")
		return ctrl.Result{}, nil
	}

	// This must be checked before we check the annotations. Deletion won't be reflected there.
	if deployment.DeletionTimestamp != nil && annotations.Status != nil {
		return r.ReconcileDelete(ctx, &deployment, annotations)
	}

	if annotations.IsUpToDate() {
		logger.Info("Deployment is up-to-date.")
		return ctrl.Result{}, nil
	}

	return r.ReconcileUpdate(ctx, &deployment, annotations)
}

func (r *DeploymentReconciler) ReconcileUpdate(ctx context.Context, deployment *appsv1.Deployment, annotations *deploymentAnnotations) (ctrl.Result, error) {
	logger := ucplog.FromContextOrDiscard(ctx)

	// Ensure that our finalizer is present before we start any operations.
	if controllerutil.AddFinalizer(deployment, DeploymentFinalizer) {
		err := r.Client.Update(ctx, deployment)
		if err != nil {
			logger.Error(err, "Unable to update resource.")
			return ctrl.Result{}, err
		}
	}

	annotations.SetDefaults(deployment.Namespace, deployment.Name)

	err := EnsureApplication(ctx, r.Radius, annotations.Status.Environment, annotations.Status.Application)
	if err != nil {
		logger.Error(err, "Unable to ensure application.")
		return ctrl.Result{}, err
	}

	// Ensure that our finalizer is present before we start any operations.
	controllerutil.AddFinalizer(deployment, DeploymentFinalizer)

	poller, err := r.StartPutOperationIfNeeded(ctx, deployment, annotations)
	if err != nil {
		logger.Error(err, "Unable to update container.")
		return ctrl.Result{}, err
	}

	if poller == nil {
		err = r.HandleWaiting(ctx, deployment, annotations)
		if err != nil {
			logger.Error(err, "Unable to update resource.")
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true, RequeueAfter: r.GetRequeueDelay()}, nil
	}

	_, err = poller.Poll(ctx)
	if err != nil {
		logger.Error(err, "Unable to check operation status.")
		return ctrl.Result{}, err
	}

	if !poller.Done() {
		err = r.HandleUpdating(ctx, deployment, annotations, poller)
		if err != nil {
			logger.Error(err, "Unable to update resource.")
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true, RequeueAfter: r.GetRequeueDelay()}, nil
	}

	err = r.HandleReady(ctx, deployment, annotations, poller)
	if err != nil {
		logger.Error(err, "Unable to update resource.")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *DeploymentReconciler) ReconcileDelete(ctx context.Context, deployment *appsv1.Deployment, annotations *deploymentAnnotations) (ctrl.Result, error) {
	logger := ucplog.FromContextOrDiscard(ctx)

	poller, err := r.StartDeleteOperationIfNeeded(ctx, deployment, annotations)
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
		err = r.HandleDeleting(ctx, deployment, annotations, poller)
		if err != nil {
			logger.Error(err, "Unable to update resource.")
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true, RequeueAfter: r.GetRequeueDelay()}, nil
	}

	err = r.HandleDeleted(ctx, deployment, annotations, poller)
	if err != nil {
		logger.Error(err, "Unable to update resource.")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *DeploymentReconciler) StartPutOperationIfNeeded(ctx context.Context, deployment *appsv1.Deployment, annotations *deploymentAnnotations) (Poller[v20220315privatepreview.ContainersClientCreateOrUpdateResponse], error) {
	logger := ucplog.FromContextOrDiscard(ctx)
	if annotations.Status.Operation != "" && annotations.Status.State == DeploymentStateUpdating {
		logger.Info("Resuming operation.")
		return r.Radius.Containers(annotations.Status.Scope).ContinueCreateOperation(ctx, annotations.Status.Operation)
	}

	logger.Info("Starting operation.")
	properties := v20220315privatepreview.ContainerProperties{
		Application:          to.Ptr(annotations.Status.Application),
		ResourceProvisioning: to.Ptr(v20220315privatepreview.ResourceProvisioningManual),
		Connections:          map[string]*v20220315privatepreview.ConnectionProperties{},
		Container: &v20220315privatepreview.Container{
			Image: to.Ptr("none"),
		},
		Resources: []*v20220315privatepreview.ResourceReference{
			{
				ID: to.Ptr("/planes/kubernetes/local/namespaces/" + deployment.Namespace + "/providers/apps/Deployment/" + deployment.Name),
			},
		},
	}

	for name, source := range annotations.Configuration.Connections {
		recipe := radappv1alpha3.Recipe{}
		err := r.Client.Get(ctx, client.ObjectKey{Namespace: deployment.Namespace, Name: source}, &recipe)
		if apierrors.IsNotFound(err) {
			logger.Info("Recipe is not ready.", "recipe", source)
			return nil, nil
		} else if err != nil {
			return nil, fmt.Errorf("failed to fetch recipe %s: %w", source, err)
		} else if recipe.Status.Resource == "" {
			logger.Info("Recipe is not ready.", "recipe", source)
			return nil, nil
		}

		properties.Connections[name] = &v20220315privatepreview.ConnectionProperties{
			Source: to.Ptr(recipe.Status.Resource),
		}
	}

	return UpdateContainer(ctx, r.Radius, annotations.Status.Container, &properties)
}

func (r *DeploymentReconciler) HandleWaiting(ctx context.Context, deployment *appsv1.Deployment, annotations *deploymentAnnotations) error {
	logger := ucplog.FromContextOrDiscard(ctx)
	logger.Info("Waiting on dependencies.")

	annotations.Status.Operation = ""
	annotations.Status.State = DeploymentStateWaiting
	err := annotations.ApplyToDeployment(deployment)
	if err != nil {
		return fmt.Errorf("unable to apply annotations: %w", err)
	}

	return r.Client.Update(ctx, deployment)
}

func (r *DeploymentReconciler) HandleUpdating(ctx context.Context, deployment *appsv1.Deployment, annotations *deploymentAnnotations, poller Poller[v20220315privatepreview.ContainersClientCreateOrUpdateResponse]) error {
	logger := ucplog.FromContextOrDiscard(ctx)
	logger.Info("Resource is still updating.")

	token, err := poller.ResumeToken()
	if err != nil {
		return fmt.Errorf("unable to get resume token: %w", err)
	}

	annotations.Status.Operation = token
	annotations.Status.State = DeploymentStateUpdating
	err = annotations.ApplyToDeployment(deployment)
	if err != nil {
		return fmt.Errorf("unable to apply annotations: %w", err)
	}

	return r.Client.Update(ctx, deployment)
}

func (r *DeploymentReconciler) HandleReady(ctx context.Context, deployment *appsv1.Deployment, annotations *deploymentAnnotations, poller Poller[v20220315privatepreview.ContainersClientCreateOrUpdateResponse]) error {
	logger := ucplog.FromContextOrDiscard(ctx)
	logger.Info("Resource is updated.")

	annotations.Status.Operation = ""
	annotations.Status.State = DeploymentStateReady

	envVars := map[string]corev1.EnvVar{}

	for name, source := range annotations.Configuration.Connections {
		recipe := radappv1alpha3.Recipe{}
		err := r.Client.Get(ctx, client.ObjectKey{Namespace: deployment.Namespace, Name: source}, &recipe)
		if err != nil {
			return fmt.Errorf("failed to fetch recipe %s: %w", source, err)
		}

		if recipe.Status.Resource == "" {
			return fmt.Errorf("recipe %s is not ready", source)
		}

		id, err := resources.Parse(recipe.Status.Resource)
		if err != nil {
			return err
		}

		response, err := r.Radius.Resources(id.RootScope(), id.Type()).Get(ctx, id.Name())
		if err != nil {
			return fmt.Errorf("failed to fetch resource %s: %w", id, err)
		}

		secrets, err := r.Radius.Resources(id.RootScope(), id.Type()).ListSecrets(ctx, id.Name())
		if clients.Is404Error(err) {
			// This is fine. The resource doesn't have any secrets.
			secrets.Value = map[string]*string{}
		} else if err != nil {
			return fmt.Errorf("failed to fetch secrets for resource %s: %w", id, err)
		}

		values, err := resourceToConnectionEnvVars(name, response.GenericResource, secrets)
		if err != nil {
			return fmt.Errorf("failed to read values resource %s: %w", id, err)
		}

		for k, v := range values {
			envVars[k] = corev1.EnvVar{Name: k, Value: v}
		}
	}

	if len(envVars) > 0 {
		// Preserve any values set by the user.
		for _, env := range deployment.Spec.Template.Spec.Containers[0].Env {
			envVars[env.Name] = env
		}

		keys := []string{}
		for k := range envVars {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		deployment.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{}
		for _, key := range keys {
			deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, envVars[key])
		}
	}

	err := annotations.ApplyToDeployment(deployment)
	if err != nil {
		return fmt.Errorf("unable to apply annotations: %w", err)
	}

	return r.Client.Update(ctx, deployment)
}

func (r *DeploymentReconciler) StartDeleteOperationIfNeeded(ctx context.Context, deployment *appsv1.Deployment, annotations *deploymentAnnotations) (Poller[v20220315privatepreview.ContainersClientDeleteResponse], error) {
	logger := ucplog.FromContextOrDiscard(ctx)
	if annotations.Status.Operation != "" && annotations.Status.State == DeploymentStateDeleting {
		logger.Info("Resuming operation.")
		return r.Radius.Containers(annotations.Status.Scope).ContinueDeleteOperation(ctx, annotations.Status.Operation)
	}

	logger.Info("Starting operation.")
	return DeleteContainer(ctx, r.Radius, annotations.Status.Container)
}

func (r *DeploymentReconciler) HandleDeleting(ctx context.Context, deployment *appsv1.Deployment, annotations *deploymentAnnotations, poller Poller[v20220315privatepreview.ContainersClientDeleteResponse]) error {
	logger := ucplog.FromContextOrDiscard(ctx)
	logger.Info("Resource is still deleting.")

	token, err := poller.ResumeToken()
	if err != nil {
		return fmt.Errorf("unable to get resume token: %w", err)
	}

	annotations.Status.Operation = token
	annotations.Status.State = DeploymentStateDeleting
	err = annotations.ApplyToDeployment(deployment)
	if err != nil {
		return fmt.Errorf("unable to apply annotations: %w", err)
	}

	return r.Client.Update(ctx, deployment)
}

func (r *DeploymentReconciler) HandleDeleted(ctx context.Context, deployment *appsv1.Deployment, annotations *deploymentAnnotations, poller Poller[v20220315privatepreview.ContainersClientDeleteResponse]) error {
	logger := ucplog.FromContextOrDiscard(ctx)
	logger.Info("Resource is deleted.")

	annotations.Status.Operation = ""
	annotations.Status.State = DeploymentStateDeleted
	_ = controllerutil.RemoveFinalizer(deployment, DeploymentFinalizer)

	err := annotations.ApplyToDeployment(deployment)
	if err != nil {
		return fmt.Errorf("unable to apply annotations: %w", err)
	}

	return r.Client.Update(ctx, deployment)
}

func (r *DeploymentReconciler) findDeploymentsForRecipe(ctx context.Context, obj client.Object) []reconcile.Request {
	recipe := obj.(*radappv1alpha3.Recipe)

	deployments := &appsv1.DeploymentList{}
	options := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(indexField, recipe.Name),
		Namespace:     recipe.Namespace,
	}
	err := r.Client.List(ctx, deployments, options)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := []reconcile.Request{}
	for _, item := range deployments.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		})
	}
	return requests
}

func (r *DeploymentReconciler) GetRequeueDelay() time.Duration {
	delay := r.Delay
	if delay == 0 {
		delay = PollingDelay
	}

	return delay
}

const indexField = "spec.recipe-reference"

func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &appsv1.Deployment{}, indexField, func(rawObj client.Object) []string {
		deployment := rawObj.(*appsv1.Deployment)
		annotations, err := readAnnotations(deployment)
		if err != nil {
			return []string{}
		} else if annotations == nil {
			return []string{}
		}

		recipes := []string{}
		for _, recipe := range annotations.Configuration.Connections {
			recipes = append(recipes, recipe)
		}

		return recipes
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Watches(&radappv1alpha3.Recipe{}, handler.EnqueueRequestsFromMapFunc(r.findDeploymentsForRecipe), builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Complete(r)
}
