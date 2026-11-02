---
name: ship-pr
description: Branch, commit, push, and open a pull request following the repo's SDLC flow. Manual only; run when the user asks to ship or open a PR.
disable-model-invocation: true
---

# Ship a PR

Flow is defined in `docs/sdlc.md`. A human merges; CI deploys on merge to `main`. Never push to `main`, force-push, or merge (the guard hook blocks these).

## Steps

1. **Check state**: `git status`, `git branch --show-current`, `git log --oneline main..HEAD`. If on `main`, create a branch first: `feat/<slug>`, `fix/<slug>`, or `chore/<slug>` (one concern per branch).
2. **Verify**: `make check` must pass. For data or UI changes confirm the `verify-local` steps were done; if not, say so in the PR instead of implying they were.
3. **Review the diff** yourself (`git diff main...HEAD --stat`, then the files). Remove debug output, stray files, and anything unrelated. Confirm no secrets and no `.env*` are staged.
4. **Commit** with explicit paths (no `git add -A` blindly): `type(scope): subject`, type in `feat|fix|chore|perf|refactor`, imperative, under ~70 characters, with a body explaining why when it isn't obvious. End the message with the attribution line from the session's system reminder.
5. **Push the branch** (`git push -u origin <branch>`), then **create the PR** against `main` with `gh pr create`. Only do this when the user asked to ship; pushing and opening a PR are outward-facing.
   - Title = the squash commit subject (GitHub adds ` (#N)`).
   - Body: what changed and why (2-4 bullets), how it was verified (commands run, data used, pages exercised), anything not verified, migration number if one was added, and the PR attribution line from the session's system reminder.
6. **After opening**: report the PR link. Do not merge. Do not poll CI; the app's PR tools handle checks.

## Notes

- If the change adds a migration, call out that CI runs `migrate up` before the backend deploy and that the down file exists.
- If a lasting decision came out of the work, add it with the `decision-record` skill in the same PR.
