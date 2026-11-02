#!/usr/bin/env bash
# SessionStart: print short, dynamic repo context (stdout is added to context)
# and warn when user-level Claude config could leak into this repo's sessions.
set -uo pipefail
root="${CLAUDE_PROJECT_DIR:-$(git rev-parse --show-toplevel)}"
cd "$root" || exit 0

echo "Repo state:"
echo "- branch: $(git branch --show-current) ($(git status --porcelain | wc -l | tr -d ' ') uncommitted paths)"

last="$(ls backend/migrations/*.up.sql 2>/dev/null | xargs -n1 basename | cut -d_ -f1 | sort | tail -1)"
if [ -n "$last" ]; then
  next=$(printf '%03d' $((10#$last + 1)))
  echo "- migrations: highest is $last, next free is $next (012 is an intentional gap; check unmerged branches)"
fi

if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
  ports="$(docker ps --format '{{.Names}} {{.Ports}}' 2>/dev/null | grep -E 'db|postgres' | head -1)"
  echo "- docker: up; postgres: ${ports:-not running}"
else
  echo "- docker: not running"
fi
[ -d frontend/node_modules ] || echo "- frontend/node_modules missing: run 'npm ci' in frontend/"

# Isolation check: anything here changes behaviour outside the repo's own setup.
home="${HOME:-}"
leaks=()
[ -f "$home/.claude/CLAUDE.md" ] && leaks+=("~/.claude/CLAUDE.md")
if [ -f "$home/.claude/settings.json" ]; then
  jq -e '(.hooks // {}) != {} or (.permissions.allow // []) != [] or (.enabledPlugins // {}) != {}' \
    "$home/.claude/settings.json" >/dev/null 2>&1 && leaks+=("~/.claude/settings.json (hooks/permissions/plugins)")
fi
[ -d "$home/.claude/agents" ] && leaks+=("~/.claude/agents")
[ -d "$home/.claude/commands" ] && leaks+=("~/.claude/commands")
if [ ${#leaks[@]} -gt 0 ]; then
  echo "- WARNING isolation: user-level config is active and may affect this repo: ${leaks[*]}"
fi
exit 0
