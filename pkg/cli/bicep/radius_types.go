/*
Copyright 2023 The Radius Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bicep

import "strings"

// Radius resource type providers
const (
	ProviderCore       = "Applications.Core"
	ProviderDatastores = "Applications.Datastores"
	ProviderMessaging  = "Applications.Messaging"
	ProviderDapr       = "Applications.Dapr"
	ProviderNetworking = "Applications.Networking"
	ProviderExtenders  = "Applications.Extenders"
	ProviderConnector  = "Applications.Connector"
)

// radiusProviders is the list of known Radius resource providers
var radiusProviders = []string{
	ProviderCore,
	ProviderDatastores,
	ProviderMessaging,
	ProviderDapr,
	ProviderNetworking,
	ProviderExtenders,
	ProviderConnector,
}

// RadiusResourceTypes maps resource type short names to their full type paths
var RadiusResourceTypes = map[string]string{
	// Applications.Core
	"applications": "Applications.Core/applications",
	"containers":   "Applications.Core/containers",
	"gateways":     "Applications.Core/gateways",
	"environments": "Applications.Core/environments",
	"secretStores": "Applications.Core/secretStores",
	"volumes":      "Applications.Core/volumes",
	"extenders":    "Applications.Core/extenders",

	// Applications.Datastores
	"mongoDatabases": "Applications.Datastores/mongoDatabases",
	"redisCaches":    "Applications.Datastores/redisCaches",
	"sqlDatabases":   "Applications.Datastores/sqlDatabases",

	// Applications.Messaging
	"rabbitMQQueues": "Applications.Messaging/rabbitMQQueues",

	// Applications.Dapr
	"daprPubSubBrokers":       "Applications.Dapr/pubSubBrokers",
	"daprSecretStores":        "Applications.Dapr/secretStores",
	"daprStateStores":         "Applications.Dapr/stateStores",
	"daprConfigurationStores": "Applications.Dapr/configurationStores",
}

// IsRadiusProvider checks if a provider name is a Radius provider.
func IsRadiusProvider(provider string) bool {
	for _, p := range radiusProviders {
		if strings.EqualFold(p, provider) {
			return true
		}
	}
	return false
}

// GetRadiusResourceCategory returns the category of a Radius resource type.
// Returns empty string for non-Radius resources.
func GetRadiusResourceCategory(resourceType string) string {
	provider := GetResourceProvider(resourceType)

	switch provider {
	case ProviderCore:
		return "core"
	case ProviderDatastores:
		return "data"
	case ProviderMessaging:
		return "messaging"
	case ProviderDapr:
		return "dapr"
	case ProviderNetworking:
		return "networking"
	case ProviderExtenders, ProviderConnector:
		return "extender"
	default:
		return ""
	}
}

// GetMermaidShape returns the Mermaid diagram shape for a resource type.
// Different resource types use different shapes for visual distinction.
func GetMermaidShape(resourceType string) string {
	typeName := GetResourceTypeName(resourceType)
	category := GetRadiusResourceCategory(resourceType)

	// Shape mapping based on category and type
	switch {
	case typeName == "gateways":
		return "diamond" // {{text}} - gateway/router shape
	case typeName == "environments":
		return "hexagon" // {{{{text}}}} - infrastructure shape
	case category == "data":
		return "cylinder" // [(text)] - database shape
	case category == "messaging":
		return "parallelogram" // [/text/] - queue shape
	case category == "dapr":
		return "stadium" // ([text]) - Dapr component shape
	case typeName == "containers":
		return "rectangle" // [text] - container shape
	default:
		return "rectangle" // Default shape
	}
}

// GetMermaidNodeSyntax returns the Mermaid node syntax for a resource.
func GetMermaidNodeSyntax(id, label, resourceType string) string {
	shape := GetMermaidShape(resourceType)

	// Sanitize the label for Mermaid (escape quotes)
	safeLabel := strings.ReplaceAll(label, `"`, `'`)

	switch shape {
	case "diamond":
		return id + "{" + safeLabel + "}"
	case "hexagon":
		return id + "{{" + safeLabel + "}}"
	case "cylinder":
		return id + "[(" + safeLabel + ")]"
	case "parallelogram":
		return id + "[/" + safeLabel + "/]"
	case "stadium":
		return id + "([" + safeLabel + "])"
	default:
		return id + "[" + safeLabel + "]"
	}
}

// PortableResourceTypes lists resource types that support portable abstraction.
// These types can be backed by Recipes that provision concrete infrastructure.
var PortableResourceTypes = []string{
	"Applications.Datastores/mongoDatabases",
	"Applications.Datastores/redisCaches",
	"Applications.Datastores/sqlDatabases",
	"Applications.Messaging/rabbitMQQueues",
	"Applications.Dapr/pubSubBrokers",
	"Applications.Dapr/secretStores",
	"Applications.Dapr/stateStores",
	"Applications.Dapr/configurationStores",
}

// IsPortableResourceType checks if a resource type is a portable (Recipe-backed) type.
func IsPortableResourceType(resourceType string) bool {
	for _, t := range PortableResourceTypes {
		if strings.EqualFold(t, resourceType) {
			return true
		}
	}
	return false
}

// ExtractApplicationID extracts the application resource ID from resource properties.
// Returns empty string if not found.
func ExtractApplicationID(properties map[string]any) string {
	if properties == nil {
		return ""
	}

	if app, ok := properties["application"].(string); ok {
		return app
	}

	return ""
}

// ExtractEnvironmentID extracts the environment resource ID from resource properties.
// Returns empty string if not found.
func ExtractEnvironmentID(properties map[string]any) string {
	if properties == nil {
		return ""
	}

	if env, ok := properties["environment"].(string); ok {
		return env
	}

	return ""
}
