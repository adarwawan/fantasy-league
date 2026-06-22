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
