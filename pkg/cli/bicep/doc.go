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

/*
Package bicep provides functionality for working with the Bicep compiler and ARM-JSON templates.

# Overview

This package enables the Radius CLI to generate static application graphs from Bicep files
without requiring deployment. It wraps the Bicep CLI and provides utilities for parsing,
validating, and extracting information from ARM JSON templates.

# Static Graph Generation

The main workflow for static graph generation is:

 1. Build Bicep to ARM JSON using [Executor]
 2. Parse the ARM JSON using [ParseARMJSON] or [ParseARMTemplate]
 3. Extract Radius resources using [ResourceExtractor]
 4. Detect connections using [ConnectionDetector]
 5. Compute source hashes for change detection using [ComputeSourceHash]

# Key Types

  - [Executor]: Wraps the Bicep CLI for executing build commands
  - [ARMTemplate]: Represents a parsed ARM JSON template
  - [ARMResource]: Represents a resource definition in an ARM template
  - [ResourceExtractor]: Extracts Radius resources from ARM templates
  - [ConnectionDetector]: Detects connections between resources

# Bicep CLI Integration

The package automatically handles Bicep CLI installation and version management.
Use [IsBicepInstalled] to check if Bicep is available, and [DownloadBicep] to
install it if needed.

# Error Handling

Build errors from the Bicep CLI are parsed and reported with file, line, and
column information when available. See [ParseBicepError] for error parsing.
*/
package bicep
