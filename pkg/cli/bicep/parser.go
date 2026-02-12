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
	"encoding/json"
	"strings"
)

// ARMTemplate represents the structure of an ARM JSON template.
// This is a subset of the full ARM template schema, focused on the parts
// needed for static graph generation.
type ARMTemplate struct {
	// Schema is the ARM template schema URL
	Schema string `json:"$schema,omitempty"`

	// ContentVersion is the template version
	ContentVersion string `json:"contentVersion,omitempty"`

	// Parameters defines the template parameters
	Parameters map[string]ARMParameter `json:"parameters,omitempty"`

	// Variables defines the template variables
	Variables map[string]any `json:"variables,omitempty"`

	// Resources is the array of resource definitions
	Resources []ARMResource `json:"resources,omitempty"`

	// Outputs defines the template outputs
	Outputs map[string]any `json:"outputs,omitempty"`
}

// ARMParameter represents a parameter definition in an ARM template.
type ARMParameter struct {
	// Type is the parameter type (string, int, bool, object, array, secureString, secureObject)
	Type string `json:"type"`

	// DefaultValue is the default value if not provided
	DefaultValue any `json:"defaultValue,omitempty"`

	// AllowedValues restricts the parameter to specific values
	AllowedValues []any `json:"allowedValues,omitempty"`

	// Description is a human-readable description
	Metadata *ARMParameterMetadata `json:"metadata,omitempty"`
}

// ARMParameterMetadata contains metadata for a parameter.
type ARMParameterMetadata struct {
	Description string `json:"description,omitempty"`
}

// ARMResource represents a single resource in an ARM template.
type ARMResource struct {
	// Type is the resource type (e.g., "Applications.Core/containers")
	Type string `json:"type"`

	// APIVersion is the API version for the resource type
	APIVersion string `json:"apiVersion"`

	// Name is the resource name (may contain expressions)
	Name string `json:"name"`

	// Location is the resource location
	Location string `json:"location,omitempty"`

	// DependsOn lists explicit dependencies
	DependsOn []string `json:"dependsOn,omitempty"`

	// Properties contains the resource-specific configuration
	Properties map[string]any `json:"properties,omitempty"`

	// Tags are resource tags
	Tags map[string]string `json:"tags,omitempty"`

	// Condition determines if the resource is deployed
	Condition any `json:"condition,omitempty"`

	// Copy defines resource iteration
	Copy *ARMCopy `json:"copy,omitempty"`

	// Comments for documentation
	Comments string `json:"comments,omitempty"`
}

// ARMCopy defines resource copy/iteration settings.
type ARMCopy struct {
	Name      string `json:"name"`
	Count     any    `json:"count"`
	Mode      string `json:"mode,omitempty"`
	BatchSize int    `json:"batchSize,omitempty"`
}

// ParseARMTemplate parses a raw ARM JSON map into a structured ARMTemplate.
func ParseARMTemplate(template map[string]any) (*ARMTemplate, error) {
	result := &ARMTemplate{}

	// Parse schema
	if schema, ok := template["$schema"].(string); ok {
		result.Schema = schema
	}

	// Parse content version
	if cv, ok := template["contentVersion"].(string); ok {
		result.ContentVersion = cv
	}

	// Parse parameters
	if params, ok := template["parameters"].(map[string]any); ok {
		result.Parameters = make(map[string]ARMParameter)
		for name, param := range params {
			if paramMap, ok := param.(map[string]any); ok {
				result.Parameters[name] = parseARMParameter(paramMap)
			}
		}
	}

	// Parse variables
	if vars, ok := template["variables"].(map[string]any); ok {
		result.Variables = vars
	}

	// Parse resources - handle both languageVersion 1.0 (array) and 2.0 (map) formats
	if resources, ok := template["resources"].([]any); ok {
		// v1 format: resources is an array of objects
		for _, res := range resources {
			if resMap, ok := res.(map[string]any); ok {
				result.Resources = append(result.Resources, parseARMResource(resMap))
			}
		}
	} else if resources, ok := template["resources"].(map[string]any); ok {
		// v2 format (languageVersion 2.0): resources is a map keyed by symbolic name
		for symbolicName, res := range resources {
			if resMap, ok := res.(map[string]any); ok {
				result.Resources = append(result.Resources, parseARMResourceV2(symbolicName, resMap))
			}
		}
	}

	// Parse outputs
	if outputs, ok := template["outputs"].(map[string]any); ok {
		result.Outputs = outputs
	}

	return result, nil
}

// parseARMParameter parses a parameter definition map.
func parseARMParameter(m map[string]any) ARMParameter {
	p := ARMParameter{}

	if t, ok := m["type"].(string); ok {
		p.Type = t
	}
	if dv, ok := m["defaultValue"]; ok {
		p.DefaultValue = dv
	}
	if av, ok := m["allowedValues"].([]any); ok {
		p.AllowedValues = av
	}
	if meta, ok := m["metadata"].(map[string]any); ok {
		p.Metadata = &ARMParameterMetadata{}
		if desc, ok := meta["description"].(string); ok {
			p.Metadata.Description = desc
		}
	}

	return p
}

// parseARMResource parses a resource definition map.
func parseARMResource(m map[string]any) ARMResource {
	r := ARMResource{}

	if t, ok := m["type"].(string); ok {
		r.Type = t
	}
	if av, ok := m["apiVersion"].(string); ok {
		r.APIVersion = av
	}
	if n, ok := m["name"].(string); ok {
		r.Name = n
	}
	if loc, ok := m["location"].(string); ok {
		r.Location = loc
	}
	if deps, ok := m["dependsOn"].([]any); ok {
		for _, dep := range deps {
			if depStr, ok := dep.(string); ok {
				r.DependsOn = append(r.DependsOn, depStr)
			}
		}
	}
	if props, ok := m["properties"].(map[string]any); ok {
		r.Properties = props
	}
	if tags, ok := m["tags"].(map[string]any); ok {
		r.Tags = make(map[string]string)
		for k, v := range tags {
			if vStr, ok := v.(string); ok {
				r.Tags[k] = vStr
			}
		}
	}
	if cond, ok := m["condition"]; ok {
		r.Condition = cond
	}
	if comments, ok := m["comments"].(string); ok {
		r.Comments = comments
	}

	return r
}

// parseARMResourceV2 parses a resource definition from ARM JSON v2.0 (languageVersion 2.0) format.
// In v2.0, the type field includes the API version (e.g., "Applications.Core/containers@2023-10-01-preview")
// and the name is inside properties rather than at the top level.
func parseARMResourceV2(symbolicName string, m map[string]any) ARMResource {
	r := ARMResource{}

	if t, ok := m["type"].(string); ok {
		// In v2.0, type includes apiVersion: "Type@ApiVersion"
		parts := strings.SplitN(t, "@", 2)
		r.Type = parts[0]
		if len(parts) == 2 {
			r.APIVersion = parts[1]
		}
	}

	// In v2.0, name is inside properties
	if props, ok := m["properties"].(map[string]any); ok {
		r.Properties = props
		if n, ok := props["name"].(string); ok {
			r.Name = n
		}
	}

	// Fall back to top-level name if present
	if r.Name == "" {
		if n, ok := m["name"].(string); ok {
			r.Name = n
		}
	}

	// If still no name, use the symbolic name
	if r.Name == "" {
		r.Name = symbolicName
	}

	if loc, ok := m["location"].(string); ok {
		r.Location = loc
	}
	if deps, ok := m["dependsOn"].([]any); ok {
		for _, dep := range deps {
			if depStr, ok := dep.(string); ok {
				r.DependsOn = append(r.DependsOn, depStr)
			}
		}
	}
	if tags, ok := m["tags"].(map[string]any); ok {
		r.Tags = make(map[string]string)
		for k, v := range tags {
			if vStr, ok := v.(string); ok {
				r.Tags[k] = vStr
			}
		}
	}
	if cond, ok := m["condition"]; ok {
		r.Condition = cond
	}
	if comments, ok := m["comments"].(string); ok {
		r.Comments = comments
	}

	return r
}

// GetResources returns all resources from the template.
func (t *ARMTemplate) GetResources() []ARMResource {
	return t.Resources
}

// GetResourceByType returns all resources of a specific type.
func (t *ARMTemplate) GetResourceByType(resourceType string) []ARMResource {
	var result []ARMResource
	for _, r := range t.Resources {
		if r.Type == resourceType {
			result = append(result, r)
		}
	}
	return result
}

// GetRequiredParameterNames returns the names of parameters without default values.
func (t *ARMTemplate) GetRequiredParameterNames() []string {
	var required []string
	for name, param := range t.Parameters {
		if param.DefaultValue == nil {
			required = append(required, name)
		}
	}
	return required
}

// UnmarshalJSON unmarshals JSON bytes into the target value.
// This is a convenience wrapper around json.Unmarshal.
func UnmarshalJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// IsModuleDeployment returns true if the resource type represents a module deployment.
// In ARM JSON, Bicep modules are represented as Microsoft.Resources/deployments.
func IsModuleDeployment(resourceType string) bool {
	return resourceType == "Microsoft.Resources/deployments"
}
