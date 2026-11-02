#!/usr/bin/env bash
# Stop: run scripts/check.sh for the areas touched in this working tree.
# Exit 2 sends the failure back to Claude so it fixes it before finishing.
set -uo pipefail
input="$(cat)"
# Already continuing because of this hook: don't loop.
[ "$(jq -r '.stop_hook_active // false' <<<"$input")" = "true" ] && exit 0

root="${CLAUDE_PROJECT_DIR:-$(git rev-parse --show-toplevel)}"
cd "$root" || exit 0

changed="$(git status --porcelain --untracked-files=all | cut -c4-)"
[ -z "$changed" ] && exit 0

targets=()
grep -q '^backend/migrations/' <<<"$changed" && targets+=(migrations)
grep -q '^backend/.*\.\(go\|mod\|sum\)$' <<<"$changed" && targets+=(backend)
grep -q '^frontend/' <<<"$changed" && targets+=(frontend)
[ ${#targets[@]} -eq 0 ] && exit 0

out=""
for t in "${targets[@]}"; do
  if ! res="$(scripts/check.sh "$t" 2>&1)"; then
    out+=$'\n'"--- scripts/check.sh $t ---"$'\n'"$(tail -40 <<<"$res")"
  fi
done

if [ -n "$out" ]; then
  echo "Checks failed for the files you changed. Fix before finishing:$out" >&2
  exit 2
fi
exit 0
