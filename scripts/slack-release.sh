#!/bin/bash
# ---
# name: slack-release
# description: Generate a Slack announcement for a tap release
# category: release
# parameters:
#   - name: version
#     type: string
#     required: false
#     description: "Version tag (e.g. v0.2.0). Defaults to latest git tag."
#   - name: copy
#     type: bool
#     required: false
#     default: "true"
#     description: "Copy output to clipboard (macOS)"
# ---
set -euo pipefail

cd "$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"

VERSION="${TAP_PARAM_VERSION:-$(git describe --tags --abbrev=0 2>/dev/null || echo "")}"
COPY="${TAP_PARAM_COPY:-true}"

if [[ -z "$VERSION" ]]; then
  echo "Error: No version tag found. Pass one with: version=v0.2.0" >&2
  exit 1
fi

# Find the previous tag
PREV_TAG=$(git tag --sort=-v:refname | grep -E '^v[0-9]' | grep -A1 "^${VERSION}$" | tail -1)
if [[ "$PREV_TAG" == "$VERSION" || -z "$PREV_TAG" ]]; then
  RANGE="$VERSION"
else
  RANGE="${PREV_TAG}..${VERSION}"
fi

# Collect commits, grouped by type
FEATS=$(git log "$RANGE" --pretty=format:"%s" --no-merges 2>/dev/null | grep -iE "^feat(\(.+\))?:" | sed -E 's/^feat(\([^)]*\))?: *//' || true)
FIXES=$(git log "$RANGE" --pretty=format:"%s" --no-merges 2>/dev/null | grep -iE "^fix(\(.+\))?:" | sed -E 's/^fix(\([^)]*\))?: *//' || true)

# --- Build "What's new" from RELEASE_NOTES.md if present, else from commits ---
NOTES_FILE="RELEASE_NOTES.md"
ITEMS=""

if [[ -f "$NOTES_FILE" ]]; then
  # Use hand-written notes (skip empty lines at start/end)
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    ITEMS="${ITEMS}${line}\n"
  done < "$NOTES_FILE"
else
  # Fall back to commit messages
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    line="$(echo "${line:0:1}" | tr '[:lower:]' '[:upper:]')${line:1}"
    ITEMS="${ITEMS}• ${line}\n"
  done <<< "$FEATS"

  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    line="$(echo "${line:0:1}" | tr '[:lower:]' '[:upper:]')${line:1}"
    ITEMS="${ITEMS}• :bug: ${line}\n"
  done <<< "$FIXES"

  if [[ -z "$ITEMS" ]]; then
    ITEMS="• Various improvements and fixes\n"
  fi
fi

# Strip the 'v' prefix for display
DISPLAY_VERSION="${VERSION#v}"

# Build the Slack message
# Note: single code block with both commands, separated sections
MSG=":rocket: *tap v${DISPLAY_VERSION}* is out!

*What's new:*
$(echo -e "$ITEMS")
*New user?*
\`\`\`brew install shaiu/tap/tap\`\`\`

*Already installed?*
\`\`\`brew upgrade tap\`\`\`

<https://github.com/shaiungar/tap/releases/tag/${VERSION}|:page_facing_up: Release notes>"

echo "─── Slack message ───"
echo ""
echo "$MSG"
echo ""
echo "─────────────────────"

# Hint about RELEASE_NOTES.md
if [[ ! -f "$NOTES_FILE" ]]; then
  echo ""
  echo "Tip: Create RELEASE_NOTES.md with user-friendly feature descriptions"
  echo "     and usage examples. The script will use it instead of commit messages."
  echo ""
  echo "     Example RELEASE_NOTES.md:"
  echo "     • *View source code* — press \`v\` to see any script's code with line numbers"
  echo "     • *Edit in place* — press \`e\` to open the script in your \$EDITOR"
fi

# Copy to clipboard on macOS
if [[ "$COPY" == "true" ]] && command -v pbcopy &>/dev/null; then
  echo "$MSG" | pbcopy
  echo ""
  echo "✓ Copied to clipboard"
fi
