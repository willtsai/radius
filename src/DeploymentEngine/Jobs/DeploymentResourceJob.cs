using Microsoft.WindowsAzure.ResourceStack.Common.BackgroundJobs;

namespace DeploymentEngine.Jobs
{
    [JobCallback(Name = "DeploymentResourceJob")]
    public class DeploymentResourceJob : JobBase<JobMetadata>
    {
        public DeploymentResourceJob(JobConfiguration configuration) :
            base(configuration)
        {
        }

        protected override Task OnConfigure()
        {
            return base.OnConfigure();
        }

        protected override Task<JobExecutionResult> OnExecute()
        {
            // TODO create radius resource for deployment
            // OR call RP contract on update resource.
            // Would love some sort of way to poll better here for deployment.
            throw new NotImplementedException();
        }
    }
}
