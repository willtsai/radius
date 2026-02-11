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
)

// ParameterValidator validates parameters against template requirements.
type ParameterValidator struct {
	// Template is the parsed ARM template
	Template *ARMTemplate
}

// NewParameterValidator creates a validator for the given template.
func NewParameterValidator(template *ARMTemplate) *ParameterValidator {
	return &ParameterValidator{
		Template: template,
	}
}

// ValidateParameters checks that all required parameters are provided.
// Returns an error describing any missing parameters.
func (v *ParameterValidator) ValidateParameters(provided map[string]any) error {
	required := v.Template.GetRequiredParameterNames()
	if len(required) == 0 {
		return nil
	}

	var missing []string
	for _, name := range required {
		if _, ok := provided[name]; !ok {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		return &MissingParametersError{
			Parameters: missing,
		}
	}

	return nil
}

// GetParameterInfo returns information about a specific parameter.
func (v *ParameterValidator) GetParameterInfo(name string) (*ParameterInfo, error) {
	param, ok := v.Template.Parameters[name]
	if !ok {
		return nil, fmt.Errorf("parameter %q not found in template", name)
	}

	info := &ParameterInfo{
		Name:         name,
		Type:         param.Type,
		Required:     param.DefaultValue == nil,
		DefaultValue: param.DefaultValue,
	}

	if param.Metadata != nil {
		info.Description = param.Metadata.Description
	}

	if len(param.AllowedValues) > 0 {
		info.AllowedValues = param.AllowedValues
	}

	return info, nil
}

// ListParameters returns information about all template parameters.
func (v *ParameterValidator) ListParameters() []*ParameterInfo {
	var params []*ParameterInfo

	for name := range v.Template.Parameters {
		info, _ := v.GetParameterInfo(name)
		if info != nil {
			params = append(params, info)
		}
	}

	return params
}

// ParameterInfo describes a template parameter.
type ParameterInfo struct {
	Name          string
	Type          string
	Required      bool
	DefaultValue  any
	Description   string
	AllowedValues []any
}

// MissingParametersError is returned when required parameters are missing.
type MissingParametersError struct {
	Parameters []string
}

// Error implements the error interface.
func (e *MissingParametersError) Error() string {
	return fmt.Sprintf("missing required parameters: %s. Use --parameters <file.bicepparam> to provide parameter values",
		strings.Join(e.Parameters, ", "))
}

// IsMissingParametersError checks if an error is a MissingParametersError.
func IsMissingParametersError(err error) bool {
	_, ok := err.(*MissingParametersError)
	return ok
}

// GetMissingParameters returns the list of missing parameters from the error.
func GetMissingParameters(err error) []string {
	if mpe, ok := err.(*MissingParametersError); ok {
		return mpe.Parameters
	}
	return nil
}

// ValidateParameterValue validates a parameter value against its type constraints.
func ValidateParameterValue(param ARMParameter, value any) error {
	// Type validation
	switch param.Type {
	case "string", "secureString":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string value")
		}
	case "int":
		switch v := value.(type) {
		case int, int32, int64, float64:
			_ = v // Valid numeric types
		default:
			return fmt.Errorf("expected integer value")
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean value")
		}
	case "object", "secureObject":
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("expected object value")
		}
	case "array":
		if _, ok := value.([]any); !ok {
			return fmt.Errorf("expected array value")
		}
	}

	// Allowed values validation
	if len(param.AllowedValues) > 0 {
		found := false
		for _, allowed := range param.AllowedValues {
			if value == allowed {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value not in allowed values: %v", param.AllowedValues)
		}
	}

	return nil
}
