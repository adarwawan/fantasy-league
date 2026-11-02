# 005 — Shot zone rankings (open-play, Understat)

Shipped in #67, #70, #72. Open-play companion to the set-piece board.

**Spec**
- **Zones (box-first, 4 exclusive buckets):** outside box = `x < 0.83`; inside the box (`x >= 0.83`) split `y` into thirds: left flank / center / right flank.
- **Filter:** keep `situation == "OpenPlay"` only. Exclude Penalty, SetPiece, DirectFreekick, FromCorner, and `result == "OwnGoal"`.
- **Window:** rolling last 6 matches, no season toggle (see 001).
- **UI:** single-team focused view. A team dropdown selects one team (~20 teams; never show all at once). The pitch diagram *is* the zone filter: it shows the selected team's shots/goals/xG per zone and clicking a zone filters the panel; an "Overall" chip resets. Do not add a separate tab row or revert to an all-teams grid. Default sort Shots, top 5 players.
- **Taken vs conceded:** top-level toggle. Taken = per-player leaderboard. Conceded = team aggregate only (Shots/Goals/xG/xG-per-shot/G−xG conceded) per zone, worst defense first, to spot a weak team to target. No fixture-list tie-in yet (deferred).

**Implementation:** `backend/internal/shotzone/` with its own small client that shares the Redis key `setpiece:match:{id}`; migration `013_shot_zone` (`012` skipped); `SZ_*` config; routes `/api/shot-zones`; page `/:game/shot-zones`. Verified against live data: 50 matches, 1047 open-play events, 100 conceded rows (20 teams × 5 zones incl. overall).

**Bugs found (still relevant)**
- chi `URLParam` returns the percent-encoded segment when a path mixes escaped/unescaped characters (`Nott'm Forest`). Shotzone decodes with `url.PathUnescape`. **Unfixed in `/api/setpiece/teams/{understat_team}`** — open follow-up.
- Local Postgres can lose its host port if another container grabs 5432; use 5433 (see root `CLAUDE.md`).

**Data reality, not a bug:** the Understat 2026 team set includes a few non-PL teams (Coventry, Hull, Ipswich), shared with `sp_events` (23 distinct teams, not 20).
