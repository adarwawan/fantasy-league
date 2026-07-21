package fpl

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"fantasy-league/internal/fantasy"
)

// TopNOptions are the supported Top-N tiers for FPL.
var TopNOptions = []int{1000, 10000, 100000}

// Source implements fantasy.Source for the FPL API.
type Source struct {
	client   *client
	leagueID int
}

func NewSource(leagueID int) *Source {
	return &Source{
		client:   newClient(),
		leagueID: leagueID,
	}
}

func (s *Source) TopNOptions() []int { return TopNOptions }

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

// FetchEntrySummary fetches a manager's team value and bank from the FPL entry
// endpoint. entryID is the external FPL manager ID as a string. Implementing
// this (plus CurrentGW, already present) makes FPL satisfy fantasy.EntryLoader.
func (s *Source) FetchEntrySummary(ctx context.Context, entryID string) (fantasy.EntrySummary, error) {
	extID, err := strconv.Atoi(entryID)
	if err != nil {
		return fantasy.EntrySummary{}, fmt.Errorf("fpl FetchEntrySummary: invalid entryID %q: %w", entryID, err)
	}
	resp, err := s.client.fetchEntry(ctx, extID)
	if err != nil {
		return fantasy.EntrySummary{}, fmt.Errorf("fpl FetchEntrySummary entry=%d: %w", extID, err)
	}
	return fantasy.EntrySummary{
		TeamValue: float64(resp.LastDeadlineValue) / 10,
		Bank:      float64(resp.LastDeadlineBank) / 10,
	}, nil
}

// FetchGWStats returns every player's stat line for a single gameweek.
func (s *Source) FetchGWStats(ctx context.Context, gw int) ([]fantasy.PlayerGWStat, error) {
	resp, err := s.client.fetchEventLive(ctx, gw)
	if err != nil {
		return nil, fmt.Errorf("fpl FetchGWStats gw=%d: %w", gw, err)
	}
	return mapGWStats(gw, resp.Elements), nil
}

// CurrentGW fetches the bootstrap to determine the active gameweek.
func (s *Source) CurrentGW(ctx context.Context) (int, error) {
	boot, err := s.client.fetchBootstrap(ctx)
	if err != nil {
		return 0, err
	}
	return currentGW(boot.Events), nil
}

// FetchDeadline returns the current GW and the next GW's deadline time.
func (s *Source) FetchDeadline(ctx context.Context) (int, time.Time, error) {
	boot, err := s.client.fetchBootstrap(ctx)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("fpl FetchDeadline: %w", err)
	}
	gw := currentGW(boot.Events)
	for _, e := range boot.Events {
		if e.IsNext && e.DeadlineTime != "" {
			t, err := time.Parse(time.RFC3339, e.DeadlineTime)
			if err != nil {
				return gw, time.Time{}, fmt.Errorf("fpl FetchDeadline parse: %w", err)
			}
			return gw, t, nil
		}
	}
	return gw, time.Time{}, nil
}
