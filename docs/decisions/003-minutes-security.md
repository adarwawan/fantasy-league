# 003 — Minutes security (rotation risk)

Shipped: FPL `starts` captured per GW (migration 011, `player_gw_stats.starts`); `recent_minutes`, `start_rate`, `avg_minutes` on the players payload; a chip in `PlayerDrawer` and in `TeamFormTable` team-expand rows. The flat players list was deliberately left unchanged (scan vs decide).

**Decisions**
- Window: last 5 finished GWs (matches `recentPointsWindow`).
- Chip: `status` override first (injured → Out, doubt → Doubt), else `start_rate` ≥ 80% Nailed, ≥ 50% Rotation, else Bench risk.
- The chip is derived on the **frontend** (`utils/minutes.ts`); the backend stays raw.
- **Everything is fixture-level.** `start_rate` and `avg_minutes` divide by fixtures the club played in the window: a blank GW drops out of the denominator, a double counts twice.

**Gotcha:** average minutes must be per *fixture*, not per GW, or a double GW (178' in one GW) reads as a 112' average.

**Known gap (deferred, low priority):** FPL has current `status`/`news` but no per-GW historical availability, so past suspensions/injuries still dilute the rate. Closing it needs a `player_status_history(player_id, gw, status)` table snapshotted each sync.

**Verification lesson:** seeded-data runs caught an ambiguous `gw` in the `club_fixtures` CTE (SQLSTATE 42702) that built and passed unit tests but returned empty minutes. Seed by marking GWs finished in `fixtures`, inserting `player_gw_stats`, and flushing Redis.
