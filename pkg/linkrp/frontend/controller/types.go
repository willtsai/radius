// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package controller

import (
	ctrl "github.com/project-radius/radius/pkg/armrpc/frontend/controller"
	"github.com/project-radius/radius/pkg/linkrp/frontend/deployment"
)

// Options is the options to configure LinkRP controller.
type Options struct {
	ctrl.Options

	// DeployProcessor is the deployment processor for LinkRP
	DeployProcessor deployment.DeploymentProcessor
}

const (
	DaprInvokeHttpRoutesResourceTypeName  = "Applications.Link/daprInvokeHttpRoutes"
	DaprPubSubBrokersResourceTypeName     = "Applications.Link/daprPubSubBrokers"
	DaprSecretStoresResourceTypeName      = "Applications.Link/daprSecretStores"
	DaprStateStoresResourceTypeName       = "Applications.Link/daprStateStores"
	ExtendersResourceTypeName             = "Applications.Link/extenders"
	MongoDatabasesResourceTypeName        = "Applications.Link/mongoDatabases"
	RabbitMQMessageQueuesResourceTypeName = "Applications.Link/rabbitMQMessageQueues"
	RedisCachesResourceTypeName           = "Applications.Link/redisCaches"
	SqlDatabasesResourceTypeName          = "Applications.Link/sqlDatabases"

	// User defined operation names
	OperationListSecret = "LISTSECRETS"
)
