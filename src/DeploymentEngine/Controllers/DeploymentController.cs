using System.Text;
using Azure.Deployments.Core.Collections;
using Azure.Deployments.Core.Definitions;
using Azure.Deployments.Core.Helpers;
using Azure.Deployments.Core.Json;
using Azure.Deployments.Engine.Dependencies;
using Azure.Deployments.Engine.Helpers;
using Microsoft.AspNetCore.Mvc;
using Newtonsoft.Json.Linq;
using Azure.Deployments.Core.Definitions.Extensibility;
using Azure.Deployments.Core.Extensions;
using Azure.Deployments.Templates.Extensions;
using Azure.Deployments.Core.Entities;
using System.Net.Http.Formatting;
using System.Text.Json;
using Microsoft.WindowsAzure.ResourceStack.Common.BackgroundJobs;
using Microsoft.WindowsAzure.ResourceStack.Common.Storage.Volatile;
using Microsoft.WindowsAzure.ResourceStack.Common.EventSources;
using Microsoft.WindowsAzure.ResourceStack.Common.Collections;
using DeploymentEngine.Jobs;

namespace DeploymentEngine.Controllers;

[ApiController]
[Route("[controller]")]
public class DeploymentController : ControllerBase
{
    private readonly ILogger<DeploymentController> _logger;

    public DeploymentController(ILogger<DeploymentController> logger)
    {
        _logger = logger;
    }

    // TODO tenant, subscription, management group
    // Default to subscription
    [HttpPut(Name = "PutDeployment")]
    public async Task<ContentResult> PutDeployment([FromBody] JsonElement content, string subscriptionId, string resourceGroupName, string deploymentName, CancellationToken cancellationToken)
    {
        string json = System.Text.Json.JsonSerializer.Serialize(content);

        // Code is copied from arm, no functionality change
        // Anthoney Uploaded extensibility package on nuget. Azure.Deployments.Extensibility
        // Ask anothony about how to determine which rp to contact with extensibility
        // Really talk to anthony about this.
        var httpContent = new StringContent(json, Encoding.UTF8, "application/json");

        // Deserialize deployment http request payload into a DeploymentContent object.
        var deploymentContent = await GetDeploymentContentAndTryCalculateHash(httpContent);
        // Overwrite newGuid() template function result in order to compare against pre-recorded baseline result.
        var overwrites = new InsensitiveConcurrentDictionary<JToken>() {
                };

        return await ProcessResourceGroupDeploymentRequest(subscriptionId, resourceGroupName, deploymentName, deploymentContent, overwrites, cancellationToken);
    }

    private async Task<ContentResult> ProcessResourceGroupDeploymentRequest(
        string subscriptionId,
        string resourceGroupName,
        string deploymentName, 
        DeploymentContent definition,
        InsensitiveConcurrentDictionary<JToken> deploymentResourceGroups,
        CancellationToken cancellationToken)
    {
        // Get ResourceGroup
        // Get Subscription
        // tenant?
        // Create deploymentContext, just a context object
        var deploymentContext = DeploymentRequestContext.CreateAtResourceGroup(
            tenantId: "temp", // TODO what is tenantId?
            subscriptionId: subscriptionId,
            resourceGroupName: resourceGroupName,
            deploymentName: deploymentName);

        return await ProcessDeploymentRequest(deploymentContext, definition, deploymentResourceGroups, cancellationToken);
    }

    private async Task<ContentResult> ProcessDeploymentRequest(DeploymentRequestContext context, DeploymentContent deploymentContent, InsensitiveConcurrentDictionary<JToken> overwrites, CancellationToken cancellationToken)
    {
        // Validation here

        // PrepareDeploymentDefinition

        // PopulateDeploymentMetadata

        //PrepareParametersForDeployment
        var deploymentEngine = new DeploymentEngine.Jobs.DeploymentEngine(location: "test", callbackFactory: new DeploymentJobCallbackFactory(new JobConfiguration()), new BackgroundEventSource());

        var deploymentState = await deploymentEngine.ProcessDeployment(context, deploymentContent, cancellationToken);

        return await this.StartDeployment(deploymentState);
    }

    private async Task<ContentResult> StartDeployment(DeploymentState deploymentState)
    {
        //await this
        //    .GetJobsDataProvider(location: deploymentState.DeploymentLocation)
        //    .CreateDeploymentJob(
        //        deploymentJob: deploymentState.DeploymentJob,
        //        onCommitJobDefinitionDelegate: () => this.SaveDeploymentAndMapping(deploymentState))
        //    .ConfigureAwait(continueOnCapturedContext: false);
        return this.CreateAsyncResponse();
    }

    private ContentResult CreateAsyncResponse()
    {
        // TODO make async response
        return new ContentResult();
    }

    /// <summary>
    /// A helper class for comparing dependencies.
    /// </summary>
    private class DependencyTuple
    {
        public string Predecessor { get; set; }

        public string Successor { get; set; }

        public static IEqualityComparer<DependencyTuple> Comparer => new DependencyTupleComparer();

        private sealed class DependencyTupleComparer : IEqualityComparer<DependencyTuple>
        {
            public bool Equals(DependencyTuple x, DependencyTuple y)
            {
                if (x is null || y is null)
                {
                    return x is null && y is null;
                }

                return string.Equals(x.Predecessor, y.Predecessor, StringComparison.OrdinalIgnoreCase)
                    && string.Equals(x.Successor, y.Successor, StringComparison.OrdinalIgnoreCase);
            }

            public int GetHashCode(DependencyTuple obj)
            {
                return StringComparer.OrdinalIgnoreCase.GetHashCode(obj);
            }
        }
    }

    public static Stream GetStreamFromString(string s)
    {
        var stream = new MemoryStream();
        var writer = new StreamWriter(stream);
        writer.Write(s);
        writer.Flush();
        stream.Position = 0;
        return stream;
    }

    /// <summary>
    /// Deserialize a deployment http request payload into a DeploymentContent object and try to calculate the hash.
    /// </summary>
    /// <param name="httpContent">The HTTP Content.</param>
    /// <param name="httpConfiguration">The HTTP Configuration.</param>
    /// <returns>The requested deployment definition.</returns>
    private static async Task<DeploymentContent> GetDeploymentContentAndTryCalculateHash(HttpContent httpContent)
    {
        var deploymentContent = await ReadAsJsonAsyncWithRewind<DeploymentContent>(httpContent)
            .ConfigureAwait(continueOnCapturedContext: false);

        deploymentContent.Properties.TemplateHash = deploymentContent.Properties.Template != null
            ? TemplateHelpers.ComputeTemplateHash(deploymentContent.Properties.Template.ToJToken())
            : null;

        return deploymentContent;
    }

    private static async Task<T> ReadAsJsonAsyncWithRewind<T>(HttpContent httpContent)
    {
        var contentStream = await httpContent.ReadAsStreamAsync().ConfigureAwait(continueOnCapturedContext: false);
        var streamPosition = contentStream.Position;

        try
        {
            var formatters = new MediaTypeFormatter[] {
                new JsonMediaTypeFormatter { SerializerSettings = SerializerSettings.SerializerMediaTypeSettings, UseDataContractJsonSerializer = false } };

            return await httpContent.ReadAsAsync<T>(formatters)
                .ConfigureAwait(continueOnCapturedContext: false);
        }
        finally
        {
            if (streamPosition != contentStream.Position)
            {
                contentStream.Seek(streamPosition, SeekOrigin.Begin);
            }
        }
    }

    /// <summary>
    /// Populate test deployment metadata, which contains JToken for evaluating template scope functions
    /// such as resourceGroup(), etc.
    /// It's up to the deployment engine host, such as ARM frontdoor or deployment micro-service, to populate
    /// the metadata and provide it to deployment engine.
    /// </summary>
    //private DeploymentMetadata GetDeploymentMetadata(DeploymentContent deploymentContent)
    //{
    //    var metadata = new DeploymentMetadata();

    //    metadata["name"] = DeploymentName;
    //    metadata[DeploymentMetadata.DeploymentKey] = deploymentContent.ToJToken();

    //    var testRgDefinition = new Dictionary<string, string> {
    //            { "subscriptionId", SubscriptionId },
    //            { "id", $"/subscriptions/{SubscriptionId}/resourceGroups/{ResourceGroupName}"},
    //            { "name", ResourceGroupName },
    //            { "location", ResourceGroupLocation } };

    //    metadata[DeploymentMetadata.ResourceGroupKey] = JObject.FromObject(testRgDefinition);

    //    var testSubscriptionDefinition = new Dictionary<string, string> {
    //            { "tenantId", TenantId},
    //            { "subscriptionId", SubscriptionId },
    //            { "id", $"/subscriptions/{SubscriptionId}"},
    //            { "displayName", "Test Subscription" } };
    //    metadata[DeploymentMetadata.SubscriptionKey] = JObject.FromObject(testSubscriptionDefinition);

    //    return metadata;
    //}
}
