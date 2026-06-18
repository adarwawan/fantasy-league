package fpl

import (
	"context"
	"fmt"
	"strconv"

	"fantasy-league/internal/fantasy"
)

// Source implements fantasy.Source for the FPL API.
type Source struct {
	client   *client
	leagueID int
	topN     int
}

func NewSource(leagueID, topN int) *Source {
	return &Source{
		client:   newClient(),
		leagueID: leagueID,
		topN:     topN,
	}
}

func (s *Source) GameID() string { return gameID }

func (s *Source) FetchTeams(ctx context.Context) ([]fantasy.Team, error) {
	boot, err := s.client.fetchBootstrap(ctx)
	if err != nil {
		return nil, fmt.Errorf("fpl FetchTeams: %w", err)
	}
	return mapTeams(boot.Teams), nil
}

func (s *Source) FetchFixtures(ctx context.Context) ([]fantasy.Fixture, error) {
	raw, err := s.client.fetchFixtures(ctx)
	if err != nil {
		return nil, fmt.Errorf("fpl FetchFixtures: %w", err)
	}
	return mapFixtures(raw), nil
}

func (s *Source) FetchPlayers(ctx context.Context) ([]fantasy.Player, error) {
	boot, err := s.client.fetchBootstrap(ctx)
	if err != nil {
		return nil, fmt.Errorf("fpl FetchPlayers: %w", err)
	}
	return mapPlayers(boot.Elements), nil
}

func (s *Source) FetchManagers(ctx context.Context, topN int) ([]fantasy.Manager, error) {
	entries, err := s.client.fetchLeagueStandings(ctx, s.leagueID, topN)
	if err != nil {
		return nil, fmt.Errorf("fpl FetchManagers: %w", err)
	}
	return mapManagers(entries), nil
}

// FetchPicks fetches picks for a manager. managerID is the external FPL manager ID as a string.
func (s *Source) FetchPicks(ctx context.Context, managerID string, gw int) ([]fantasy.ManagerPick, error) {
	extID, err := strconv.Atoi(managerID)
	if err != nil {
		return nil, fmt.Errorf("fpl FetchPicks: invalid managerID %q: %w", managerID, err)
	}
	resp, err := s.client.fetchPicks(ctx, extID, gw)
	if err != nil {
		return nil, fmt.Errorf("fpl FetchPicks manager=%d gw=%d: %w", extID, gw, err)
	}
	return mapPicks(extID, gw, resp.Picks), nil
}

// CurrentGW fetches the bootstrap to determine the active gameweek.
func (s *Source) CurrentGW(ctx context.Context) (int, error) {
	boot, err := s.client.fetchBootstrap(ctx)
	if err != nil {
		return 0, err
	}
	return currentGW(boot.Events), nil
}
