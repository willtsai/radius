#!/bin/bash

# ============================================================================
# Detect app-graph.json files in the repository
# Supports both single file path and monorepo glob patterns
# ============================================================================

set -euo pipefail

# Optional environment variables with defaults
GRAPH_PATH="${GRAPH_PATH:-.radius/app-graph.json}"
MONOREPO_GLOB="${MONOREPO_GLOB:-}"

main() {
    local files=""
    
    if [[ -n "${MONOREPO_GLOB}" ]]; then
        echo "Searching for graph files with pattern: ${MONOREPO_GLOB}"
        
        # Use find with glob pattern for monorepo support
        while IFS= read -r -d '' file; do
            [[ -n "${files}" ]] && files="${files},"
            files="${files}${file}"
        done < <(find . -path "./${MONOREPO_GLOB}" -type f -print0 2>/dev/null || true)
        
        # Fallback to bash globbing if find doesn't work
        if [[ -z "${files}" ]]; then
            shopt -s nullglob globstar
            for file in ${MONOREPO_GLOB}; do
                [[ -f "${file}" ]] || continue
                [[ -n "${files}" ]] && files="${files},"
                files="${files}${file}"
            done
            shopt -u nullglob globstar
        fi
    else
        echo "Using graph path: ${GRAPH_PATH}"
        
        if [[ -f "${GRAPH_PATH}" ]]; then
            files="${GRAPH_PATH}"
        else
            echo "Warning: Graph file not found at ${GRAPH_PATH}"
            echo "The action will compare against empty graphs for new files"
            files="${GRAPH_PATH}"
        fi
    fi
    
    if [[ -z "${files}" ]]; then
        echo "No graph files detected"
        files="${GRAPH_PATH}"
    fi
    
    # Count files
    local count
    count=$(echo "${files}" | tr ',' '\n' | wc -l | tr -d ' ')
    echo "Detected ${count} graph file(s)"
    
    # Output for GitHub Actions
    echo "files=${files}" >> "${GITHUB_OUTPUT:-/dev/stdout}"
    echo "count=${count}" >> "${GITHUB_OUTPUT:-/dev/stdout}"
}

main "$@"
