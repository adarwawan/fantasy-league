package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"fantasy-league/internal/fantasy"
	"fantasy-league/internal/store"
)

// GWProvider is implemented by sources that can report the current gameweek.
type GWProvider interface {
	CurrentGW(ctx context.Context) (int, error)
}

// Store is the subset of store.Store used by the syncer.
type Store interface {
	UpsertTeams(ctx context.Context, teams []fantasy.Team) error
	UpsertFixtures(ctx context.Context, fixtures []fantasy.Fixture) error
	UpsertPlayers(ctx context.Context, players []fantasy.Player) error
	ResetManagerRanks(ctx context.Context, gameID string) error
	UpsertManagers(ctx context.Context, managers []fantasy.Manager) error
	UpsertPicks(ctx context.Context, picks []fantasy.ManagerPick) error
	RecomputeTopNOwnerships(ctx context.Context, gameID string, topNOptions []int, gw int) error
	RecomputeTeamForm(ctx context.Context, gameID string, gwWindow int) error
}

// DeadlineProvider is implemented by sources that can report the next deadline.
type DeadlineProvider interface {
	FetchDeadline(ctx context.Context) (currentGW int, nextDeadline time.Time, err error)
}

// Cache is the subset of store.Cache used by the syncer.
type Cache interface {
	InvalidateGame(ctx context.Context, gameID string) error
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
}

type Syncer struct {
	sources      []fantasy.Source
	store        Store
	cache        Cache
	formGWWindow int
}

func New(sources []fantasy.Source, store Store, cache Cache, formGWWindow int) *Syncer {
	return &Syncer{sources: sources, store: store, cache: cache, formGWWindow: formGWWindow}
}

// RunAll syncs all registered sources in parallel.
func (s *Syncer) RunAll(ctx context.Context) {
	var wg sync.WaitGroup
	for _, src := range s.sources {
		wg.Add(1)
		go func(src fantasy.Source) {
			defer wg.Done()
			if err := s.run(ctx, src); err != nil {
				slog.Error("sync failed", "game", src.GameID(), "err", err)
			}
		}(src)
	}
	wg.Wait()
}

func (s *Syncer) run(ctx context.Context, src fantasy.Source) error {
	gameID := src.GameID()
	slog.Info("sync start", "game", gameID)

	// Teams
	teams, err := src.FetchTeams(ctx)
	if err != nil {
		return fmt.Errorf("FetchTeams: %w", err)
	}
	if err := s.store.UpsertTeams(ctx, teams); err != nil {
		return fmt.Errorf("UpsertTeams: %w", err)
	}
	slog.Info("teams synced", "game", gameID, "count", len(teams))

	// Fixtures
	fixtures, err := src.FetchFixtures(ctx)
	if err != nil {
		return fmt.Errorf("FetchFixtures: %w", err)
	}
	if err := s.store.UpsertFixtures(ctx, fixtures); err != nil {
		return fmt.Errorf("UpsertFixtures: %w", err)
	}
	slog.Info("fixtures synced", "game", gameID, "count", len(fixtures))

	if err := s.store.RecomputeTeamForm(ctx, gameID, s.formGWWindow); err != nil {
		return fmt.Errorf("RecomputeTeamForm: %w", err)
	}

	// Players
	players, err := src.FetchPlayers(ctx)
	if err != nil {
		return fmt.Errorf("FetchPlayers: %w", err)
	}
	if err := s.store.UpsertPlayers(ctx, players); err != nil {
		return fmt.Errorf("UpsertPlayers: %w", err)
	}
	slog.Info("players synced", "game", gameID, "count", len(players))

	// Managers — fetch up to the largest configured tier.
	topNOptions := src.TopNOptions()
	topNMax := topNOptions[len(topNOptions)-1]
	managers, err := src.FetchManagers(ctx, topNMax)
	if err != nil {
		return fmt.Errorf("FetchManagers: %w", err)
	}
	if err := s.store.ResetManagerRanks(ctx, gameID); err != nil {
		return fmt.Errorf("ResetManagerRanks: %w", err)
	}
	if err := s.store.UpsertManagers(ctx, managers); err != nil {
		return fmt.Errorf("UpsertManagers: %w", err)
	}
	slog.Info("managers synced", "game", gameID, "count", len(managers))

	// Determine current GW
	gw := 1
	if gp, ok := src.(GWProvider); ok {
		if g, err := gp.CurrentGW(ctx); err == nil && g > 0 {
			gw = g
		}
	}

	// Picks
	var pickErrs int
	for _, m := range managers {
		managerID := strconv.Itoa(m.ExternalID)
		picks, err := src.FetchPicks(ctx, managerID, gw)
		if err != nil {
			pickErrs++
			slog.Warn("FetchPicks failed", "game", gameID, "manager", m.ExternalID, "err", err)
			continue
		}
		if err := s.store.UpsertPicks(ctx, picks); err != nil {
			pickErrs++
			slog.Warn("UpsertPicks failed", "game", gameID, "manager", m.ExternalID, "err", err)
		}
	}
	slog.Info("picks synced", "game", gameID, "managers", len(managers), "errors", pickErrs)

	// Top-N ownership recompute for every configured tier.
	if err := s.store.RecomputeTopNOwnerships(ctx, gameID, topNOptions, gw); err != nil {
		return fmt.Errorf("RecomputeTopNOwnerships: %w", err)
	}

	// Invalidate cache
	if err := s.cache.InvalidateGame(ctx, gameID); err != nil {
		slog.Warn("cache invalidation failed", "game", gameID, "err", err)
	}

	// Deadline — set after invalidation so it survives the wipe.
	if dp, ok := src.(DeadlineProvider); ok {
		dgw, deadline, err := dp.FetchDeadline(ctx)
		if err != nil {
			slog.Warn("FetchDeadline failed", "game", gameID, "err", err)
		} else {
			type deadlinePayload struct {
				CurrentGW    int       `json:"current_gw"`
				NextDeadline time.Time `json:"next_deadline"`
				CachedAt     time.Time `json:"cached_at"`
			}
			b, _ := json.Marshal(deadlinePayload{CurrentGW: dgw, NextDeadline: deadline, CachedAt: time.Now().UTC()})
			if err := s.cache.Set(ctx, store.CacheKey(gameID, "deadline"), b, 0); err != nil {
				slog.Warn("cache set deadline failed", "game", gameID, "err", err)
			}
		}
	}

	slog.Info("sync complete", "game", gameID)
	return nil
}
