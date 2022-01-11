using Azure.Deployments.Core.Collections;
using Azure.Deployments.Core.Definitions;
using Azure.Deployments.Core.Definitions.Extensibility;
using Azure.Deployments.Core.Entities;
using Azure.Deployments.Core.Extensions;
using Azure.Deployments.Engine.Dependencies;
using Azure.Deployments.Engine.Helpers;
using Azure.Deployments.Templates.Extensions;
using DeploymentEngine.Controllers;
using Microsoft.WindowsAzure.ResourceStack.Common.BackgroundJobs;
using Microsoft.WindowsAzure.ResourceStack.Common.EventSources;
using Microsoft.WindowsAzure.ResourceStack.Common.Storage.Volatile;
using Newtonsoft.Json.Linq;

namespace DeploymentEngine.Jobs
{
    public class DeploymentEngine
    {
        private const string DeploymentApiVersion = "2020-10-01";

        public DeploymentEngine(string location, JobCallbackFactory callbackFactory, IBackgroundJobsEventSource eventSource)
        {
            jobDispatcherClient = new JobDispatcherClient(memoryStorage: new VolatileMemoryStorage(), executionAffinity: location, eventSource: eventSource, factory: callbackFactory, secretThumbprint: null);

            jobDispatcherClient.RegisterJobCallbackAssembly(typeof(DeploymentEngine).Assembly);
            jobDispatcherClient.Start(); // TODO should start be here? 
        }

        private readonly JobDispatcherClient jobDispatcherClient;

        public JobManagementClient JobManagementClient => this.jobDispatcherClient.JobManagement;

        public async Task<DeploymentState> ProcessDeployment(DeploymentRequestContext context, DeploymentContent deploymentContent, CancellationToken cancellationToken, InsensitiveDictionary<JToken> overwrites = null)
        {
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
                metadata: this.GetDeploymentMetadata(context, deploymentContent),
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
                subscriptionId: context.SubscriptionId,
                resourceGroupName: context.ResourceGroupName,
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


            var dependencyProvider = new DependencyProcessor(DeploymentApiVersion, new EventSource());

            // Calculate predecessor/successor dependencies.
            var dependencies = dependencyProvider.GetDeploymentDependencies(
                managementGroupId: null,
                subscriptionId: context.SubscriptionId,
                resourceGroupName: context.ResourceGroupName,
                metadata: templateMetadata,
                deploymentResources: deploymentResources,
                extensibleResources: new Dictionary<string, ExtensibleResource>());


            // Before worker job, preflight
            // Kind of like whatif, failfast, send resource definition to RPs, RP will validate payload as much as possible.
            // 

            // Create a sequencer job with these dependencies and resources,
            // Sequencer job could either be extensible OR same one with different endpoint
            // When resource stack package is created, start to use it and create the deployment job here.
            var deploymentJob = await CreateNewDeploymentJob();

            // TODO need to check which resource to deploy and ordering of them

            // For now, let's do a single action in the sequence.
            foreach (var resources in deploymentResources)
            {
                // METADATA -> RESOURCE METADATA
                deploymentJob.WithAction("ID", "DeploymentResourceJob", "METADATA");
            }

            // TODO for now, let's do linear
            await this.JobManagementClient.CreateSequencer(SequencerType.Linear, deploymentJob);

            return new DeploymentState
            {
                DeploymentJob = deploymentJob,
            };
        }

        private DeploymentMetadata GetDeploymentMetadata(DeploymentRequestContext context, DeploymentContent deploymentContent)
        {
            var metadata = new DeploymentMetadata();

            metadata["name"] = context.DeploymentName;
            metadata[DeploymentMetadata.DeploymentKey] = deploymentContent.ToJToken();

            var testRgDefinition = new Dictionary<string, string> {
                { "subscriptionId", context.SubscriptionId },
                { "id", $"/subscriptions/{context.SubscriptionId}/resourceGroups/{context.ResourceGroupName}"},
                { "name", context.ResourceGroupName },
                { "location", "TODO" } };

            metadata[DeploymentMetadata.ResourceGroupKey] = JObject.FromObject(testRgDefinition);

            var testSubscriptionDefinition = new Dictionary<string, string> {
                { "tenantId", context.TenantId},
                { "subscriptionId", context.SubscriptionId },
                { "id", $"/subscriptions/{context.SubscriptionId}"},
                { "displayName", "Test Subscription" } };
            metadata[DeploymentMetadata.SubscriptionKey] = JObject.FromObject(testSubscriptionDefinition);

            return metadata;
        }

        public Task<SequencerBuilder> CreateNewDeploymentJob()
        {
            var deploymentJob = CreateDeploymentSequencer();

            return Task.FromResult(deploymentJob);
        }

        private SequencerBuilder CreateDeploymentSequencer()
        {
            //var deploymentSequencerPartition = DeploymentJobMetadata.GetSequencerPartition(deployment: deployment);
            //             var deploymentSequencerId = DeploymentJobMetadata.GetSequencerId(deployment: deployment);

            // What is a partition: even spread across sequencer jobs, can use subscriptionId + resourcegroupname maybe?
            // ID is just an id
            var deploymentJob = SequencerBuilder.Create("test", "SOMEID");

            return deploymentJob;
        }
    }
}
