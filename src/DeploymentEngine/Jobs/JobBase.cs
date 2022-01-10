using Microsoft.WindowsAzure.ResourceStack.Common.BackgroundJobs;

namespace DeploymentEngine.Jobs
{
    public abstract class JobBase<TMetadata> : JobCallback<TMetadata>
    {
        protected JobConfiguration jobConfiguration { get; }
        public JobBase(JobConfiguration configuration)
        {
            this.jobConfiguration = configuration;
        }
    }
}
