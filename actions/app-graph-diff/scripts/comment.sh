#!/bin/bash

# ============================================================================
# Post or update PR comment with graph diff
# Creates new comment or updates existing one matching the header marker
# ============================================================================

set -euo pipefail

# Check jq availability
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed" >&2
    exit 1
fi

# Check curl availability
if ! command -v curl &> /dev/null; then
    echo "Error: curl is required but not installed" >&2
    exit 1
fi

# Required environment variables
: "${GITHUB_TOKEN:?GITHUB_TOKEN environment variable is required}"
: "${COMMENT_BODY:?COMMENT_BODY environment variable is required}"
: "${COMMENT_HEADER:?COMMENT_HEADER environment variable is required}"
: "${PR_NUMBER:?PR_NUMBER environment variable is required}"
: "${REPO:?REPO environment variable is required}"

readonly API_BASE="https://api.github.com"

# Make authenticated API request
gh_api() {
    local method="$1"
    local endpoint="$2"
    local data="${3:-}"
    
    local args=(-s -H "Authorization: token ${GITHUB_TOKEN}" -H "Accept: application/vnd.github.v3+json")
    
    if [[ -n "${data}" ]]; then
        args+=(-H "Content-Type: application/json" -d "${data}")
    fi
    
    curl "${args[@]}" -X "${method}" "${API_BASE}${endpoint}"
}

# Find existing comment by header marker
find_existing_comment() {
    local comments
    comments=$(gh_api GET "/repos/${REPO}/issues/${PR_NUMBER}/comments")
    
    # Find comment containing our header marker
    echo "${comments}" | jq -r --arg header "${COMMENT_HEADER}" \
        '.[] | select(.body | contains($header)) | .id' | head -1
}

# Create new comment
create_comment() {
    local body="$1"
    
    # Prepend header marker
    local full_body="${COMMENT_HEADER}"$'\n'"${body}"
    
    local payload
    payload=$(jq -n --arg body "${full_body}" '{body: $body}')
    
    gh_api POST "/repos/${REPO}/issues/${PR_NUMBER}/comments" "${payload}"
}

# Update existing comment
update_comment() {
    local comment_id="$1"
    local body="$2"
    
    # Prepend header marker
    local full_body="${COMMENT_HEADER}"$'\n'"${body}"
    
    local payload
    payload=$(jq -n --arg body "${full_body}" '{body: $body}')
    
    gh_api PATCH "/repos/${REPO}/issues/comments/${comment_id}" "${payload}"
}

main() {
    echo "Posting PR comment to ${REPO}#${PR_NUMBER}"
    
    # Check for existing comment
    local existing_id
    existing_id=$(find_existing_comment)
    
    if [[ -n "${existing_id}" ]]; then
        echo "Updating existing comment: ${existing_id}"
        update_comment "${existing_id}" "${COMMENT_BODY}"
    else
        echo "Creating new comment"
        create_comment "${COMMENT_BODY}"
    fi
    
    echo "Comment posted successfully"
}

main "$@"
