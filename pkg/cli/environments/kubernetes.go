// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package environments

import (
	"context"
	"net/http"
	"time"

	"github.com/Azure/go-autorest/autorest"
	azclients "github.com/project-radius/radius/pkg/azure/clients"
	"github.com/project-radius/radius/pkg/azure/radclient"
	"github.com/project-radius/radius/pkg/cli/azure"
	"github.com/project-radius/radius/pkg/cli/clients"
	"github.com/project-radius/radius/pkg/cli/kubernetes"
)

// KubernetesEnvironment represents a Kubernetes Radius environment.
type KubernetesEnvironment struct {
	Name                       string `mapstructure:"name" validate:"required"`
	Kind                       string `mapstructure:"kind" validate:"required"`
	Context                    string `mapstructure:"context" validate:"required"`
	Namespace                  string `mapstructure:"namespace" validate:"required"`
	DefaultApplication         string `mapstructure:"defaultapplication,omitempty"`
	APIServerBaseURL           string `mapstructure:"apiserverbaseurl,omitempty"`
	APIDeploymentEngineBaseURL string `mapstructure:"apideploymentenginebaseurl,omitempty"`
	EnableUCP                  bool   `mapstructure:"enableucp,omitempty"`

	// We tolerate and allow extra fields - this helps with forwards compat.
	Properties map[string]interface{} `mapstructure:",remain"`
}

func (e *KubernetesEnvironment) GetName() string {
	return e.Name
}

func (e *KubernetesEnvironment) GetKind() string {
	return e.Kind
}

func (e *KubernetesEnvironment) GetEnableUCP() bool {
	return e.EnableUCP
}

func (e *KubernetesEnvironment) GetDefaultApplication() string {
	return e.DefaultApplication
}

func (e *KubernetesEnvironment) GetContainerRegistry() *Registry {
	return nil
}

// No Status Link for kubernetes
func (e *KubernetesEnvironment) GetStatusLink() string {
	return ""
}

var _ autorest.Sender = (*sender)(nil)

type sender struct {
	RoundTripper http.RoundTripper
}

func (s *sender) Do(request *http.Request) (*http.Response, error) {
	return s.RoundTripper.RoundTrip(request)
}

func (e *KubernetesEnvironment) CreateDeploymentClient(ctx context.Context) (clients.DeploymentClient, error) {
	url, roundTripper, err := kubernetes.GetBaseUrlAndRoundTripperForDeploymentEngine(e.APIDeploymentEngineBaseURL, e.Context, e.EnableUCP)

	if err != nil {
		return nil, err
	}

	if e.EnableUCP {
		ucpClient := azclients.NewUCPDeploymentClient(url)
		ucpClient.PollingDelay = 5 * time.Second
		ucpClient.Sender = &sender{RoundTripper: roundTripper}

		ucpOperationClient := azclients.NewUCPOperationClient(url)
		ucpOperationClient.PollingDelay = 5 * time.Second
		ucpOperationClient.Sender = &sender{RoundTripper: roundTripper}

		return &azure.ARMDeploymentClient{
			Client:           ucpClient,
			OperationsClient: ucpOperationClient,
			ResourceGroup:    e.Namespace,
		}, nil
	} else {
		dc := azclients.NewDeploymentsClientWithBaseURI(url, e.Namespace)

		// Poll faster than the default, many deployments are quick
		dc.PollingDelay = 5 * time.Second

		dc.Sender = &sender{RoundTripper: roundTripper}

		op := azclients.NewOperationsClientWithBaseUri(url, e.Namespace)
		op.PollingDelay = 5 * time.Second
		op.Sender = &sender{RoundTripper: roundTripper}
		return &azure.ARMDeploymentClient{
			Client:           dc,
			OperationsClient: op,
			ResourceGroup:    e.Namespace,
		}, nil
	}
}

func (e *KubernetesEnvironment) CreateDiagnosticsClient(ctx context.Context) (clients.DiagnosticsClient, error) {
	k8sClient, config, err := kubernetes.CreateTypedClient(e.Context)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.CreateRuntimeClient(e.Context, kubernetes.Scheme)
	if err != nil {
		return nil, err
	}

	_, con, err := kubernetes.CreateAPIServerConnection(e.Context, e.APIServerBaseURL)
	if err != nil {
		return nil, err
	}

	return &azure.ARMDiagnosticsClient{
		K8sTypedClient:   k8sClient,
		RestConfig:       config,
		K8sRuntimeClient: client,
		ResourceClient:   *radclient.NewRadiusResourceClient(con, e.Namespace),
		ResourceGroup:    e.Namespace,
		SubscriptionID:   e.Namespace,
	}, nil
}

func (e *KubernetesEnvironment) CreateManagementClient(ctx context.Context) (clients.ManagementClient, error) {
	_, connection, err := kubernetes.CreateAPIServerConnection(e.Context, e.APIServerBaseURL)
	if err != nil {
		return nil, err
	}

	return &azure.ARMManagementClient{
		EnvironmentName: e.Name,
		Connection:      connection,
		ResourceGroup:   e.Namespace, // Temporarily set resource group and subscription id to the namespace
		SubscriptionID:  e.Namespace,
	}, nil
}
