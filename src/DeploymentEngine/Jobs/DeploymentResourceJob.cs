using Azure.Deployments.Core.Constants;
using Azure.Deployments.Core.Definitions;
using Azure.Deployments.Core.Definitions.Extensibility;
using Azure.Deployments.Core.Definitions.Resources;
using Azure.Deployments.Core.Definitions.Schema;
using Azure.Deployments.Core.Entities;
using Azure.Deployments.Core.ErrorResponses;
using Azure.Deployments.Core.Exceptions;
using Azure.Deployments.Core.Extensions;
using Azure.Deployments.Core.Instrumentation.Extensions;
using Azure.Deployments.Core.Json;
using Azure.Deployments.Core.Resources;
using Azure.Deployments.Expression.Expressions;
using Azure.Deployments.Templates.Exceptions;
using Microsoft.WindowsAzure.ResourceStack.Common.BackgroundJobs;
using Microsoft.WindowsAzure.ResourceStack.Common.Extensions;
using Microsoft.WindowsAzure.ResourceStack.Common.Instrumentation;
using Microsoft.WindowsAzure.ResourceStack.Frontdoor.Data.Engines;
using Microsoft.WindowsAzure.ResourceStack.Frontdoor.Data.Extensions;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;
using System.Net;
using System.Net.Http.Formatting;
using System.Net.Http.Headers;
using ErrorResponseCode = global::Azure.Deployments.Core.ErrorResponses.DeploymentsErrorResponseCode;
using DeploymentsProvisioningState = global::Azure.Deployments.Core.Definitions.ProvisioningState;


namespace DeploymentEngine.Jobs
{
    [JobCallback(Name = "DeploymentResourceJob")]
    public class DeploymentResourceJob : JobBase<DeploymentResourceJobMetadata>
    {
        public DeploymentResourceJob(JobConfiguration configuration) :
            base(configuration)
        {
        }

        /// <summary>
        /// Gets the default MediaTypeFormatterSettings, notice that 'JsonExtensions' references AzureUX-Deployments type
        /// which means this formatter will be able to properly serialize / deserialize AzureUX-Deployments types,
        /// any duplicate types (e.g. TagsDictionary) from ARM would not be properly serialize / deserialize.
        /// </summary>
        private JsonMediaTypeFormatter JsonMediaTypeFormatter
            => new JsonMediaTypeFormatter
            {
                SerializerSettings = JsonExtensions.MediaTypeFormatterSettings,
                UseDataContractJsonSerializer = false
            };
        private JsonMediaTypeFormatter JsonObjectTypeFormatter
            => new JsonMediaTypeFormatter
            {
                SerializerSettings = JsonExtensions.ObjectSerializationSettings,
                UseDataContractJsonSerializer = false
            };

        private MediaTypeFormatter[] JsonObjectTypeFormatters
    => new MediaTypeFormatter[] { JsonObjectTypeFormatter };

        /// <summary>
        /// Gets the default resource operation timeout.
        /// </summary>
        private TimeSpan DefaultResourceOperationTimeout
        {
            get
            {
                return TimeSpan.FromHours(2);
            }
        }

        protected override Task OnConfigure()
        {
            return base.OnConfigure();
        }

        protected override async Task<JobExecutionResult> OnExecute()
        {
            var provisionResult = await ProvisionResource();

            // TODO create radius resource for deployment
            // OR call RP contract on update resource.
            // Would love some sort of way to poll better here for deployment.
            Metadata = JToken.Parse(this.BackgroundJob.Metadata).ToObject<DeploymentResourceJobMetadata>();

            return provisionResult;
        }


        /// <summary>
        /// Provisions the resource.
        /// </summary>
        private async Task<JobExecutionResult> ProvisionResource()
        {
            try
            {
                var deployment = await this.GetCurrentDeploymentSequence().ConfigureAwait(continueOnCapturedContext: false);
                if (deployment.ProvisioningState != ProvisioningState.Running)
                {
                    return new JobExecutionResult
                    {
                        Status = JobExecutionStatus.Failed,
                        Message = string.Format("Deployment provisioning state '{0}' is not expected. Current deployment operation is failed.", deployment.ProvisioningState),
                        Details = deployment.ProvisioningState.ToString(),
                    };
                }

                var operationTimeout = await this
                    .GetOperationTimeout(resourceOperation: this.Metadata.ResourceOperation)
                    .ConfigureAwait(continueOnCapturedContext: false);

                if (this.BackgroundJob.StartTime.Value.Add(operationTimeout) < DateTime.UtcNow)
                {
                    var errorResponseMessage = new ErrorResponseMessage(
                        code: ErrorResponseCode.ResourceDeploymentFailure,
                        message: Microsoft.WindowsAzure.ResourceStack.Common.Instrumentation.LocalizationExtensions.ToLocalizedMessage(ErrorResponseMessages.DeploymentResourceOperationTimeout));

                    this.Metadata.ResourceOperationStatusCode = HttpStatusCode.RequestTimeout;
                    this.Metadata.ResourceOperationStatusMessage = errorResponseMessage.ToJToken();

                    return new JobExecutionResult
                    {
                        Status = JobExecutionStatus.Failed,
                        Message = string.Format(
                            format: "The resource provision operation did not complete within the allowed timeout period of '{0}'. Current deployment operation is failed.",
                            arg0: operationTimeout),
                        NextMetadata = this.Metadata.ToJson()
                    };
                }

                return await this.HandleResourceOperationState(deployment).ConfigureAwait(continueOnCapturedContext: false);
            }
            catch (JobExecutionResultException ex)
            {
                return ex.ToJobExecutionResult();
            }
            catch (Exception ex)
            {
                //if (ex.IsFatal())
                //{
                //    throw;
                //}

                //if (ex.IsTransientException())
                //{
                //    this.Logger.LogWarning(
                //        exception: ex,
                //        operationName: "DeploymentResourceJob.DeploymentResource",
                //        message: "The transient exception is encountered and handled. Current deployment operation is postponed.");

                //    return new JobExecutionResult
                //    {
                //        Status = JobExecutionStatus.Postponed,
                //        Message = "The transient exception is encountered and handled. Current deployment operation is postponed.",
                //    };
                //}
                //else
                //{
                this.Logger.LogError(
                    exception: ex,
                    operationName: "DeploymentResourceJob.DeploymentResource",
                    message: string.Format("The unknown exception is encountered. Current deployment operation is failed. Error message: '{0}'. Exception: '{1}'", ex.Message, ex.InnerException));

                var errorResponseMessage = new ErrorResponseMessage(
                    code: ErrorResponseCode.ResourceDeploymentFailure,
                    message: Microsoft.WindowsAzure.ResourceStack.Common.Instrumentation.LocalizationExtensions.ToLocalizedMessage(DateTime.UtcNow.ToString(), RequestCorrelationContext.Current.SubscriptionId, RequestCorrelationContext.Current.CurrentActivityId, RequestCorrelationContext.Current.CorrelationId));

                this.Metadata.ResourceOperationStatusCode = HttpStatusCode.InternalServerError;
                this.Metadata.ResourceOperationStatusMessage = errorResponseMessage.ToJToken();

                return new JobExecutionResult
                {
                    Status = JobExecutionStatus.Failed,
                    Message = "The unknown exception is encountered. Current deployment operation is failed.",
                    NextMetadata = this.Metadata.ToJson()
                };
                //}
            }
        }

        /// <summary>
        /// Executes the provisioning operation's state machine.
        /// </summary>
        /// <param name="deployment">The deployment.</param>
        private async Task<JobExecutionResult> HandleResourceOperationState(IDeploymentEntity deployment)
        {
            switch (this.Metadata.ResourceOperation)
            {
                case ProvisioningOperation.Create:
                    return await this.CreateResource(deployment).ConfigureAwait(continueOnCapturedContext: false);

                //case ProvisioningOperation.Action:
                //    return await this.StartResourceAction(deployment).ConfigureAwait(continueOnCapturedContext: false);

                //case ProvisioningOperation.Read:
                //case ProvisioningOperation.Waiting:
                //    return await this.CheckResourceOperationStatus(deployment).ConfigureAwait(continueOnCapturedContext: false);

                //case ProvisioningOperation.AzureAsyncOperationWaiting:
                //    return await this.HandleAzureAsyncOperation(deployment).ConfigureAwait(continueOnCapturedContext: false);

                //case ProvisioningOperation.ResourceCacheWaiting:
                //    return await this.HandleResourceCacheWaitingOperation().ConfigureAwait(continueOnCapturedContext: false);

                default:
                    return new JobExecutionResult
                    {
                        Status = JobExecutionStatus.Failed,
                        Message = string.Format("The unknown resource operation '{0}' is encountered. Current deployment operation is failed.", this.Metadata.ResourceOperation),
                        Details = this.Metadata.ResourceOperation.ToString(),
                    };
            }
        }

        /// <summary>
        /// Creates the resource.
        /// </summary>
        /// <param name="deployment">The deployment.</param>
        private async Task<JobExecutionResult> CreateResource(IDeploymentEntity deployment)
        {
            await this
                .ProcessResourceLanguageExpressions(
                    resource: this.Metadata.Resource,
                    deployment: deployment)
                .ConfigureAwait(continueOnCapturedContext: false);

            var definition = new ResourceProxyDefinition
            {
                Location = this.Metadata.Resource.Location,
                ExtendedLocation = this.Metadata.Resource.ExtendedLocation,
                Tags = this.Metadata.Resource.Tags,
                Scale = this.Metadata.Resource.Scale,
                Sku = this.Metadata.Resource.Sku,
                Kind = this.Metadata.Resource.Kind,
                ManagedBy = this.Metadata.Resource.ManagedBy,
                ManagedByExtended = this.Metadata.Resource.ManagedByExtended,
                Plan = this.Metadata.Resource.Plan,
                Identity = this.Metadata.Resource.Identity,
                Zones = this.Metadata.Resource.Zones,
                Properties = this.Metadata.Resource.Properties
            };

            this.UpdateMetadataOnOperationRequest(JToken.FromObject(definition, JsonExtensions.JsonMediaTypeSerializer));

            using (var httpContent = CreateJsonContent(definition, JsonMediaTypeFormatter))
            using (var response = await this.CallFrontdoorService(
                requestMethod: HttpMethod.Put,
                requestUri: this.Metadata.ResourceOperationUri,
                cancellationToken: this.CancellationToken,
                content: httpContent,
                addHeadersFunc: this.AddNotificationToken)
                .ConfigureAwait(continueOnCapturedContext: false))
            {
                this.Metadata.ServiceRequestId = response.Headers?.GetServiceRequestId(defaultValue: null);

                return await this.HandleResponse(response, deployment).ConfigureAwait(continueOnCapturedContext: false);
            }
        }

        /// <summary>
        /// Update metadata based on operation request.
        /// </summary>
        /// <param name="content">The request content.</param>
        private void UpdateMetadataOnOperationRequest(JToken content)
        {
            if (this.Metadata.DebugSetting != null && this.Metadata.DebugSetting.DetailLevel.HasFlag(DeploymentDebugDetailLevel.RequestContent))
            {
                this.Metadata.ResourceOperationRequest = new DeploymentOperationHttpMessage
                {
                    Content = content
                };
            }
        }

        /// <summary>
        /// Processes the resource language expressions.
        /// </summary>
        /// <param name="deployment">The deployment.</param>
        /// <param name="resource">The deployment resource.</param>
        private async Task ProcessResourceLanguageExpressions(IDeploymentEntity deployment, DeploymentResource resource)
        {
            if (this.Metadata.AreTemplateExpressionsEvaluated.HasValue &&
                this.Metadata.AreTemplateExpressionsEvaluated.Value)
            {
                return;
            }

            try
            {
                //if (resource.Properties != null)
                //{
                //    var references = await FetchResourcesFromSequencerActions(deployment, resource.References).ConfigureAwait(false);

                //    var sequencerActions = await this
                //        .GetDeploymentSequencerActions(deployment)
                //        .ConfigureAwait(continueOnCapturedContext: false);

                //    var evaluationContext = this.GetTemplateExpressionEvaluationContext(
                //        deployment: deployment,
                //        referenceValueLookup: references,
                //        copyContext: resource.CopyContext,
                //        sequencerActions: sequencerActions,
                //        hasSymbolicName: !string.IsNullOrWhiteSpace(resource.SymbolicName));

                //    resource.Properties = ExpressionsEngine.EvaluateLanguageExpressionsRecursive(
                //        root: resource.Properties,
                //        evaluationContext: evaluationContext,
                //        skipEvaluationPaths: resource.SkipEvaluationPaths);
                //}

                //if (this.Metadata.Resource.IsDeploymentType())
                //{
                //    this.UpdateNestedDeploymentRelativePath(deployment);
                //}

                this.Metadata.AreTemplateExpressionsEvaluated = true;
            }
            catch (Exception ex)
            {
                if (!(ex is TemplateException || ex is ExpressionException))
                {
                    throw;
                }

                this.Logger.LogDebug(
                    exception: ex,
                    operationName: "DeploymentResourceJob.ProcessResourceLanguageExpressions",
                    format: "Unable to process template language expressions for resource '{0}'.",
                    arg0: resource.GetFullyQualifiedResourceId());

                var lineNumber = resource.DeploymentResourceLineInfo != null ? resource.DeploymentResourceLineInfo.LineNumber : null;
                var linePosition = resource.DeploymentResourceLineInfo != null ? resource.DeploymentResourceLineInfo.LinePosition : null;
                var additionalInfo = Microsoft.WindowsAzure.ResourceStack.Common.Extensions.ObjectExtensions.AsArray(new TemplateViolationErrorInfo(
                    new TemplateErrorAdditionalInfo(lineNumber: lineNumber, positionNumber: linePosition)));

                var errorResponseMessage = new ErrorResponseMessage(
                    code: ErrorResponseCode.InvalidTemplate,
                    message: Microsoft.WindowsAzure.ResourceStack.Common.Instrumentation.LocalizationExtensions.ToLocalizedMessage(ErrorResponseMessages.InvalidTemplateLanguageExpression, resource.GetFullyQualifiedResourceId(), lineNumber, linePosition, ex.Message),
                    additionalInfo: additionalInfo);

                this.Metadata.ResourceOperationStatusCode = HttpStatusCode.BadRequest;
                this.Metadata.ResourceOperationStatusMessage = errorResponseMessage.ToJToken();

                throw new JobExecutionResultException(
                    status: JobExecutionStatus.Failed,
                    message: string.Format("Unable to process template language expressions for resource '{0}' at line '{1}' and column '{2}'. Current deployment operation failed. Exception: '{3}'.", resource.GetResourceName(), lineNumber, linePosition, ex.Message),
                    details: ex.GetType().Name,
                    nextMetadata: this.Metadata.ToJson());
            }
        }

        // TODO required for reference.
        ///// <summary>
        ///// Gets the deployment sequencer actions.
        ///// </summary>
        ///// <param name="deployment">The deployment.</param>
        //protected Task<SequencerAction[]> GetDeploymentSequencerActions(IDeploymentEntity deployment)
        //{
        //    return this
        //        .GetJobsDataProvider(this.Metadata.GetDeploymentLocation())
        //        .GetDeploymentJobActions(deployment: deployment);
        //}

        /// <summary>
        /// Update metadata based on operation response.
        /// </summary>
        /// <param name="response">The response message.</param>
        private async Task UpdateMetadataOnOperationResponse(HttpResponseMessage response)
        {
            if (this.Metadata.DebugSetting != null && this.Metadata.DebugSetting.DetailLevel.HasFlag(DeploymentDebugDetailLevel.ResponseContent))
            {
                this.Metadata.ResourceOperationResponse = new DeploymentOperationHttpMessage
                {
                    Content = await this
                        .GetOperationResponseContent(response.Content)
                        .ConfigureAwait(continueOnCapturedContext: false)
                };
            }
        }

        /// <summary>
        /// Gets operation response content as JToken.
        /// </summary>
        /// <param name="content">The response content.</param>
        private async Task<JToken> GetOperationResponseContent(HttpContent content)
        {
            if (content == null)
            {
                return null;
            }

            if (content.Headers.ContentType != null && Microsoft.WindowsAzure.ResourceStack.Common.Extensions.StringExtensions.EqualsInsensitively(content.Headers.ContentType.MediaType, "application/json"))
            {
                return await content
                    .TryReadAsJsonAsync<JToken>(new Newtonsoft.Json.JsonSerializer())
                    .ConfigureAwait(continueOnCapturedContext: false);
            }
            else
            {
                // TODO 
                return null;
                //return await content
                //    .TryReadAsStringAsync(
                //        eventSource: this.FrontdoorConfiguration.EventSource,
                //        rewindContentStream: true)
                //    .ConfigureAwait(continueOnCapturedContext: false);
            }
        }


        /// <summary>
        /// Handles the resource operation response.
        /// </summary>
        /// <param name="resourceResponse">The resource response.</param>
        /// <param name="deployment">The deployment.</param>
        private async Task<JobExecutionResult> HandleResponse(HttpResponseMessage resourceResponse, IDeploymentEntity deployment)
        {
            await this
                .UpdateMetadataOnOperationResponse(resourceResponse)
                .ConfigureAwait(continueOnCapturedContext: false);

            this.Metadata.IsAsyncNotificationEnabled = this.Metadata.IsAsyncNotificationEnabled ?? this.TryGetAsyncOperationCallbackStatus(resourceResponse);
            //this.TrySetOperationTimeout(response: resourceResponse);

            var azureAsyncHeaderUri = TryGetAzureAsyncOperationUri(resourceResponse.Headers);

            this.Metadata.ResourceAsyncOperationUri = azureAsyncHeaderUri ?? this.Metadata.ResourceAsyncOperationUri;
            this.Metadata.ResourceOperationStatusCode = resourceResponse.StatusCode;
            this.Metadata.ResourceOperationStatusMessage = null;

            if (!resourceResponse.IsSuccessStatusCode)
            {
                //this.Metadata.ResourceOperationStatusMessage = await ErrorResponseHandling
                //    .GenerateErrorResponseMessageFromResponse(
                //        eventSource: this.FrontdoorConfiguration.EventSource,
                //        configuration: this.HttpConfiguration,
                //        response: resourceResponse,
                //        rewindContentStream: false)
                //    .ConfigureAwait(continueOnCapturedContext: false);

                //if (HttpUtility.IsServerFailureRequest(resourceResponse.StatusCode) || resourceResponse.StatusCode == HttpUtility.TooManyRequestsStatusCode)
                //{
                //    // First, check if we're encountering consecutive failures, if so fail the resource deployment early.
                //    if (this.Metadata.UnsuccessAndThrottleCounter >= this.UnsuccessAndThrottleMaximumConsecutiveRetryLimit)
                //    {
                //        return new JobExecutionResult
                //        {
                //            Status = JobExecutionStatus.Failed,
                //            Message = string.Format("The deployment operation encountered sustained unsuccessful status code: '{0}'. Current deployment operation is failed.", resourceResponse.StatusCode),
                //            Details = resourceResponse.StatusCode.ToString(),
                //            NextMetadata = this.Metadata.ToJson()
                //        };
                //    }

                //    var providerRetryAfter = resourceResponse.Headers.RetryAfter != null
                //        ? resourceResponse.Headers.RetryAfter.Delta
                //        : null;

                //    // Note(fitopata): If RetryAfter value is not provided in the headers, calculate retry interval using exponential backoff strategy.
                //    var retryAfter = providerRetryAfter.HasValue
                //        ? this.GetResourceOperationRetryInterval(retryAfter: providerRetryAfter, isAsyncOperationCallbackEnabled: this.Metadata.IsAsyncNotificationEnabled ?? false, isThrottledResponse: resourceResponse.StatusCode == HttpUtility.TooManyRequestsStatusCode)
                //        : this.GetResourceOperationRetryIntervalWithExponentialBackoff();

                //    await this
                //        .TryUpdateDeploymentRetryAfter(deployment: deployment, retryAfter: retryAfter, providerRetryAfter: providerRetryAfter)
                //        .ConfigureAwait(continueOnCapturedContext: false);

                //    return new JobExecutionResult
                //    {
                //        Status = JobExecutionStatus.Postponed,
                //        Message = string.Format("The deployment operation completed with unsuccessful status code: '{0}'. Current deployment operation is postponed.", resourceResponse.StatusCode),
                //        Details = resourceResponse.StatusCode.ToString(),
                //        NextMetadata = this.Metadata.ToJson(),
                //        NextExecutionTime = DateTime.UtcNow.Add(retryAfter),
                //    };
                //}
                //else
                //{
                //    return new JobExecutionResult
                //    {
                //        Status = JobExecutionStatus.Failed,
                //        Message = string.Format("The deployment operation completed with unsuccessful status code: '{0}'. Current deployment operation is failed.", resourceResponse.StatusCode),
                //        Details = resourceResponse.StatusCode.ToString(),
                //        NextMetadata = this.Metadata.ToJson()
                //    };
                //}
            }

            if (resourceResponse.StatusCode == HttpStatusCode.Accepted)
            {
                var providerRetryAfter = resourceResponse.Headers.RetryAfter != null
                    ? resourceResponse.Headers.RetryAfter.Delta
                    : null;

                var resourceUri = UriTemplateEngine.GetResourceUri(
                    endpoint: this.Metadata.FrontdoorEndpoint,
                    subscriptionId: this.Metadata.Resource.SubscriptionId,
                    resourceGroupName: this.Metadata.Resource.ResourceGroupName,
                    resourceId: this.Metadata.Resource.GetUnqualifiedResourceId(),
                    apiVersion: this.Metadata.Resource.ApiVersion);

                var fullyQualifiedScopeResourceUri = GetFullyQualifiedScope(uri: resourceUri);
                //var managementGroupHierarchy = await this
                //   .GetManagementGroupDataProvider()
                //   .GetManagementGroupHierarchyFromCache(
                //        tenantId: RequestCorrelationContext.Current.TenantId,
                //        entityId: IResourceIdentifiableExtensions.GetSubscriptionId(fullyQualifiedScopeResourceUri) ?? IResourceIdentifiableExtensions.GetManagementGroupId(fullyQualifiedScopeResourceUri))
                //   .ConfigureAwait(continueOnCapturedContext: false);

                //this.LogPercentCompletionEvent(
                //    response: resourceResponse,
                //    resourceUri: resourceUri,
                //    managementGroupHierarchy: managementGroupHierarchy);

                this.Metadata.ResourceOperationUri = resourceResponse.Headers.Location ?? this.Metadata.ResourceOperationUri;
                this.Metadata.ResourceOperation = this.Metadata.ResourceAsyncOperationUri != null
                    ? ProvisioningOperation.AzureAsyncOperationWaiting
                    : ProvisioningOperation.Waiting;

                this.Logger.LogDebug(
                      operationName: "DeploymentResourceJob.HandleResponse",
                      format: "Async operation response received. Location header value is '{0}'; current ResourceOperationUri value is '{1}'",
                      arg0: resourceResponse.Headers.Location?.ToString() ?? string.Empty,
                      arg1: this.Metadata.ResourceOperationUri.ToString());

                //if (this.ShouldInlineRetry(providerRetryAfter))
                //{
                //    await this
                //        .TryUpdateDeploymentRetryAfter(deployment: deployment, retryAfter: providerRetryAfter, providerRetryAfter: providerRetryAfter)
                //        .ConfigureAwait(continueOnCapturedContext: false);

                //    this.Logger.LogDebug(
                //        operationName: "DeploymentResourceJob.HandleResponse",
                //        format: "Delaying the job inline for short retry-after. Delay period: '{0}'",
                //        arg0: providerRetryAfter.Value.ToString());

                //    await AsyncTimer.Stable.Delay(providerRetryAfter.Value).ConfigureAwait(continueOnCapturedContext: false);
                //    return new JobExecutionResult
                //    {
                //        Status = JobExecutionStatus.Postponed,
                //        Message = "Deployment operation is not completed yet. Continuation is requested.",
                //        NextMetadata = this.Metadata.ToJson(),
                //        NextExecutionTime = DateTime.UtcNow,
                //    };
                //}

                //var retryAfter = this.GetResourceOperationRetryInterval(providerRetryAfter, this.Metadata.IsAsyncNotificationEnabled ?? false);
                //await this
                //    .TryUpdateDeploymentRetryAfter(deployment: deployment, retryAfter: retryAfter, providerRetryAfter: providerRetryAfter)
                //    .ConfigureAwait(continueOnCapturedContext: false);

                return new JobExecutionResult
                {
                    Status = JobExecutionStatus.Postponed,
                    Message = "Deployment operation is not completed yet. Continuation is requested.",
                    NextMetadata = this.Metadata.ToJson(),
                    NextExecutionTime = DateTime.UtcNow.Add(TimeSpan.FromSeconds(5)),
                };
            }

            if (resourceResponse.StatusCode != HttpStatusCode.OK && resourceResponse.StatusCode != HttpStatusCode.Created)
            {
                //this.FrontdoorConfiguration.EventSource.ProviderError(
                //    operationName: "DeploymentResourceJob.HandleResourceResponse",
                //    format: "The unknown resource response code '{0}' is encountered, but 200 or 201 expected.",
                //    arg0: resourceResponse.StatusCode,
                //    providerNamespace: this.Metadata.Resource.GetResourceProviderNamespace(),
                //    resourceType: this.Metadata.Resource.GetResourceType());

                var errorResponseMessage = new ErrorResponseMessage(
                    code: ErrorResponseCode.ResourceDeploymentFailure,
                    message: Microsoft.WindowsAzure.ResourceStack.Common.Instrumentation.LocalizationExtensions.ToLocalizedMessage(ErrorResponseMessages.DeploymentInvalidResourceStatusCode, resourceResponse.StatusCode));

                this.Metadata.ResourceOperationStatusCode = HttpStatusCode.InternalServerError;
                this.Metadata.ResourceOperationStatusMessage = errorResponseMessage.ToJToken();

                return new JobExecutionResult
                {
                    Status = JobExecutionStatus.Failed,
                    Message = string.Format("The unknown resource response code '{0}' is encountered, but 200 or 201 expected. Current deployment operation is failed.", resourceResponse.StatusCode),
                    Details = resourceResponse.StatusCode.ToString(),
                    NextMetadata = this.Metadata.ToJson()
                };
            }

            return this.Metadata.Resource.IsAction
                ? await this.HandleResponseWithResourceAction(resourceResponse).ConfigureAwait(continueOnCapturedContext: false)
                : await this.HandleResponseWithResourceDefinition(resourceResponse, deployment).ConfigureAwait(continueOnCapturedContext: false);
        }

        /// <summary>
        /// Handles the response with resource action.
        /// </summary>
        /// <param name="resourceResponse">The resource response.</param>
        private async Task<JobExecutionResult> HandleResponseWithResourceAction(HttpResponseMessage resourceResponse)
        {
            // TODO(ilygre): Find a way to store action result instead of resource properties.
            this.Metadata.Resource.Properties = await this
                .TryReadResourceResponse<JToken>(resourceResponse)
                .ConfigureAwait(continueOnCapturedContext: false);

            if (this.Metadata.Resource.Properties == null)
            {
                var errorResponseMessage = new ErrorResponseMessage(
                    code: ErrorResponseCode.ResourceDeploymentFailure,
                    message: ErrorResponseMessages.DeploymentInvalidResourceProperties);

                this.Metadata.ResourceOperationStatusCode = HttpStatusCode.InternalServerError;
                this.Metadata.ResourceOperationStatusMessage = errorResponseMessage.ToJToken();

                return new JobExecutionResult
                {
                    Status = JobExecutionStatus.Failed,
                    Message = "The response for resource had empty or invalid content. Current deployment operation is failed.",
                    NextMetadata = this.Metadata.ToJson()
                };
            }

            return new JobExecutionResult
            {
                Status = JobExecutionStatus.Succeeded,
                Message = "The resource operation completed successfully.",
                NextMetadata = this.Metadata.ToJson()
            };
        }

        /// <summary>
        /// Handles the resource operation response.
        /// </summary>
        /// <param name="resourceResponse">The resource response.</param>
        /// <param name="deployment">The deployment.</param>
        private async Task<JobExecutionResult> HandleResponseWithResourceDefinition(HttpResponseMessage resourceResponse, IDeploymentEntity deployment)
        {
            var resourceDefinition = await this
                .TryReadResourceResponse<ResourceProxyDefinition>(resourceResponse)
                .ConfigureAwait(continueOnCapturedContext: false);

            if (resourceDefinition == null)
            {
                var errorResponseMessage = new ErrorResponseMessage(
                    code: ErrorResponseCode.ResourceDeploymentFailure,
                    message: ErrorResponseMessages.DeploymentInvalidResourceProperties);

                this.Metadata.ResourceOperationStatusCode = HttpStatusCode.InternalServerError;
                this.Metadata.ResourceOperationStatusMessage = errorResponseMessage.ToJToken();

                return new JobExecutionResult
                {
                    Status = JobExecutionStatus.Failed,
                    Message = "The response for resource had empty or invalid content. Current deployment operation is failed.",
                    NextMetadata = this.Metadata.ToJson()
                };
            }

            var provisioningState =
                Microsoft.WindowsAzure.ResourceStack.Common.Extensions.EnumExtensions.ToEnum<DeploymentsProvisioningState>( resourceDefinition
                    .ToArmResourceProxyDefinition()
                    .GetResourceProvisioningState().ToString());

            this.Logger.LogDebug(
                operationName: "DeploymentResourceJob.HandleResponseWithResourceDefinition",
                format: "The response for resource Id '{0}' has provisioning state '{1}'.",
                arg0: this.Metadata.Resource.GetFullyQualifiedResourceId(),
                arg1: provisioningState.ToString());

            if (provisioningState != ProvisioningState.NotSpecified)
            {
                if (provisioningState == ProvisioningState.Succeeded)
                {
                    this.Metadata.Resource = DeploymentResource.FromResourceProxyDefinition(
                        definition: resourceDefinition,
                        baseResource: this.Metadata.Resource);

                    return await this
                        .CompleteResourceProvisioning(resourceDefinition, deployment)
                        .ConfigureAwait(continueOnCapturedContext: false);
                }
                else if (provisioningState == ProvisioningState.Failed || provisioningState == ProvisioningState.Canceled)
                {
                    var errorResponseMessage = new OperationResult
                    {
                        Status = provisioningState.ToString(),
                        Error = new ExtendedErrorInfo
                        {
                            Code = ErrorResponseCode.ResourceDeploymentFailure.ToString(),
                            Message = ErrorResponseMessages.DeploymentInvalidResourceProvisioningState
                        }
                    };

                    var asyncOperationResult = await this
                        .TryGetAsyncOperationResult(this.Metadata.ResourceAsyncOperationUri)
                        .ConfigureAwait(continueOnCapturedContext: false);

                    if (asyncOperationResult != null && asyncOperationResult.Error != null)
                    {
                        errorResponseMessage.Error.Details = Microsoft.WindowsAzure.ResourceStack.Common.Extensions.ObjectExtensions.AsArray(asyncOperationResult.Error);
                    }

                    this.Metadata.ResourceOperationStatusCode = HttpStatusCode.Conflict;
                    this.Metadata.ResourceOperationStatusMessage = errorResponseMessage.ToJToken();

                    return new JobExecutionResult
                    {
                        Status = JobExecutionStatus.Failed,
                        Message = string.Format("The resource operation completed with terminal provisioning state '{0}'. Current deployment operation is failed.", provisioningState),
                        NextMetadata = this.Metadata.ToJson(),
                    };
                }
                else
                {
                    this.Metadata.ResourceOperation = this.Metadata.ResourceAsyncOperationUri != null
                        ? ProvisioningOperation.AzureAsyncOperationWaiting
                        : ProvisioningOperation.Waiting;

                    var providerRetryAfter = resourceResponse.Headers.RetryAfter != null
                        ? resourceResponse.Headers.RetryAfter.Delta
                        : null;

                    this.Metadata.ResourceOperationUri = UriTemplateEngine.GetResourceUri(
                        endpoint: this.Metadata.FrontdoorEndpoint,
                        subscriptionId: this.Metadata.Resource.SubscriptionId,
                        resourceGroupName: this.Metadata.Resource.ResourceGroupName,
                        resourceId: this.Metadata.Resource.GetUnqualifiedResourceId(),
                        apiVersion: this.Metadata.Resource.ApiVersion);

                    //if (this.ShouldInlineRetry(providerRetryAfter))
                    //{
                    //    await this
                    //        .TryUpdateDeploymentRetryAfter(deployment: deployment, retryAfter: providerRetryAfter, providerRetryAfter: providerRetryAfter)
                    //        .ConfigureAwait(continueOnCapturedContext: false);

                    //    this.Logger.LogDebug(
                    //        operationName: "DeploymentResourceJob.HandleResponseWithResourceDefinition",
                    //        format: "Delaying the job inline for short retry-after. Delay period: '{0}'",
                    //        arg0: providerRetryAfter.Value.ToString());

                    //    await AsyncTimer.Stable.Delay(providerRetryAfter.Value).ConfigureAwait(continueOnCapturedContext: false);
                    //    return new JobExecutionResult
                    //    {
                    //        Status = JobExecutionStatus.Postponed,
                    //        Message = string.Format("The resource operation completed with non terminal provisioning state '{0}'. Current deployment operation is postponed.", provisioningState),
                    //        NextMetadata = this.Metadata.ToJson(),
                    //        NextExecutionTime = DateTime.UtcNow,
                    //    };
                    //}

                    //var retryAfter = this.GetResourceOperationRetryInterval(providerRetryAfter, this.Metadata.IsAsyncNotificationEnabled ?? false);
                    //await this
                    //    .TryUpdateDeploymentRetryAfter(deployment: deployment, retryAfter: retryAfter, providerRetryAfter: providerRetryAfter)
                    //    .ConfigureAwait(continueOnCapturedContext: false);

                    return new JobExecutionResult
                    {
                        Status = JobExecutionStatus.Postponed,
                        Message = string.Format("The resource operation completed with non terminal provisioning state '{0}'. Current deployment operation is postponed.", provisioningState),
                        NextMetadata = this.Metadata.ToJson(),
                        NextExecutionTime = DateTime.UtcNow.Add(TimeSpan.FromSeconds(5)),
                    };
                }
            }
            else
            {
                this.Metadata.Resource = DeploymentResource.FromResourceProxyDefinition(
                    definition: resourceDefinition,
                    baseResource: this.Metadata.Resource);

                return await this
                    .CompleteResourceProvisioning(resourceDefinition, deployment)
                    .ConfigureAwait(continueOnCapturedContext: false);
            }
        }

        /// <summary>
        /// Gets the async operation status.
        /// </summary>
        /// <param name="resourceAsyncOperationUri">The resource async operation URI.</param>
        private async Task<OperationResult> TryGetAsyncOperationResult(Uri resourceAsyncOperationUri)
        {
            try
            {
                if (resourceAsyncOperationUri != null)
                {
                    using (var response = await this
                        .CallFrontdoorService(
                            requestMethod: HttpMethod.Get,
                            requestUri: resourceAsyncOperationUri,
                            cancellationToken: this.CancellationToken)
                        .ConfigureAwait(continueOnCapturedContext: false))
                    {
                        if (response.StatusCode == HttpStatusCode.OK)
                        {
                            return await this
                                .TryReadResourceResponse<OperationResult>(response)
                                .ConfigureAwait(continueOnCapturedContext: false);
                        }
                    }
                }
            }
            catch (Exception ex)
            {
                // TODO
            }

            return null;
        }

        /// <summary>
        /// Tries to read resource definition.
        /// </summary>
        /// <typeparam name="T">response type</typeparam>
        /// <param name="resourceResponse">The resource response.</param>
        private async Task<T> TryReadResourceResponse<T>(HttpResponseMessage resourceResponse)
        {
            try
            {
                return await resourceResponse.Content
                    .ReadAsAsync<T>(JsonObjectTypeFormatters)
                    .ConfigureAwait(continueOnCapturedContext: false);
            }
            catch (Exception ex)
            {
                if (ex is ArgumentException || ex is FormatException || ex is JsonException || ex is UnsupportedMediaTypeException)
                {
                    this.Logger.LogDebug(
                        exception: ex,
                        operationName: "DeploymentResourceJob.TryReadResourceResponse",
                        message: "Unable to read resource response. The Json exception is encountered.");

                    return default(T);
                }

                throw;
            }
        }


        /// <summary>
        /// Completes the resource provisioning.
        /// </summary>
        /// <param name="resourceDefinition">The resource definition.</param>
        /// <param name="deployment">The deployment.</param>
        private Task<JobExecutionResult> CompleteResourceProvisioning(ResourceProxyDefinition resourceDefinition, IDeploymentEntity deployment)
        {
            // NOTE(ilygre): Check if provisioning operation corresponds to template resource.
            //if (this.Metadata.Resource.IsTemplateResource)
            //{
            //    var resourceTypeRegistration = await this
            //        .GetRegistrationCacheProvider()
            //        .FindMostSpecificRegistration(
            //            subscriptionId: this.Metadata.Resource.SubscriptionId,
            //            resourceProviderNamespace: this.Metadata.Resource.GetResourceProviderNamespace(),
            //            resourceType: this.Metadata.Resource.GetResourceType(),
            //            location: this.Metadata.Resource.Location,
            //            apiVersion: this.Metadata.Resource.ApiVersion)
            //        .ConfigureAwait(continueOnCapturedContext: false);

            //    // NOTE(ilygre): Check if template resource is ARM registered resource and should be visible in cache.
            //    if (resourceTypeRegistration != null && !resourceTypeRegistration.IsProxyOnly)
            //    {
            //        // Note(vinarvek): To improve TDP number try inline retry if resource is not available in cache instead of job postponed for 5sec.
            //        this.Metadata.ResourceCacheWaitTimeout = DateTime.UtcNow.Add(this.ResourceCacheWaitTimeout);
            //        this.Metadata.ResourceOperation = ProvisioningOperation.ResourceCacheWaiting;

            //        this.Logger.LogDebug(
            //           operationName: "DeploymentResourceJob.CompleteResourceProvisioning",
            //           message: "Resource operation state set to ResourceCacheWaiting, Check if the resource is present in cache");

            //        return await this.HandleResourceOperationState(deployment).ConfigureAwait(continueOnCapturedContext: false);
            //    }
            //}

            return Task.FromResult<JobExecutionResult>(new JobExecutionResult
            {
                Status = JobExecutionStatus.Succeeded,
                Message = "The resource operation completed successfully.",
                NextMetadata = this.Metadata.ToJson()
            });
        }

        /// <summary>
        /// Gets the scope of request from non-provider request url.
        /// </summary>
        /// <param name="uri">The uri that will be parsed.</param>
        public static string GetFullyQualifiedScope(Uri uri)
        {
            var segments = uri.LocalPath.Split(new char[] { '/' }, StringSplitOptions.RemoveEmptyEntries);
            var scope = segments.Length % 2 == 0 ? Microsoft.WindowsAzure.ResourceStack.Common.Extensions.StringExtensions.ConcatStrings(segments, "/") : Microsoft.WindowsAzure.ResourceStack.Common.Extensions.StringExtensions.ConcatStrings(segments.Take(segments.Length - 1), "/");
            return string.Format("/{0}", scope);
        }

        private Uri TryGetAzureAsyncOperationUri(HttpResponseHeaders headers)
        {
            Uri azureAsyncOperationUri = null;
            if (headers.Contains(RequestCorrelationContext.HeaderAzureAsyncOperation))
            {
                if (headers.GetValues(RequestCorrelationContext.HeaderAzureAsyncOperation).Count() != 1 ||
                    !Uri.TryCreate(headers.GetValues(RequestCorrelationContext.HeaderAzureAsyncOperation).Single(), UriKind.Absolute, out azureAsyncOperationUri))
                {
                    //""
                }
            }

            return azureAsyncOperationUri;
        }

        /// <summary>
        /// Gets the URI from the azure async operation response header.
        /// </summary>
        /// <param name="headers">The HTTP response headers.</param>
        /// <param name="providerNamespace">The resource provider namespace.</param>
        /// <param name="resourceType">The resource type.</param>
        /// <param name="eventSource">The event source.</param>
        public static AsyncOperationCallbackStatus TryGetAzureAsyncOperationCallbackStatus(HttpResponseHeaders headers)
        {
            var azureAsyncOperationCallback = AsyncOperationCallbackStatus.Disabled;
            if (headers.Contains(RequestCorrelationContext.HeaderAzureAsyncNotification))
            {
                if (headers.GetValues(RequestCorrelationContext.HeaderAzureAsyncNotification).Count() != 1 ||
                    !Enum.TryParse(headers.GetValues(RequestCorrelationContext.HeaderAzureAsyncNotification).Single(), out azureAsyncOperationCallback))
                {
                    //
                }
            }

            return azureAsyncOperationCallback;
        }

        //
        // Summary:
        //     The async notification status
        public enum AsyncOperationCallbackStatus
        {
            //
            // Summary:
            //     The async notification status is not specified
            NotSpecified,
            //
            // Summary:
            //     The provider has enabled notification based polling
            Enabled,
            //
            // Summary:
            //     The Provider has not enabled notification based polling
            Disabled
        }

        /// <summary>
        /// Try to get the async notification status.
        /// </summary>
        /// <param name="response">The response.</param>
        /// <returns>Whether or not the async notification should be enabled in this job.</returns>
        private bool TryGetAsyncOperationCallbackStatus(HttpResponseMessage response)
        {
            var azureAsyncOperationCallbackStatus = TryGetAzureAsyncOperationCallbackStatus(response.Headers);
            // FrontdoorConfiguration.AllowedProvidersForAsyncCallback.ContainsInsensitively(this.Metadata.Resource.GetResourceProviderNamespace())
            return azureAsyncOperationCallbackStatus == AsyncOperationCallbackStatus.Enabled;
        }

        /// <summary>
        /// Get the timeout value for the asynchronous action.
        /// </summary>
        /// <param name="resourceOperation">The resource operation.</param>
        private async Task<TimeSpan> GetOperationTimeout(ProvisioningOperation resourceOperation)
        {
            if (this.Metadata.AsyncOperationTimeout.HasValue)
            {
                this.Logger.LogDebug(
                    operationName: "DeploymentResourceJob.GetOperationTimeout",
                    format: "Async operation timeout was set in job. Timeout value: '{0}'",
                    arg0: this.Metadata.AsyncOperationTimeout.Value.ToString());

                return this.Metadata.AsyncOperationTimeout.Value;
            }

            //var resourceTypeRegistration = await this
            //    .GetRegistrationCacheProvider()
            //    .FindMostSpecificRegistration(
            //        subscriptionId: this.Metadata.Resource.SubscriptionId,
            //        resourceProviderNamespace: this.Metadata.Resource.GetResourceProviderNamespace(),
            //        resourceType: this.Metadata.Resource.GetResourceType(),
            //        location: this.Metadata.Resource.Location,
            //        apiVersion: this.Metadata.Resource.ApiVersion)
            //    .ConfigureAwait(continueOnCapturedContext: false);

            //if (resourceTypeRegistration != null)
            //{
            //    if (resourceTypeRegistration.AsyncTimeoutRules.CoalesceEnumerable().Any())
            //    {
            //        this.Logger.LogDebug(
            //            operationName: "DeploymentResourceJob.GetOperationTimeout",
            //            format: "Found following async timeout rules in the manifest '{0}'.",
            //            arg0: resourceTypeRegistration.AsyncTimeoutRules.Select(rule => rule.ActionName).ConcatStrings(","));
            //    }

            //    var asyncActionTimeoutRule = resourceTypeRegistration.GetAsyncActionTimeoutRule(
            //        providerNamespace: this.Metadata.Resource.GetResourceProviderNamespace(),
            //        resourceType: this.Metadata.Resource.GetResourceType(),
            //        actionVerb: ActionVerb.Write);

            //    if (asyncActionTimeoutRule != null)
            //    {
            //        return asyncActionTimeoutRule.Timeout;
            //    }
            //}
            //else
            //{
            //    this.Logger.LogWarning(
            //        operationName: "DeploymentResourceJob.GetOperationTimeout",
            //        format: "Could not find resource type registration for subscriptionId '{0}', provider namespace '{1}', resource type '{2}', location '{3}', apiVersion '{4}'",
            //        arg0: this.Metadata.Resource.SubscriptionId,
            //        arg1: this.Metadata.Resource.GetResourceProviderNamespace(),
            //        arg2: this.Metadata.Resource.GetResourceType(),
            //        arg3: this.Metadata.Resource.Location,
            //        arg4: this.Metadata.Resource.ApiVersion);
            //}

            //this.Logger.LogDebug(
            //    operationName: "DeploymentResourceJob.GetOperationTimeout",
            //    format: "Could not find async timeout rule for subscriptionId '{0}', provider namespace '{1}', resource type '{2}', location '{3}', apiVersion '{4}'. Returning default timeout.",
            //    arg0: this.Metadata.Resource.SubscriptionId,
            //    arg1: this.Metadata.Resource.GetResourceProviderNamespace(),
            //    arg2: this.Metadata.Resource.GetResourceType(),
            //    arg3: this.Metadata.Resource.Location,
            //    arg4: this.Metadata.Resource.ApiVersion);

            return this.DefaultResourceOperationTimeout;
        }


        /// <summary>
        /// Gets the current deployment sequence.
        /// </summary>
        protected async Task<IDeploymentEntity> GetCurrentDeploymentSequence()
        {
            var deployment = await this.GetLatestDeploymentSequence().ConfigureAwait(continueOnCapturedContext: false);
            if (deployment.IsDifferentSequence(this.Metadata.SequenceId))
            {
                var errorMessage = deployment != null
                    ? "Terminating deployment job: deployment sequence mismatch."
                    : "Terminating deployment job: deployment not found.";

                throw new JobExecutionResultException(status: JobExecutionStatus.Failed, message: errorMessage);
            }

            return deployment;
        }

        /// <summary>
        /// Gets the latest deployment sequence.
        /// </summary>
        protected async Task<IDeploymentEntity> GetLatestDeploymentSequence()
        {
            var deployment = await this.FindLatestDeploymentSequence().ConfigureAwait(continueOnCapturedContext: false);

            // NOTE(ilygre): the job is created before deployment object is saved, so we should wait for at least 10 minutes if current sequence is orphan.
            if (deployment.IsDifferentSequence(this.Metadata.SequenceId) && this.BackgroundJob.CreatedTime.Add(TimeSpan.FromMinutes(10)) > DateTime.UtcNow)
            {
                var errorMessage = deployment != null
                    ? "Postponing deployment job: deployment sequence mismatch."
                    : "Postponing deployment job: deployment not found.";

                throw new JobExecutionResultException(
                    status: JobExecutionStatus.Postponed,
                    message: errorMessage,
                    nextExecutionTime: DateTime.UtcNow.AddSeconds(this.BackgroundJob.TotalExecutedCount * this.BackgroundJob.TotalExecutedCount));
            }

            return deployment;
        }


        /// <summary>
        /// Gets the current deployment sequence.
        /// </summary>
        protected async Task<IDeploymentEntity> FindLatestDeploymentSequence()
        {
            if (this.Metadata.IsTenantDeployment)
            {
                return await this
                    .GetDeploymentDataProvider(location: this.Metadata.GetDeploymentLocation())
                    .FindTenantDeployment(
                        tenantId: this.Metadata.TenantId,
                        managementGroupId: this.Metadata.ManagementGroupId,
                        deploymentName: this.Metadata.DeploymentName)
                    .ConfigureAwait(continueOnCapturedContext: false);
            }

            return await this
                .GetDeploymentDataProvider(location: this.Metadata.GetDeploymentLocation())
                .FindDeployment(
                    subscriptionId: this.Metadata.SubscriptionId,
                    resourceGroupName: this.Metadata.ResourceGroupName,
                    deploymentName: this.Metadata.DeploymentName)
                .ConfigureAwait(continueOnCapturedContext: false);
        }

        private DeploymentDataProvider GetDeploymentDataProvider(string location)
        {
        }

        /// <summary>
        /// Calls the front door service.
        /// </summary>
        /// <param name="requestMethod">The request method.</param>
        /// <param name="requestUri">The request Uri.</param>
        /// <param name="cancellationToken">The cancellation token.</param>
        /// <param name="content">The HTTP message content.</param>
        /// <param name="addHeadersFunc">Callback to add headers.</param>
        protected async Task<HttpResponseMessage> CallFrontdoorService(HttpMethod requestMethod, Uri requestUri, CancellationToken cancellationToken, HttpContent content = null, Action<HttpRequestHeaders> addHeadersFunc = null)
        {
            var frontdoorClient = new LocalFrontdoorRequestClient(this.FrontdoorConfiguration, this.HttpConfiguration);

            return await frontdoorClient
                .SendAsync(
                    method: requestMethod,
                    requestUri: requestUri,
                    cancellationToken: cancellationToken,
                    content: content,
                    addHeadersFunc: this.AddDefaultHeadersForBackgroundJobs + addHeadersFunc)
                .ConfigureAwait(continueOnCapturedContext: false);
        }

        /// <summary>
        /// Adds a notification token header to the headers collection.
        /// </summary>
        /// <param name="httpHeaders">The HTTP headers collection.</param>
        private void AddNotificationToken(HttpRequestHeaders httpHeaders)
        {
            if (FrontdoorConfiguration.AllowedProvidersForAsyncCallback.ContainsInsensitively(this.Metadata.Resource.GetResourceProviderNamespace()))
            {
                var notificationTokenDefinition = new AsyncOperationCallbackTokenDefinition
                {
                    JobDefinitions = new List<AsyncOperationCallbackJobDefinition>
                    {
                        new AsyncOperationCallbackJobDefinition
                        {
                            JobId = this.BackgroundJob.JobId,
                            JobIdPrefix = this.BackgroundJob.JobId,
                            JobPartition = this.BackgroundJob.JobPartition,
                            Location = this.BackgroundJob.CurrentExecutionAffinity
                        },
                    },
                    ResourceType = this.Metadata.Resource.GetFullyQualifiedResourceType()
                };

                var notificationUri = UriTemplateEngine.CreateAsyncOperationCallbackUri(
                    endpoint: this.Metadata.FrontdoorEndpoint,
                    asyncOperationCallbackTokenData: notificationTokenDefinition.ToJson().EncodeToBase64String(),
                    apiVersion: FrontdoorConstants.ApiVersion20180201);

                httpHeaders.AddAsyncCallbackTokenHeader(notificationUri.AbsoluteUri);
            }
        }

        /// <summary>
        /// Creates an HttpContent object serialized as JSON, returning null if the object is null.
        /// </summary>
        /// <typeparam name="T">The type of the object to serialize.</typeparam>
        /// <param name="body">The object to serialize.</param>
        /// <param name="mediaTypeFormatter">The media type formatter.</param>
        public static HttpContent CreateJsonContent<T>(T body, MediaTypeFormatter mediaTypeFormatter = null)
            where T : class
            => (body != null) ? new ObjectContent<T>(body, mediaTypeFormatter ?? HttpHelper.JsonMediaTypeFormatter) : null;

        /// <summary>
        /// Fetches a set of resource references from deployment sequencer job metadata.
        /// </summary>
        /// <param name="deployment">The deployment.</param>
        /// <param name="references">The resource references.</param>
        protected async Task<IReadOnlyDictionary<DeploymentResourceReference, JToken>> FetchResourcesFromSequencerActions(IDeploymentEntity deployment, IEnumerable<DeploymentResourceReference> references)
        {
            var results = await references
                .CoalesceEnumerable()
                .Select(async reference => await FetchResourceReferenceValueFromSequencerAction(deployment, reference).ConfigureAwait(false))
                .WhenAllForAwait()
                .ConfigureAwait(false);

            return results.ToDictionary(
                r => r.reference,
                r => r.value,
                DeploymentResourceReference.EqualityComparer.Instance);
        }


        /// <summary>
        /// Fetches a resource reference from deployment sequencer job metadata.
        /// </summary>
        /// <param name="deployment">The deployment.</param>
        /// <param name="reference">The resource reference.</param>
        private async Task<(DeploymentResourceReference reference, JToken value, bool hasSymbolicName)> FetchResourceReferenceValueFromSequencerAction(IDeploymentEntity deployment, DeploymentResourceReference reference)
        {
            var action = await this
                .GetJobsDataProvider(this.Metadata.GetDeploymentLocation())
                .GetDeploymentJobAction(
                    deployment: deployment,
                    actionId: DeploymentEngine.GetDeploymentJobSequencerId(reference))
                .ConfigureAwait(continueOnCapturedContext: false);

            if (action is null)
            {
                throw new InvalidOperationException($"The referenced deployment operation '{reference.ToJson()}' could not be found.");
            }

            var output = TryGetResourceReferenceFromSequencerAction(action);
            if (!output.HasValue)
            {
                // Note(antmarti): We don't expect to hit this in practice, because DeploymentEngine.GetDeploymentJobSequencerId will only
                // ever return a DeploymentResourceJob or DeploymentExtensibleResourceJob sequencer action - in both cases, TryGetResourceReferenceFromSequencerAction
                // will return non-null values.
                throw new InvalidOperationException($"Unable to fetch resource reference from callback {action.Callback}");
            }

            return output.Value;
        }


        /// <summary>
        /// Gets resource reference metadata from deployment sequencer job metadata.
        /// </summary>
        /// <param name="action">The sequencer action</param>
        private static (DeploymentResourceReference reference, JToken value, bool hasSymbolicName)? TryGetResourceReferenceFromSequencerAction(SequencerAction action)
        {
            if (action.Result != SequencerActionResult.Succeeded)
            {
                throw new InvalidOperationException($"The referenced deployment operation '{action.ActionId}' is not completed successfully (current state: '{action.Result}'");
            }

            if (action.Callback.EqualsOrdinalInsensitively("DeploymentResourceJob"))
            {
                var metadata = action.Metadata.FromJson<DeploymentResourceJobMetadata>();

                var hasSymbolicName = !string.IsNullOrWhiteSpace(metadata.Resource.SymbolicName);
                return (reference: metadata.Resource, value: metadata.Resource.ToJToken(), hasSymbolicName: hasSymbolicName);
            }

            //if (action.Callback.EqualsOrdinalInsensitively(DeploymentExtensibleResourceJob.JobName))
            //{
            //    var metadata = action.Metadata.FromJson<DeploymentExtensibleResourceJobMetadata>();

            //    var reference = DeploymentResourceReference.ForExtensibleResource(metadata.Resource.SymbolicName);

            //    var value = new JObject
            //    {
            //        ["Properties"] = metadata.ReturnedProperties.DeepClone(),
            //    };

            //    // Note(antmarti): Extensible resources always use symbolic names.
            //    return (reference: reference, value: value, hasSymbolicName: true);
            //}

            // Note(antmarti): This method is called in DeploymentLastJob with the full set of job sequencer actions
            // - including those not corresponding to DeploymentResourceJob / DeploymentExtensibleResourceJob instances.
            // We don't want to throw an exception in this case, so return null and let the caller decide.
            return null;
        }


        /// <summary>
        /// Gets the template expression evaluation helper.
        /// </summary>
        /// <param name="deployment">The deployment.</param>
        /// <param name="referenceValueLookup">The reference value lookup.</param>
        /// <param name="copyContext">The copy context.</param>
        /// <param name="sequencerActions">sequencer actions.</param>
        /// <param name="hasSymbolicName">Indicates if resource has symbolic name.</param>
        protected ExpressionEvaluationContext GetTemplateExpressionEvaluationContext(
            IDeploymentEntity deployment,
            IReadOnlyDictionary<DeploymentResourceReference, JToken> referenceValueLookup,
            TemplateCopyContext copyContext,
            SequencerAction[] sequencerActions,
            bool hasSymbolicName)
        {
            var symbolicNameLookup = hasSymbolicName ? this.GetSymbolicNameLookupFromSequencerActions(sequencerActions) : null;

            //var extensibleResources = this.GetExtensibleResourcesFromSequencerActions(sequencerActions);

            return DeploymentEngineUtils
                .GetTemplateExpressionEvaluationContext(
                    deployment: deployment,
                    symbolicNameLookup: symbolicNameLookup,
                    referenceValueLookup: referenceValueLookup,
                    extensibleResources: new Dictionary<string, ExtensibleResource>(),
                    copyContext: copyContext,
                    eventSource: new DeploymentsEventSource(eventSource: this.FrontdoorConfiguration.EventSource));
        }

        /// <summary>
        /// Gets resource symbolic name lookup from sequencer actions.
        /// </summary>
        /// <param name="resourceSequencerActions">The sequencer actions.</param>
        private IReadOnlyDictionary<string, DeploymentResource> GetSymbolicNameLookupFromSequencerActions(SequencerAction[] resourceSequencerActions)
        {
            var resourceActionsMetadata = resourceSequencerActions
                .Where(resourceSequencerAction => resourceSequencerAction.Callback.EqualsOrdinalInsensitively(nameof(DeploymentResourceJob)))
                .Select(resourceSequencerAction => resourceSequencerAction.Metadata.FromJson<DeploymentResourceJobMetadata>())
                .Where(metadata => metadata.Resource.IsTemplateResource);

            // Symbolic names are case sensitive.
            return resourceActionsMetadata.ToDictionary(
                keySelector: metadata => metadata.Resource.SymbolicName,
                elementSelector: metadata => metadata.Resource,
                comparer: CoreConstants.SymbolicNameComparer);
        }
    }
}
