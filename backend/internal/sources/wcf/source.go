package wcf

import (
	"context"
	"fmt"
	"strconv"

	"fantasy-league/internal/fantasy"
)

// Source implements fantasy.Source for the FIFA World Cup Fantasy API.
// Set WCFSyncEnabled=true only during active World Cup tournaments.
type Source struct {
	client   *client
	leagueID int // unused for WCF (uses global rankings), kept for interface parity
	topN     int
}

// NewSource creates a WCF source. cookie is the value of WC_COOKIE from env (WCF_AUTH_TOKEN).
func NewSource(topN int, cookie string) *Source {
	return &Source{
		client: newClient(cookie),
		topN:   topN,
	}
}

func (s *Source) GameID() string { return gameID }

func (s *Source) FetchTeams(ctx context.Context) ([]fantasy.Team, error) {
	squads, err := s.client.fetchSquads(ctx)
	if err != nil {
		return nil, fmt.Errorf("wcf FetchTeams: %w", err)
	}
	return mapTeams(squads), nil
}

func (s *Source) FetchFixtures(ctx context.Context) ([]fantasy.Fixture, error) {
	rounds, err := s.client.fetchRounds(ctx)
	if err != nil {
		return nil, fmt.Errorf("wcf FetchFixtures: %w", err)
	}
	return mapFixtures(rounds), nil
}

func (s *Source) FetchPlayers(ctx context.Context) ([]fantasy.Player, error) {
	raw, err := s.client.fetchPlayers(ctx)
	if err != nil {
		return nil, fmt.Errorf("wcf FetchPlayers: %w", err)
	}
	return mapPlayers(raw), nil
}

func (s *Source) FetchManagers(ctx context.Context, topN int) ([]fantasy.Manager, error) {
	entries, err := s.client.fetchRankings(ctx, topN)
	if err != nil {
		return nil, fmt.Errorf("wcf FetchManagers: %w", err)
	}
	return mapManagers(extractRanks(entries)), nil
}

// FetchPicks fetches picks for a manager via the history endpoint.
// managerID is the external WCF user ID as a string.
func (s *Source) FetchPicks(ctx context.Context, managerID string, gw int) ([]fantasy.ManagerPick, error) {
	extID, err := strconv.Atoi(managerID)
	if err != nil {
		return nil, fmt.Errorf("wcf FetchPicks: invalid managerID %q: %w", managerID, err)
	}
	resp, err := s.client.fetchHistory(ctx, gw, extID)
	if err != nil {
		return nil, fmt.Errorf("wcf FetchPicks manager=%d gw=%d: %w", extID, gw, err)
	}
	return mapPicks(extID, gw, &resp.Success), nil
}

// CurrentGW is not exposed by the WCF public API; the syncer must pass the
// current round number explicitly via the sync orchestration layer.
func (s *Source) CurrentGW(_ context.Context) (int, error) {
	return 1, nil
}
