//-----------------------------------------------------------
// Copyright (c) Microsoft Corporation.  All rights reserved.
//-----------------------------------------------------------

namespace Microsoft.WindowsAzure.ResourceStack.Frontdoor.Data.Extensions
{
    using System.Net;
    using Azure.Deployments.Core.Definitions;
    using Azure.Deployments.Core.Definitions.Resources;
    using Microsoft.WindowsAzure.ResourceStack.Common.Extensions;
    using Microsoft.WindowsAzure.ResourceStack.Common.Instrumentation;
    using Microsoft.WindowsAzure.ResourceStack.Common.Json;
    using Microsoft.WindowsAzure.ResourceStack.Common.Utilities;
    using Newtonsoft.Json.Linq;
    using DeploymentsResourceProxyDefinition = global::Azure.Deployments.Core.Definitions.Resources.ResourceProxyDefinition;

    /// <summary>
    /// The resource proxy definition extension helper.
    /// </summary>
    public static class ResourceProxyDefinitionExtensions
    {
        /// <summary>
        /// Gets the provisioning state.
        /// </summary>
        /// <param name="resourceDefinition">The resource definition.</param>
        public static ProvisioningState GetResourceProvisioningState(this ResourceProxyDefinition resourceDefinition)
        {
            if (resourceDefinition == null || resourceDefinition.Properties == null || resourceDefinition.Properties.Type != JTokenType.Object)
            {
                return ProvisioningState.NotSpecified;
            }

            var provisioningStateProperty = resourceDefinition.Properties.GetProperty("provisioningState");
            if (provisioningStateProperty == null || provisioningStateProperty.Type != JTokenType.String)
            {
                return ProvisioningState.NotSpecified;
            }

            var provisioningStateAsString = provisioningStateProperty.ToObject<string>();
            return provisioningStateAsString.ParseWithDefault<ProvisioningState>(ProvisioningState.Running);
        }

        /// <summary>
        /// Sets the fields derived from other places like uri
        /// </summary>
        /// <param name="definition">The resource definition.</param>
        /// <param name="resourceName">The resource name.</param>
        /// <param name="resourceType">The resource type.</param>
        /// <param name="subscriptionId">The subscription Id.</param>
        /// <param name="resourceGroupName">The resource group name.</param>
        public static void SetAndNormalizeProperties(this ResourceProxyDefinition definition, string resourceName, string resourceType, string subscriptionId, string resourceGroupName)
        {
            definition.SetDerivedProperties(
                subscriptionId: subscriptionId,
                resourceGroupName: resourceGroupName);

            definition.Name = definition.Name ?? resourceName;
            definition.Type = definition.Type ?? resourceType;

            definition.NormalizeNestedResourceTypeAndNameRecursive();
        }

        /// <summary>
        /// Sets the fields derived from other places like uri
        /// </summary>
        /// <param name="resource">The resource.</param>
        /// <param name="subscriptionId">The subscription Id.</param>
        /// <param name="resourceGroupName">The resource group name.</param>
        public static void SetDerivedProperties(this ResourceProxyDefinition resource, string subscriptionId, string resourceGroupName)
        {
            resource.SubscriptionId = subscriptionId;
            resource.ResourceGroup = resourceGroupName;
            resource.Resources
                .CoalesceEnumerable()
                .ForEach(nestedResource => nestedResource.SetDerivedProperties(subscriptionId: subscriptionId, resourceGroupName: resourceGroupName));
        }

        /// <summary>
        /// Normalizes resource type and name, in recursive manner.
        /// The 'definition' type must have fully qualified type and name.
        /// ex: Type: PROVIDERNAMESPACE/PARENTRESOURCETYPE/NESTEDRESOURCETYPE1/NESTEDRESOURCETYPE2
        ///     Name: PARENTNAME/NESTEDNAME1/NESTEDNAME2
        /// </summary>
        /// <param name="definition">The resource definition.</param>
        public static void NormalizeNestedResourceTypeAndNameRecursive(this ResourceProxyDefinition definition)
        {
            definition.Resources?.ForEach(nestedResource =>
            {
                definition.NormalizeNestedResourceTypeAndName(nestedResource: nestedResource);
                nestedResource.NormalizeNestedResourceTypeAndNameRecursive();
            });
        }

        /// <summary>
        /// Normalizes (fully qualify) resource type and name
        /// ex: Type: PROVIDERNAMESPACE/PARENTRESOURCETYPE/NESTEDRESOURCETYPE1/NESTEDRESOURCETYPE2
        ///     Name: PARENTNAME/NESTEDNAME1/NESTEDNAME2
        /// </summary>
        /// <param name="definition">The resource definition.</param>
        /// <param name="nestedResource">The nested resource being processed (could be nested/child resource).</param>
        private static void NormalizeNestedResourceTypeAndName(this ResourceProxyDefinition definition, ResourceProxyDefinition nestedResource)
        {
            var typeSegmentsLength = HttpUtility.GetPathSegments(nestedResource.Type).Length;
            var nameSegmentsLength = HttpUtility.GetPathSegments(nestedResource.Name).Length;

            // The path is already normalized
            if (typeSegmentsLength == nameSegmentsLength + 1)
            {
                return;
            }

            if (typeSegmentsLength == nameSegmentsLength)
            {
                nestedResource.Type = $"{definition.Type}/{nestedResource.Type}";
                nestedResource.Name = $"{definition.Name}/{nestedResource.Name}";
                return;
            }

            throw new Exception();
        }

        /// <summary>
        /// Converts <see cref="global::Azure.Deployments.Core.Definitions.Resources.ResourceProxyDefinition"/> to <see cref="ResourceProxyDefinition"/>
        /// </summary>
        /// <param name="resourceDefinition">The deployments resource proxy definition to convert.</param>
        public static ResourceProxyDefinition ToArmResourceProxyDefinition(this DeploymentsResourceProxyDefinition resourceDefinition)
        {
            if (resourceDefinition == null)
            {
                return default;
            }

            return new ResourceProxyDefinition
            {
                ApiVersion = resourceDefinition.ApiVersion,
                ETag = resourceDefinition.ETag,
                //ExtendedLocation = resourceDefinition.ExtendedLocation.ToArmResourceExtendedLocation(),
                Id = resourceDefinition.Id,
                //Identity = resourceDefinition.Identity.ToArmResourceIdentity(),
                Kind = resourceDefinition.Kind,
                Location = resourceDefinition.Location,
                ManagedBy = resourceDefinition.ManagedBy,
                ManagedByExtended = resourceDefinition.ManagedByExtended,
                ManagementGroupId = resourceDefinition.ManagementGroupId,
                Name = resourceDefinition.Name,
                //Plan = resourceDefinition.Plan.ToArmResourcePlan(),
                Properties = resourceDefinition.Properties,
                ResourceGroup = resourceDefinition.ResourceGroup,
                Resources = resourceDefinition.Resources?.SelectArray(resource => resource.ToArmResourceProxyDefinition()),
                //Scale = resourceDefinition.Scale.ToArmResourceScale(),
                //Sku = resourceDefinition.Sku.ToArmSku(),
                SubscriptionId = resourceDefinition.SubscriptionId,
                //SystemData = resourceDefinition.SystemData.ToArmSystemData(),
                //Tags = resourceDefinition.Tags.ToArmTagsDictionary(),
                Type = resourceDefinition.Type,
                Zones = resourceDefinition.Zones
            };
        }
    }
}
