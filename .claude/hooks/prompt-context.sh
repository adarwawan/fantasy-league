#!/usr/bin/env bash
# UserPromptSubmit: (1) refuse prompts that contain pasted secrets, (2) add
# pointers to the repo rules relevant to what the prompt is about.
# Exit 2 blocks the prompt (stderr is shown to the user and the prompt is
# dropped); stdout on exit 0 is added to Claude's context.
set -uo pipefail
prompt="$(jq -r '.prompt // empty')"
[ -z "$prompt" ] && exit 0
root="${CLAUDE_PROJECT_DIR:-$(git rev-parse --show-toplevel)}"

# --- 1. Secret scan --------------------------------------------------------
# Placeholders and the local docker credentials are fine; real-looking values are not.
secret_patterns=(
  'sk-[A-Za-z0-9_-]{20,}'                       # API keys
  'AKIA[0-9A-Z]{16}'                            # AWS / R2 access key id
  'gh[pousr]_[A-Za-z0-9]{30,}'                  # GitHub tokens
  'github_pat_[A-Za-z0-9_]{30,}'
  'FlyV1 [A-Za-z0-9+/=_,-]{20,}'                # Fly tokens
  '-----BEGIN [A-Z ]*PRIVATE KEY-----'
  'postgres(ql)?://[^[:space:]:@/]+:[^[:space:]@]{4,}@[^[:space:]/]*(neon\.tech|fly\.dev|amazonaws\.com)'
  'rediss://[^[:space:]]*:[^[:space:]@]{8,}@'   # Upstash
  '(API_KEY|API_TOKEN|AUTH_TOKEN|SECRET_ACCESS_KEY|SYNC_SECRET|SYNC_ENDPOINT_SECRET|PASSWORD)[A-Z_]*[=:][[:space:]]*["'"'"']?[A-Za-z0-9+/_.-]{16,}'
)
for p in "${secret_patterns[@]}"; do
  if grep -Eq -- "$p" <<<"$prompt"; then
    {
      echo "Prompt blocked: it looks like it contains a secret."
      echo "Remove the value and refer to it by name (e.g. FLY_API_TOKEN). Secrets live in .env / .env.infra, never in chat."
    } >&2
    exit 2
  fi
done

# --- 2. Context routing ----------------------------------------------------
lc="$(tr '[:upper:]' '[:lower:]' <<<"$prompt")"
hints=()
has() { grep -Eq -- "$1" <<<"$lc"; }

if has 'migrat|schema|alter table|create table|new table|new column|add column'; then
  last="$(ls "$root"/backend/migrations/*.up.sql 2>/dev/null | xargs -n1 basename | cut -d_ -f1 | sort | tail -1)"
  next="$(printf '%03d' $((10#${last:-0} + 1)))"
  hints+=("Migrations: append-only, up+down pair, next free number is $next (012 is an intentional gap, leave it). Landed migrations can't be edited (guard hook). See backend/CLAUDE.md.")
fi
if has 'stat|metric|form\b|window|season|average|avg|rate|rolling|leaderboard|ranking'; then
  hints+=("Metrics: default to a rolling window, not season totals (docs/decisions/001). Raw values from the backend, labels derived on the frontend. Rates/averages are per fixture, not per GW, so double/blank GWs stay correct (docs/decisions/003).")
fi
if has 'finished|current gw|current gameweek|upcoming fixture|provisional'; then
  hints+=("Fixture completion uses FPL finished, not finished_provisional; the post-match lag is intended (docs/decisions/002).")
fi
if has 'shot.?zone|set.?piece|understat|new module|new data feature|isolated'; then
  hints+=("Isolated module shape: own tables (sp_*/sz_*), own config block + env prefix, own /api route, never touches the players/fixtures pipeline. Shot-zone spec: docs/decisions/005.")
fi
if has 'minutes|rotation|nailed|start rate|benched'; then
  hints+=("Minutes security: chip derived on the frontend (utils/minutes.ts), start_rate and avg_minutes are per fixture over the last 5 finished GWs (docs/decisions/003).")
fi
if has 'page|component|\bui\b|frontend|tailwind|mobile|responsive|layout|drawer|table'; then
  hints+=("Frontend: reuse components/ui primitives and QueryBoundary, add api/ + hooks/ + types/, check a ~375px viewport. See frontend/CLAUDE.md.")
fi
if has 'new game|source adapter|wcf|uclf|fantasy\.source'; then
  hints+=("A new game is a fantasy.Source adapter under internal/sources/{id}/, namespaced by game_id; no handler or routing changes.")
fi
if has 'commit|pull request|\bpr\b|ship|release|deploy|branch'; then
  hints+=("Ship flow (docs/sdlc.md): feat|fix|chore/<slug> branch, 'type(scope): subject' title, PR to main, human merges, CI deploys on merge. Never push to main or deploy by hand.")
fi

if [ ${#hints[@]} -gt 0 ]; then
  echo "Relevant repo context (from .claude/hooks/prompt-context.sh):"
  for h in "${hints[@]}"; do echo "- $h"; done
fi
exit 0
