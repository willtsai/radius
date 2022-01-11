using Azure.Deployments.Core.Definitions;
using Azure.Deployments.Core.Definitions.Extensibility;
using Azure.Deployments.Core.Entities;
using Microsoft.WindowsAzure.ResourceStack.Common.BackgroundJobs;

namespace DeploymentEngine.Jobs
{

    /// <summary>
    /// The deployment state.
    /// </summary>
    public class DeploymentState// : IDeploymentState<IDeploymentEntity, DeploymentMappingEntity>
    {
        /// <summary>
        /// Gets or sets the deployment location.
        /// </summary>
        public string DeploymentLocation { get; set; }

        /// <summary>
        /// Gets or sets the old deployment.
        /// </summary>
        public IDeploymentEntity OldDeployment { get; set; }

        /// <summary>
        /// Gets or sets the new deployment.
        /// </summary>
        public IDeploymentEntity NewDeployment { get; set; }

        ///// <summary>
        ///// Gets or sets the new deployment mapping.
        ///// </summary>
        //public DeploymentMappingEntity NewDeploymentMapping { get; set; }

        ///// <summary>
        ///// Gets or sets the new resource groups to create.
        ///// </summary>
        //public ResourceGroup[] NewResourceGroups { get; set; }

        /// <summary>
        /// Gets or sets the deployment job.
        /// </summary>
        public SequencerBuilder DeploymentJob { get; set; }

        /// <summary>
        /// Gets or sets the preflight resources.
        /// </summary>
        public DeploymentPreflightResource[] PreflightResources { get; set; }

        /// <summary>
        /// The extensible resources to deploy.
        /// </summary>
        public ExtensibleResource[] ExtensibleResources { get; set; }

        ///// <summary>
        ///// Gets or sets the preflight requests.
        ///// </summary>
        //public DeploymentPreflightRequest[] PreflightRequests { get; set; }

        ///// <summary>
        ///// Gets the preflight diagnostics.
        ///// </summary>
        //public DeploymentPreflightDiagnostics PreflightDiagnostics { get; } = new DeploymentPreflightDiagnostics();

        /// <summary>
        /// Gets or sets the deployment validation sequencer job.
        /// </summary>
        public SequencerBuilder DeploymentValidationSequencerJob { get; set; }
    }
}
