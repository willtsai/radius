#!/bin/bash

# ============================================================================
# Render graph diff as Markdown for PR comments
# Generates formatted Markdown with optional Mermaid diagrams
# ============================================================================

set -euo pipefail

# Check jq availability
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed" >&2
    exit 1
fi

# Required environment variables
: "${DIFF_JSON:?DIFF_JSON environment variable is required}"

# Optional variables with defaults
INCLUDE_MERMAID="${INCLUDE_MERMAID:-true}"

# Extract resource name from a resource ID (last path segment)
name_from_id() {
    echo "$1" | awk -F'/' '{print $NF}'
}

# Render connections changes section
render_connections_section() {
    local diff="$1"
    local added_count
    local removed_count
    added_count=$(echo "${diff}" | jq '.addedConnections | length')
    removed_count=$(echo "${diff}" | jq '.removedConnections | length')

    [[ "${added_count}" -eq 0 && "${removed_count}" -eq 0 ]] && return

    echo "### 🔗 Connection Changes"
    echo ""

    if [[ "${added_count}" -gt 0 ]]; then
        echo "**Added Connections**"
        echo ""
        echo "| Source | Target | Type |"
        echo "|--------|--------|------|"
        echo "${diff}" | jq -r '.addedConnections[] | "| \(.sourceId | split("/") | last) | \(.targetId | split("/") | last) | \(.type // "-") |"'
        echo ""
    fi

    if [[ "${removed_count}" -gt 0 ]]; then
        echo "**Removed Connections**"
        echo ""
        echo "| Source | Target | Type |"
        echo "|--------|--------|------|"
        echo "${diff}" | jq -r '.removedConnections[] | "| ~~\(.sourceId | split("/") | last)~~ | ~~\(.targetId | split("/") | last)~~ | \(.type // "-") |"'
        echo ""
    fi
}

# Render summary badges
render_badges() {
    local diff="$1"
    local badges=""
    
    local added
    local removed
    local modified
    added=$(echo "${diff}" | jq -r '.summary.resourcesAdded // 0')
    removed=$(echo "${diff}" | jq -r '.summary.resourcesRemoved // 0')
    modified=$(echo "${diff}" | jq -r '.summary.resourcesModified // 0')
    
    local conn_added
    local conn_removed
    conn_added=$(echo "${diff}" | jq -r '.summary.connectionsAdded // 0')
    conn_removed=$(echo "${diff}" | jq -r '.summary.connectionsRemoved // 0')
    
    [[ "${added}" -gt 0 ]] && badges="${badges}➕ ${added} added | "
    [[ "${removed}" -gt 0 ]] && badges="${badges}➖ ${removed} removed | "
    [[ "${modified}" -gt 0 ]] && badges="${badges}🔄 ${modified} modified | "
    [[ "${conn_added}" -gt 0 ]] && badges="${badges}🔗 ${conn_added} connections added | "
    [[ "${conn_removed}" -gt 0 ]] && badges="${badges}🔗 ${conn_removed} connections removed | "
    
    # Remove trailing separator
    badges="${badges% | }"
    echo "${badges}"
}

# Render added resources section
render_added_section() {
    local diff="$1"
    local count
    count=$(echo "${diff}" | jq '.addedResources | length')
    
    [[ "${count}" -eq 0 ]] && return
    
    cat << 'EOF'
### ➕ Added Resources

| Resource | Type |
|----------|------|
EOF
    
    echo "${diff}" | jq -r '.addedResources[] | "| `\(.name // .id)` | \(.type // "unknown") |"'
    echo ""
}

# Render removed resources section
render_removed_section() {
    local diff="$1"
    local count
    count=$(echo "${diff}" | jq '.removedResources | length')
    
    [[ "${count}" -eq 0 ]] && return
    
    cat << 'EOF'
### ➖ Removed Resources

| Resource | Type |
|----------|------|
EOF
    
    echo "${diff}" | jq -r '.removedResources[] | "| ~~`\(.name // .id)`~~ | \(.type // "unknown") |"'
    echo ""
}

# Render modified resources section
render_modified_section() {
    local diff="$1"
    local count
    count=$(echo "${diff}" | jq '.modifiedResources | length')
    
    [[ "${count}" -eq 0 ]] && return
    
    echo "### 🔄 Modified Resources"
    echo ""
    
    while IFS= read -r resource; do
        local name
        name=$(echo "${resource}" | jq -r '.name // .id')
        echo "**${name}**"
        echo ""
        
        local changes_count
        changes_count=$(echo "${resource}" | jq '.changedProperties | length')
        if [[ "${changes_count}" -gt 0 ]]; then
            echo "| Property | Change |"
            echo "|----------|--------|"
            echo "${resource}" | jq -r '.changedProperties[] | "| `\(.path)` | \(.oldValue // "none") → \(.newValue // "none") |"'
            echo ""
        fi
    done < <(echo "${diff}" | jq -c '.modifiedResources[]')
}

# Sanitize a resource ID into a valid Mermaid node identifier
mermaid_node_id() {
    echo "$1" | sed 's/[^a-zA-Z0-9]/_/g'
}

# Render "Before" Mermaid diagram (unchanged + removed resources/connections)
render_mermaid_before() {
    local diff="$1"

    # Skip if nothing to show
    local removed_res unchanged_res modified_res
    removed_res=$(echo "${diff}" | jq '.removedResources | length')
    unchanged_res=$(echo "${diff}" | jq '[.unchangedResources[]?] | length')
    modified_res=$(echo "${diff}" | jq '.modifiedResources | length')
    [[ "${removed_res}" -eq 0 && "${unchanged_res}" -eq 0 && "${modified_res}" -eq 0 ]] && return

    cat << 'EOF'
#### Before

```mermaid
graph LR
EOF

    # Unchanged resources
    echo "${diff}" | jq -r '.unchangedResources[]? | "    " + (.id | gsub("[^a-zA-Z0-9]"; "_")) + "[\"" + (.name // .id) + " (" + (.type // "") + ")\"]"'
    # Modified resources (shown as unchanged in Before view, since they existed before)
    echo "${diff}" | jq -r '.modifiedResources[] | "    " + (.id | gsub("[^a-zA-Z0-9]"; "_")) + "[\"" + (.name // .id) + " (" + (.type // "") + ")\"]"'
    # Removed resources (highlighted red)
    echo "${diff}" | jq -r '.removedResources[] | "    " + (.id | gsub("[^a-zA-Z0-9]"; "_")) + "[\"" + (.name // .id) + " (" + (.type // "") + ")\"]:::removed"'

    local link_index=0
    local link_styles=""

    # Removed connections (red)
    while IFS= read -r conn; do
        [[ -z "${conn}" ]] && continue
        local src tgt label
        src=$(echo "${conn}" | jq -r '.sourceId | gsub("[^a-zA-Z0-9]"; "_")')
        tgt=$(echo "${conn}" | jq -r '.targetId | gsub("[^a-zA-Z0-9]"; "_")')
        label=$(echo "${conn}" | jq -r '.type // ""')
        if [[ -n "${label}" ]]; then
            echo "    ${src} -->|${label}| ${tgt}"
        else
            echo "    ${src} --> ${tgt}"
        fi
        link_styles+="    linkStyle ${link_index} stroke:#DC143C,stroke-width:2px"$'\n'
        link_index=$((link_index + 1))
    done < <(echo "${diff}" | jq -c '.removedConnections[]?')

    # Unchanged connections
    while IFS= read -r conn; do
        [[ -z "${conn}" ]] && continue
        local src tgt label
        src=$(echo "${conn}" | jq -r '.sourceId | gsub("[^a-zA-Z0-9]"; "_")')
        tgt=$(echo "${conn}" | jq -r '.targetId | gsub("[^a-zA-Z0-9]"; "_")')
        label=$(echo "${conn}" | jq -r '.type // ""')
        if [[ -n "${label}" ]]; then
            echo "    ${src} -->|${label}| ${tgt}"
        else
            echo "    ${src} --> ${tgt}"
        fi
        link_index=$((link_index + 1))
    done < <(echo "${diff}" | jq -c '.unchangedConnections[]?')

    echo ""
    if [[ -n "${link_styles}" ]]; then
        printf '%s' "${link_styles}"
    fi

    cat << 'EOF'
    classDef removed fill:#FFB6C1,stroke:#DC143C
```

EOF
}

# Render "After" Mermaid diagram (unchanged + added + modified resources/connections)
render_mermaid_after() {
    local diff="$1"

    # Skip if nothing to show
    local added_res unchanged_res modified_res
    added_res=$(echo "${diff}" | jq '.addedResources | length')
    unchanged_res=$(echo "${diff}" | jq '[.unchangedResources[]?] | length')
    modified_res=$(echo "${diff}" | jq '.modifiedResources | length')
    [[ "${added_res}" -eq 0 && "${unchanged_res}" -eq 0 && "${modified_res}" -eq 0 ]] && return

    cat << 'EOF'
#### After

```mermaid
graph LR
EOF

    # Unchanged resources
    echo "${diff}" | jq -r '.unchangedResources[]? | "    " + (.id | gsub("[^a-zA-Z0-9]"; "_")) + "[\"" + (.name // .id) + " (" + (.type // "") + ")\"]:::unchanged"'
    # Added resources (highlighted green)
    echo "${diff}" | jq -r '.addedResources[] | "    " + (.id | gsub("[^a-zA-Z0-9]"; "_")) + "[\"" + (.name // .id) + " (" + (.type // "") + ")\"]:::added"'
    # Modified resources (highlighted yellow)
    echo "${diff}" | jq -r '.modifiedResources[] | "    " + (.id | gsub("[^a-zA-Z0-9]"; "_")) + "[\"" + (.name // .id) + " (" + (.type // "") + ")\"]:::modified"'

    local link_index=0
    local link_styles=""

    # Added connections (green)
    while IFS= read -r conn; do
        [[ -z "${conn}" ]] && continue
        local src tgt label
        src=$(echo "${conn}" | jq -r '.sourceId | gsub("[^a-zA-Z0-9]"; "_")')
        tgt=$(echo "${conn}" | jq -r '.targetId | gsub("[^a-zA-Z0-9]"; "_")')
        label=$(echo "${conn}" | jq -r '.type // ""')
        if [[ -n "${label}" ]]; then
            echo "    ${src} -->|${label}| ${tgt}"
        else
            echo "    ${src} --> ${tgt}"
        fi
        link_styles+="    linkStyle ${link_index} stroke:#228B22,stroke-width:2px"$'\n'
        link_index=$((link_index + 1))
    done < <(echo "${diff}" | jq -c '.addedConnections[]?')

    # Unchanged connections
    while IFS= read -r conn; do
        [[ -z "${conn}" ]] && continue
        local src tgt label
        src=$(echo "${conn}" | jq -r '.sourceId | gsub("[^a-zA-Z0-9]"; "_")')
        tgt=$(echo "${conn}" | jq -r '.targetId | gsub("[^a-zA-Z0-9]"; "_")')
        label=$(echo "${conn}" | jq -r '.type // ""')
        if [[ -n "${label}" ]]; then
            echo "    ${src} -->|${label}| ${tgt}"
        else
            echo "    ${src} --> ${tgt}"
        fi
        link_index=$((link_index + 1))
    done < <(echo "${diff}" | jq -c '.unchangedConnections[]?')

    echo ""
    if [[ -n "${link_styles}" ]]; then
        printf '%s' "${link_styles}"
    fi

    cat << 'EOF'
    classDef added fill:#90EE90,stroke:#228B22
    classDef modified fill:#FFFACD,stroke:#DAA520
    classDef unchanged fill:#F0F0F0,stroke:#808080
```

EOF
}

main() {
    local diff="${DIFF_JSON}"
    
    # Build comment body
    local body=""
    
    # Header
    body+="## 📊 Application Graph Changes"$'\n\n'
    
    # Summary badges
    local badges
    badges=$(render_badges "${diff}")
    [[ -n "${badges}" ]] && body+="${badges}"$'\n\n'
    
    # Mermaid diagrams rendered outside <details> to avoid GitHub rendering issues
    # Note: $() strips trailing newlines, so we append them explicitly
    # to prevent sections from running together.
    if [[ "${INCLUDE_MERMAID}" == "true" ]]; then
        body+="$(render_mermaid_before "${diff}")"$'\n\n'
        body+="$(render_mermaid_after "${diff}")"$'\n\n'
    fi
    
    # Collapsible details for change tables
    body+="<details>"$'\n'
    body+="<summary>View Details</summary>"$'\n\n'
    
    # Change sections
    body+="$(render_added_section "${diff}")"$'\n\n'
    body+="$(render_removed_section "${diff}")"$'\n\n'
    body+="$(render_modified_section "${diff}")"$'\n\n'
    body+="$(render_connections_section "${diff}")"$'\n\n'
    
    body+="</details>"$'\n\n'
    
    # Footer
    body+="---"$'\n'
    body+="*Generated by Radius App Graph Action*"$'\n'
    
    # Output for GitHub Actions using heredoc delimiter
    {
        echo "comment-body<<RENDER_EOF"
        echo "${body}"
        echo "RENDER_EOF"
    } >> "${GITHUB_OUTPUT:-/dev/stdout}"
    
    echo "Render complete"
}

main "$@"
