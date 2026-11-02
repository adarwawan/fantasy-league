---
name: new-migration
description: Add a Postgres migration to backend/migrations (new table, column, index). Use whenever the schema changes. Picks the next safe number and writes the up/down pair.
---

# New migration

Migrations are golang-migrate files in `backend/migrations/`, applied by CI before each backend deploy. They are append-only: a migration that exists in `HEAD` can never be edited (the guard hook blocks it).

## Steps

1. **Pick the number.**
   - `ls backend/migrations | tail -4` for the highest local number.
   - `git branch -a` and `git log --all --oneline -- backend/migrations | head` to see numbers claimed by unmerged branches.
   - `012` is an intentional gap: never use it. Take the highest number + 1 (the SessionStart hook prints it).
2. **Create the pair**, same number and name: `NNN_short_name.up.sql` and `NNN_short_name.down.sql`.
3. **Write the up** to be safe on a populated production table: `CREATE TABLE IF NOT EXISTS`, `ADD COLUMN IF NOT EXISTS`, nullable or defaulted new columns, no table rewrites on large tables, indexes for the new query paths.
   - Core tables hold FPL/WCF data keyed by `game_id`; new columns on `player_gw_stats` etc. need a default so existing rows stay valid.
   - Isolated modules use their own prefix (`sp_*`, `sz_*`), no foreign keys to players/teams (see `013_shot_zone.up.sql`).
4. **Write the down** to exactly reverse the up (`DROP TABLE IF EXISTS`, `DROP COLUMN IF EXISTS`), in reverse order.
5. **Apply and round-trip locally** against the docker database (credentials are the local compose ones):
   ```bash
   cd backend
   DATABASE_URL="postgres://fantasy:fantasy@localhost:5433/fantasy?sslmode=disable" make migrate-up
   DATABASE_URL="postgres://fantasy:fantasy@localhost:5433/fantasy?sslmode=disable" migrate -path ./migrations -database "$DATABASE_URL" down 1
   DATABASE_URL="postgres://fantasy:fantasy@localhost:5433/fantasy?sslmode=disable" make migrate-up
   ```
   Local Postgres is on 5433 (see root `CLAUDE.md`). Never run `down` against a remote database.
6. `make check-migrations` (pairs + unique numbers), then update the store code and any store test.

## Don't

- Don't edit or renumber a migration already on `main`.
- Don't put data backfills that take minutes inside a migration; do them in the sync.
