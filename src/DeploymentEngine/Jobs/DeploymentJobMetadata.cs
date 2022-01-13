using Azure.Deployments.Core.Algorithms;
using Azure.Deployments.Core.Entities;
using Azure.Deployments.Core.Extensions;
using Azure.Deployments.Core.Resources;
using Azure.Deployments.Core.Storage.Helpers;
using Newtonsoft.Json;

namespace DeploymentEngine.Jobs
{
    public class DeploymentJobMetadata : FrontdoorJobMetadata
    {
        /// <summary>
        /// The sequencer job id storage key limit.
        /// </summary>
        private const int SequencerIdStorageKeyLimit = 64;

        /// <summary>
        /// Gets or sets the tenant id.
        /// </summary>
        [JsonProperty]
        public string TenantId { get; set; }

        /// <summary>
        /// Gets or sets the management group id.
        /// </summary>
        [JsonProperty]
        public string ManagementGroupId { get; set; }

        /// <summary>
        /// Gets or sets the subscription id.
        /// </summary>
        [JsonProperty]
        public string SubscriptionId { get; set; }

        /// <summary>
        /// Gets or sets the name of the resource group.
        /// </summary>
        [JsonProperty]
        public string ResourceGroupName { get; set; }

        /// <summary>
        /// Gets or sets the name of the deployment.
        /// </summary>
        [JsonProperty]
        public string DeploymentName { get; set; }

        /// <summary>
        /// Gets or sets the deployment sequence identifier.
        /// </summary>
        [JsonProperty]
        public string SequenceId { get; set; }

        /// <summary>
        /// Gets or sets the deployment template hash value.
        /// </summary>
        [JsonProperty]
        public string TemplateHash { get; set; }

        /// <summary>
        /// Gets or sets the resource group location.
        /// </summary>
        [JsonProperty]
        public string ResourceGroupLocation { get; set; }

        /// <summary>
        /// Gets or sets the deployment location.
        /// </summary>
        [JsonProperty]
        public string DeploymentLocation { get; set; }

        /// <summary>
        /// Gets or sets the count of unsuccessful and throttled request(without a retryAfter header).
        /// </summary>
        [JsonProperty]
        public int UnsuccessAndThrottleCounter { get; set; }

        /// <summary>
        /// Gets a value indicating whether the job is for a tenant deployment.
        /// </summary>
        public bool IsTenantDeployment => string.IsNullOrEmpty(this.SubscriptionId);

        /// <summary>
        /// Gets the deployment location.
        /// </summary>
        public string GetDeploymentLocation()
        {
            return this.DeploymentLocation ?? this.ResourceGroupLocation;
        }

        /// <summary>
        /// Gets the sequencer partition.
        /// </summary>
        /// <param name="subscriptionId">The subscription identifier.</param>
        public static string GetSequencerPartition(string subscriptionId)
        {
            return StorageUtility.EscapeAndTrimSubscriptionId(subscriptionId);
        }

        /// <summary>
        /// Gets the sequencer partition.
        /// </summary>
        /// <param name="deployment">The deployment.</param>
        public static string GetSequencerPartition(IDeploymentEntity deployment)
        {
            return string.IsNullOrEmpty(deployment.SubscriptionId) ?
                StorageUtility.EscapeAndTrimTenantId(deployment.TenantId) :
                StorageUtility.EscapeAndTrimSubscriptionId(deployment.SubscriptionId);
        }

        /// <summary>
        /// Gets the sequencer identifier.
        /// </summary>
        /// <param name="resourceGroupName">Name of the resource group.</param>
        /// <param name="deploymentName">Name of the deployment.</param>
        /// <param name="sequenceId">The sequence identifier.</param>
        public static string GetSequencerId(string resourceGroupName, string deploymentName, string sequenceId)
        {
            var uniqueSequencerId = StorageUtility.CombineStorageKeys(
                    "DeploymentJob",
                    StorageUtility.EscapeStorageKey(resourceGroupName.CoalesceString().ToUpperInvariant()),
                    StorageUtility.EscapeStorageKey(deploymentName.ToUpperInvariant()),
                    sequenceId);

            return StorageUtility.EscapeAndTrimStorageKey(uniqueSequencerId, DeploymentJobMetadata.SequencerIdStorageKeyLimit);
        }

        /// <summary>
        /// Gets the sequencer identifier.
        /// </summary>
        /// <param name="deployment">Name of the deployment.</param>
        public static string GetSequencerId(IDeploymentEntity deployment)
        {
            var uniqueSequencerId = deployment.IsTenantDeployment
                ? StorageUtility.CombineStorageKeys(
                    "DeploymentJob",
                    StorageUtility.EscapeStorageKey(deployment.ManagementGroupId.CoalesceString().ToUpperInvariant()),
                    StorageUtility.EscapeStorageKey(deployment.DeploymentName.ToUpperInvariant()),
                    deployment.SequenceId)
                : StorageUtility.CombineStorageKeys(
                    "DeploymentJob",
                    StorageUtility.EscapeStorageKey(deployment.ResourceGroupName.CoalesceString().ToUpperInvariant()),
                    StorageUtility.EscapeStorageKey(deployment.DeploymentName.ToUpperInvariant()),
                    deployment.SequenceId);

            return StorageUtility.EscapeAndTrimStorageKey(uniqueSequencerId, DeploymentJobMetadata.SequencerIdStorageKeyLimit);
        }

        /// <summary>
        /// Gets the sequencer operation identifier.
        /// </summary>
        /// <param name="reference">The resource reference.</param>
        public static string GetSequencerOperationId(DeploymentResourceReference reference)
        {
            var resourceGroupNameSegment = !string.IsNullOrEmpty(reference.ResourceGroupName)
                ? string.Concat("/", reference.ResourceGroupName)
                : string.Empty;

            var resourceIdSegment = string.Concat("/", reference.GetUnqualifiedResourceId());

            var referenceActionSegment = !reference.IsTemplateResource
                ? string.Concat("/", reference.ReferenceAction.CoalesceString(), "?", reference.ReferenceApiVersion.CoalesceString(), reference.GetReferenceRequestContentInJson())
                : string.Empty;

            // NOTE(ilygre): keeping this version of 'qualifiedResourceId' for template resources for backward compatibility.
            var qualifiedResourceId = string.IsNullOrEmpty(reference.SubscriptionId)
                ? string.Concat(resourceIdSegment, referenceActionSegment)
                : string.Concat(reference.SubscriptionId, resourceGroupNameSegment, resourceIdSegment, referenceActionSegment);

            return ComputeHash.MurmurHash64(qualifiedResourceId.ToUpperInvariant()).ToString("X16");
        }
    }
}
