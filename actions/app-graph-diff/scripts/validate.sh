#!/bin/bash

# ============================================================================
# Validate graph freshness by comparing committed hash with regenerated hash
# Fails if the committed graph is out of sync with Bicep source files
# ============================================================================

set -euo pipefail

# Check jq availability
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed" >&2
    exit 1
fi

# Required environment variables
: "${GRAPH_FILES:?GRAPH_FILES environment variable is required}"

# Optional variables with defaults
FAIL_ON_STALE="${FAIL_ON_STALE:-false}"

# Extract source hash from graph JSON
get_committed_hash() {
    local file="$1"
    
    if [[ ! -f "${file}" ]]; then
        echo ""
        return
    fi
    
    jq -r '.generatedAt.sourceHash // .metadata.sourceHash // ""' "${file}" 2>/dev/null || echo ""
}

# Compute hash of source files (placeholder - would need Bicep tooling)
compute_source_hash() {
    local file="$1"
    
    # Get directory of graph file
    local dir
    dir=$(dirname "${file}")
    
    # Look for Bicep files in same directory or parent
    local bicep_dir="${dir}"
    if [[ ! -f "${bicep_dir}/app.bicep" ]] && [[ ! -f "${bicep_dir}/main.bicep" ]]; then
        bicep_dir=$(dirname "${dir}")
    fi
    
    # Hash all .bicep files
    local hash=""
    if command -v sha256sum &> /dev/null; then
        hash=$(find "${bicep_dir}" -maxdepth 2 -name "*.bicep" -type f -exec sha256sum {} \; 2>/dev/null | sort | sha256sum | cut -d' ' -f1)
    elif command -v shasum &> /dev/null; then
        hash=$(find "${bicep_dir}" -maxdepth 2 -name "*.bicep" -type f -exec shasum -a 256 {} \; 2>/dev/null | sort | shasum -a 256 | cut -d' ' -f1)
    fi
    
    echo "${hash}"
}

main() {
    echo "Validating graph freshness"
    
    local stale_files=""
    local all_valid=true
    
    IFS=',' read -ra files <<< "${GRAPH_FILES}"
    for file in "${files[@]}"; do
        [[ -z "${file}" ]] && continue
        [[ ! -f "${file}" ]] && continue
        
        echo "Checking: ${file}"
        
        local committed_hash
        local current_hash
        committed_hash=$(get_committed_hash "${file}")
        current_hash=$(compute_source_hash "${file}")
        
        if [[ -z "${committed_hash}" ]]; then
            echo "  Warning: No source hash found in graph file"
            continue
        fi
        
        if [[ -z "${current_hash}" ]]; then
            echo "  Warning: Could not compute source hash"
            continue
        fi
        
        if [[ "${committed_hash}" != "${current_hash}" ]]; then
            echo "  STALE: Hash mismatch"
            echo "    Committed: ${committed_hash:0:16}..."
            echo "    Current:   ${current_hash:0:16}..."
            stale_files="${stale_files}${file},"
            all_valid=false
        else
            echo "  OK: Hash matches"
        fi
    done
    
    # Output results
    echo "is-stale=$( [[ "${all_valid}" == "false" ]] && echo "true" || echo "false" )" >> "${GITHUB_OUTPUT:-/dev/stdout}"
    echo "stale-files=${stale_files%,}" >> "${GITHUB_OUTPUT:-/dev/stdout}"
    
    # Fail if configured and stale
    if [[ "${FAIL_ON_STALE}" == "true" ]] && [[ "${all_valid}" == "false" ]]; then
        echo ""
        echo "Error: Stale graph files detected"
        echo "Please regenerate the graph files by running:"
        echo "  rad app graph <bicep-file>"
        echo ""
        exit 1
    fi
    
    echo "Validation complete"
}

main "$@"
