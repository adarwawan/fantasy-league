package fantasy

import "context"

type Source interface {
	FetchPlayers(ctx context.Context) ([]Player, error)
	FetchTeams(ctx context.Context) ([]Team, error)
	FetchFixtures(ctx context.Context) ([]Fixture, error)
	FetchManagers(ctx context.Context, topN int) ([]Manager, error)
	FetchPicks(ctx context.Context, managerID string, gw int) ([]ManagerPick, error)
	GameID() string
	TopNOptions() []int
}

// EntrySummary is a manager's budget snapshot: team value (squad selling value
// plus bank) and bank, both in the game's currency units (£m for FPL).
type EntrySummary struct {
	TeamValue float64
	Bank      float64
}

// EntryLoader is an optional source capability: games whose public API can
// resolve a manager's current squad and budget from their entry ID implement
// it. The entry handler enables the load-team route only for sources that
// satisfy this interface, so adding a new game is purely a matter of
// implementing these methods — no handler or routing changes required.
type EntryLoader interface {
	Source
	CurrentGW(ctx context.Context) (int, error)
	FetchEntrySummary(ctx context.Context, entryID string) (EntrySummary, error)
}
