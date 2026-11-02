#!/usr/bin/env bash
# PostToolUse: gofmt edited Go files. Never blocks.
set -uo pipefail
path="$(jq -r '.tool_input.file_path // empty' 2>/dev/null)"
case "$path" in
  *.go) [ -f "$path" ] && gofmt -w "$path" 2>/dev/null ;;
esac
exit 0
