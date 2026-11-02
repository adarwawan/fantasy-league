---
name: decision-record
description: Record a decision that future work must respect in docs/decisions. Use when the user settles a product rule, data-source choice, trade-off, or finds a non-obvious gotcha worth keeping. Replaces personal memory for project knowledge.
---

# Decision record

Project knowledge lives in the repo, not in `~/.claude` (auto memory is off). The repo keeps only two kinds of docs: `docs/sdlc.md` (flow) and `docs/decisions/` (binding decisions). No plans, task lists, or backlogs.

## When to write one

Yes: a product rule ("form, not season"), a deliberate non-fix ("use `finished`, lag is accepted"), a data-source or architecture choice with a reason, a gotcha whose cause is not visible in the code.

No: what the code already shows, what git history records, a plan for work not yet done, a one-off bug fix with no lasting lesson.

## Steps

1. `ls docs/decisions` and read any related record. Update it instead of creating a duplicate; delete a record that turned out to be wrong.
2. Next number = highest + 1 (numbers are not reused; `004` is intentionally absent).
3. Create `docs/decisions/NNN-short-name.md`:
   ```markdown
   # NNN — Title

   **Decision:** what is true, in one or two sentences.

   **Why:** the reason or constraint, ideally who/what drove it.

   **How to apply:** what to do or avoid in future work; code paths involved.
   ```
   Add **Gotcha** / **Known gap** sections only if real. Keep it under ~40 lines.
4. If the decision is an always-on invariant, add one line to the Invariants list in the root `CLAUDE.md` pointing at the record. If it only matters to some work, add a keyword hint to `.claude/hooks/prompt-context.sh` instead.
5. Mention the new record in your reply so the user can correct the wording.
