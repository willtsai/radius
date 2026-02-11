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

// BicepError represents an error from Bicep compilation with source location.
type BicepError struct {
	// Message is the error message
	Message string

	// File is the source file path
	File string

	// Line is the 1-based line number (0 if unknown)
	Line int

	// Column is the 1-based column number (0 if unknown)
	Column int

	// Code is the Bicep error/warning code (e.g., "BCP001")
	Code string

	// Severity is "error", "warning", or "info"
	Severity string
}

// Error implements the error interface.
func (e *BicepError) Error() string {
	var sb strings.Builder

	if e.File != "" {
		sb.WriteString(e.File)
		if e.Line > 0 {
			sb.WriteString(fmt.Sprintf(":%d", e.Line))
			if e.Column > 0 {
				sb.WriteString(fmt.Sprintf(":%d", e.Column))
			}
		}
		sb.WriteString(": ")
	}

	if e.Code != "" {
		sb.WriteString(fmt.Sprintf("[%s] ", e.Code))
	}

	sb.WriteString(e.Message)

	return sb.String()
}

// BicepErrors is a collection of Bicep errors.
type BicepErrors struct {
	Errors []*BicepError
}

// Error implements the error interface.
func (e *BicepErrors) Error() string {
	if len(e.Errors) == 0 {
		return "unknown Bicep error"
	}

	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d Bicep errors:\n", len(e.Errors)))
	for i, err := range e.Errors {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("  %d. %s", i+1, err.Error()))
	}
	return sb.String()
}

// HasErrors returns true if there are any errors (not just warnings).
func (e *BicepErrors) HasErrors() bool {
	for _, err := range e.Errors {
		if err.Severity == "error" || err.Severity == "" {
			return true
		}
	}
	return false
}

// GetErrors returns only error-severity items.
func (e *BicepErrors) GetErrors() []*BicepError {
	var result []*BicepError
	for _, err := range e.Errors {
		if err.Severity == "error" || err.Severity == "" {
			result = append(result, err)
		}
	}
	return result
}

// GetWarnings returns only warning-severity items.
func (e *BicepErrors) GetWarnings() []*BicepError {
	var result []*BicepError
	for _, err := range e.Errors {
		if err.Severity == "warning" {
			result = append(result, err)
		}
	}
	return result
}

// ParseBicepOutput parses the output from the Bicep CLI to extract errors.
// Bicep CLI outputs errors in the format:
// /path/to/file.bicep(line,col) : severity BCP001: message
func ParseBicepOutput(output string) *BicepErrors {
	errors := &BicepErrors{}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		err := parseBicepErrorLine(line)
		if err != nil {
			errors.Errors = append(errors.Errors, err)
		}
	}

	return errors
}

// parseBicepErrorLine parses a single Bicep error line.
func parseBicepErrorLine(line string) *BicepError {
	// Format: /path/to/file.bicep(line,col) : severity BCP001: message
	err := &BicepError{Severity: "error"}

	// Find the file path and location
	locStart := strings.Index(line, "(")
	locEnd := strings.Index(line, ")")

	if locStart > 0 && locEnd > locStart {
		err.File = line[:locStart]

		// Parse line,col
		locStr := line[locStart+1 : locEnd]
		parts := strings.Split(locStr, ",")
		if len(parts) >= 1 {
			_, _ = fmt.Sscanf(parts[0], "%d", &err.Line)
		}
		if len(parts) >= 2 {
			_, _ = fmt.Sscanf(parts[1], "%d", &err.Column)
		}

		line = strings.TrimSpace(line[locEnd+1:])
	}

	// Remove leading colon
	line = strings.TrimPrefix(line, ":")
	line = strings.TrimSpace(line)

	// Parse severity and code
	if strings.HasPrefix(line, "error") {
		err.Severity = "error"
		line = strings.TrimPrefix(line, "error")
	} else if strings.HasPrefix(line, "warning") {
		err.Severity = "warning"
		line = strings.TrimPrefix(line, "warning")
	} else if strings.HasPrefix(line, "info") {
		err.Severity = "info"
		line = strings.TrimPrefix(line, "info")
	}

	line = strings.TrimSpace(line)

	// Extract error code (BCP001, etc.)
	if strings.HasPrefix(line, "BCP") || strings.HasPrefix(line, "RPS") {
		colonIdx := strings.Index(line, ":")
		if colonIdx > 0 {
			err.Code = strings.TrimSpace(line[:colonIdx])
			line = strings.TrimSpace(line[colonIdx+1:])
		}
	}

	err.Message = line

	return err
}

// WrapBicepError wraps a generic error with Bicep-specific context.
func WrapBicepError(err error, file string, operation string) error {
	if err == nil {
		return nil
	}

	// Check if it's already a BicepError
	if _, ok := err.(*BicepError); ok {
		return err
	}

	return &BicepError{
		Message:  fmt.Sprintf("%s: %s", operation, err.Error()),
		File:     file,
		Severity: "error",
	}
}

// IsResourceNotFoundError checks if the error indicates a resource was not found.
func IsResourceNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "not found") ||
		strings.Contains(msg, "does not exist") ||
		strings.Contains(msg, "BCP081") // Bicep: resource not found
}

// IsExtensionError checks if the error is related to a missing Bicep extension.
func IsExtensionError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "extension") ||
		strings.Contains(msg, "BCP084") || // Missing provider
		strings.Contains(msg, "BCP203") // Extension not found
}
