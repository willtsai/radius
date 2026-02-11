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

import (
	"fmt"
	"strings"

	"github.com/radius-project/radius/pkg/corerp/api/v20231001preview"
)

// ResourceExtractor extracts Radius resources from parsed ARM templates
// and converts them to static app graph resources.
type ResourceExtractor struct {
	// ResourceGroup is the default resource group for resource ID construction
	ResourceGroup string
}

// NewResourceExtractor creates a new ResourceExtractor.
func NewResourceExtractor(resourceGroup string) *ResourceExtractor {
	if resourceGroup == "" {
		resourceGroup = "default"
	}
	return &ResourceExtractor{
		ResourceGroup: resourceGroup,
	}
}

// ExtractResources extracts all resources from an ARM template and converts them
// to static app graph resources.
func (e *ResourceExtractor) ExtractResources(template *ARMTemplate, sourceFile string) ([]v20231001preview.StaticAppGraphResource, error) {
	var resources []v20231001preview.StaticAppGraphResource

	for i, armResource := range template.Resources {
		resource, err := e.convertARMResource(armResource, sourceFile, i+1)
		if err != nil {
			// Log warning but continue processing other resources
			continue
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

// convertARMResource converts an ARM resource to a static app graph resource.
func (e *ResourceExtractor) convertARMResource(armResource ARMResource, sourceFile string, lineNumber int) (v20231001preview.StaticAppGraphResource, error) {
	// Construct the resource ID
	resourceID := e.constructResourceID(armResource.Type, armResource.Name)

	// Extract the resource name (handle ARM expressions)
	name := extractResourceName(armResource.Name)

	resource := v20231001preview.StaticAppGraphResource{
		ID:   resourceID,
		Name: name,
		Type: armResource.Type,
		SourceLocation: v20231001preview.SourceLocation{
			File: sourceFile,
			Line: lineNumber, // Note: ARM JSON doesn't preserve original line numbers
		},
		Properties: armResource.Properties,
	}

	return resource, nil
}

// constructResourceID builds a fully qualified Radius resource ID.
func (e *ResourceExtractor) constructResourceID(resourceType, name string) string {
	// Handle ARM expressions in the name
	cleanName := extractResourceName(name)

	// Format: /planes/radius/local/resourceGroups/{rg}/providers/{type}/{name}
	return fmt.Sprintf("/planes/radius/local/resourceGroups/%s/providers/%s/%s",
		e.ResourceGroup, resourceType, cleanName)
}

// extractResourceName extracts the actual name from an ARM template name expression.
// ARM names can be:
// - Simple strings: "myapp"
// - Expressions: "[parameters('appName')]"
// - Concatenations: "[concat(parameters('prefix'), '-app')]"
func extractResourceName(name string) string {
	// If it's not an expression, return as-is
	if !strings.HasPrefix(name, "[") {
		return name
	}

	// Try to extract a simple parameter reference
	// e.g., "[parameters('appName')]" -> "<appName>"
	if strings.HasPrefix(name, "[parameters('") && strings.HasSuffix(name, "')]") {
		paramName := name[len("[parameters('") : len(name)-len("')]")]
		return fmt.Sprintf("<%s>", paramName)
	}

	// For complex expressions, return a placeholder
	return "<dynamic>"
}

// IsRadiusResource checks if a resource type is a Radius-managed resource.
func IsRadiusResource(resourceType string) bool {
	radiusPrefixes := []string{
		"Applications.Core/",
		"Applications.Datastores/",
		"Applications.Messaging/",
		"Applications.Dapr/",
		"Applications.Networking/",
	}

	for _, prefix := range radiusPrefixes {
		if strings.HasPrefix(resourceType, prefix) {
			return true
		}
	}
	return false
}

// GetResourceProvider extracts the provider from a resource type.
// e.g., "Applications.Core/containers" -> "Applications.Core"
func GetResourceProvider(resourceType string) string {
	parts := strings.SplitN(resourceType, "/", 2)
	if len(parts) >= 1 {
		return parts[0]
	}
	return ""
}

// GetResourceTypeName extracts the type name from a resource type.
// e.g., "Applications.Core/containers" -> "containers"
func GetResourceTypeName(resourceType string) string {
	parts := strings.SplitN(resourceType, "/", 2)
	if len(parts) >= 2 {
		return parts[1]
	}
	return resourceType
}
