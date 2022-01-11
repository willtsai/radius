using Azure.Deployments.Core.EventSources;

namespace DeploymentEngine.Controllers
{
    public class EventSource : IGeneralEventSource
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

        public void ProviderDebug(string providerNamespace, string resourceType, string operationName, string message, Exception exception = null)
        {
            return;
        }

        public void ProviderError(string providerNamespace, string resourceType, string operationName, string message, Exception exception = null)
        {
            return;
        }

        public void Warning(string operationName, string message, Exception exception = null)
        {
            return;
        }
    }
}
