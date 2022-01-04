using System.Text;
using Azure.Deployments.Core.Definitions;
using Azure.Deployments.Core.Helpers;
using Azure.Deployments.Core.Json;
using Microsoft.AspNetCore.Mvc;

namespace DeploymentEngine.Controllers;

[ApiController]
[Route("[controller]")]
public class WeatherForecastController : ControllerBase
{
    private static readonly string[] Summaries = new[]
    {
        "Freezing", "Bracing", "Chilly", "Cool", "Mild", "Warm", "Balmy", "Hot", "Sweltering", "Scorching"
    };

    private readonly ILogger<WeatherForecastController> _logger;

    public WeatherForecastController(ILogger<WeatherForecastController> logger)
    {
        _logger = logger;
    }

    [HttpGet(Name = "GetWeatherForecast")]
    public async Task Get()
    {
        var deploymentPayload = System.IO.File.ReadAllText("");
        var httpContent = new StringContent(deploymentPayload, Encoding.UTF8, "application/json");

    }

    /// <summary>
    /// Deserialize a deployment http request payload into a DeploymentContent object and try to calculate the hash.
    /// </summary>
    /// <param name="httpContent">The HTTP Content.</param>
    /// <param name="httpConfiguration">The HTTP Configuration.</param>
    /// <returns>The requested deployment definition.</returns>
    private static async Task<DeploymentContent> GetDeploymentContentAndTryCalculateHash(HttpContent httpContent)
    {
        var deploymentContent = await ReadAsJsonAsyncWithRewind<DeploymentContent>(httpContent)
            .ConfigureAwait(continueOnCapturedContext: false);

        deploymentContent.Properties.TemplateHash = deploymentContent.Properties.Template != null
            ? TemplateHelpers.ComputeTemplateHash(deploymentContent.Properties.Template.ToJToken())
            : null;

        return deploymentContent;
    }

    private static async Task<T> ReadAsJsonAsyncWithRewind<T>(HttpContent httpContent)
    {
        var contentStream = await httpContent.ReadAsStreamAsync().ConfigureAwait(continueOnCapturedContext: false);
        var streamPosition = contentStream.Position;

        try
        {
            var formatters = new MediaTypeFormatter[] {
                new JsonMediaTypeFormatter { SerializerSettings = SerializerSettings.SerializerMediaTypeSettings, UseDataContractJsonSerializer = false } };

            return await httpContent.ReadAsAsync<T>(formatters)
                .ConfigureAwait(continueOnCapturedContext: false);
        }
        finally
        {
            if (streamPosition != contentStream.Position)
            {
                contentStream.Seek(streamPosition, SeekOrigin.Begin);
            }
        }
    }
}
