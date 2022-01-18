using Azure.Deployments.Core.Algorithms;
using Azure.Deployments.Core.Collections;
using Azure.Deployments.Core.Constants;
using Azure.Deployments.Core.Definitions;
using Azure.Deployments.Core.Definitions.Extensibility;
using Azure.Deployments.Core.Definitions.Schema;
using Azure.Deployments.Core.Entities;
using Azure.Deployments.Core.ErrorResponses;
using Azure.Deployments.Core.Extensions;
using Azure.Deployments.Core.Instrumentation.Extensions;
using Azure.Deployments.Core.Json;
using Azure.Deployments.Core.Resources;
using Azure.Deployments.Core.Storage.Helpers;
using Azure.Deployments.Engine.Dependencies;
using Azure.Deployments.Engine.Helpers;
using Azure.Deployments.Templates.Exceptions;
using Azure.Deployments.Templates.Extensions;
using DeploymentEngine.Controllers;
using Microsoft.WindowsAzure.ResourceStack.Common.BackgroundJobs;
using Microsoft.WindowsAzure.ResourceStack.Common.EventSources;
using Microsoft.WindowsAzure.ResourceStack.Common.Storage.Volatile;
using Microsoft.WindowsAzure.ResourceStack.Frontdoor.Data.Engines;
using Newtonsoft.Json.Linq;
using System.Globalization;
using DeploymentsProvisioningState = global::Azure.Deployments.Core.Definitions.ProvisioningState;


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
            dependencyProcessor = new DependencyProcessor(deploymentApiVersion: DeploymentApiVersion, new EventSource());
        }

        private readonly JobDispatcherClient jobDispatcherClient;

        // URI to radius service.
        //private readonly Uri frontendUri = new Uri("https://radius-service.radius-system.svc.cluster.local/apis/api.radius.dev/v1alpha1/");
        private readonly Uri frontendUri = new Uri("http://localhost:7443/apis/api.radius.dev/v1alpha1/");
        private DependencyProcessor dependencyProcessor;

        public JobManagementClient JobManagementClient => this.jobDispatcherClient.JobManagement;

        public async Task<DeploymentState> ProcessDeployment(DeploymentRequestContext context, DeploymentContent deploymentContent, CancellationToken cancellationToken, InsensitiveDictionary<JToken> overwrites = null)
        {
            var inputParameters = deploymentContent.Properties.Parameters is null
            ? new InsensitiveDictionary<DeploymentParameterDefinition>()
            : deploymentContent.Properties.Parameters.ToInsensitiveDictionary(
                keySelector: parameterKvp => parameterKvp.Key,
                elementSelector: parameterKvp => parameterKvp.Value);

            //var metadata = await this.deploymentEngineHost.PopulateDeploymentMetadata(
            //    context: context,
            //    deploymentDefinition: deploymentContent,
            //    originalDeploymentName: originalDeploymentName);

            //Evaluate template language expressions(excluding those runtime functions)
            // and perform static (as compared to RP resource validation) validations.

            // TODO temporary for now to get something working.
            var metadata = GetDeploymentMetadata(context, deploymentContent);
            var template = DeploymentUtils.PrepareTemplateForDeployment(
                deploymentContent: deploymentContent,
                template: deploymentContent.Properties.Template,
                deploymentParameters: inputParameters,
                metadata: metadata,
                apiVersion: DeploymentApiVersion,
                functionEvaluationOverwrites: overwrites);

            var templateMetadata = TemplateMetadata.Build(deploymentContent.Properties.Template);

            // Populate deployment resources from template.Resources.
            var deploymentResources = await DeploymentUtils.GetDeploymentResources(
                managementGroupId: null, // TODO fill this in
                subscriptionId: context.SubscriptionId,
                resourceGroupName: context.ResourceGroupName,
                apiVersion: DeploymentApiVersion,
                metadata: templateMetadata);

            var dependencyProvider = new DependencyProcessor(DeploymentApiVersion, new EventSource());

            // Calculate predecessor/successor dependencies.
            var deploymentDependencies = dependencyProvider.GetDeploymentDependencies(
                managementGroupId: null,
                subscriptionId: context.SubscriptionId,
                resourceGroupName: context.ResourceGroupName,
                metadata: templateMetadata,
                deploymentResources: deploymentResources,
                extensibleResources: new Dictionary<string, ExtensibleResource>());

            // Before worker job, preflight
            // Kind of like whatif, failfast, send resource definition to RPs, RP will validate payload as much as possible.
            // 

            var newDeployment = CreateNewDeployment(
                tenantId: context.TenantId,
                managementGroupId: context.ManagementGroupId,
                subscriptionId: context.SubscriptionId,
                resourceGroupName: context.ResourceGroupName,
                deploymentName: context.DeploymentName,
                template: template,
                templateHash: deploymentContent.Properties.TemplateHash,
                templateParametersHash: deploymentContent.Properties.TemplateParametersHash,
                definition: deploymentContent,
                metadata: metadata,
                parameters: inputParameters,
                resources: deploymentResources,
                dependencies: deploymentDependencies);

            // Create a sequencer job with these dependencies and resources,
            // Sequencer job could either be extensible OR same one with different endpoint
            // When resource stack package is created, start to use it and create the deployment job here.
            var deploymentJob = await CreateNewDeploymentJob(
                        frontdoorEndpoint: frontendUri,
                        frontdoorLocation: "test",
                        subscriptionId: context.SubscriptionId,
                        resourceGroupName: context.ResourceGroupName,
                        deploymentLocation: "westus2", // TODO temp
                        template: template,
                        metadata: templateMetadata,
                        deployment: newDeployment,
                        deploymentResources: deploymentResources,
                        deploymentDependencies: deploymentDependencies);
                        //resourceProvidersToRegister: resourceProvidersToRegister,
                        //frontdoorEndpointUri: frontendUri,);

            return new DeploymentState
            {
                DeploymentJob = deploymentJob,
            };
        }

        private IDeploymentEntity CreateNewDeployment(
            string tenantId,
            string managementGroupId,
            string subscriptionId,
            string resourceGroupName,
            string deploymentName,
            Template template,
            string templateHash,
            string templateParametersHash,
            DeploymentContent definition,
            DeploymentMetadata metadata,
            InsensitiveDictionary<DeploymentParameterDefinition> parameters,
            DeploymentResource[] resources,
            HashSet<DeploymentDependency> dependencies)
        {
            var sequenceId = GetReverseSequenceId();

            return new DeploymentInfo
            {
                TenantId = tenantId,
                ManagementGroupId = managementGroupId,
                SubscriptionId = subscriptionId,
                ResourceGroupName = resourceGroupName,
                DeploymentName = deploymentName,
                SequenceId = sequenceId,
                TemplateHash = templateHash,
                TemplateParametersHash = templateParametersHash,
                ProvisioningState = DeploymentsProvisioningState.Accepted,
                ValidationLevel = definition.Properties.ValidationLevel,
                DeploymentMode = definition.Properties.Mode,
                DeploymentSecurityMode = definition.Properties.SecurityMode.HasValue ? definition.Properties.SecurityMode.Value : DeploymentSecurityMode.NotSpecified,
                TemplateLink = definition.Properties.TemplateLink,
                ParametersLink = definition.Properties.ParametersLink,
                Providers = this.CreateMinResourceProviders(resources),
                Dependencies = dependencies.ToArray(),
                Imports = CreateDeploymentImports(template.Imports),
                Variables = this.CreateDeploymentVariables(template.Variables),
                Parameters = this.CreateDeploymentParameters(template.Parameters, parameters),
                Functions = this.CreateDeploymentFunctions(template),
                Outputs = this.CreateDeploymentParameters(template.Outputs),
                Metadata = metadata,
                DebugSetting = definition.Properties.DebugSetting,
                OnErrorDeployment = definition.Properties.OnErrorDeployment,
                Tags = definition.Tags
            };
        }

        public static string GetReverseSequenceId(string oldSequenceId = null)
        {
            if (!string.IsNullOrEmpty(oldSequenceId))
            {
                return Math.Min(long.MaxValue - DateTimeExtensions.PreciseUtcNow.Ticks, long.Parse(oldSequenceId, CultureInfo.InvariantCulture) - 1).ToString("D20", CultureInfo.InvariantCulture);
            }

            return (long.MaxValue - DateTimeExtensions.PreciseUtcNow.Ticks).ToString("D20", CultureInfo.InvariantCulture);
        }

        public async Task<SequencerBuilder> CreateNewDeploymentJob(
            Uri frontdoorEndpoint,
            string frontdoorLocation,
            string subscriptionId,
            string resourceGroupName,
            string deploymentLocation,
            Template template,
            ITemplateMetadata metadata,
            IDeploymentEntity deployment,
            DeploymentResource[] deploymentResources,
            HashSet<DeploymentDependency> deploymentDependencies)
        {
            var deploymentJob = this.CreateDeploymentSequencer(
               frontdoorEndpoint: frontdoorEndpoint,
               frontdoorLocation: frontdoorLocation,
               deploymentLocation: deploymentLocation,
               deployment: deployment);

            this.PopulateDeploymentRegistrationJobs(
                deploymentLocation: deploymentLocation,
                deployment: deployment,
                resourceProvidersToRegister: null,
                deploymentJob: deploymentJob);

            await this
                .PopulateDeploymentResourceJobs(
                    frontdoorEndpoint: frontdoorEndpoint,
                    subscriptionId: subscriptionId,
                    resourceGroupName: resourceGroupName,
                    deploymentLocation: deploymentLocation,
                    template: template,
                    metadata: metadata,
                    deployment: deployment,
                    deploymentResources: deploymentResources,
                    extensibleResources: new Dictionary<string, ExtensibleResource>(),
                    deploymentJob: deploymentJob,
                    deploymentDependencies: deploymentDependencies)
                .ConfigureAwait(continueOnCapturedContext: false);

            this.PopulateDeploymentJobDependencies(
                deploymentResources: deploymentResources,
                deploymentDependencies: deploymentDependencies,
                deploymentJob: deploymentJob);

            //if (deployment.DeploymentMode == DeploymentMode.Complete)
            //{
            //    this.PopulateDeploymentCleanupJob(
            //        deploymentLocation: deploymentLocation,
            //        deployment: deployment,
            //        deploymentJob: deploymentJob);
            //}

            // Validate dependencies here as we don't have all the dependencies until this point.
            //this.HandleCircularDependencies(deploymentDependencies, deploymentJob);

            await this.JobManagementClient.CreateSequencer(SequencerType.Linear, deploymentJob);
            return deploymentJob;
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

        /// <summary>
        /// Creates deployment resource provider.
        /// </summary>
        /// <param name="deploymentResources">The deployment resources.</param>
        private MinResourceProvider[] CreateMinResourceProviders(DeploymentResource[] deploymentResources)
        {
            if (deploymentResources == null)
            {
                return null;
            }

            return deploymentResources
                .Where(resource => resource.IsConditionTrue)
                .GroupByInsensitively(resource => resource.GetResourceProviderNamespace())
                .SelectArray(resources => new MinResourceProvider
                {
                    Namespace = resources.Key,
                    ResourceTypes = resources
                        .GroupByInsensitively(resource => resource.GetResourceType())
                        .SelectArray(resourcesByType => new MinResourceType
                        {
                            ResourceType = resourcesByType.Key,
                            Locations = resourcesByType.DistinctArray(resourceByLocation => StorageUtility.NormalizeLocationForDisplay(resourceByLocation.Location))
                        })
                });
        }

        private static Import[] CreateDeploymentImports(Dictionary<string, TemplateImport> imports)
        {
            if (imports is null)
            {
                return null;
            }

            return imports.SelectArray(kvp => new Import
            {
                Alias = kvp.Key,
                Provider = kvp.Value.Provider.Value,
                Version = kvp.Value.Version.Value,
                Config = kvp.Value.Config?.Value,
            });
        }


        /// <summary>
        /// Creates the deployment variables.
        /// </summary>
        /// <param name="variables">The template variables.</param>
        private DeploymentVariable[] CreateDeploymentVariables(Dictionary<string, TemplateGenericProperty<JToken>> variables)
        {
            if (variables == null)
            {
                return null;
            }

            return variables
                .Select(variable => this.CreateDeploymentVariable(variable))
                .ToArray();
        }

        /// <summary>
        /// Creates the deployment variable.
        /// </summary>
        /// <param name="variable">The template variable.</param>
        private DeploymentVariable CreateDeploymentVariable(KeyValuePair<string, TemplateGenericProperty<JToken>> variable)
        {
            return new DeploymentVariable { Key = variable.Key, Value = variable.Value.Value };
        }


        /// <summary>
        /// Creates the deployment parameters.
        /// </summary>
        /// <typeparam name="TTemplateParameter">The type of the template parameter.</typeparam>
        /// <param name="parameters">The template parameters.</param>
        /// <param name="parameterDefinitions">The parameters definition.</param>
        private DeploymentParameter[] CreateDeploymentParameters<TTemplateParameter>(
            Dictionary<string, TTemplateParameter> parameters,
            Dictionary<string, DeploymentParameterDefinition> parameterDefinitions = null)
            where TTemplateParameter : TemplateParameter
        {
            if (parameters == null)
            {
                return null;
            }

            return parameters.SelectArray(
                parameterKvp => new DeploymentParameter
                {
                    Key = parameterKvp.Key,
                    Type = parameterKvp.Value.Type.Value,
                    Value = parameterKvp.Value.Value.Value,
                    Reference = parameterDefinitions.CoalesceDictionary().ContainsKey(parameterKvp.Key) ? parameterDefinitions[parameterKvp.Key].Reference : null
                });
        }


        /// <summary>
        /// Creates the deployment functions.
        /// </summary>
        /// <param name="template">The template.</param>
        private DeploymentFunctionDefinition[] CreateDeploymentFunctions(Template template)
        {
            return template.Functions != null ? template.GetDeploymentFunctionDefinitions() : null;
        }

        /// <summary>
        /// Populate the deployment registration jobs.
        /// </summary>
        /// <param name="deploymentLocation">The deployment location.</param>
        /// <param name="deployment">The deployment.</param>
        /// <param name="resourceProvidersToRegister">The resource providers to register.</param>
        /// <param name="deploymentJob">The deployment job.</param>
        private void PopulateDeploymentRegistrationJobs(
            string deploymentLocation,
            IDeploymentEntity deployment,
            string[] resourceProvidersToRegister,
            SequencerBuilder deploymentJob)
        {
            // TODO figure out if this is needed
            //foreach (var resourceProviderNamespace in resourceProvidersToRegister)
            //{
            //    this.AddDeploymentRegistrationJob(
            //        deploymentLocation: deploymentLocation,
            //        deployment: deployment,
            //        deploymentJob: deploymentJob,
            //        resourceProviderNamespace: resourceProviderNamespace);
            //}
        }

        /// <summary>
        /// Adds the deployment registration job.
        /// </summary>
        /// <param name="deploymentLocation">The deployment location.</param>
        /// <param name="deployment">The deployment.</param>
        /// <param name="deploymentJob">The deployment job.</param>
        /// <param name="resourceProviderNamespace">The resource provider namespace.</param>
        private void AddDeploymentRegistrationJob(
            string deploymentLocation,
            IDeploymentEntity deployment,
            SequencerBuilder deploymentJob,
            string resourceProviderNamespace)
        {
            //var deploymentRegistrationOperationId = DeploymentRegistrationJobMetadata.GetSequencerOperationId(resourceProviderNamespace);
            //if (!deploymentJob.ContainsAction(deploymentRegistrationOperationId))
            //{
            //    var deploymentRegistrationJobMetadata = this.GetDeploymentRegistrationJobMetadata(
            //        deploymentLocation: deploymentLocation,
            //        deployment: deployment,
            //        resourceProviderNamespace: resourceProviderNamespace);

            //    deploymentJob.WithAction(
            //        actionId: deploymentRegistrationOperationId,
            //        callback: "DeploymentRegistrationJob",
            //        metadata: deploymentRegistrationJobMetadata.ToJson());
            //}
        }

        private async Task PopulateDeploymentResourceJobs(
            Uri frontdoorEndpoint,
            string subscriptionId,
            string resourceGroupName,
            string deploymentLocation,
            Template template,
            ITemplateMetadata metadata,
            IDeploymentEntity deployment,
            DeploymentResource[] deploymentResources,
            IReadOnlyDictionary<string, ExtensibleResource> extensibleResources,
            SequencerBuilder deploymentJob,
            HashSet<DeploymentDependency> deploymentDependencies)
        {
            var templateResourcesLookup = new InsensitiveDictionary<DeploymentResource>();

            var symbolicNameLookup = template.HasSymbolicName() ?
                deploymentResources.ToDictionary(resource => resource.SymbolicName, CoreConstants.SymbolicNameComparer) :
                null;

            foreach (var deploymentResource in deploymentResources.CoalesceEnumerable())
            {
                if (templateResourcesLookup.ContainsKey(deploymentResource.GetFullyQualifiedResourceId()))
                {
                    throw new TemplateValidationException(
                        message: ErrorResponseMessages.TemplateResourceAlreadyDefined.ToLocalizedMessage(deploymentResource.GetUnqualifiedResourceId(), deploymentResource.DeploymentResourceLineInfo.LineNumber, deploymentResource.DeploymentResourceLineInfo.LinePosition),
                        additionalInfo: new TemplateErrorAdditionalInfo(
                            lineNumber: deploymentResource.DeploymentResourceLineInfo.LineNumber,
                            positionNumber: deploymentResource.DeploymentResourceLineInfo.LinePosition));
                }

                var deploymentOperation = deploymentResource.Existing ?
                    ProvisioningOperation.Read :
                    ProvisioningOperation.Create;

                this.AddDeploymentResourceJob(
                    frontdoorEndpoint: frontdoorEndpoint,
                    deploymentLocation: deploymentLocation,
                    deployment: deployment,
                    deploymentJob: deploymentJob,
                    deploymentResource: deploymentResource,
                    deploymentOperation: deploymentOperation);

                templateResourcesLookup.Add(deploymentResource.GetFullyQualifiedResourceId(), deploymentResource);
            }

            foreach (var deploymentResource in deploymentResources.CoalesceEnumerable())
            {
                foreach (var resourceReference in deploymentResource.References.CoalesceEnumerable())
                {
                    if (!resourceReference.IsTemplateResource)
                    {
                        var deploymentResourceLineInfo = new DeploymentResourceLineInfo
                        {
                            LineNumber = deploymentResource.DeploymentResourceLineInfo.LineNumber,
                            LinePosition = deploymentResource.DeploymentResourceLineInfo.LinePosition,
                        };

                        var referencedDeploymentResource = await this
                            .AddDeploymentResourceReferenceJob(
                                frontdoorEndpoint: frontdoorEndpoint,
                                deploymentLocation: deploymentLocation,
                                deployment: deployment,
                                deploymentJob: deploymentJob,
                                deploymentResourceLineInfo: deploymentResourceLineInfo,
                                resourceReference: resourceReference)
                            .ConfigureAwait(continueOnCapturedContext: false);

                        this.AddResourceReferenceDependencies(
                            deploymentDependencies: deploymentDependencies,
                            templateResourcesLookup: templateResourcesLookup,
                            resourceReference: resourceReference,
                            deploymentResource: referencedDeploymentResource);
                    }
                    else if (templateResourcesLookup.ContainsKey(resourceReference.GetFullyQualifiedResourceId())
                        && !templateResourcesLookup[resourceReference.GetFullyQualifiedResourceId()].IsConditionTrue)
                    {
                        // Reference to template resource (missing API version) is invalid and will throw if the template resource condition is false
                        throw new TemplateValidationException(
                             message: ErrorResponseMessages.ResourceReferenceApiVersionRequired.ToLocalizedMessage(deploymentResource.GetUnqualifiedResourceId(), resourceReference.GetUnqualifiedResourceId()));
                    }
                }
            }

            //var dependenciesBySuccessor = deploymentDependencies.ToLookup(x => x.Successor, x => x.Predecessor, DeploymentResourceReference.EqualityComparer.Instance);
            //foreach (var extensibleResource in extensibleResources.Values)
            //{
            //    AddDeploymentExtensibleResourceJob(
            //        deploymentLocation: deploymentLocation,
            //        deployment: deployment,
            //        deploymentJob: deploymentJob,
            //        metadata: metadata,
            //        dependenciesBySuccessor: dependenciesBySuccessor,
            //        resource: extensibleResource);
            //}

            var referencesLookup = deploymentResources.ToLookupInsensitively(
                keySelector: resource => resource.GetResourceNames().Last(),
                elementSelector: resource => resource.Cast<DeploymentResourceReference>());

            foreach (var resourceReference in template.OutputsReferences
                .CoalesceEnumerable()
                .Select(reference => this.dependencyProcessor.GetDeploymentResourceReference(subscriptionId, resourceGroupName, reference, referencesLookup, extensibleResources, symbolicNameLookup))
                .DistinctArray(DeploymentResourceReference.EqualityComparer.Instance))
            {
                if (resourceReference.ExtensibleReference != null &&
                    extensibleResources.ContainsKey(resourceReference.ExtensibleReference.Name))
                {
                    // Extensible resource - there is no reference() function equivalent
                    continue;
                }

                if (!resourceReference.IsTemplateResource)
                {
                    // TODO(ilygre): it is currently impossible to get line info for Dictionary<,>, but we should fix this some day.
                    var referencedDeploymentResource = await this
                        .AddDeploymentResourceReferenceJob(
                            frontdoorEndpoint: frontdoorEndpoint,
                            deploymentLocation: deploymentLocation,
                            deployment: deployment,
                            deploymentJob: deploymentJob,
                            deploymentResourceLineInfo: null,
                            resourceReference: resourceReference)
                        .ConfigureAwait(continueOnCapturedContext: false);

                    this.AddResourceReferenceDependencies(
                        deploymentDependencies: deploymentDependencies,
                        templateResourcesLookup: templateResourcesLookup,
                        resourceReference: resourceReference,
                        deploymentResource: referencedDeploymentResource);
                }
                else if (!templateResourcesLookup.ContainsKey(resourceReference.GetFullyQualifiedResourceId()))
                {
                    throw new TemplateValidationException(
                        message: ErrorResponseMessages.OutputReferenceTemplateResourceNotDefined.ToLocalizedMessage(resourceReference.GetUnqualifiedResourceId()));
                }
                else if (!templateResourcesLookup[resourceReference.GetFullyQualifiedResourceId()].IsConditionTrue)
                {
                    // Reference to template resource (missing API version) is invalid if the template resource condition is false
                    throw new TemplateValidationException(
                         message: ErrorResponseMessages.OutputReferenceApiVersionRequired.ToLocalizedMessage(resourceReference.GetUnqualifiedResourceId()));
                }
            }
        }

        /// <summary>
        /// Adds the deployment resource job.
        /// </summary>
        /// <param name="frontdoorEndpoint">The front door endpoint.</param>
        /// <param name="deploymentLocation">The deployment location.</param>
        /// <param name="deployment">The deployment.</param>
        /// <param name="deploymentJob">The deployment job.</param>
        /// <param name="deploymentResourceLineInfo">The deployment resource line information.</param>
        /// <param name="resourceReference">The resource reference.</param>
        private async Task<DeploymentResource> AddDeploymentResourceReferenceJob(
            Uri frontdoorEndpoint,
            string deploymentLocation,
            IDeploymentEntity deployment,
            SequencerBuilder deploymentJob,
            DeploymentResourceLineInfo deploymentResourceLineInfo,
            DeploymentResourceReference resourceReference)
        {
            if (resourceReference.IsTemplateResource)
            {
                throw new InvalidOperationException("Resource reference job could not be created for template defined resource object.");
            }

            var referencedDeploymentResource = new DeploymentResource
            {
                SubscriptionId = resourceReference.SubscriptionId,
                ResourceGroupName = resourceReference.ResourceGroupName,
                ResourceId = resourceReference.ResourceId,
                Scope = resourceReference.Scope,
                ReferenceAction = resourceReference.ReferenceAction,
                ReferenceApiVersion = resourceReference.ReferenceApiVersion,
                ReferenceRequestContent = resourceReference.ReferenceRequestContent,
                ApiVersion = resourceReference.ReferenceApiVersion,
                DeploymentResourceLineInfo = deploymentResourceLineInfo,
                Condition = true
            };

            //await this
            //    .deploymentEngineHost
            //    .ValidateReferencedDeploymentResource(referencedDeploymentResource)
            //    .ConfigureAwait(continueOnCapturedContext: false);

            var deploymentOperation = resourceReference.IsAction
                ? ProvisioningOperation.Action
                : ProvisioningOperation.Waiting;

            this.AddDeploymentResourceJob(
                frontdoorEndpoint: frontdoorEndpoint,
                deploymentLocation: deploymentLocation,
                deployment: deployment,
                deploymentJob: deploymentJob,
                deploymentResource: referencedDeploymentResource,
                deploymentOperation: deploymentOperation);

            return referencedDeploymentResource;
        }

        /// <summary>
        /// Adds the resource reference dependencies.
        /// </summary>
        /// <param name="deploymentDependencies">The deployment dependencies.</param>
        /// <param name="templateResourcesLookup">The template resources lookup.</param>
        /// <param name="resourceReference">The resource reference.</param>
        /// <param name="deploymentResource">The deployment resource.</param>
        private void AddResourceReferenceDependencies(
            HashSet<DeploymentDependency> deploymentDependencies,
            Dictionary<string, DeploymentResource> templateResourcesLookup,
            DeploymentResourceReference resourceReference,
            DeploymentResource deploymentResource)
        {
            // Note(antmarti): We have a reference() or a list*() function. Look up any resources with matching id, and any parent resources which are being deployed in this template.
            // Here we add dependencies on any matching resources, so that we wait for them to be deployed BEFORE executing the reference()/list*().
            // The exception to this rule is when we have a reference() function which EXACTLY correlates to a resource being deployed (same id, same api-version)
            // - in this case, we've already added a dependency on the resource creation job, and can skip this logic as we don't need to perform a GET - we just use the result of the PUT.
            foreach (var resourceId in resourceReference
                .GetParentRoutingResourceIds()
                .Concat(resourceReference.ResourceId)
                .OrderByDescending(resourceId => resourceId.Length))
            {
                var fullyQualifiedResourceId = IDeploymentResourceIdentifiableExtensions.GetFullyQualifiedResourceId(
                    subscriptionId: resourceReference.SubscriptionId,
                    resourceGroupName: resourceReference.ResourceGroupName,
                    scope: resourceReference.Scope,
                    resourceId: resourceId);

                if (templateResourcesLookup.ContainsKey(fullyQualifiedResourceId))
                {
                    var referencedResource = templateResourcesLookup[fullyQualifiedResourceId];
                    this.AddDeploymentDependency(deploymentDependencies, predecessor: referencedResource, successor: deploymentResource);

                    break;
                }
            }
        }


        /// <summary>
        /// Populates the deployment job dependencies.
        /// </summary>
        /// <param name="deploymentResources">The deployment resources.</param>
        /// <param name="deploymentDependencies">The deployment dependencies.</param>
        /// <param name="deploymentJob">The deployment job.</param>
        private void PopulateDeploymentJobDependencies(
            DeploymentResource[] deploymentResources,
            HashSet<DeploymentDependency> deploymentDependencies,
            SequencerBuilder deploymentJob)
        {
            foreach (var deploymentDependency in deploymentDependencies.CoalesceEnumerable())
            {
                var predecessorSequencerId = GetDeploymentJobSequencerId(deploymentDependency.Predecessor);
                var successorSequencerId = GetDeploymentJobSequencerId(deploymentDependency.Successor);

                if (!deploymentJob.ContainsAction(predecessorSequencerId))
                {
                    throw new TemplateValidationException(
                        message: ErrorResponseMessages.TemplateResourceNotDefined.ToLocalizedMessage(deploymentDependency.Predecessor.GetUnqualifiedResourceId()));
                }

                if (!deploymentJob.ContainsAction(successorSequencerId))
                {
                    throw new TemplateValidationException(
                        message: ErrorResponseMessages.TemplateResourceNotDefined.ToLocalizedMessage(deploymentDependency.Successor.GetUnqualifiedResourceId()));
                }

                if (DeploymentResourceReference.EqualityComparer.Instance.Equals(deploymentDependency.Predecessor, deploymentDependency.Successor))
                {
                    throw new TemplateValidationException(
                        message: ErrorResponseMessages.TemplateResourceSelfReference.ToLocalizedMessage(deploymentDependency.Successor.GetUnqualifiedResourceId()));
                }

                deploymentJob.WithDependency(
                    runBeforeActionId: predecessorSequencerId,
                    runAfterActionId: successorSequencerId);
            }

            foreach (var deploymentResource in deploymentResources)
            {
                var deploymentRegistrationOperationId = ComputeHash.MurmurHash64(deploymentResource.GetResourceProviderNamespace().ToUpperInvariant()).ToString("X16");

                if (deploymentJob.ContainsAction(deploymentRegistrationOperationId))
                {
                    deploymentJob.WithDependency(
                        runBeforeActionId: deploymentRegistrationOperationId,
                        runAfterActionId: DeploymentJobMetadata.GetSequencerOperationId(deploymentResource));
                }
            }
        }

        private void AddDeploymentResourceJob(Uri frontdoorEndpoint,
            string deploymentLocation,
            IDeploymentEntity deployment,
            SequencerBuilder deploymentJob,
            DeploymentResource deploymentResource,
            ProvisioningOperation deploymentOperation)
        {
            var deploymentSequencerOperationId = DeploymentJobMetadata.GetSequencerOperationId(deploymentResource);
            if (!deploymentJob.ContainsAction(deploymentSequencerOperationId))
            {
                var deploymentResourceJobMetadata = GetDeploymentResourceJobMetadata(
                    frontdoorEndpoint: frontdoorEndpoint,
                    deploymentLocation: deploymentLocation,
                    deployment: deployment,
                    deploymentResource: deploymentResource,
                    deploymentOperation: deploymentOperation);

                if (deploymentResource.IsConditionTrue)
                {
                    deploymentJob.WithAction(
                        actionId: deploymentSequencerOperationId,
                        callback: "DeploymentResourceJob",
                        metadata: deploymentResourceJobMetadata.ToJson());
                }
                else
                {
                    deploymentJob.WithAction(
                        actionId: deploymentSequencerOperationId,
                        callback: "DeploymentResourceNoOperationJob",
                        metadata: deploymentResourceJobMetadata.ToJson());
                }
            }
        }


        /// <summary>
        /// Gets the deployment job sequencer Id for a resource reference.
        /// </summary>
        /// <param name="reference">The resource reference.</param>
        public static string GetDeploymentJobSequencerId(DeploymentResourceReference reference)
        {
            //if (reference.ExtensibleReference != null)
            //{
            //    return DeploymentExtensibleResourceJobMetadata.GetSequencerOperationId(reference.ExtensibleReference.Name);
            //}

            return DeploymentJobMetadata.GetSequencerOperationId(reference);
        }


        /// <summary>
        /// Adds the deployment dependency.
        /// </summary>
        /// <param name="deploymentDependencies">The deployment dependencies.</param>
        /// <param name="predecessor">The predecessor.</param>
        /// <param name="successor">The successor.</param>
        private void AddDeploymentDependency(HashSet<DeploymentDependency> deploymentDependencies, DeploymentResourceReference predecessor, DeploymentResourceReference successor)
        {
            deploymentDependencies.Add(new DeploymentDependency { Predecessor = predecessor.GetReference(), Successor = successor.GetReference() });
        }



        private DeploymentResourceJobMetadata GetDeploymentResourceJobMetadata(
                Uri frontdoorEndpoint, // where the server should contact the RP
                string deploymentLocation, // Not sure what this is.
                IDeploymentEntity deployment,
                DeploymentResource deploymentResource,
                ProvisioningOperation deploymentOperation)
        {
            return new DeploymentResourceJobMetadata
            {
                TenantId = deployment.TenantId,
                ManagementGroupId = deployment.ManagementGroupId,
                SubscriptionId = deployment.SubscriptionId,
                ResourceGroupName = deployment.ResourceGroupName,
                ResourceGroupLocation = deploymentLocation,
                DeploymentLocation = deploymentLocation,
                DeploymentName = deployment.DeploymentName,
                SequenceId = deployment.SequenceId,

                Resource = deploymentResource,
                ResourceOperation = deploymentOperation,
                ResourceOperationUri = this.GetResourceOperationUri(frontdoorEndpoint, deploymentResource),
                ResourceOperationRequestContent = deploymentResource.ReferenceRequestContent,
                AreTemplateExpressionsEvaluated = false,

                DebugSetting = deployment.DebugSetting,
            };
        }

        /// <summary>
        /// Gets the resource operation URI.
        /// </summary>
        /// <param name="frontdoorEndpoint">The front door endpoint.</param>
        /// <param name="deploymentResource">The deployment resource.</param>
        private Uri GetResourceOperationUri(Uri frontdoorEndpoint, DeploymentResource deploymentResource)
        {
            var normalizedResourceId = deploymentResource.IsAction
                ? deploymentResource.GetUnqualifiedResourceId().Trim('/') + '/' + deploymentResource.ReferenceAction
                : deploymentResource.GetUnqualifiedResourceId().Trim('/');

            return UriTemplateEngine.GetResourceUri(
                endpoint: frontdoorEndpoint,
                subscriptionId: deploymentResource.SubscriptionId,
                resourceGroupName: deploymentResource.ResourceGroupName,
                resourceId: normalizedResourceId,
                apiVersion: deploymentResource.ApiVersion);
        }

        private SequencerBuilder CreateDeploymentSequencer(Uri frontdoorEndpoint,
            string frontdoorLocation,
            string deploymentLocation,
            IDeploymentEntity deployment)
        {
            var deploymentSequencerPartition = DeploymentJobMetadata.GetSequencerPartition(deployment: deployment);
            var deploymentSequencerId = DeploymentJobMetadata.GetSequencerId(deployment: deployment);


            var deploymentJob = SequencerBuilder.Create(
                sequencerPartition: deploymentSequencerPartition,
                sequencerId: deploymentSequencerId);

            //var deploymentFrontdoorLocation = this.deploymentSettings.DeploymentFrontdoorLocationEnabled
            //    && !string.IsNullOrEmpty(deployment.SubscriptionId)
            //    && this.deploymentSettings.DeploymentFrontdoorLocationEnabledSubscriptions.ContainsOrdinalInsensitively(deployment.SubscriptionId)
            //        ? deploymentLocation : frontdoorLocation;
            //var deploymentFrontdoorLocation = this.deploymentSettings.DeploymentFrontdoorLocationEnabled
            //    && !string.IsNullOrEmpty(deployment.SubscriptionId)
            //    && this.deploymentSettings.DeploymentFrontdoorLocationEnabledSubscriptions.ContainsOrdinalInsensitively(deployment.SubscriptionId)
            //        ? deploymentLocation : frontdoorLocation;

            var sharedMetadata = new FrontdoorJobMetadata
            {
                FrontdoorEndpoint = frontdoorEndpoint,
            };

            deploymentJob.WithSharedMetadata(sharedMetadata.ToJson());

            //var deploymentJobMetadata = this.GetDeploymentJobMetadata(
            //    deploymentLocation: deploymentLocation,
            //    deployment: deployment);


            //deploymentJob.WithFirstAction(
            //    callback: "DeploymentFirstJob",
            //    metadata: deploymentJobMetadata.ToJson());

            //deploymentJob.WithLastAction(
            //    callback: "DeploymentLastJob",
            //    metadata: deploymentJobMetadata.ToJson());

            return deploymentJob;
        }
    }
}
