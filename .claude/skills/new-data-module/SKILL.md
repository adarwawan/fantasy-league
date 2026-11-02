---
name: new-data-module
description: Scaffold a new isolated data module (like setpiece, shotzone) end to end - migration, Go package, config, scheduler, routes, CLI, and frontend page. Use for a new PL-wide data feature that must not touch the players/fixtures pipeline.
---

# New isolated data module

Reference implementations: `backend/internal/shotzone/` (newest, smallest) and `backend/internal/setpiece/`. Read `shotzone` first and copy its shape; do not invent a new one. Decisions to respect: `docs/decisions/001` (rolling windows) and `005`.

A module is isolated when it has: its own tables, config block, env prefix, routes, sync, and never imports or writes the players/fixtures pipeline. Pick a 2-letter prefix `xx` (`sp`, `sz` are taken).

## Checklist

1. **Spec first.** Write down (in the reply, not a file) the data source, rolling window, what is excluded, and the UI. Get the user to confirm anything ambiguous; if a lasting rule comes out of it, use the `decision-record` skill.
2. **Migration** `xx_*` tables via the `new-migration` skill. No foreign keys to core tables. Raw events table + derived board tables, unique keys so re-syncs are idempotent, plus a down that drops them all.
3. **Go package** `backend/internal/<name>/`:
   - `types.go`, `client.go` (HTTP; cache raw responses in Redis through `store.Cache`; reuse the shared match cache key if the source is the same), `parse.go`, `aggregate.go` (pure, table-driven tests with `testdata/`), `store.go` (pgx), `service.go` (`Sync(ctx)`), `handler.go`.
   - Handlers: clamp params, decode path params with `url.PathUnescape` (chi returns them percent-encoded when mixed with unescaped characters), use `respond.go` helpers.
   - Window-scoped by default; values raw, no composite scores.
4. **Config**: block in `backend/config.yaml` (`enabled`, `season`, `window_matches`, `sync_cron`), fields in `internal/config/config.go`, env overrides with prefix `XX_` via `envBoolOr/envIntOr/envStringOr`.
5. **Wire** in `cmd/server/main.go` like `szHandler`: build client/store/service when enabled, `gocron.CronJob(cfg.XXSyncCron, false)`, an optional startup sync goroutine, and `r.Route("/api/<kebab-name>", ...)` registered only if the handler is non-nil. Stagger the cron after the existing ones (06:00 set-piece, 06:30 shot-zone).
6. **CLI** `cmd/sync-<name>/main.go` copying `cmd/sync-shotzone` (forces `Enabled: true`).
7. **Frontend** via the `add-page` skill: `api/<name>.ts` (game-less route, like `api/shotzone.ts`), `hooks/use<Name>.ts` (`staleTime` 30 min), `types/<name>.ts`, page, route, and a `games: ['fpl']` nav entry in `AppShell`.
8. **Verify** with the `verify-local` skill: migrate, run the sync CLI against the live source, hit the endpoints, load the page. Report real row counts.
9. `make check`, then `ship-pr`.

## Pitfalls seen before

- A source's team set may include non-current clubs (Coventry, Hull, Ipswich in Understat 2026); don't assert 20 teams.
- Exclude penalties and own goals deliberately and say so in the UI copy.
- Ambiguous column names in CTE joins fail silently as empty results; always run against data, not just `go test`.
