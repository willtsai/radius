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

namespace DeploymentEngine.Controllers;

[ApiController]
[Route("[controller]")]
public class DeploymentController : ControllerBase
{
    private const string GuidFunctionEvaluationResultOverwrites = "6ae62a0e-7f70-445d-b3e0-8adf273fda31";

    private const string TenantId = "be07e532-cc1d-4a62-8543-b2ed399fb8c4";

    private const string SubscriptionId = "c105bc0f-ed77-493b-b9e1-1b78d672a578";

    private const string ResourceGroupName = "testRg";

    private const string ResourceGroupLocation = "DevBox";

    private const string DeploymentName = "testDeployment";

    private const string DeploymentApiVersion = "2020-10-01";

    private readonly ILogger<DeploymentController> _logger;

    public DeploymentController(ILogger<DeploymentController> logger)
    {
        _logger = logger;
    }

    [HttpPost(Name = "PostDeployment")]
    public async Task Post([FromBody] JsonElement content)
    {
        string json = System.Text.Json.JsonSerializer.Serialize(content);

        var httpContent = new StringContent(json, Encoding.UTF8, "application/json");

        // Deserialize deployment http request payload into a DeploymentContent object.
        var deploymentContent = await GetDeploymentContentAndTryCalculateHash(httpContent);
        // Overwrite newGuid() template function result in order to compare against pre-recorded baseline result.
        var overwrites = new InsensitiveDictionary<JToken>() {
                    { "newGuid", GuidFunctionEvaluationResultOverwrites }
                };

        var inputParameters = deploymentContent.Properties.Parameters is null
            ? new InsensitiveDictionary<DeploymentParameterDefinition>()
            : deploymentContent.Properties.Parameters.ToInsensitiveDictionary(
                keySelector: parameterKvp => parameterKvp.Key,
                elementSelector: parameterKvp => parameterKvp.Value);

        // Evaluate template language expressions (excluding those runtime functions)
        // and perform static (as compared to RP resource validation) validations.
        DeploymentUtils.PrepareTemplateForDeployment(
            deploymentContent: deploymentContent,
            template: deploymentContent.Properties.Template,
            deploymentParameters: inputParameters,
            metadata: this.GetDeploymentMetadata(deploymentContent),
            apiVersion: DeploymentApiVersion,
            functionEvaluationOverwrites: overwrites);

        var templateMetadata = TemplateMetadata.Build(deploymentContent.Properties.Template);

        // Populate deployment resources from template.Resources.
        var deploymentResources = await DeploymentUtils.GetDeploymentResources(
            managementGroupId: null,
            subscriptionId: SubscriptionId,
            resourceGroupName: ResourceGroupName,
            apiVersion: DeploymentApiVersion,
            metadata: templateMetadata);

        var dependencyProvider = new DependencyProcessor(DeploymentApiVersion, new TempEventSource());

        // Calculate predecessor/successor dependencies.
        var dependencies = dependencyProvider.GetDeploymentDependencies(
            managementGroupId: null,
            subscriptionId: SubscriptionId,
            resourceGroupName: ResourceGroupName,
            metadata: templateMetadata,
            deploymentResources: deploymentResources,
            extensibleResources: new Dictionary<string, ExtensibleResource>());
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
    private DeploymentMetadata GetDeploymentMetadata(DeploymentContent deploymentContent)
    {
        var metadata = new DeploymentMetadata();

        metadata["name"] = DeploymentName;
        metadata[DeploymentMetadata.DeploymentKey] = deploymentContent.ToJToken();

        var testRgDefinition = new Dictionary<string, string> {
                { "subscriptionId", SubscriptionId },
                { "id", $"/subscriptions/{SubscriptionId}/resourceGroups/{ResourceGroupName}"},
                { "name", ResourceGroupName },
                { "location", ResourceGroupLocation } };

        metadata[DeploymentMetadata.ResourceGroupKey] = JObject.FromObject(testRgDefinition);

        var testSubscriptionDefinition = new Dictionary<string, string> {
                { "tenantId", TenantId},
                { "subscriptionId", SubscriptionId },
                { "id", $"/subscriptions/{SubscriptionId}"},
                { "displayName", "Test Subscription" } };
        metadata[DeploymentMetadata.SubscriptionKey] = JObject.FromObject(testSubscriptionDefinition);

        return metadata;
    }
}
