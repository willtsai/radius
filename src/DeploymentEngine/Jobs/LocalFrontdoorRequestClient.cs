//-----------------------------------------------------------
// Copyright (c) Microsoft Corporation.  All rights reserved.
//-----------------------------------------------------------

namespace Microsoft.WindowsAzure.ResourceStack.Frontdoor.Worker.Engines
{
    using System;
    using System.Net.Http;
    using System.Net.Http.Headers;
    using System.Threading;
    using System.Threading.Tasks;
    using System.Web.Http;
    using Microsoft.WindowsAzure.ResourceStack.Common.EventSources;
    using Microsoft.WindowsAzure.ResourceStack.Common.Extensions;
    using Microsoft.WindowsAzure.ResourceStack.Common.Instrumentation;

    /// <summary>
    /// A client that makes HTTP requests using the local front door pipeline. These requests are not routed externally.
    /// </summary>
    internal class LocalFrontdoorRequestClient
    {
        /// <summary>
        /// Initializes a new instance of the <see cref="LocalFrontdoorRequestClient" /> class.
        /// </summary>
        /// <param name="frontdoorConfiguration">The front door configuration.</param>
        /// <param name="httpConfiguration">The HTTP configuration.</param>
        public LocalFrontdoorRequestClient()
        {
        }

        /// <summary>
        /// Makes an HTTP request to front door and returns the result.
        /// </summary>
        /// <param name="method">The HTTP method.</param>
        /// <param name="requestUri">The HTTP request URI.</param>
        /// <param name="cancellationToken">The cancellation token.</param>
        /// <param name="content">The HTTP message content.</param>
        /// <param name="addHeadersFunc">Callback to add headers.</param>
        public async Task<HttpResponseMessage> SendAsync(HttpMethod method, Uri requestUri, CancellationToken cancellationToken, HttpContent content = null, Action<HttpRequestHeaders> addHeadersFunc = null)
        {
            using (var request = new HttpRequestMessage(method, requestUri))
            {
                try
                {
                    request.Content = content ?? request.Content;
                    var response = await this.SendLocalFrontdoorRequest(request, cancellationToken);
                    //request.Headers.AddSystemUserAgent();
                    //request.Headers.AddOutgoingHeaders(
                    //    deploymentLocation: this.GetDeploymentLocation(),
                    //    frontdoorLocation: this.GetFrontdoorLocation());

                    ////Note(antmarti): This is to avoid content being passed by-reference, so that the handlers/ controllers do not modify the object being passed by the client.
                    //request.Content = await HttpHelper
                    //    .CloneContent(content)
                    //    .ConfigureAwait(continueOnCapturedContext: false);

                    //addHeadersFunc?.Invoke(request.Headers);

                    //var operationName = request.GetRESTfulRequestOperationName();

                    //var originalAuthorizationEvidence = RequestCorrelationContext.Current.AuthenticationIdentity?.AuthorizationEvidence;
                    //var response = await this.FrontdoorConfiguration.EventSource
                    //    .TraceHttpOutgoingRequest(
                    //        request: request,
                    //        operationName: operationName,
                    //        action: HttpHelper.TlsVersionHttpActionDecorator(this.FrontdoorConfiguration.EventSource, () => this.SendLocalFrontdoorRequest(request, cancellationToken)),
                    //        targetResourceProvider: FrontdoorEventHelpers.GetTargetResourceProvider(operationName),
                    //        targetResourceType: FrontdoorEventHelpers.GetTargetResourceType(operationName),
                    //        activity: activity)
                    //    .ConfigureAwait(continueOnCapturedContext: false);

                    //if (RequestCorrelationContext.Current.AuthenticationIdentity != null)
                    //{
                    //    RequestCorrelationContext.Current.AuthenticationIdentity.SetAuthorizationEvidence(originalAuthorizationEvidence);
                    //}

                    return response;
                }
                catch (Exception ex) when (!ex.IsFatal())
                {
                    cancellationToken.ThrowIfCancellationRequested();
                    
                }
                return null;
            }
        }

        /// <summary>
        /// Sends a request to the local front door pipeline.
        /// </summary>
        /// <param name="request">The request.</param>
        /// <param name="cancellationToken">The cancellation token.</param>
        private async Task<HttpResponseMessage> SendLocalFrontdoorRequest(HttpRequestMessage request, CancellationToken cancellationToken)
        {
            var prevContext = RequestCorrelationContext.Current;

            // Note(ilygre): Run request in new request correlation context. Original context will be saved and restored later.
            using (RequestCorrelationContext.NewCorrelationContextScope())
            {
                // Note(ilygre): Use original correlation id, client IP and client identity when handling this request.

                // Note(elpere): Initialize the http client with message handlers identical to the ones in the frontdoor http configuration. This simulates a frontdoor call without actually sending a request.
                // Note(elpere): In case of requests sent by the worker, the request doesn't have any authorization header and this is the only way to access the resource.
                using (var client = new HttpClient())
                {
                    return await client.SendAsync(request, cancellationToken).ConfigureAwait(continueOnCapturedContext: false);
                }
            }
        }
    }
}