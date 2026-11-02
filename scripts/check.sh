#!/usr/bin/env bash
# One verify command, shared by local runs, Claude hooks and CI.
# Usage: scripts/check.sh [backend|frontend|migrations|all]   (default: all)
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
target="${1:-all}"
failed=0

step() { printf '\n==> %s\n' "$*"; }
run() { "$@" || { echo "FAILED: $*" >&2; failed=1; }; }

check_backend() {
  # gofmt is gated on files changed vs origin/main only: some older files
  # predate the rule, and reformatting them would bury unrelated diffs.
  step "backend: gofmt (changed files)"
  local changed unformatted
  changed="$(cd "$ROOT" && { git diff --name-only --diff-filter=d origin/main -- 'backend/*.go' 2>/dev/null; \
    git ls-files --others --exclude-standard -- 'backend/*.go'; } | sort -u)"
  if [ -n "$changed" ]; then
    unformatted="$(cd "$ROOT" && echo "$changed" | xargs gofmt -l)"
    if [ -n "$unformatted" ]; then
      echo "Not gofmt'd:" >&2; echo "$unformatted" >&2; failed=1
    fi
  fi
  step "backend: go vet"
  (cd "$ROOT/backend" && run go vet ./...)
  step "backend: go test -short (live Understat tests skipped)"
  (cd "$ROOT/backend" && run go test -short ./...)
}

check_frontend() {
  step "frontend: tsc --noEmit"
  if [ ! -d "$ROOT/frontend/node_modules" ]; then
    echo "frontend/node_modules missing; run 'npm ci' in frontend/ first" >&2
    failed=1
    return
  fi
  (cd "$ROOT/frontend" && run npx tsc --noEmit)
}

# Migrations are append-only: every number needs an up+down pair, and no
# number may be used twice. Gaps are allowed (a number can be reserved by an
# unmerged branch).
check_migrations() {
  step "migrations: numbering and up/down pairs"
  local dir="$ROOT/backend/migrations" f n
  for f in "$dir"/*.up.sql; do
    [ -e "$f" ] || continue
    [ -e "${f%.up.sql}.down.sql" ] || { echo "missing down for $(basename "$f")" >&2; failed=1; }
  done
  for f in "$dir"/*.down.sql; do
    [ -e "$f" ] || continue
    [ -e "${f%.down.sql}.up.sql" ] || { echo "missing up for $(basename "$f")" >&2; failed=1; }
  done
  n="$(ls "$dir"/*.up.sql 2>/dev/null | xargs -n1 basename | cut -d_ -f1 | sort | uniq -d)"
  if [ -n "$n" ]; then
    echo "duplicate migration numbers: $n" >&2; failed=1
  fi
}

case "$target" in
  backend)    check_backend ;;
  frontend)   check_frontend ;;
  migrations) check_migrations ;;
  all)        check_migrations; check_backend; check_frontend ;;
  *) echo "usage: $0 [backend|frontend|migrations|all]" >&2; exit 2 ;;
esac

if [ "$failed" -ne 0 ]; then
  printf '\ncheck: FAILED\n' >&2
  exit 1
fi
printf '\ncheck: OK\n'
