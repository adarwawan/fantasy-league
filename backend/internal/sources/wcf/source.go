package wcf

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"fantasy-league/internal/fantasy"
)

// TopNOptions are the supported Top-N tiers for WCF. The API caps at 2 000
// entries, so the maximum tier is 1 000.
var TopNOptions = []int{100, 1000}

// Source implements fantasy.Source for the FIFA World Cup Fantasy API.
// Set WCFSyncEnabled=true only during active World Cup tournaments.
type Source struct {
	client   *client
	leagueID int // unused for WCF (uses global rankings), kept for interface parity
}

// NewSource creates a WCF source. cookie is the value of WC_COOKIE from env (WCF_AUTH_TOKEN).
func NewSource(cookie string) *Source {
	return &Source{
		client: newClient(cookie),
	}
}

func (s *Source) TopNOptions() []int { return TopNOptions }

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
	rounds, err := s.client.fetchRounds(context.Background())
	if err != nil {
		return 0, fmt.Errorf("wcf CurrentGW: %w", err)
	}
	for _, r := range rounds {
		if r.Status == "playing" {
			return r.ID, nil
		}
	}
	return 0, fmt.Errorf("wcf CurrentGW: no round with status playing")
}

// FetchDeadline returns the current GW and the start date of the next scheduled round.
func (s *Source) FetchDeadline(ctx context.Context) (int, time.Time, error) {
	rounds, err := s.client.fetchRounds(ctx)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("wcf FetchDeadline: %w", err)
	}
	currentGW := 0
	for _, r := range rounds {
		if r.Status == "playing" {
			currentGW = r.ID
			break
		}
	}
	for _, r := range rounds {
		if r.Status == "scheduled" && r.StartDate != "" {
			t, err := time.Parse(time.RFC3339, r.StartDate)
			if err != nil {
				return currentGW, time.Time{}, fmt.Errorf("wcf FetchDeadline parse: %w", err)
			}
			return currentGW, t, nil
		}
	}
	return currentGW, time.Time{}, nil
}
