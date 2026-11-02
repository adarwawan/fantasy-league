# Backend (Go)

Module `fantasy-league`, Go 1.22+ (toolchain 1.25 locally). chi router, pgx, go-redis, gocron.

## Layout

- `cmd/server` — API + scheduler; routes are registered in `cmd/server/main.go`. `cmd/sync*` — one-shot sync runners. (`cmd/seed-gwstats` is a gitignored local helper.)
- `internal/fantasy` — `Source` interface + domain types (`types.go`). Optional capabilities are separate interfaces (e.g. `EntryLoader`).
- `internal/sources/{fpl,wcf,odds}` — adapters: `client.go` (HTTP), `mapper.go` (raw → domain, unit-tested), `source.go`.
- `internal/store` — Postgres (`postgres.go`, `queries.go`) and Redis (`cache.go`). Store takes **external IDs as strings** and resolves UUIDs internally via subqueries.
- `internal/sync/syncer.go` — orchestration; must-have/watchlist flags are computed at sync time and served from cache.
- `internal/handler` — thin HTTP handlers; `respond.go` for JSON/errors. Clamp every query param (bounds cache-key cardinality).
- `internal/{musthave,watchlist}` — pure rule evaluators, driven by `config.yaml`.
- `internal/{setpiece,shotzone}` — isolated Understat/FPL-derived modules (client/parse/aggregate/store/service/handler/types).

## Conventions

- Cache keys via `CacheKey(gameID, resource, params...)`, e.g. `fpl:players:MID:form:10000`. Isolated modules use their own prefix; `shotzone` deliberately shares `setpiece:match:{id}` so a match is scraped once.
- Config: add a field to `internal/config/config.go`, a block in `config.yaml`, and an env override (`envBoolOr`/`envIntOr`/`envStringOr`). Module env prefixes: `SP_`, `SZ_`.
- Tests: table-driven, next to the code; fixtures in `testdata/`. Tests that hit the network must call `testing.Short()` → `t.Skip` (see `setpiece/parse_integration_test.go`).
- FPL rate limits: bootstrap-static is cached; `FetchPicks` runs with bounded workers (`sync.picks_workers`) and jittered 429 backoff. Don't add unbounded fan-out.
- Fixture "finished" = FPL `finished`. Current GW = `MIN(gw) WHERE NOT finished`.
- Sync endpoint is secret-guarded (`X-Sync-Secret`); keep new control endpoints behind the same pattern.

## Migrations

- Files: `NNN_name.up.sql` + `NNN_name.down.sql` in `migrations/`, applied with golang-migrate (`make migrate-up`; CI runs it before deploy).
- Next free number: highest in `migrations/` **and** any number named in unmerged branches/docs (`012` is an intentional gap, don't fill it).
- Never edit a migration that has landed on `main`. `scripts/check.sh migrations` enforces pairs and unique numbers.

## Verify

```bash
cd backend && gofmt -l . && go vet ./... && go test -short ./...    # or: make check-backend (repo root)
```

For data features, also run the stack with seeded data and `curl localhost:8080/api/...` — mark GWs finished in `fixtures`, insert `player_gw_stats`, flush Redis.
