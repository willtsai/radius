using Microsoft.WindowsAzure.ResourceStack.Common.Instrumentation;
using Newtonsoft.Json;

namespace DeploymentEngine.Jobs
{
    /// <summary>
    /// The front door base job metadata.
    /// </summary>
    public class FrontdoorJobMetadata
    {
        /// <summary>
        /// Gets or sets the request correlation context.
        /// </summary>
        [JsonProperty]
        public RequestCorrelationContext RequestCorrelationContext { get; set; }

        /// <summary>
        /// Gets or sets the front door role location.
        /// </summary>
        [JsonProperty]
        public string FrontdoorLocation { get; set; }

        /// <summary>
        /// Gets or sets the front door endpoint.
        /// </summary>
        [JsonProperty]
        public Uri FrontdoorEndpoint { get; set; }

        /// <summary>
        /// Gets or sets the front door job maximum life time.
        /// </summary>
        public TimeSpan? FrontdoorJobMaxLifetime { get; set; }
    }
}
