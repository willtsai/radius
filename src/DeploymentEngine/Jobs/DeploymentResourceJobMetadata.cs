using Azure.Deployments.Core.Definitions;
using Azure.Deployments.Core.Entities;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;
using System.Net;

namespace DeploymentEngine.Jobs
{
    // TODO populate job metadata here with all necessary info for a deployment of a resource.
    public class DeploymentResourceJobMetadata : DeploymentJobMetadata
    {
        /// <summary>
        /// Gets or sets the resource.
        /// </summary>
        [JsonProperty(Required = Required.Always)]
        public DeploymentResource Resource { get; set; }

        /// <summary>
        /// Gets or sets the resource operation.
        /// </summary>
        [JsonProperty(Required = Required.Always)]
        public ProvisioningOperation ResourceOperation { get; set; }

        /// <summary>
        /// Gets or sets the resource operation URI.
        /// </summary>
        [JsonProperty(Required = Required.Always)]
        public Uri ResourceOperationUri { get; set; }

        /// <summary>
        /// Gets or sets the resource operation status code.
        /// </summary>
        [JsonProperty(Required = Required.Default)]
        public HttpStatusCode? ResourceOperationStatusCode { get; set; }

        /// <summary>
        /// Gets or sets the resource operation status message.
        /// </summary>
        [JsonProperty(Required = Required.Default)]
        public JToken ResourceOperationStatusMessage { get; set; }

        /// <summary>
        /// Gets or sets the azure async operation uri for this resource.
        /// </summary>
        [JsonProperty(Required = Required.Default)]
        public Uri ResourceAsyncOperationUri { get; set; }

        /// <summary>
        /// Gets or sets the content of deployment operation request.
        /// </summary>
        [JsonProperty(Required = Required.Default)]
        public JToken ResourceOperationRequestContent { get; set; }

        /// <summary>
        /// Gets or sets the resource cache wait timeout.
        /// </summary>
        [JsonProperty(Required = Required.Default)]
        public DateTime? ResourceCacheWaitTimeout { get; set; }

        /// <summary>
        /// Gets or sets the debug setting.
        /// </summary>
        [JsonProperty(Required = Required.Default)]
        public DeploymentDebugSetting DebugSetting { get; set; }

        /// <summary>
        /// Gets or sets the resource operation service request Id.
        /// </summary>
        [JsonProperty(Required = Required.Default)]
        public string ServiceRequestId { get; set; }

        /// <summary>
        /// Gets or sets the deployment operation request.
        /// </summary>
        [JsonProperty(Required = Required.Default)]
        public DeploymentOperationHttpMessage ResourceOperationRequest { get; set; }

        /// <summary>
        /// Gets or sets the deployment operation response.
        /// </summary>
        [JsonProperty(Required = Required.Default)]
        public DeploymentOperationHttpMessage ResourceOperationResponse { get; set; }

        /// <summary>
        /// Gets or sets the async notification status.
        /// </summary>
        [JsonProperty(Required = Required.Default)]
        public bool? IsAsyncNotificationEnabled { get; set; }

        /// <summary>
        /// Get or set whether or not references have already been resolved
        /// </summary>
        [JsonProperty(Required = Required.Default)]
        public bool? AreTemplateExpressionsEvaluated { get; set; }

        /// <summary>
        /// Get or set the inline wait attempts on resource cache
        /// </summary>
        [JsonProperty(Required = Required.Default)]
        public int? ResourceCacheInlineWaitAttempts { get; set; }

        /// <summary>
        /// Gets or sets the async operation timeout.
        /// </summary>
        public TimeSpan? AsyncOperationTimeout { get; set; }
    }
}
