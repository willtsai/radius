namespace DeploymentEngine.Controllers
{
    public enum TemplateDeploymentScope
    {
        NotSpecified,
        ResourceGroup,
        Subscription,
        ManagementGroup,
        Tenant
    }

    public class DeploymentRequestContext
    {
        private DeploymentRequestContext()
        {
        }

        public static DeploymentRequestContext CreateAtResourceGroup(
            string tenantId,
            string subscriptionId,
            string resourceGroupName,
            string deploymentName) // TODO CachedSubscription, ResourceGroup, DeploymentTenant
        {
            return new DeploymentRequestContext()
            {
                TenantId = tenantId,
                SubscriptionId = subscriptionId,
                ResourceGroupName = resourceGroupName,
                DeploymentName = deploymentName,
                Scope = TemplateDeploymentScope.ResourceGroup
            };
        }

        /// <summary>
        /// Gets the tenant id.
        /// </summary>
        public string TenantId { get; private set; }

        /// <summary>
        /// Gets the management group id.
        /// </summary>
        public string ManagementGroupId { get; private set; }

        /// <summary>
        /// Gets the subscription id.
        /// </summary>
        public string SubscriptionId { get; private set; }

        /// <summary>
        /// Gets the resource group name.
        /// </summary>
        public string ResourceGroupName { get; private set; }

        /// <summary>
        /// Gets the name of the deployment.
        /// </summary>
        public string DeploymentName { get; private set; }

        /// <summary>
        /// Gets the scope of the deployment.
        /// </summary>
        public TemplateDeploymentScope Scope { get; private set; }

    }
}
