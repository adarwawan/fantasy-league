# Fantasy League Dashboard

Full-stack dashboard aggregating FPL (live), WCF and UCLF (dormant) data. Go API + Postgres + Redis, React SPA.
Purpose: help pick transfers/captains **before a deadline**, so it shows recent form and trustworthy (settled) data.

## Map

| Path | What |
|---|---|
| `backend/` | Go module `fantasy-league`. `cmd/` binaries (`server`, `sync*`), `internal/` packages, `migrations/`, `config.yaml` |
| `frontend/` | Vite + React 18 + TS + Tailwind + TanStack Query. Deployed to Cloudflare Pages |
| `infra/terraform/`, `fly.toml` | Fly.io API, Neon Postgres, Upstash Redis, Cloudflare. Deploys run from CI on merge to `main` |
| `docs/` | `sdlc.md` (how a change ships) and `decisions/` (binding decisions) |
| `scripts/check.sh` | The one verify command (also used by hooks/CI) |

Area-specific rules: `backend/CLAUDE.md`, `frontend/CLAUDE.md` (loaded when working in those dirs).

## Commands

```bash
make check                 # migrations + gofmt(changed) + go vet + go test -short + tsc. Run before finishing.
make check-backend | check-frontend | check-migrations

docker compose up -d       # Postgres 16 + Redis 7
cd backend && make run     # API on :8080 (reads backend/.env)
cd backend && DATABASE_URL=... make migrate-up
cd frontend && npm run dev # SPA on :5173 (also via .claude/launch.json "frontend")
```

Live-network tests (Understat) are skipped by `-short`; run explicitly with `go test ./internal/setpiece -run Live -v`.

## Invariants (product + architecture)

1. **Form, not season.** Every stat defaults to a rolling window (last N GWs/matches). No season-total variants unless asked. See `docs/decisions/001-form-not-season.md`.
2. **Backend returns raw values.** No composite scores server-side; chips/labels (e.g. minutes Nailed/Rotation) are derived on the frontend.
3. **Fixture-level math, not gameweek-level.** Averages/rates divide by fixtures played so double/blank GWs stay correct.
4. **Settled data over fresh data.** Use FPL `finished`, never `finished_provisional`. The post-match lag is intended (`docs/decisions/002-finished-flag.md`).
5. **Multi-game by `game_id`.** New game = new adapter under `internal/sources/{id}/` implementing `fantasy.Source`; no handler/routing changes.
6. **Isolated modules** (`setpiece`, `shotzone`) are PL-wide, own tables (`sp_*`, `sz_*`), own config block + env (`SP_*`, `SZ_*`), own `/api/...` routes. They never touch the players/fixtures pipeline. Copy this shape for new data features.
7. **Config split.** Non-secret → `backend/config.yaml`; secrets → `.env` (gitignored). Env vars override yaml.
8. **Migrations are append-only.** Never edit an applied migration; add the next number with an `.up.sql` and `.down.sql`. `012` is an intentional gap: leave it unused. Unmerged branches may reserve numbers too: run `git branch -a` + `ls backend/migrations` before picking one.

## Secrets & safety

- Never read, print, or commit `.env`, `backend/.env`, `.env.infra`, `*.tfvars`. `.env.infra` holds live Fly/Cloudflare/Neon/R2 tokens.
- Don't run `terraform apply|destroy`, `fly deploy`, force-push, or merge PRs. Deploys happen via CI on push to `main`.
- Don't push directly to `main`; work on a branch and open a PR.

## Hooks (`.claude/hooks/`, wired in `.claude/settings.json`)

- `guard.sh` (PreToolUse) blocks secret files, infra apply/destroy, fly deploy/secrets, force-push, push to `main`, PR merge, remote `migrate down`, and edits to migrations already in `HEAD`. If blocked, don't work around it; ask the user.
- `prompt-context.sh` (UserPromptSubmit) rejects prompts containing pasted secrets/tokens and adds pointers to the relevant rules (migrations, metrics, UI, isolated modules, ship flow) based on what the prompt is about. Keyword routing is a hint, not a substitute for reading the linked rule.
- `format.sh` gofmts edited Go files. `stop-check.sh` runs `scripts/check.sh` for the areas you changed and sends failures back to you. `session-start.sh` prints branch, next migration number, Postgres state and warns about user-level config leaking in.

## Skills (`.claude/skills/`)

| Skill | Use for |
|---|---|
| `new-migration` | any schema change: picks the number, writes the up/down pair, round-trips locally |
| `new-data-module` | a new isolated PL-wide data feature (migration, Go package, config, cron, routes, CLI, page) |
| `add-page` | a new or substantially changed frontend page |
| `verify-local` | running the stack and checking a change against real or seeded data |
| `decision-record` | saving a binding decision or gotcha to `docs/decisions/` |
| `ship-pr` | branch, commit, push, PR (manual only) |

## Conventions

- Commits/PR titles: `type(scope): subject (#PR)` (`feat`, `fix`, `chore`, `perf`, `refactor`), e.g. `feat(shotzone): show team shot totals (#72)`.
- Comments explain *why* (design decisions, gotchas), not what. Match the surrounding code's density.
- Verify against real or seeded data, not only unit tests: a silent SQL error (ambiguous column in a CTE) once passed build + tests and produced empty output.

## Local environment gotchas

- Local Postgres is reached on **5433** (`backend/.env`), while `docker-compose.yml` and `INSTALLATION.md` say 5432. Another project's container can grab 5432, leaving ours running with no published port. If Postgres is unreachable but `docker compose ps` is healthy, check `docker port`.
- chi `URLParam` returns the still-percent-encoded segment when a path mixes escaped and unescaped characters (`Nott'm Forest`). Decode with `url.PathUnescape`. Fixed in `shotzone`; **still present** in `/api/setpiece/teams/{understat_team}`.
- Local personal state lives in `.claude/settings.local.json` and `CLAUDE.local.md` (both gitignored). Project knowledge belongs in this repo, not in `~/.claude`.
