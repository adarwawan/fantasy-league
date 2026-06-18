package sync

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"

	"fantasy-league/internal/fantasy"
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
	UpsertManagers(ctx context.Context, managers []fantasy.Manager) error
	UpsertPicks(ctx context.Context, picks []fantasy.ManagerPick) error
	RecomputeTopNOwnership(ctx context.Context, gameID string, topN int, gw int) error
}

// Cache is the subset of store.Cache used by the syncer.
type Cache interface {
	InvalidateGame(ctx context.Context, gameID string) error
}

type Syncer struct {
	sources []fantasy.Source
	store   Store
	cache   Cache
	topN    int
}

func New(sources []fantasy.Source, store Store, cache Cache, topN int) *Syncer {
	return &Syncer{sources: sources, store: store, cache: cache, topN: topN}
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

	// Players
	players, err := src.FetchPlayers(ctx)
	if err != nil {
		return fmt.Errorf("FetchPlayers: %w", err)
	}
	if err := s.store.UpsertPlayers(ctx, players); err != nil {
		return fmt.Errorf("UpsertPlayers: %w", err)
	}
	slog.Info("players synced", "game", gameID, "count", len(players))

	// Managers
	managers, err := src.FetchManagers(ctx, s.topN)
	if err != nil {
		return fmt.Errorf("FetchManagers: %w", err)
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

	// Top-N ownership recompute
	if err := s.store.RecomputeTopNOwnership(ctx, gameID, s.topN, gw); err != nil {
		return fmt.Errorf("RecomputeTopNOwnership: %w", err)
	}

	// Invalidate cache
	if err := s.cache.InvalidateGame(ctx, gameID); err != nil {
		slog.Warn("cache invalidation failed", "game", gameID, "err", err)
	}

	slog.Info("sync complete", "game", gameID)
	return nil
}
