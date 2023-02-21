// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/project-radius/radius/pkg/ucp/ucplog"
)

// UseLogValues appends logging values to the context based on the request.
func UseLogValues(h http.Handler, basePath string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		values := []any{}

		values = append(values,
			ucplog.LogFieldServiceID, ucplog.UCPServiceName,
			ucplog.LogFieldCorrelationID, r.Header.Get(ucplog.LogFieldCorrelationID),
			ucplog.LogFieldHTTPMethod, r.Method,
			ucplog.LogFieldHTTPRequestURI, r.RequestURI,
			ucplog.LogFieldContentLength, r.ContentLength,
		)

		clientIP := r.Header.Get(ucplog.XForwardedForHeader)
		if clientIP == "" {
			clientIP = r.RemoteAddr
		}
		if net.ParseIP(clientIP) != nil {
			values = append(values,
				ucplog.LogFieldClientIP, clientIP,
			)
		}

		correlationID := r.Header.Get(ucplog.LogFieldCorrelationID)
		if IsValidUUID(correlationID) {
			values = append(values,
				ucplog.LogFieldClientIP, clientIP,
			)
		}

		ua := r.Header.Get(ucplog.UserAgent)
		if ua != "" {
			values = append(values,
				ucplog.LogFieldUserAgent, ua,
			)
		}

		logger := logr.FromContextOrDiscard(r.Context()).WithValues(values...)
		r = r.WithContext(logr.NewContext(r.Context(), logger))
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func GetRelativePath(basePath string, path string) string {
	trimmedPath := strings.TrimPrefix(path, basePath)
	return trimmedPath
}

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
