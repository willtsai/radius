using Microsoft.WindowsAzure.ResourceStack.Common.BackgroundJobs;

namespace DeploymentEngine.Jobs
{
    public class DeploymentJobCallbackFactory : JobCallbackFactory
    {
        private JobConfiguration jobConfiguration;

        public DeploymentJobCallbackFactory(JobConfiguration jobConfiguration)
        {
            this.jobConfiguration = jobConfiguration;
        }

        public override JobDelegate CreateInstance(JobLogger jobLogger, Type callbackType, BackgroundJob job)
        {
            if (this.ShouldCreateJob(callbackType))
            {
                return (JobDelegate)Activator.CreateInstance(
                    type: callbackType,
                    args: new object[] { this.jobConfiguration });
            }

            return base.CreateInstance(jobLogger, callbackType, job);
        }

        private bool ShouldCreateJob(Type callbackType)
        {
            if (typeof(JobBase<JobMetadata>).IsAssignableFrom(callbackType))
            {
                return true;
            }

            return false;
        }
    }
}
