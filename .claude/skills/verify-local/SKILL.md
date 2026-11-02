---
name: verify-local
description: Bring up the local stack (Postgres, Redis, API, frontend) and verify a change against real or seeded data. Use before calling a backend, data, or UI change done; unit tests and tsc alone are not enough.
---

# Verify locally

Past failures this prevents: a SQL error (ambiguous `gw` in a CTE) passed build and tests but returned empty minutes; a per-GW average read 112' on double gameweeks. Both only showed up with real data.

## 1. Infrastructure

```bash
docker compose up -d
docker port fantasy-league-01-db-1        # expect 5432/tcp -> 0.0.0.0:5433
lsof -iTCP:5432 -iTCP:5433 -iTCP:8080 -sTCP:LISTEN
```

If compose reports healthy but `docker port` shows nothing, another project's container owns 5432; local Postgres is expected on **5433** (`backend/.env`). Never print or `cat` `.env` files; the server reads `backend/.env` itself.

## 2. Schema and API

```bash
cd backend
DATABASE_URL="postgres://fantasy:fantasy@localhost:5433/fantasy?sslmode=disable" make migrate-up
make run                                  # :8080, run in the background
curl -s localhost:8080/health
```

## 3. Data

Pick what the change needs:

- **Real data**: `cd backend && go run ./cmd/sync` (FPL), `go run ./cmd/sync-setpiece`, `go run ./cmd/sync-shotzone`. These hit live APIs; the FPL API rate-limits, so don't loop them.
- **Seeded data** for logic that needs a specific shape (double GWs, finished fixtures): mark GWs finished in `fixtures`, insert `player_gw_stats`, then flush Redis so cached payloads don't hide the result:
  ```bash
  docker exec fantasy-league-01-db-1 psql -U fantasy -d fantasy -c "<sql>"
  docker exec fantasy-league-01-cache-1 redis-cli FLUSHALL
  ```
  Say in your report that the data was seeded, and revert or note it.

## 4. Exercise the change

- API: `curl -s 'localhost:8080/api/fpl/players?...'` and inspect fields, counts, edge cases (blank/double GW, missing odds, empty window). Compare against the numbers you expect, not just "200 OK".
- UI: start the `frontend` launch config (`:5173`), open the page, exercise filters and every state, resize to ~375px, read the console and network panel for errors.
- Isolated modules: confirm row counts after a sync (e.g. matches processed, events stored) and report them.

## 5. Finish

`make check`, stop background servers you started, and report what you actually observed, including what you could not verify.
