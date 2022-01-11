using System;
using Azure.Deployments.Core.EventSources;
using Microsoft.WindowsAzure.ResourceStack.Common.EventSources;

namespace DeploymentEngine.Controllers
{
    public class BackgroundEventSource : IBackgroundJobsEventSource
    {
        public void Critical(string operationName, string message, Exception exception = null)
        {
            return;
        }

        public void Debug(string operationName, string message, Exception exception = null)
        {
            return;
        }

        public void Error(string operationName, string message, Exception exception = null)
        {
            return;
        }

        public void JobCritical(string subscriptionId, string correlationId, string principalOid, string principalPuid, string tenantId, string operationName, string jobPartition, string jobId, string message, string exception, string organizationId, string activityVector, string realPuid, string altSecId, string additionalProperties)
        {
            return;
        }

        public void JobDebug(string subscriptionId, string correlationId, string principalOid, string principalPuid, string tenantId, string operationName, string jobPartition, string jobId, string message, string exception, string organizationId, string activityVector, string realPuid, string altSecId, string additionalProperties)
        {
            return;
        }

        public void JobDefinition(string jobPartition, string jobId, string version, string callback, string location, string locationsAffinity, string flags, string state, string executionState, string startTime, string endTime, int repeatCount, long repeatInterval, string repeatUnit, string repeatSchedule, int currentRepeatCount, int retryCount, long retryInterval, string retryUnit, int currentRetryCount, int currentExecutionCount, string timeout, string retention, string nextExecutionTime, string lastExecutionTime, string lastExecutionStatus, string createdTime, string changedTime, string subscriptionId, string correlationId, string principalOid, string principalPuid, string tenantId, int totalSucceededCount, int totalCompletedCount, int totalFailedCount, int totalFaultedCount, int totalPostponedCount, string parentJobCompletionTrigger, string organizationId, string activityVector, string realPuid, string altSecId, string additionalProperties)
        {
            return;
        }

        public void JobDispatchingError(string operationName, string jobPartition, string jobId, string message, string exception, string subscriptionId, string correlationId, string principalOid, string principalPuid, string tenantId, string organizationId, string activityVector, string realPuid, string altSecId, string additionalProperties)
        {
            return;
        }

        public void JobError(string subscriptionId, string correlationId, string principalOid, string principalPuid, string tenantId, string operationName, string jobPartition, string jobId, string message, string exception, string organizationId, string activityVector, string realPuid, string altSecId, string additionalProperties)
        {
            return;
        }

        public void JobHistory(string jobPartition, string jobId, string callback, string startTime, string endTime, string executionTimeInMilliseconds, string executionDelayInMilliseconds, string executionIntervalInMilliseconds, string executionStatus, string executionMessage, string executionDetails, string nextExecutionTime, string subscriptionId, string correlationId, string principalOid, string principalPuid, string tenantId, string dequeueCount, string advanceVersion, string triggerId, string messageId, string state, string organizationId, string activityVector, string realPuid, string altSecId, string additionalProperties, string jobDurabilityLevel)
        {
            return;
        }

        public void JobOperation(string subscriptionId, string correlationId, string principalOid, string principalPuid, string tenantId, string operationName, string jobPartition, string jobId, string message, string exception, string organizationId, string activityVector, string realPuid, string altSecId, string additionalProperties)
        {
            return;
        }

        public void JobWarning(string subscriptionId, string correlationId, string principalOid, string principalPuid, string tenantId, string operationName, string jobPartition, string jobId, string message, string exception, string organizationId, string activityVector, string realPuid, string altSecId, string additionalProperties)
        {
            return;
        }

        public void ProviderDebug(string providerNamespace, string resourceType, string operationName, string message, Exception exception = null)
        {
            return;
        }

        public void ProviderError(string providerNamespace, string resourceType, string operationName, string message, Exception exception = null)
        {
            return;
        }

        public void StorageOperation(string subscriptionId, string correlationId, string principalOid, string principalPuid, string tenantId, string operationName, string accountName, string resourceType, string resourceName, string clientRequestId, string operationStatus, long durationInMilliseconds, string exceptionMessage, int requestsStarted, int requestsCompleted, int requestsTimedout, string requestsDetails, string organizationId, string activityVector, long ingressBytes, long egressBytes, string realPuid, string altSecId, string additionalProperties)
        {
            return;
        }

        public void StorageRequestEndWithClientFailure(string subscriptionId, string correlationId, string principalOid, string principalPuid, string tenantId, string operationName, string accountName, string resourceType, string resourceName, string clientRequestId, string serviceRequestId, long durationInMilliseconds, int httpStatusCode, string exceptionMessage, string errorCode, string errorMessage, string organizationId, string activityVector, string realPuid, string altSecId, string additionalProperties)
        {
            return;
        }

        public void StorageRequestEndWithServerFailure(string subscriptionId, string correlationId, string principalOid, string principalPuid, string tenantId, string operationName, string accountName, string resourceType, string resourceName, string clientRequestId, string serviceRequestId, long durationInMilliseconds, int httpStatusCode, string exceptionMessage, string errorCode, string errorMessage, string organizationId, string activityVector, string realPuid, string altSecId, string additionalProperties)
        {
            return;
        }

        public void StorageRequestEndWithSuccess(string subscriptionId, string correlationId, string principalOid, string principalPuid, string tenantId, string operationName, string accountName, string resourceType, string resourceName, string clientRequestId, string serviceRequestId, long durationInMilliseconds, int httpStatusCode, string exceptionMessage, string errorCode, string errorMessage, string organizationId, string activityVector, string realPuid, string altSecId, string additionalProperties)
        {
            return;
        }

        public void StorageRequestStart(string subscriptionId, string correlationId, string principalOid, string principalPuid, string tenantId, string operationName, string accountName, string resourceType, string resourceName, string clientRequestId, string serviceRequestId, long durationInMilliseconds, int httpStatusCode, string exceptionMessage, string errorCode, string errorMessage, string organizationId, string activityVector, string realPuid, string altSecId, string additionalProperties)
        {
            return;
        }

        public void Warning(string operationName, string message, Exception exception = null)
        {
            return;
        }
    }
}
