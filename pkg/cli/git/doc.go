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
Package git provides functionality for extracting git metadata from source files.

# Overview

This package enables the Radius CLI to enrich application graph resources with git
commit information, including author, date, and commit message for each resource's
source location. This helps teams track when and why resources were added or modified.

# Main Features

  - Repository detection: Check if a directory is within a git repository
  - Git blame: Get commit information for specific lines of a file
  - Git log: Get commit history and details
  - Uncommitted changes detection: Identify pending changes not yet committed

# Key Types

  - [Repository]: Represents a git repository with its root path
  - [BlameExecutor]: Executes git blame for line-level commit info
  - [LogExecutor]: Executes git log for commit metadata
  - [MetadataEnricher]: Enriches resources with git metadata
  - [StatusChecker]: Detects uncommitted changes

# Usage

Typical workflow for enriching resources with git metadata:

 1. Detect repository using [DetectRepository]
 2. Create enricher with [NewMetadataEnricher]
 3. Call enricher.Enrich() to add git info to resources

# Graceful Degradation

The package handles edge cases gracefully:

  - Non-git directories: Resources are marked as having no git info
  - Shallow clones: Resources may be marked as "history unavailable"
  - Uncommitted files: Resources show "uncommitted" status
  - Files outside repository: Resources are marked appropriately

# Integration

This package is used by the static graph generator in pkg/cli/cmd/app/graph
to add git metadata to each resource in the application topology graph.
*/
package git
