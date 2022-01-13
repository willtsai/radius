//-----------------------------------------------------------
// Copyright (c) Microsoft Corporation.  All rights reserved.
//-----------------------------------------------------------

namespace Microsoft.WindowsAzure.ResourceStack.Frontdoor.Data.Engines
{
    using System;
    using System.Collections.Generic;
    using System.Collections.Specialized;
    using System.Linq;
    using System.Net.Http;
    using Azure.Deployments.Core.Resources;
    using Azure.Deployments.Core.Uri;
    using Microsoft.WindowsAzure.ResourceStack.Common.Extensions;
    using Microsoft.WindowsAzure.ResourceStack.Common.Instrumentation;

    /// <summary>
    /// The Uri template engine.
    /// </summary>
    public static class UriTemplateEngine
    {
        /// <summary>
        /// The resource uri template.
        /// </summary>
        private static readonly UriTemplate ResourceUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/{resourceId}?api-version={api-version}");

        /// <summary>
        /// The subscription level resource uri template.
        /// </summary>
        private static readonly UriTemplate SubscriptionResourceUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/providers/{resourceId}?api-version={api-version}");

        /// <summary>
        /// The tenant level resource uri template.
        /// </summary>
        private static readonly UriTemplate TenantResourceUriTemplate = new UriTemplate(
            template: "providers/{resourceId}?api-version={api-version}");

        /// <summary>
        /// The fully qualified resource Id uri template.
        /// </summary>
        private static readonly UriTemplate FullyQualifiedResourceIdUriTemplate = new UriTemplate(
            template: "{fullyQualifiedResourceId}?api-version={api-version}");

        /// <summary>
        /// The resource action uri template.
        /// </summary>
        private static readonly UriTemplate ResourceActionUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/{resourceId}/{actionVerb}?api-version={api-version}");

        /// <summary>
        /// The subscription level resource action uri template.
        /// </summary>
        private static readonly UriTemplate SubscriptionResourceActionUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/providers/{resourceId}/{actionVerb}?api-version={api-version}");

        /// <summary>
        /// The move resource uri template.
        /// </summary>
        private static readonly UriTemplate ResourceMoveUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/{resourceId}/move?api-version={api-version}");

        /// <summary>
        /// The resource group uri template.
        /// </summary>
        private static readonly UriTemplate ResourceGroupUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}?api-version={api-version}");

        /// <summary>
        /// The resource groups collection uri template.
        /// </summary>
        private static readonly UriTemplate ResourceGroupsUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/resourcegroups?api-version={api-version}");

        /// <summary>
        /// The resource batch move identities move Uri template.
        /// </summary>
        private static readonly UriTemplate ResourceBatchMoveIdentitiesMoveUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/moveIdentities?api-version={api-version}");

        /// <summary>
        /// The resource batch move provider notification Uri template.
        /// </summary>
        private static readonly UriTemplate ResourceBatchMoveProviderNotificationUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/moveResources?api-version={api-version}");

        /// <summary>
        /// The resource batch move provider validation Uri template.
        /// </summary>
        private static readonly UriTemplate ResourceBatchMoveProviderValidationUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/validateMoveResources?api-version={api-version}");

        /// <summary>
        /// The async operation uri template.
        /// </summary>
        private static readonly UriTemplate OperationResultsUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/operationresults/{operationId}?api-version={api-version}");

        /// <summary>
        /// The subscription uri template.
        /// </summary>
        private static readonly UriTemplate SubscriptionUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}?api-version={api-version}");

        /// <summary>
        /// The management group uri template.
        /// </summary>
        private static readonly UriTemplate ManagementGroupUriTemplate = new UriTemplate(
            template: "providers/Microsoft.Management/managementGroups/{managementGroupId}?api-version={api-version}");

        /// <summary>
        /// The subscription resources uri template.
        /// </summary>
        private static readonly UriTemplate SubscriptionResourcesUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/resources?api-version={api-version}");

        /// <summary>
        /// The subscription resources uri with filter template.
        /// </summary>
        private static readonly UriTemplate SubscriptionResourcesUriWithFilterTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/resources?$filter={filter}&api-version={api-version}");

        /// <summary>
        /// The role assignment uri template.
        /// </summary>
        private static readonly UriTemplate AssignmentUriTemplate = new UriTemplate(
            template: "{scope}/providers/Microsoft.Authorization/roleAssignments/{roleAssignmentName}?api-version={api-version}");

        /// <summary>
        /// The role assignment uri template for principal.
        /// </summary>
        private static readonly UriTemplate AssignmentsForPrincipalAtScopeUriTemplate = new UriTemplate(
            template: "{scope}/providers/Microsoft.Authorization/roleAssignments?$filter={principalIdFilter}&api-version={api-version}");

        /// <summary>
        /// The subscription level resource provider registration uri template.
        /// </summary>
        private static readonly UriTemplate SubscriptionResourceProviderRegisterUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/providers/{resourceProviderNamespace}/register?api-version={api-version}");

        /// <summary>
        /// The subscription level resource query for a resource provider.
        /// </summary>
        private static readonly UriTemplate SubscriptionResourceProviderResourcesUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/providers/{resourceProviderNamespace}/{resourceType}?api-version={api-version}");

        /// <summary>
        /// The resource group level resource query for a resource provider.
        /// </summary>
        private static readonly UriTemplate ResourceGroupProviderResourcesUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/{resourceProviderNamespace}/{resourceType}?api-version={api-version}");

        /// <summary>
        /// The resource group level nested resource query for a resource provider.
        /// </summary>
        private static readonly UriTemplate ResourceGroupProviderNestedResourcesUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/{parentResourceId}/{nestedResourceType}?api-version={api-version}");

        /// <summary>
        /// The linked resource provider notification uri template.
        /// </summary>
        private static readonly UriTemplate ResourceProviderLinkedNotificationUriTemplate = new UriTemplate(
            template: "{scope}/providers/{linkedResourceProvider}/notify?api-version={api-version}");

        /// <summary>
        /// The tenant deployment operation status uri.
        /// </summary>
        private static readonly UriTemplate TenantDeploymentAzureAsyncOperationsUriTemplate = new UriTemplate(
            template: "providers/Microsoft.Resources/deployments/{deploymentName}/operationStatuses/{deploymentSequence}?api-version={api-version}");

        /// <summary>
        /// The management group deployment operation status uri.
        /// </summary>
        private static readonly UriTemplate ManagementGroupDeploymentAzureAsyncOperationsUriTemplate = new UriTemplate(
            template: "providers/Microsoft.Management/managementGroups/{managementGroupId}/providers/Microsoft.Resources/deployments/{deploymentName}/operationStatuses/{deploymentSequence}?api-version={api-version}");

        /// <summary>
        /// The subscription deployment operation status uri.
        /// </summary>
        private static readonly UriTemplate SubscriptionDeploymentAzureAsyncOperationsUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/providers/Microsoft.Resources/deployments/{deploymentName}/operationStatuses/{deploymentSequence}?api-version={api-version}");

        /// <summary>
        /// The resource group deployment operation status uri.
        /// </summary>
        private static readonly UriTemplate ResourceGroupDeploymentAzureAsyncOperationsUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/Microsoft.Resources/deployments/{deploymentName}/operationStatuses/{deploymentSequence}?api-version={api-version}");

        /// <summary>
        /// The deployment preflight URI template for tenant level resources.
        /// </summary>
        private static readonly UriTemplate TemplateDeploymentTenantPreflightUriTemplate = new UriTemplate(
            template: "providers/{resourceProviderNamespace}/deployments/{deploymentName}/preflight?api-version={api-version}");

        /// <summary>
        /// The deployment preflight URI template for management group level resources.
        /// </summary>
        private static readonly UriTemplate TemplateDeploymentManagementGroupPreflightUriTemplate = new UriTemplate(
            template: "providers/Microsoft.Management/managementGroups/{managementGroupId}/providers/{resourceProviderNamespace}/deployments/{deploymentName}/preflight?api-version={api-version}");

        /// <summary>
        /// The deployment preflight URI template for subscription level resources.
        /// </summary>
        private static readonly UriTemplate TemplateDeploymentSubscriptionPreflightUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/providers/{resourceProviderNamespace}/deployments/{deploymentName}/preflight?api-version={api-version}");

        /// <summary>
        /// The deployment preflight URI template.
        /// </summary>
        private static readonly UriTemplate TemplateDeploymentPreflightUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/{resourceProviderNamespace}/deployments/{deploymentName}/preflight?api-version={api-version}");

        /// <summary>
        /// The deployment validate URI template for tenant level resources.
        /// </summary>
        private static readonly UriTemplate TemplateDeploymentTenantValidateUriTemplate = new UriTemplate(
            template: "providers/Microsoft.Resources/deployments/{deploymentName}/validate?api-version={api-version}");

        /// <summary>
        /// The deployment validate URI template for management group level resources.
        /// </summary>
        private static readonly UriTemplate TemplateDeploymentManagementGroupValidateUriTemplate = new UriTemplate(
            template: "providers/Microsoft.Management/managementGroups/{managementGroupId}/providers/Microsoft.Resources/deployments/{deploymentName}/validate?api-version={api-version}");

        /// <summary>
        /// The deployment validate URI template for subscription level resources.
        /// </summary>
        private static readonly UriTemplate TemplateDeploymentSubscriptionValidateUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/providers/Microsoft.Resources/deployments/{deploymentName}/validate?api-version={api-version}");

        /// <summary>
        /// The deployment resource group validate URI template.
        /// </summary>
        private static readonly UriTemplate TemplateDeploymentResourceGroupValidateUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/Microsoft.Resources/deployments/{deploymentName}/validate?api-version={api-version}");

        /// <summary>
        /// The deployment redeploy URI template.
        /// </summary>
        private static readonly UriTemplate TemplateDeploymentRedeployUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/Microsoft.Resources/deployments/{deploymentName}/redeploy?api-version={api-version}");

        /// <summary>
        /// The resource identity uri template.
        /// </summary>
        private static readonly UriTemplate ResourceIdentityUriTemplate = new UriTemplate(
            template: "{scope}/providers/Microsoft.ManagedIdentity/identities/default?api-version={api-version}");

        /// <summary>
        /// The event grid system topic uri template.
        /// </summary>
        private static readonly UriTemplate EventGridSystemTopicUriTemplate = new UriTemplate(
            template: "eventGrid/api/events?api-version={api-version}");

        /// <summary>
        /// The event grid custom topic uri template.
        /// </summary>
        private static readonly UriTemplate EventGridCustomTopicUriTemplate = new UriTemplate(
            template: "?api-version={api-version}",
            ignoreTrailingSlash: true);

        /// <summary>
        /// The tenant provider operations uri template.
        /// </summary>
        private static readonly UriTemplate TenantProviderOperationsUriTemplate = new UriTemplate(
            template: "providers/{resourceProviderNamespace}/operations?api-version={api-version}");

        /// <summary>
        /// The async batch operation results uri template.
        /// </summary>
        private static readonly UriTemplate AsyncBatchOperationResultsUriTemplate = new UriTemplate(
            template: "batch/{batchOperationId}?api-version={apiVersion}");

        /// <summary>
        /// The async batch operation results uri template.
        /// </summary>
        private static readonly UriTemplate BatchUriTemplate = new UriTemplate(
            template: "batch?api-version={apiVersion}");

        /// <summary>
        /// The async bulk deletion operation results uri template.
        /// </summary>
        private static readonly UriTemplate AsyncBulkDeletionOperationResultsUriTemplate = new UriTemplate(
            template: "bulkDelete/{bulkDeletionOperationId}?api-version={apiVersion}");

        /// <summary>
        /// The async notification uri template.
        /// </summary>
        private static readonly UriTemplate AsyncOperationCallbackUriTemplate = new UriTemplate(
            template: "providers/Microsoft.Resources/notifyResourceJobs?api-version={api-version}&asyncNotificationToken={asyncNotificationToken}");

        /// <summary>
        /// The tenant level async operation uri template.
        /// </summary>
        private static readonly UriTemplate TenantOperationResultsUriTemplate = new UriTemplate(
            template: "providers/Microsoft.Resources/operationResults/{operationId}?api-version={api-version}");

        /// <summary>
        /// The policy cleanup uri template.
        /// </summary>
        private static readonly UriTemplate PolicyCleanupUriTemplate = new UriTemplate(
            template: "{scope}/providers/Microsoft.Authorization/PolicyDataCleanup?api-version={api-version}");

        /// <summary>
        /// The trigger policy driven deployment uri template.
        /// </summary>
        private static readonly UriTemplate TriggerPolicyDeploymentUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/Microsoft.Authorization/triggerPolicyDeployment?api-version={api-version}");

        /// <summary>
        /// The trigger policy driven subscription level deployment uri template.
        /// </summary>
        private static readonly UriTemplate TriggerPolicySubscriptionDeploymentUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/providers/Microsoft.Authorization/triggerPolicyDeployment?api-version={api-version}");

        /// <summary>
        /// The policy patch pass-through uri template.
        /// </summary>
        private static readonly UriTemplate PolicyPassthroughPatchOperationUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/providers/Microsoft.Authorization/patchResource?api-version={api-version}");

        /// <summary>
        /// The policy put pass-through uri template.
        /// </summary>
        private static readonly UriTemplate PolicyPassthroughPutOperationUriTemplate = new UriTemplate(
            template: "subscriptions/{subscriptionId}/providers/Microsoft.Authorization/putResource?api-version={api-version}");

        /// <summary>
        /// The API version of Microsoft.Authorization provider.
        /// </summary>
        public static readonly string AuthorizationProviderVersion = "2018-07-01";

        /// <summary>
        /// The API version for the resource provider contract; used for communicating with RPs when
        /// we do not know which <c>api</c>-version to use.
        /// </summary>
        public static readonly string ProviderContractVersion = "2.0";

        /// <summary>
        /// Get tenant level async operation URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="operationId">The operation identifier.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetTenantOperationResultsUri(Uri endpoint, string operationId, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "operationId", operationId },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.TenantOperationResultsUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Get async operation URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="operationId">The operation identifier.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetOperationResultsUri(Uri endpoint, string subscriptionId, string operationId, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "operationId", operationId },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.OperationResultsUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the subscription resource URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="resourceId">The resource identifier.</param>
        /// <param name="apiVersion">The API version.</param>
        private static Uri GetResourceUri(Uri endpoint, string subscriptionId, string resourceId, string apiVersion)
        {
            if (IResourceIdentifiableExtensions.IsResourceGroupType(resourceId))
            {
                var resourceGroupName = resourceId.SplitRemoveEmpty('/').Last();

                return UriTemplateEngine.GetResourceGroupUri(
                    endpoint: endpoint,
                    subscriptionId: subscriptionId,
                    resourceGroupName: resourceGroupName,
                    apiVersion: apiVersion);
            }

            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceId", resourceId },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.SubscriptionResourceUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the tenant resource URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="resourceId">The resource identifier.</param>
        /// <param name="apiVersion">The API version.</param>
        private static Uri GetTenantResourceUri(Uri endpoint, string resourceId, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "resourceId", resourceId },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.TenantResourceUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the resource URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="resourceGroupName">Name of the resource group.</param>
        /// <param name="resourceId">The resource identifier.</param>
        /// <param name="apiVersion">The API version.</param>
        /// <param name="extraParametersInURL">The extra parameters in the URL.</param>
        public static Uri GetResourceUri(Uri endpoint, string subscriptionId, string resourceGroupName, string resourceId, string apiVersion, Dictionary<string, string> extraParametersInURL = null)
        {
            if (string.IsNullOrEmpty(subscriptionId))
            {
                return UriTemplateEngine.GetTenantResourceUri(
                    endpoint: endpoint,
                    resourceId: resourceId,
                    apiVersion: apiVersion);
            }

            if (string.IsNullOrEmpty(resourceGroupName))
            {
                return UriTemplateEngine.GetResourceUri(
                    endpoint: endpoint,
                    subscriptionId: subscriptionId,
                    resourceId: resourceId,
                    apiVersion: apiVersion);
            }

            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceGroupName", resourceGroupName },
                { "resourceId", resourceId },
                { "api-version", apiVersion },
            };

            parameters = extraParametersInURL.CoalesceDictionary().Any()
                ? parameters.Union(extraParametersInURL).ToDictionary(k => k.Key, v => v.Value)
                : parameters;

            return UriTemplateEngine.ResourceUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the deployment URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="fullyQualifiedResourceId">The fully qualified resource Id.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetResourceUri(Uri endpoint, string fullyQualifiedResourceId, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "fullyQualifiedResourceId", fullyQualifiedResourceId.TrimStart('/') },
                { "api-version", apiVersion }
            };

            return UriTemplateEngine.FullyQualifiedResourceIdUriTemplate.BindByName(
                baseAddress: endpoint,
                parameters: parameters);
        }

        /// <summary>
        /// Gets the resource action URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="resourceGroupName">Name of the resource group.</param>
        /// <param name="resourceId">The resource identifier.</param>
        /// <param name="actionVerb">The action verb.</param>
        /// <param name="apiVersion">the API version.</param>
        public static Uri GetResourceActionUri(Uri endpoint, string subscriptionId, string resourceGroupName, string resourceId, string actionVerb, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceGroupName", resourceGroupName },
                { "resourceId", resourceId },
                { "actionVerb", actionVerb },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.ResourceActionUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the subscription level resource action URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="resourceId">The resource identifier.</param>
        /// <param name="actionVerb">The action verb.</param>
        /// <param name="apiVersion">the API version.</param>
        public static Uri GetSubscriptionResourceActionUri(Uri endpoint, string subscriptionId, string resourceId, string actionVerb, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceId", resourceId },
                { "actionVerb", actionVerb },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.SubscriptionResourceActionUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the move resource URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="resourceGroupName">Name of the resource group.</param>
        /// <param name="resourceId">The resource identifier.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetResourceMoveUri(Uri endpoint, string subscriptionId, string resourceGroupName, string resourceId, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceGroupName", resourceGroupName },
                { "resourceId", resourceId },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.ResourceMoveUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the resource group URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="resourceGroupName">Name of the resource group.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetResourceGroupUri(Uri endpoint, string subscriptionId, string resourceGroupName, string apiVersion)
        {
            if (string.IsNullOrEmpty(resourceGroupName))
            {
                var parameters = new Dictionary<string, string>()
                {
                    { "subscriptionId", subscriptionId },
                    { "api-version", apiVersion },
                };

                return UriTemplateEngine.ResourceGroupsUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
            }

            var singleGroupParameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceGroupName", resourceGroupName },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.ResourceGroupUriTemplate.BindByName(baseAddress: endpoint, parameters: singleGroupParameters);
        }

        /// <summary>
        /// Gets the resource batch move provider notification Uri.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription Id.</param>
        /// <param name="resourceGroupName">The resource group name.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetResourceBatchMoveIdentitiesMoveUri(Uri endpoint, string subscriptionId, string resourceGroupName, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceGroupName", resourceGroupName },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.ResourceBatchMoveIdentitiesMoveUriTemplate.BindByName(
                baseAddress: endpoint,
                parameters: parameters);
        }

        /// <summary>
        /// Gets the resource batch move provider notification Uri.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription Id.</param>
        /// <param name="resourceGroupName">The resource group name.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetResourceBatchMoveProviderNotificationUri(Uri endpoint, string subscriptionId, string resourceGroupName, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceGroupName", resourceGroupName },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.ResourceBatchMoveProviderNotificationUriTemplate.BindByName(
                baseAddress: endpoint,
                parameters: parameters);
        }

        /// <summary>
        /// Gets the resource batch move provider validation Uri.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription Id.</param>
        /// <param name="resourceGroupName">The resource group name.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetResourceBatchMoveProviderValidationUri(Uri endpoint, string subscriptionId, string resourceGroupName, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceGroupName", resourceGroupName },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.ResourceBatchMoveProviderValidationUriTemplate.BindByName(
                baseAddress: endpoint,
                parameters: parameters);
        }

        /// <summary>
        /// Gets the subscription URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        public static Uri GetSubscriptionUri(Uri endpoint, string subscriptionId)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "api-version", UriTemplateEngine.ProviderContractVersion },
            };

            return UriTemplateEngine.SubscriptionUriTemplate.BindByName(
                baseAddress: endpoint,
                parameters: parameters);
        }

        /// <summary>
        /// Gets the management group URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="managementGroupId">The management group identifier.</param>
        public static Uri GetManagementGroupUri(Uri endpoint, string managementGroupId)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "managementGroupId", managementGroupId },
                { "api-version", UriTemplateEngine.ProviderContractVersion },
            };

            return UriTemplateEngine.ManagementGroupUriTemplate.BindByName(
                baseAddress: endpoint,
                parameters: parameters);
        }

        /// <summary>
        /// Gets tenant provider operations uri.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="providerNamespace">The provider namespace.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetTenantProviderOperationsUri(Uri endpoint, string providerNamespace, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "resourceProviderNamespace", providerNamespace },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.TenantProviderOperationsUriTemplate.BindByName(
                baseAddress: endpoint,
                parameters: parameters);
        }

        /// <summary>
        /// Gets the subscription resources URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="apiVersion">The <c>api</c> version.</param>
        public static Uri GetSubscriptionResourcesUri(Uri endpoint, string subscriptionId, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.SubscriptionResourcesUriTemplate.BindByName(
                baseAddress: endpoint,
                parameters: parameters);
        }

        /// <summary>
        /// Gets the subscription resources URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="filter">The filter string.</param>
        /// <param name="apiVersion">The <c>api</c> version.</param>
        public static Uri GetSubscriptionResourcesUriWithFilter(Uri endpoint, string subscriptionId, string filter, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "api-version", apiVersion },
                { "filter", filter },
            };

            return UriTemplateEngine.SubscriptionResourcesUriWithFilterTemplate.BindByName(
                baseAddress: endpoint,
                parameters: parameters);
        }

        /// <summary>
        /// Gets the subscription level role assignment URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="scope">The scope.</param>
        /// <param name="roleAssignmentName">The role assignment name.</param>
        public static Uri GetRoleAssignmentUri(Uri endpoint, string scope, string roleAssignmentName)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "scope", scope },
                { "roleAssignmentName", roleAssignmentName },
                { "api-version", UriTemplateEngine.AuthorizationProviderVersion },
            };

            return UriTemplateEngine.AssignmentUriTemplate.BindByName(
                baseAddress: endpoint,
                parameters: parameters);
        }

        /// <summary>
        /// Gets the subscription level role assignment URI for principal.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="scope">The scope.</param>
        /// <param name="principalId">The principal id.</param>
        public static Uri GetAssignmentForPrincipalUriAtScope(Uri endpoint, string scope, string principalId)
        {
            var principalIdFilter = string.Format("principalId eq '{0}'", principalId);

            var parameters = new Dictionary<string, string>()
            {
                { "scope", scope },
                { "principalIdFilter", principalIdFilter },
                { "api-version", UriTemplateEngine.AuthorizationProviderVersion },
            };

            return UriTemplateEngine.AssignmentsForPrincipalAtScopeUriTemplate.BindByName(
                baseAddress: endpoint,
                parameters: parameters);
        }

        /// <summary>
        /// Gets the subscription level resources by resource provider namespace and type.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="resourceProviderNamespace">The resource provider namespace.</param>
        /// <param name="resourceType">The resource type.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetSubscriptionResourceProviderResourcesUri(Uri endpoint, string subscriptionId, string resourceProviderNamespace, string resourceType, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceProviderNamespace", resourceProviderNamespace },
                { "resourceType", resourceType },
                { "api-version", apiVersion }
            };

            return UriTemplateEngine.SubscriptionResourceProviderResourcesUriTemplate.BindByName(
                baseAddress: endpoint,
                parameters: parameters);
        }

        /// <summary>
        /// Gets the resource group level nested resources by resource provider namespace and type.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="resourceGroupName">The resource group name.</param>
        /// <param name="parentResourceId">The parent resource Id.</param>
        /// <param name="nestedResourceType">The resource type.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetResourceGroupProviderNestedResourcesUri(Uri endpoint, string subscriptionId, string resourceGroupName, string parentResourceId, string nestedResourceType, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceGroupName", resourceGroupName },
                { "parentResourceId", parentResourceId },
                { "nestedResourceType", nestedResourceType },
                { "api-version", apiVersion }
            };

            return UriTemplateEngine.ResourceGroupProviderNestedResourcesUriTemplate.BindByName(
                baseAddress: endpoint,
                parameters: parameters);
        }

        /// <summary>
        /// Gets the resource group level resources by resource provider namespace and type.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="resourceGroupName">The resource group name.</param>
        /// <param name="resourceProviderNamespace">The resource provider namespace.</param>
        /// <param name="resourceType">The resource type.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetResourceGroupProviderResourcesUri(Uri endpoint, string subscriptionId, string resourceGroupName, string resourceProviderNamespace, string resourceType, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceGroupName", resourceGroupName },
                { "resourceProviderNamespace", resourceProviderNamespace },
                { "resourceType", resourceType },
                { "api-version", apiVersion }
            };

            return UriTemplateEngine.ResourceGroupProviderResourcesUriTemplate.BindByName(
                baseAddress: endpoint,
                parameters: parameters);
        }

        /// <summary>
        /// Gets the subscription level resource provider register URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="resourceProviderNamespace">The resource provider namespace.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetSubscriptionResourceProviderRegisterUri(Uri endpoint, string subscriptionId, string resourceProviderNamespace, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceProviderNamespace", resourceProviderNamespace },
                { "api-version", apiVersion }
            };

            return UriTemplateEngine.SubscriptionResourceProviderRegisterUriTemplate.BindByName(
                baseAddress: endpoint,
                parameters: parameters);
        }

        /// <summary>
        /// Gets the resource provider request URI.
        /// </summary>
        /// <param name="originalRequestUri">The original request.</param>
        /// <param name="resourceProviderEndpoint">The resource provider endpoint.</param>
        public static Uri GetResourceProviderRequestUri(Uri originalRequestUri, Uri resourceProviderEndpoint)
        {
            var uri = new Uri(baseUri: resourceProviderEndpoint, relativeUri: originalRequestUri.AbsolutePath.Substring(1));

            var builder = new UriBuilder(
                scheme: uri.Scheme,
                host: uri.Host,
                port: uri.Port,
                path: uri.AbsolutePath,
                extraValue: originalRequestUri.Query);

            return builder.Uri;
        }

        ///// <summary>
        ///// Rewrites the resource provider endpoint URI to use host-based routing.
        ///// </summary>
        ///// <param name="resourceTypeRegistration">The resource type registration.</param>
        ///// <param name="resourceId">The resource Id.</param>
        //public static Uri GetRoutedEndpointUri(ResourceTypeRegistration resourceTypeRegistration, string resourceId)
        //{
        //    var resourceProviderEndpoint = new Uri(resourceTypeRegistration.EndpointUri);

        //    var newHostName = UriTemplateEngine.GetHostBasedName(resourceTypeRegistration, resourceId);

        //    if (!string.IsNullOrEmpty(newHostName))
        //    {
        //        var builder = new UriBuilder(
        //            scheme: resourceProviderEndpoint.Scheme,
        //            host: newHostName,
        //            port: resourceProviderEndpoint.Port,
        //            path: resourceProviderEndpoint.AbsolutePath,
        //            extraValue: resourceProviderEndpoint.Query);

        //        return builder.Uri;
        //    }

        //    return resourceProviderEndpoint;
        //}

        ///// <summary>
        ///// Gets the name of the host based.
        ///// </summary>
        ///// <param name="registration">The registration.</param>
        ///// <param name="resourceId">The resource Id.</param>
        //public static string GetHostBasedName(ResourceTypeRegistration registration, string resourceId)
        //{
        //    if (registration.IsNestedResourceType() && registration.IsHostBasedRoutingEnabled)
        //    {
        //        string parentResourceId = registration.RoutingRule != null && !string.IsNullOrEmpty(registration.RoutingRule.HostResourceType)
        //            ? IResourceIdentifiableExtensions.GetParentResourceId(resourceId, registration.RoutingRule.HostResourceType)
        //            : IResourceIdentifiableExtensions.GetRootResourceId(resourceId);

        //        if (!string.IsNullOrEmpty(parentResourceId))
        //        {
        //            return string.Format(
        //                format: "{0}.{1}",
        //                arg0: IResourceIdentifiableExtensions.GetResourceName(parentResourceId).SplitRemoveEmpty('/').Last(),
        //                arg1: new Uri(registration.EndpointUri).Host);
        //        }
        //    }

        //    return null;
        //}

        /// <summary>
        /// Gets the linked resource provider notification URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="scope">The base scope.</param>
        /// <param name="linkedResourceProvider">The linked resource provider.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetResourceProviderLinkedNotificationUri(Uri endpoint, string scope, string linkedResourceProvider, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "scope", scope.TrimStart('/') },
                { "linkedResourceProvider", linkedResourceProvider },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.ResourceProviderLinkedNotificationUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the resource identity URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="scope">The base scope.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetResourceIdentityUri(Uri endpoint, string scope, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "scope", scope },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.ResourceIdentityUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the event grid system topic URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetEventGridSystemTopicUri(Uri endpoint, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.EventGridSystemTopicUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the event grid custom topic URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetEventGridCustomTopicUri(Uri endpoint, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.EventGridCustomTopicUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the azure async operations URI for a deployment.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="managementGroupId">The management group id.</param>
        /// <param name="subscriptionId">The subscription Id.</param>
        /// <param name="resourceGroupName">The resource group name.</param>
        /// <param name="deploymentName">The deployment name.</param>
        /// <param name="deploymentSequence">The deployment sequence.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetDeploymentAzureAsyncOperationsUri(
            Uri endpoint,
            string managementGroupId,
            string subscriptionId,
            string resourceGroupName,
            string deploymentName,
            string deploymentSequence,
            string apiVersion)
        {
            if (!string.IsNullOrEmpty(resourceGroupName))
            {
                return UriTemplateEngine.GetResourceGroupDeploymentAzureAsyncOperationsUri(
                    endpoint: endpoint,
                    subscriptionId: subscriptionId,
                    resourceGroupName: resourceGroupName,
                    deploymentName: deploymentName,
                    deploymentSequence: deploymentSequence,
                    apiVersion: apiVersion);
            }

            if (!string.IsNullOrEmpty(subscriptionId))
            {
                return UriTemplateEngine.GetSubscriptionDeploymentAzureAsyncOperationsUri(
                    endpoint: endpoint,
                    subscriptionId: subscriptionId,
                    deploymentName: deploymentName,
                    deploymentSequence: deploymentSequence,
                    apiVersion: apiVersion);
            }

            if (!string.IsNullOrEmpty(managementGroupId))
            {
                return UriTemplateEngine.GetManagementGroupDeploymentAzureAsyncOperationsUri(
                    endpoint: endpoint,
                    managementGroupId: managementGroupId,
                    deploymentName: deploymentName,
                    deploymentSequence: deploymentSequence,
                    apiVersion: apiVersion);
            }

            return UriTemplateEngine.GetTenantDeploymentAzureAsyncOperationsUri(
                endpoint: endpoint,
                deploymentName: deploymentName,
                deploymentSequence: deploymentSequence,
                apiVersion: apiVersion);
        }

        /// <summary>
        /// Gets the policy cleanup URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="scope">The scope.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetPolicyCleanupUri(Uri endpoint, string scope, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "scope", scope },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.PolicyCleanupUriTemplate.BindByName(
                baseAddress: endpoint,
                parameters: parameters);
        }

        /// <summary>
        /// Gets the trigger policy driven deployment URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="resourceGroupName">Name of the resource group.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetTriggerPolicyDeploymentUri(Uri endpoint, string subscriptionId, string resourceGroupName, string apiVersion)
        {
            if (string.IsNullOrWhiteSpace(resourceGroupName))
            {
                var subscriptionParameters = new Dictionary<string, string>()
                {
                    { "subscriptionId", subscriptionId },
                    { "api-version", apiVersion },
                };

                return UriTemplateEngine.TriggerPolicySubscriptionDeploymentUriTemplate.BindByName(baseAddress: endpoint, parameters: subscriptionParameters);
            }

            var resourceGroupParameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceGroupName", resourceGroupName },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.TriggerPolicyDeploymentUriTemplate.BindByName(baseAddress: endpoint, parameters: resourceGroupParameters);
        }

        /// <summary>
        /// Gets the URI to perform a policy pass-through PATCH operation.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription Id.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetPolicyPassthroughPatchOperationUri(Uri endpoint, string subscriptionId, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.PolicyPassthroughPatchOperationUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the URI to perform a policy pass-through PUT operation.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription Id.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetPolicyPassthroughPutOperationUri(Uri endpoint, string subscriptionId, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.PolicyPassthroughPutOperationUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the azure async operations URI for a tenant deployment.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="deploymentName">The deployment name.</param>
        /// <param name="deploymentSequence">The deployment sequence.</param>
        /// <param name="apiVersion">The API version.</param>
        private static Uri GetTenantDeploymentAzureAsyncOperationsUri(Uri endpoint, string deploymentName, string deploymentSequence, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "deploymentName", deploymentName },
                { "deploymentSequence", deploymentSequence },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.TenantDeploymentAzureAsyncOperationsUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the azure async operations URI for a management group deployment.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="managementGroupId">The management group id.</param>
        /// <param name="deploymentName">The deployment name.</param>
        /// <param name="deploymentSequence">The deployment sequence.</param>
        /// <param name="apiVersion">The API version.</param>
        private static Uri GetManagementGroupDeploymentAzureAsyncOperationsUri(Uri endpoint, string managementGroupId, string deploymentName, string deploymentSequence, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "managementGroupId", managementGroupId },
                { "deploymentName", deploymentName },
                { "deploymentSequence", deploymentSequence },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.ManagementGroupDeploymentAzureAsyncOperationsUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the azure async operations URI for a deployment at subscription scope.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription Id.</param>
        /// <param name="deploymentName">The deployment name.</param>
        /// <param name="deploymentSequence">The deployment sequence.</param>
        /// <param name="apiVersion">The API version.</param>
        private static Uri GetSubscriptionDeploymentAzureAsyncOperationsUri(Uri endpoint, string subscriptionId, string deploymentName, string deploymentSequence, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "deploymentName", deploymentName },
                { "deploymentSequence", deploymentSequence },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.SubscriptionDeploymentAzureAsyncOperationsUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the azure async operations URI for a deployment at resource group scope.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription Id.</param>
        /// <param name="resourceGroupName">The resource group name.</param>
        /// <param name="deploymentName">The deployment name.</param>
        /// <param name="deploymentSequence">The deployment sequence.</param>
        /// <param name="apiVersion">The API version.</param>
        private static Uri GetResourceGroupDeploymentAzureAsyncOperationsUri(Uri endpoint, string subscriptionId, string resourceGroupName, string deploymentName, string deploymentSequence, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceGroupName", resourceGroupName },
                { "deploymentName", deploymentName },
                { "deploymentSequence", deploymentSequence },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.ResourceGroupDeploymentAzureAsyncOperationsUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the template deployment preflight URI for resources in resource group.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="resourceGroupName">The resource group name.</param>
        /// <param name="resourceProviderNamespace">The resource provider namespace.</param>
        /// <param name="deploymentName">Name of the deployment.</param>
        /// <param name="apiVersion">The API version.</param>
        private static Uri GetResourceGroupTemplateDeploymentPreflightUri(Uri endpoint, string subscriptionId, string resourceGroupName, string resourceProviderNamespace, string deploymentName, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceGroupName", resourceGroupName },
                { "resourceProviderNamespace", resourceProviderNamespace },
                { "deploymentName", deploymentName },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.TemplateDeploymentPreflightUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the template deployment preflight URI for subscription level resources.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="resourceProviderNamespace">The resource provider namespace.</param>
        /// <param name="deploymentName">Name of the deployment.</param>
        /// <param name="apiVersion">The API version.</param>
        private static Uri GetSubscriptionTemplateDeploymentPreflightUri(Uri endpoint, string subscriptionId, string resourceProviderNamespace, string deploymentName, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceProviderNamespace", resourceProviderNamespace },
                { "deploymentName", deploymentName },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.TemplateDeploymentSubscriptionPreflightUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the template deployment preflight URI for management group level resources.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="managementGroupId">The management group id.</param>
        /// <param name="resourceProviderNamespace">The resource provider namespace.</param>
        /// <param name="deploymentName">Name of the deployment.</param>
        /// <param name="apiVersion">The API version.</param>
        private static Uri GetManagementGroupTemplateDeploymentPreflightUri(Uri endpoint, string managementGroupId, string resourceProviderNamespace, string deploymentName, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "managementGroupId", managementGroupId },
                { "resourceProviderNamespace", resourceProviderNamespace },
                { "deploymentName", deploymentName },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.TemplateDeploymentManagementGroupPreflightUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the template deployment preflight URI for tenant level resources.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="resourceProviderNamespace">The resource provider namespace.</param>
        /// <param name="deploymentName">Name of the deployment.</param>
        /// <param name="apiVersion">The API version.</param>
        private static Uri GetTenantTemplateDeploymentPreflightUri(Uri endpoint, string resourceProviderNamespace, string deploymentName, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "resourceProviderNamespace", resourceProviderNamespace },
                { "deploymentName", deploymentName },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.TemplateDeploymentTenantPreflightUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the template deployment validate URI for tenant level resources.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="deploymentName">Name of the deployment.</param>
        /// <param name="apiVersion">The API version.</param>
        private static Uri GetTemplateDeploymentTenantValidateUri(Uri endpoint, string deploymentName, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "deploymentName", deploymentName },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.TemplateDeploymentTenantValidateUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the template deployment validate URI for management group level resources.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="managementGroupId">The management group id.</param>
        /// <param name="deploymentName">Name of the deployment.</param>
        /// <param name="apiVersion">The API version.</param>
        private static Uri GetTemplateDeploymentManagementGroupValidateUri(Uri endpoint, string managementGroupId, string deploymentName, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "managementGroupId", managementGroupId },
                { "deploymentName", deploymentName },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.TemplateDeploymentManagementGroupValidateUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the template deployment validate URI for subscription level resources.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="deploymentName">Name of the deployment.</param>
        /// <param name="apiVersion">The API version.</param>
        private static Uri GetTemplateDeploymentSubscriptionValidateUri(Uri endpoint, string subscriptionId, string deploymentName, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "deploymentName", deploymentName },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.TemplateDeploymentSubscriptionValidateUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the template deployment resource group validate URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="resourceGroupName">Name of the resource group.</param>
        /// <param name="deploymentName">Name of the deployment.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetTemplateDeploymentResourceGroupValidateUri(Uri endpoint, string subscriptionId, string resourceGroupName, string deploymentName, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceGroupName", resourceGroupName },
                { "deploymentName", deploymentName },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.TemplateDeploymentResourceGroupValidateUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the template deployment preflight URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="managementGroupId">The management group id.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="resourceGroupName">Name of the resource group.</param>
        /// <param name="resourceProviderNamespace">The resource provider namespace.</param>
        /// <param name="deploymentName">Name of the deployment.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetTemplateDeploymentPreflightUri(
            Uri endpoint,
            string managementGroupId,
            string subscriptionId,
            string resourceGroupName,
            string resourceProviderNamespace,
            string deploymentName,
            string apiVersion)
        {
            if (!string.IsNullOrEmpty(resourceGroupName))
            {
                return UriTemplateEngine.GetResourceGroupTemplateDeploymentPreflightUri(
                    endpoint: endpoint,
                    subscriptionId: subscriptionId,
                    resourceGroupName: resourceGroupName,
                    resourceProviderNamespace: resourceProviderNamespace,
                    deploymentName: deploymentName,
                    apiVersion: apiVersion);
            }

            if (!string.IsNullOrEmpty(subscriptionId))
            {
                return UriTemplateEngine.GetSubscriptionTemplateDeploymentPreflightUri(
                    endpoint: endpoint,
                    subscriptionId: subscriptionId,
                    resourceProviderNamespace: resourceProviderNamespace,
                    deploymentName: deploymentName,
                    apiVersion: apiVersion);
            }

            if (!string.IsNullOrEmpty(managementGroupId))
            {
                return UriTemplateEngine.GetManagementGroupTemplateDeploymentPreflightUri(
                    endpoint: endpoint,
                    managementGroupId: managementGroupId,
                    resourceProviderNamespace: resourceProviderNamespace,
                    deploymentName: deploymentName,
                    apiVersion: apiVersion);
            }

            return UriTemplateEngine.GetTenantTemplateDeploymentPreflightUri(
                endpoint: endpoint,
                resourceProviderNamespace: resourceProviderNamespace,
                deploymentName: deploymentName,
                apiVersion: apiVersion);
        }

        /// <summary>
        /// Gets the template deployment validate URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="managementGroupId">The management group id.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="resourceGroupName">Name of the resource group.</param>
        /// <param name="deploymentName">Name of the deployment.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetTemplateDeploymentValidateUri(
            Uri endpoint,
            string managementGroupId,
            string subscriptionId,
            string resourceGroupName,
            string deploymentName,
            string apiVersion)
        {
            if (!string.IsNullOrEmpty(resourceGroupName))
            {
                return UriTemplateEngine.GetTemplateDeploymentResourceGroupValidateUri(
                    endpoint: endpoint,
                    subscriptionId: subscriptionId,
                    resourceGroupName: resourceGroupName,
                    deploymentName: deploymentName,
                    apiVersion: apiVersion);
            }

            if (!string.IsNullOrEmpty(subscriptionId))
            {
                return UriTemplateEngine.GetTemplateDeploymentSubscriptionValidateUri(
                    endpoint: endpoint,
                    subscriptionId: subscriptionId,
                    deploymentName: deploymentName,
                    apiVersion: apiVersion);
            }

            if (!string.IsNullOrEmpty(managementGroupId))
            {
                return UriTemplateEngine.GetTemplateDeploymentManagementGroupValidateUri(
                    endpoint: endpoint,
                    managementGroupId: managementGroupId,
                    deploymentName: deploymentName,
                    apiVersion: apiVersion);
            }

            return UriTemplateEngine.GetTemplateDeploymentTenantValidateUri(
                endpoint: endpoint,
                deploymentName: deploymentName,
                apiVersion: apiVersion);
        }

        /// <summary>
        /// Gets the template redeploy URI.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="subscriptionId">The subscription identifier.</param>
        /// <param name="resourceGroupName">Name of the resource group.</param>
        /// <param name="deploymentName">Name of the deployment.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetTemplateDeploymentRedeployUri(Uri endpoint, string subscriptionId, string resourceGroupName, string deploymentName, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "subscriptionId", subscriptionId },
                { "resourceGroupName", resourceGroupName },
                { "deploymentName", deploymentName },
                { "api-version", apiVersion },
            };

            return UriTemplateEngine.TemplateDeploymentRedeployUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the async batch results URI
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="batchOperationId">The async batch operation id.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetAsyncBatchResultsUri(Uri endpoint, string batchOperationId, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "batchOperationId", batchOperationId },
                { "apiVersion", apiVersion }
            };

            return UriTemplateEngine.AsyncBatchOperationResultsUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        ///// <summary>
        ///// Gets the batch URI
        ///// </summary>
        ///// <param name="resourceTypeRegistration">The resource type registration.</param>
        //public static Uri GetResourceProviderBatchUri(ResourceTypeRegistration resourceTypeRegistration)
        //{
        //    if (resourceTypeRegistration?.ResourceManagementOptions?.BatchProvisioningSupport?.SupportedOperations != BatchSupportedOperations.NotSpecified)
        //    {
        //        var resourceProviderEndpoint = new Uri(resourceTypeRegistration.EndpointUri);
        //        var parameters = new Dictionary<string, string>()
        //        {
        //            { "apiVersion", UriTemplateEngine.ProviderContractVersion }
        //        };

        //        return UriTemplateEngine.BatchUriTemplate.BindByName(baseAddress: resourceProviderEndpoint, parameters: parameters);
        //    }

        //    return null;
        //}

        /// <summary>
        /// Gets the async bulk results URI
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="bulkDeletionOperationId">The bulk deletion operation id.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri GetAsyncBulkDeletionResultsUri(Uri endpoint, string bulkDeletionOperationId, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "bulkDeletionOperationId", bulkDeletionOperationId },
                { "apiVersion", apiVersion }
            };

            return UriTemplateEngine.AsyncBulkDeletionOperationResultsUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Creates the async notification uri that will be sent as part of the requests to providers.
        /// </summary>
        /// <param name="endpoint">The endpoint.</param>
        /// <param name="asyncOperationCallbackTokenData">The notification token data.</param>
        /// <param name="apiVersion">The API version.</param>
        public static Uri CreateAsyncOperationCallbackUri(Uri endpoint, string asyncOperationCallbackTokenData, string apiVersion)
        {
            var parameters = new Dictionary<string, string>()
            {
                { "api-Version", apiVersion },
                { "asyncNotificationToken", asyncOperationCallbackTokenData }
            };

            return UriTemplateEngine.AsyncOperationCallbackUriTemplate.BindByName(baseAddress: endpoint, parameters: parameters);
        }

        /// <summary>
        /// Gets the ARM preflight validation uri.
        /// </summary>
        /// <param name="originalRequestUri">The original request.</param>
        /// <param name="subscriptionId">The subscription id.</param>
        /// <param name="resourceGroupName">The resource group name.</param>
        /// <param name="fullyQualifiedNestedResourceType">Nested resource type.</param>
        /// <param name="fullyQualifiedNestedResourceName">Nested resource name.</param>
        /// <param name="apiVersion">API version</param>
        /// <param name="additionalQueryParameterCollection">Additional query parameters.</param>
        public static Uri GetArmPreflightValidationUri(Uri originalRequestUri, string subscriptionId, string resourceGroupName, string fullyQualifiedNestedResourceType, string fullyQualifiedNestedResourceName, string apiVersion, NameValueCollection additionalQueryParameterCollection = null)
        {
            var resourceId = IResourceIdentifiableExtensions.GetRootResourceId(fullyQualifiedResourceType: fullyQualifiedNestedResourceType, resourceName: fullyQualifiedNestedResourceName);
            var absolutePath = IResourceIdentifiableExtensions.GetFullyQualifiedResourceId(subscriptionId: subscriptionId, resourceGroupName: resourceGroupName, resourceId: resourceId);

            var builder = new UriBuilder(
                scheme: originalRequestUri.Scheme,
                host: originalRequestUri.Host,
                port: originalRequestUri.Port,
                pathValue: absolutePath);

            var queryParametersCollection = originalRequestUri.ParseQueryString();
            queryParametersCollection.Add("ARMPreflightValidation", bool.TrueString);
            if (additionalQueryParameterCollection != null)
            {
                queryParametersCollection.Add(additionalQueryParameterCollection);
            }

            builder.Query = queryParametersCollection.ToString();

            return builder.Uri;
        }

        /// <summary>
        /// Gets the resource uri.
        /// </summary>
        /// <param name="originalRequestUri">The original request.</param>
        /// <param name="subscriptionId">The subscription id.</param>
        /// <param name="resourceGroupName">The resource group name.</param>
        /// <param name="fullyQualifiedNestedResourceType">Nested resource type.</param>
        /// <param name="fullyQualifiedNestedResourceName">Nested resource name.</param>
        /// <param name="apiVersion">API version</param>
        public static Uri GetResourceUri(Uri originalRequestUri, string subscriptionId, string resourceGroupName, string fullyQualifiedNestedResourceType, string fullyQualifiedNestedResourceName, string apiVersion)
        {
            var resourceId = IResourceIdentifiableExtensions.GetRootResourceId(fullyQualifiedResourceType: fullyQualifiedNestedResourceType, resourceName: fullyQualifiedNestedResourceName);
            var absolutePath = IResourceIdentifiableExtensions.GetFullyQualifiedResourceId(subscriptionId: subscriptionId, resourceGroupName: resourceGroupName, resourceId: resourceId);

            var builder = new UriBuilder(
                scheme: originalRequestUri.Scheme,
                host: originalRequestUri.Host,
                port: originalRequestUri.Port,
                pathValue: absolutePath);

            var queryParametersCollection = originalRequestUri.ParseQueryString();
            queryParametersCollection.Add(RequestCorrelationContext.ParameterApiVersion, apiVersion);
            builder.Query = queryParametersCollection.ToString();

            return builder.Uri;
        }

        /// <summary>
        /// Wheter request is get subscrition request
        /// </summary>
        /// <param name="request">The request.</param>
        public static bool IsGetSubscriptionRequest(HttpRequestMessage request)
        {
            return UriTemplateEngine.SubscriptionUriTemplate.Match(new Uri(request.RequestUri.GetLeftPart(UriPartial.Authority)), request.RequestUri) != null;
        }
    }
}
