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

namespace DeploymentEngine.Controllers;

[ApiController]
[Route("[controller]")]
public class DeploymentController : ControllerBase
{
    private const string DeploymentApiVersion = "2020-10-01";

    private readonly ILogger<DeploymentController> _logger;

    private readonly JobDispatcherClient jobDispatcherClient;

    public JobManagementClient JobManagementClient => this.jobDispatcherClient.JobManagement;

    public DeploymentController(ILogger<DeploymentController> logger, string location, JobCallbackFactory callbackFactory, IBackgroundJobsEventSource eventSource, string secretThumbprint)
    {
        _logger = logger;
        jobDispatcherClient = new JobDispatcherClient(memoryStorage: new VolatileMemoryStorage(), executionAffinity: location, eventSource: eventSource, factory: callbackFactory, secretThumbprint: secretThumbprint);

        jobDispatcherClient.RegisterJobCallbackAssembly(typeof(DeploymentController).Assembly);
    }

    [HttpPost(Name = "PostDeployment")]
    public async Task Post([FromBody] JsonElement content)
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
        var overwrites = new InsensitiveDictionary<JToken>() {
                    { "newGuid", GuidFunctionEvaluationResultOverwrites }
                };


        await ProcessDeployment(deploymentContent, overwrites);
    }

    private async Task ProcessDeployment(DeploymentContent deploymentContent, InsensitiveDictionary<JToken> overwrites)
    {
        // Validation here

        // PrepareDeploymentDefinition

        // PopulateDeploymentMetadata

        //PrepareParametersForDeployment

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

        //var deploymentLocation = DeploymentEngine.GetDeploymentLocation(resourceGroupLocation: deploymentContext.ResourceGroupLocation, definition: definition);

    //    var oldDeploymentMapping = await this
    //.GetOldDeploymentMapping(
    //    deploymentContext: deploymentContext,
    //    deploymentDefinition: definition,
    //    deploymentName: deploymentContext.DeploymentName)
    //.ConfigureAwait(continueOnCapturedContext: false);

    //    var oldDeployment = await this
    //        .ProcessOldDeployment(
    //            deploymentContext: deploymentContext,
    //            deploymentDefinition: definition,
    //            deploymentLocation: deploymentLocation,
    //            frontdoorEndpoint: request.GetFrontdoorEndpoint().Uri,
    //            oldDeploymentMapping: oldDeploymentMapping)
    //        .ConfigureAwait(continueOnCapturedContext: false);

        // Populate deployment resources from template.Resources.
        var deploymentResources = await DeploymentUtils.GetDeploymentResources(
            managementGroupId: null, // TODO fill this in
            subscriptionId: SubscriptionId,
            resourceGroupName: ResourceGroupName,
            apiVersion: DeploymentApiVersion,
            metadata: templateMetadata);

        // ExtensibleResources

        // preflight validation
        //                if (validatePreflightResources)
        //{
        //    await this.deploymentEngineHost.PerformStaticResourceValidation(
        //                subscriptionId: deploymentContext.SubscriptionId,
        //                resourceGroupName: deploymentContext.ResourceGroupName,
        //                deploymentResources: deploymentResources)
        //        .ConfigureAwait(continueOnCapturedContext: false);
        //}


        var dependencyProvider = new DependencyProcessor(DeploymentApiVersion, new TempEventSource());

        // Calculate predecessor/successor dependencies.
        var dependencies = dependencyProvider.GetDeploymentDependencies(
            managementGroupId: null,
            subscriptionId: SubscriptionId,
            resourceGroupName: ResourceGroupName,
            metadata: templateMetadata,
            deploymentResources: deploymentResources,
            extensibleResources: new Dictionary<string, ExtensibleResource>());


        // Before worker job, preflight
        // Kind of like whatif, failfast, send resource definition to RPs, RP will validate payload as much as possible.
        // 

        // Create a sequencer job with these dependencies and resources,
        // Sequencer job could either be extensible OR same one with different endpoint
        // When resource stack package is created, start to use it and create the deployment job here.

    }

    public async Task<SequencerBuilder> CreateNewDeploymentJob()
    {
        var deploymentJob = CreateDeploymentSequencer();

        // For now, let's do a single action in the sequence.
        deploymentJob.WithAction("ID", "DeploymentResourceJob", "METADATA");

        // TODO for now, let's do linear
        await this.JobManagementClient.CreateSequencer(SequencerType.Linear, deploymentJob);
    }

    private SequencerBuilder CreateDeploymentSequencer()
    {
        //var deploymentSequencerPartition = DeploymentJobMetadata.GetSequencerPartition(deployment: deployment);
        //             var deploymentSequencerId = DeploymentJobMetadata.GetSequencerId(deployment: deployment);

        // What is a partition: even spread across sequencer jobs, can use subscriptionId + resourcegroupname maybe?
        // ID is just an id
        var deploymentJob = SequencerBuilder.Create("test", "SOMEID");


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
