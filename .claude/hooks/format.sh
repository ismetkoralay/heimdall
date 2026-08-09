#!/usr/bin/env bash
# PostToolUse(Write|Edit): format the Go file that was just written.
# Never fails the tool call — formatting is a convenience, not a gate.
set -uo pipefail

command -v jq >/dev/null 2>&1 || exit 0

input="$(cat)"
file="$(printf '%s' "$input" | jq -r '.tool_input.file_path // .tool_input.path // empty' 2>/dev/null)"
[ -z "${file:-}" ] && exit 0
[ -f "$file" ] || exit 0

case "$file" in
  *.go)
    command -v gofmt     >/dev/null 2>&1 && gofmt -w "$file"
    command -v goimports >/dev/null 2>&1 && goimports -w "$file"
    ;;
esac

exit 0
