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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	// DeploymentWebhookMutatePath is the path for the deployment mutating webhook. This must match the webhook registration.
	DeploymentWebhookMutatePath = "/mutate-apps-v1-deployment"
)

type DeploymentWebhook struct {
}

var _ admission.Handler = (*DeploymentWebhook)(nil)

func (r *DeploymentWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	mgr.GetWebhookServer().Register(DeploymentWebhookMutatePath, &webhook.Admission{Handler: &DeploymentWebhook{}})
	return nil
}

// Handle implements admission.Handler.
func (*DeploymentWebhook) Handle(context.Context, admission.Request) admission.Response {
	return admission.Allowed("YOLO")
}
