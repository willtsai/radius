using Microsoft.WindowsAzure.ResourceStack.Common.BackgroundJobs;
using Newtonsoft.Json.Linq;

namespace DeploymentEngine.Jobs
{
    [JobCallback(Name = "DeploymentResourceJob")]
    public class DeploymentResourceJob : JobBase<DeploymentResourceJobMetadata>
    {
        public DeploymentResourceJob(JobConfiguration configuration) :
            base(configuration)
        {
        }

        protected override Task OnConfigure()
        {
            return base.OnConfigure();
        }

        protected override async Task<JobExecutionResult> OnExecute()
        {
            // TODO create radius resource for deployment
            // OR call RP contract on update resource.
            // Would love some sort of way to poll better here for deployment.
            Metadata = JToken.Parse(this.BackgroundJob.Metadata).ToObject<DeploymentResourceJobMetadata>();

            // TODO do async stuff here.
            await Task.CompletedTask;

            return new JobExecutionResult()
            {
                Status = JobExecutionStatus.Succeeded,
                Message = ""
            };
        }
    }
}
