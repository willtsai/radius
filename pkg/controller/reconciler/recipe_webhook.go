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

	"github.com/radius-project/radius/pkg/controller/api/rad_app/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	// RecipeWebhookMutatePath is the path for the recipe mutating webhook. This must match the webhook registration.
	RecipeWebhookMutatePath = "/mutate-rad-app-v1alpha3-recipe"

	// RecipeWebhookValidatePath is the path for the recipe validating webhook. This must match the webhook registration.
	RecipeWebhookValidatePath = "/validate-rad-app-v1alpha3-recipe"
)

type RecipeWebhook struct {
}

var _ webhook.CustomDefaulter = (*RecipeWebhook)(nil)
var _ webhook.CustomValidator = (*RecipeWebhook)(nil)

func (r *RecipeWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha3.Recipe{}).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

// Default implements admission.CustomDefaulter.
func (*RecipeWebhook) Default(ctx context.Context, obj runtime.Object) error {
	return nil
}

// ValidateCreate implements admission.CustomValidator.
func (*RecipeWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return admission.Warnings{}, nil
}

// ValidateDelete implements admission.CustomValidator.
func (*RecipeWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return admission.Warnings{}, nil
}

// ValidateUpdate implements admission.CustomValidator.
func (*RecipeWebhook) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) (admission.Warnings, error) {
	return admission.Warnings{}, nil
}
