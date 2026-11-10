# 006 — FotMob v2 for set pieces and shot zones

Understat is in maintenance, so v2 sources the same two features from FotMob. v1 (Understat) stays untouched and serving until cutover.

**Shape**
- `internal/fotmob/`: client, parser, one sync service. One match fetch feeds both features. Storage, aggregation and read handlers are the existing `setpiece`/`shotzone` code, pointed at v2 tables via `NewStoreV2` (the stores rewrite table names; v1 SQL is unchanged).
- Tables `sp2_*`, `sz2_*` (migration `014`). Config block `fotmob:` / `FM_*` (disabled by default). Sync CLI `cmd/sync-fotmob`. Routes `/api/v2/setpiece`, `/api/v2/shot-zones` with the same response shapes as v1. Cutover: set `fotmob.serve_v2: true` (`FM_SERVE_V2`); the unversioned routes are then backed by v2 and the frontend is unchanged. Rollback = set it false.
- v2 events are keyed by FotMob's per-shot id. v1's (player, minute, zone) key merged two same-minute shots.

**FotMob API (unofficial, no auth as of 2026-11)**
- `GET /api/data/leagues?id=47&ccode3=GBR&season=2026%2F2027` -> `fixtures.allMatches[]` (`status.finished`, per decision 002).
- `GET /api/data/matchDetails?matchId=` -> `content.shotmap.shots[]`. A missing shotmap means shot data isn't published yet: the match stays unprocessed and is retried next run.
- Coordinates are metres on 105x68, already oriented so the shooter attacks x=105. The zone spec of 005 is unchanged (divide by 105/68).

**Situation mapping (to match Understat's classes)**
`Penalty`->taker/penalty, `FreeKick`->taker/dfk, `FromCorner`->target/corner, `SetPiece`->target/setpiece. Open play = `RegularPlay`, `FastBreak`, `IndividualPlay`, and `ThrowInSetPiece` (Understat has no throw-in class and counts them as open play). Own goals dropped; blocked shots kept; unknown situations dropped.

**Open items**
- Left/right: FotMob `y` is assumed to match Understat's `Y` orientation (y/68 into the same thirds). Unverified; if flanks look mirrored in the UI, flip `y` in `fotmob/parse.go`.
- No retry/backoff on FotMob timeouts (3/100 timed out on first run; the next sync picked them up).
- Team names are mapped by FotMob full name in `fotmob/namematch.go`; a newly promoted team passes through unmapped until added.
- v1 tables/routes/config can be removed after cutover (new migration, never edit 008-013).
