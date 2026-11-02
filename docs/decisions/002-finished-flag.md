# 002 — Use FPL `finished`, not `finished_provisional`

**Decision:** fixture completion uses `finished` (`backend/internal/sources/fpl/mapper.go`, `mapFixtures`). `store.CurrentGW` and the player upcoming-fixtures window derive the current GW from `MIN(gw) WHERE NOT finished`.

**Consequence:** right after a match, `finished_provisional` is true but `finished` stays false until FPL confirms bonus points (hours; overnight for late kickoffs). During that window a finished match still shows as current/upcoming.

**Why:** the app is for pre-deadline research, so settled, trustworthy data beats provisional data. The lag is accepted, not a bug.

**How to apply:** do not "fix" this by switching flags.
