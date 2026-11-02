# SDLC flow

How a change goes from idea to production in this repo. Decisions that constrain the work are in [`decisions/`](decisions/); the rules for the code itself are in the `CLAUDE.md` files.

## 1. Frame
- Check `docs/decisions/` and the invariants in the root `CLAUDE.md` first (form not season, raw backend values, fixture-level math, settled data).
- A new data feature copies the isolated-module shape (own tables, config block, env prefix, routes). A new game is a `fantasy.Source` adapter.
- If the change needs a migration, pick the number from `main` **and** unmerged branches (`git branch -a`).

## 2. Branch
- Branch from `main`, named `feat/<slug>`, `fix/<slug>`, or `chore/<slug>`. One concern per branch.
- Never commit to `main` directly.

## 3. Build
- Backend: handler stays thin, rules live in pure packages with table-driven tests, config goes in `config.yaml` with an env override.
- Frontend: reuse `components/ui` primitives, add `api/` + `hooks/` + `types/`, and handle loading/empty/error states.
- Migrations are append-only, with an up and a down file.

## 4. Verify locally
1. `make check` runs automatically when Claude finishes a turn (Stop hook) and in CI; run it yourself otherwise (migrations, gofmt on changed files, vet, `go test -short`, `tsc`).
2. Run the stack (`docker compose up -d`, `cd backend && make run`, `cd frontend && npm run dev`) and exercise the change against real or seeded data. Unit tests alone have missed silent SQL errors.
3. UI changes: check the page at a ~375px viewport as well as desktop.
4. Network-dependent tests are opt-in (`-run Live`), never part of the default check.

## 5. Ship
- Commit as `type(scope): subject` using `feat|fix|chore|perf|refactor`; the PR is squash-merged, so the PR title becomes the commit and gets ` (#N)`.
- Open a PR against `main` from the branch. The PR description states what changed and how it was verified.
- A human merges. Claude does not merge PRs, force-push, or push to `main`.

## 6. Deploy (automatic on merge to `main`)
- **Backend** (`backend/**`, `fly.toml`): CI runs `migrate up` against Neon, then `flyctl deploy` and starts the machine.
- **Frontend** (`frontend/**`): CI runs `npm ci && npm run build` and publishes to Cloudflare Pages.
- `.github/workflows/ci.yml` runs `scripts/check.sh all` on every PR and non-`main` push. Deploys do not wait on it, so merge only green PRs.
- Infra (`infra/terraform`) is applied by hand by the owner, never by an agent.

## 7. Operate
- A daily scheduled sync refreshes data. `GET /health/sync` returns 503 when the last successful sync is older than `sync.freshness_max_age`, and external monitors alert on it.
- Manual sync is `GET /api/{game}/sync` with `X-Sync-Secret`.
- Season rollover and tournament arming (WCF/Euros) are owner-run procedures.

## 8. Record
- A decision that future work must respect (a product rule, a data-source choice, a gotcha with a non-obvious cause) gets a short file in `docs/decisions/NNN-name.md`: decision, why, how to apply.
- Plans, task lists, and backlogs are not kept in the repo.
