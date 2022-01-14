using Azure.Deployments.Core.DataProviders;
using Azure.Deployments.Core.Definitions;
using Azure.Deployments.Core.Entities;
using Azure.Deployments.Core.Storage.Entities;

namespace DeploymentEngine.Jobs
{
    // TODO create inmemroy data provider to transfer betwen controller and jobs.
    public class InMemoryDataProvider : IDeploymentDataProvider
    {
        public Task CancelDeployments(string subscriptionId, string resourceGroupName)
        {
            throw new NotImplementedException();
        }

        public Task<string> DeleteDeployment(IDeploymentEntity deployment)
        {
            throw new NotImplementedException();
        }

        public Task<bool> DeleteDeployments(string subscriptionId, string resourceGroupName, int maxItemsToDelete = int.MaxValue)
        {
            throw new NotImplementedException();
        }

        public Task<bool> DeleteTenantDeployments(string tenantId, string managementGroupId, int maxItemsToDelete = int.MaxValue)
        {
            throw new NotImplementedException();
        }

        public Task<IDeploymentEntity> FindDeployment(string subscriptionId, string resourceGroupName, string deploymentName)
        {
            throw new NotImplementedException();
        }

        public Task<IDeploymentEntity> FindDeployment(string subscriptionId, string resourceGroupName, string deploymentName, string deploymentSequence)
        {
            throw new NotImplementedException();
        }

        public Task<IDeploymentEntity[]> FindDeployments(string subscriptionId, string resourceGroupName, int top, bool includeOrphanSequences = false, ProvisioningState? provisioningState = null)
        {
            throw new NotImplementedException();
        }

        public Task<SegmentedResult<IDeploymentEntity>> FindDeploymentsSegmented(string subscriptionId, string resourceGroupName, bool includeOrphanSequences = false, ProvisioningState? provisioningState = null, int? top = null, DataContinuationToken continuationToken = null)
        {
            throw new NotImplementedException();
        }

        public Task<IDeploymentEntity> FindTenantDeployment(string tenantId, string managementGroupId, string deploymentName)
        {
            throw new NotImplementedException();
        }

        public Task<IDeploymentEntity> FindTenantDeployment(string tenantId, string managementGroupId, string deploymentName, string deploymentSequence)
        {
            throw new NotImplementedException();
        }

        public Task<SegmentedResult<IDeploymentEntity>> FindTenantDeploymentsSegmented(string tenantId, string managementGroupId, bool includeOrphanSequences = false, ProvisioningState? provisioningState = null, int? top = null, DataContinuationToken continuationToken = null)
        {
            throw new NotImplementedException();
        }

        public Task<IDeploymentEntity[]> GetDeploymentBasicProperties(string subscriptionId, string resourceGroupName, int top = 1000, bool includeOrphanSequences = false, ProvisioningState? provisioningState = null)
        {
            throw new NotImplementedException();
        }

        public Task<long> GetDeploymentCount(string subscriptionId, string resourceGroupName, int top, bool includeOrphanSequences = false, ProvisioningState? provisioningState = null)
        {
            throw new NotImplementedException();
        }

        public Task<string> ReplaceDeployment(IDeploymentEntity newDeployment, IDeploymentEntity oldDeployment)
        {
            throw new NotImplementedException();
        }

        public Task<string> ReplaceDeployment(IDeploymentEntity deployment)
        {
            throw new NotImplementedException();
        }

        public Task<string> SaveDeployment(IDeploymentEntity deployment)
        {
            throw new NotImplementedException();
        }
    }
}
