# App Graph Command

The `rad app graph` command generates application topology graphs from Bicep files or deployed applications.

## Overview

This package implements the `rad app graph` command which provides two modes of operation:

1. **Static Mode**: Generate graphs from Bicep files without deployment
2. **Runtime Mode**: Fetch graphs from deployed applications via the Radius API

## Usage

### Static Graph Generation (from Bicep)

```bash
# Generate graph from Bicep file
rad app graph app.bicep

# Output to stdout instead of file
rad app graph app.bicep --stdout

# Custom output path
rad app graph app.bicep -o custom-graph.json

# Generate with Markdown preview
rad app graph app.bicep --format markdown

# Skip git metadata (faster)
rad app graph app.bicep --no-git

# Use parameter file
rad app graph app.bicep --parameters app.bicepparam
```

### Runtime Graph (from deployed app)

```bash
# Fetch graph from deployed application
rad app graph myapp

# Specify workspace/resource group
rad app graph myapp -w my-workspace -g my-resource-group
```

## Output Files

By default, static graph generation writes to `.radius/app-graph.json`.

With `--format markdown`, it also generates `.radius/app-graph.md` containing:
- Resource table with source locations and git history
- Mermaid topology diagram
- Connection details

## Package Structure

| File | Purpose |
|------|---------|
| `graph.go` | Command entry point and flag definitions |
| `static.go` | Static graph generation from Bicep files |
| `compute.go` | Runtime graph via Radius API |
| `detect.go` | Bicep file detection logic |
| `params.go` | Parameter file handling |
| `display.go` | Output formatting and display |
| `diff.go` | Graph diff computation between versions |

## Dependencies

- `pkg/cli/bicep/` - Bicep CLI execution and ARM JSON parsing
- `pkg/cli/git/` - Git metadata extraction
- `pkg/cli/output/` - JSON/Markdown/Mermaid formatting
- `pkg/corerp/api/v20231001preview/` - Type definitions

## Static Graph JSON Schema

```json
{
  "metadata": {
    "generatedAt": "2026-02-04T00:58:00Z",
    "radiusCliVersion": "0.35.0",
    "sourceFiles": ["app.bicep"],
    "sourceHash": "sha256:...",
    "gitCommit": "abc123..."
  },
  "resources": [
    {
      "id": "/planes/radius/local/resourceGroups/default/providers/Applications.Core/containers/frontend",
      "name": "frontend",
      "type": "Applications.Core/containers",
      "sourceLocation": {
        "file": "app.bicep",
        "line": 12,
        "module": null
      },
      "gitInfo": {
        "commitSha": "abc123...",
        "author": "dev@example.com",
        "authorDate": "2026-02-03T15:30:00Z",
        "message": "Add frontend container"
      }
    }
  ],
  "connections": [
    {
      "sourceId": ".../containers/frontend",
      "targetId": ".../containers/backend",
      "type": "connection"
    }
  ]
}
```

## Graph Diff

The diff functionality compares two graph versions and reports:
- Added resources
- Removed resources
- Modified resources (property changes)
- Connection changes

Used by the GitHub Action to post PR comments showing architecture changes.

## Testing

```bash
# Run unit tests
go test ./pkg/cli/cmd/app/graph/...

# Run with verbose output
go test -v ./pkg/cli/cmd/app/graph/...
```

See `test/unit/cli/graph/` for test files.
