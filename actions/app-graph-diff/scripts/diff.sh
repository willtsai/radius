#!/bin/bash

# ============================================================================
# Compute graph diff between base and head refs
# Reads JSON graph files from git history and computes differences
# ============================================================================

set -euo pipefail

# Check jq availability
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed" >&2
    exit 1
fi

# Required environment variables
: "${GRAPH_FILES:?GRAPH_FILES environment variable is required}"
: "${BASE_REF:?BASE_REF environment variable is required}"
: "${HEAD_REF:?HEAD_REF environment variable is required}"

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Read graph file content at specific ref
read_graph_at_ref() {
    local file="$1"
    local ref="$2"
    
    if git show "${ref}:${file}" 2>/dev/null; then
        return 0
    else
        echo "{}"
        return 0
    fi
}

# Compare resources between two graphs
compare_resources() {
    local base="$1"
    local head="$2"
    
    # Extract resource IDs from both graphs
    local base_ids
    local head_ids
    base_ids=$(echo "${base}" | jq -r '.resources[]?.id // empty' | sort)
    head_ids=$(echo "${head}" | jq -r '.resources[]?.id // empty' | sort)
    
    # Find added resources (in head but not in base)
    local added_ids
    added_ids=$(comm -13 <(echo "${base_ids}") <(echo "${head_ids}"))
    
    # Find removed resources (in base but not in head)
    local removed_ids
    removed_ids=$(comm -23 <(echo "${base_ids}") <(echo "${head_ids}"))
    
    # Find common resources for modification check
    local common_ids
    common_ids=$(comm -12 <(echo "${base_ids}") <(echo "${head_ids}"))
    
    # Build added resources array
    local added_json="[]"
    while IFS= read -r id; do
        [[ -z "${id}" ]] && continue
        local resource
        resource=$(echo "${head}" | jq --arg id "${id}" '.resources[] | select(.id == $id)')
        added_json=$(echo "${added_json}" | jq --argjson res "${resource}" '. + [$res]')
    done <<< "${added_ids}"
    
    # Build removed resources array
    local removed_json="[]"
    while IFS= read -r id; do
        [[ -z "${id}" ]] && continue
        local resource
        resource=$(echo "${base}" | jq --arg id "${id}" '.resources[] | select(.id == $id)')
        removed_json=$(echo "${removed_json}" | jq --argjson res "${resource}" '. + [$res]')
    done <<< "${removed_ids}"
    
    # Check modified resources
    local modified_json="[]"
    while IFS= read -r id; do
        [[ -z "${id}" ]] && continue
        local base_resource
        local head_resource
        base_resource=$(echo "${base}" | jq --arg id "${id}" '.resources[] | select(.id == $id)')
        head_resource=$(echo "${head}" | jq --arg id "${id}" '.resources[] | select(.id == $id)')
        
        if [[ "${base_resource}" != "${head_resource}" ]]; then
            # Build property changes
            local changes="[]"
            local name
            local type
            name=$(echo "${head_resource}" | jq -r '.name // .id')
            type=$(echo "${head_resource}" | jq -r '.type // ""')
            
            local modified_entry
            modified_entry=$(jq -n \
                --arg id "${id}" \
                --arg name "${name}" \
                --arg type "${type}" \
                --argjson changes "${changes}" \
                '{id: $id, name: $name, type: $type, changedProperties: $changes}')
            modified_json=$(echo "${modified_json}" | jq --argjson entry "${modified_entry}" '. + [$entry]')
        fi
    done <<< "${common_ids}"
    
    # Output combined diff structure
    jq -n \
        --argjson added "${added_json}" \
        --argjson removed "${removed_json}" \
        --argjson modified "${modified_json}" \
        '{addedResources: $added, removedResources: $removed, modifiedResources: $modified}'
}

# Compute diff summary
compute_summary() {
    local diff="$1"
    
    local added_count
    local removed_count
    local modified_count
    added_count=$(echo "${diff}" | jq '.addedResources | length')
    removed_count=$(echo "${diff}" | jq '.removedResources | length')
    modified_count=$(echo "${diff}" | jq '.modifiedResources | length')
    
    local total=$((added_count + removed_count + modified_count))
    
    jq -n \
        --argjson total "${total}" \
        --argjson added "${added_count}" \
        --argjson removed "${removed_count}" \
        --argjson modified "${modified_count}" \
        '{totalChanges: $total, resourcesAdded: $added, resourcesRemoved: $removed, resourcesModified: $modified, connectionsAdded: 0, connectionsRemoved: 0}'
}

main() {
    echo "Computing graph diff between ${BASE_REF} and ${HEAD_REF}"
    
    # Process each graph file
    local combined_diff='{"addedResources":[],"removedResources":[],"modifiedResources":[]}'
    
    IFS=',' read -ra files <<< "${GRAPH_FILES}"
    for file in "${files[@]}"; do
        [[ -z "${file}" ]] && continue
        echo "Processing: ${file}"
        
        local base_graph
        local head_graph
        base_graph=$(read_graph_at_ref "${file}" "${BASE_REF}")
        head_graph=$(read_graph_at_ref "${file}" "${HEAD_REF}")
        
        local file_diff
        file_diff=$(compare_resources "${base_graph}" "${head_graph}")
        
        # Merge into combined diff
        combined_diff=$(jq -s '.[0].addedResources += .[1].addedResources | .[0].removedResources += .[1].removedResources | .[0].modifiedResources += .[1].modifiedResources | .[0]' <(echo "${combined_diff}") <(echo "${file_diff}"))
    done
    
    # Add summary
    local summary
    summary=$(compute_summary "${combined_diff}")
    combined_diff=$(echo "${combined_diff}" | jq --argjson summary "${summary}" '. + {summary: $summary}')
    
    # Determine if there are changes
    local has_changes="false"
    local total_changes
    total_changes=$(echo "${combined_diff}" | jq '.summary.totalChanges')
    if [[ "${total_changes}" -gt 0 ]]; then
        has_changes="true"
    fi
    
    # Output for GitHub Actions
    echo "has-changes=${has_changes}" >> "${GITHUB_OUTPUT:-/dev/stdout}"
    echo "added-count=$(echo "${combined_diff}" | jq '.summary.resourcesAdded')" >> "${GITHUB_OUTPUT:-/dev/stdout}"
    echo "removed-count=$(echo "${combined_diff}" | jq '.summary.resourcesRemoved')" >> "${GITHUB_OUTPUT:-/dev/stdout}"
    echo "modified-count=$(echo "${combined_diff}" | jq '.summary.resourcesModified')" >> "${GITHUB_OUTPUT:-/dev/stdout}"
    
    # Output diff JSON (escape newlines for GitHub Actions)
    local diff_json
    diff_json=$(echo "${combined_diff}" | jq -c '.')
    echo "diff-json=${diff_json}" >> "${GITHUB_OUTPUT:-/dev/stdout}"
    
    # Generate no-changes body if needed
    if [[ "${has_changes}" == "false" ]]; then
        local no_changes_body
        no_changes_body="## 📊 Application Graph

✅ **No changes to application graph**

The application topology remains unchanged in this PR.

---
*Generated by Radius App Graph Action*"
        echo "no-changes-body<<EOF" >> "${GITHUB_OUTPUT:-/dev/stdout}"
        echo "${no_changes_body}" >> "${GITHUB_OUTPUT:-/dev/stdout}"
        echo "EOF" >> "${GITHUB_OUTPUT:-/dev/stdout}"
    fi
    
    echo "Diff computation complete. Changes detected: ${has_changes}"
}

main "$@"
