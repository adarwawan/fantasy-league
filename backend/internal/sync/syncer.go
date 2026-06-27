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
	"fantasy-league/internal/sources/odds"
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
	QueryTeams(ctx context.Context, gameID string, window int, sort string) ([]store.TeamRow, error)
	QueryFixtures(ctx context.Context, gameID string, fromGW, toGW int) ([]store.FixtureRow, error)
	DeleteMatchOdds(ctx context.Context, gameID string) error
	UpsertMatchOdds(ctx context.Context, rows []store.MatchOddsRow, cache *store.Cache, ttl time.Duration) error
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

// OddsDeps bundles the optional odds-sync dependencies. Pass nil to skip odds.
type OddsDeps struct {
	Client   *odds.Client
	Configs  map[string]odds.GameOddsConfig
	Enabled  map[string]bool
	Cache    *store.Cache
	CacheTTL time.Duration
}

type Syncer struct {
	sources      []fantasy.Source
	store        Store
	cache        Cache
	formGWWindow int
	odds         *OddsDeps
}

func New(sources []fantasy.Source, store Store, cache Cache, formGWWindow int) *Syncer {
	return &Syncer{sources: sources, store: store, cache: cache, formGWWindow: formGWWindow}
}

// WithOdds attaches odds-sync dependencies to the syncer.
func (s *Syncer) WithOdds(o *OddsDeps) *Syncer {
	s.odds = o
	return s
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

	// 1. Teams
	teams, err := src.FetchTeams(ctx)
	if err != nil {
		return fmt.Errorf("FetchTeams: %w", err)
	}
	if err := s.store.UpsertTeams(ctx, teams); err != nil {
		return fmt.Errorf("UpsertTeams: %w", err)
	}
	slog.Info("teams synced", "game", gameID, "count", len(teams))

	// 2. Fixtures
	fixtures, err := src.FetchFixtures(ctx)
	if err != nil {
		return fmt.Errorf("FetchFixtures: %w", err)
	}
	if err := s.store.UpsertFixtures(ctx, fixtures); err != nil {
		return fmt.Errorf("UpsertFixtures: %w", err)
	}
	slog.Info("fixtures synced", "game", gameID, "count", len(fixtures))

	// 3. Players
	players, err := src.FetchPlayers(ctx)
	if err != nil {
		return fmt.Errorf("FetchPlayers: %w", err)
	}
	if err := s.store.UpsertPlayers(ctx, players); err != nil {
		return fmt.Errorf("UpsertPlayers: %w", err)
	}
	slog.Info("players synced", "game", gameID, "count", len(players))

	// Determine current GW (needed by concurrent steps).
	gw := 1
	if gp, ok := src.(GWProvider); ok {
		if g, err := gp.CurrentGW(ctx); err == nil && g > 0 {
			gw = g
		}
	}

	// 4. Concurrent: recompute team form | managers+picks+ownership | odds
	var (
		wg      sync.WaitGroup
		concErr [3]error
	)

	// 4a. Recompute team form
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.store.RecomputeTeamForm(ctx, gameID, s.formGWWindow); err != nil {
			concErr[0] = fmt.Errorf("RecomputeTeamForm: %w", err)
		}
	}()

	// 4b. Fetch managers + picks + recompute ownership
	wg.Add(1)
	go func() {
		defer wg.Done()
		topNOptions := src.TopNOptions()
		topNMax := topNOptions[len(topNOptions)-1]

		managers, err := src.FetchManagers(ctx, topNMax)
		if err != nil {
			concErr[1] = fmt.Errorf("FetchManagers: %w", err)
			return
		}
		if err := s.store.ResetManagerRanks(ctx, gameID); err != nil {
			concErr[1] = fmt.Errorf("ResetManagerRanks: %w", err)
			return
		}
		if err := s.store.UpsertManagers(ctx, managers); err != nil {
			concErr[1] = fmt.Errorf("UpsertManagers: %w", err)
			return
		}
		slog.Info("managers synced", "game", gameID, "count", len(managers))

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

		if err := s.store.RecomputeTopNOwnerships(ctx, gameID, topNOptions, gw); err != nil {
			concErr[1] = fmt.Errorf("RecomputeTopNOwnerships: %w", err)
		}
	}()

	// 4c. Odds
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.syncOdds(ctx, gameID); err != nil {
			concErr[2] = err
		}
	}()

	wg.Wait()

	for _, e := range concErr {
		if e != nil {
			return e
		}
	}

	// 5. Invalidate cache
	if err := s.cache.InvalidateGame(ctx, gameID); err != nil {
		slog.Warn("cache invalidation failed", "game", gameID, "err", err)
	}

	// 6. Deadline — set after invalidation so it survives the wipe.
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

// RunOdds runs the odds sync for a single game in isolation (used by the sync-odds CLI).
func (s *Syncer) RunOdds(ctx context.Context, gameID string) error {
	return s.syncOdds(ctx, gameID)
}

// syncOdds is the per-game odds step run inside the concurrent block of run.
// It is a no-op when odds deps are not configured or the game is disabled.
func (s *Syncer) syncOdds(ctx context.Context, gameID string) error {
	o := s.odds
	if o == nil {
		return nil
	}
	if !o.Enabled[gameID] {
		slog.Debug("odds sync disabled", "game", gameID)
		return nil
	}
	cfg, ok := o.Configs[gameID]
	if !ok {
		return nil
	}

	slog.Info("odds sync start", "game", gameID)

	rawMatches, err := o.Client.FetchOdds(ctx, cfg)
	if err != nil {
		return fmt.Errorf("FetchOdds (%s): %w", gameID, err)
	}

	computed := odds.AggregateBookmakers(rawMatches)

	teamRows, err := s.store.QueryTeams(ctx, gameID, 5, "ovr_form")
	if err != nil {
		return fmt.Errorf("QueryTeams (%s): %w", gameID, err)
	}
	fantasyTeams := make([]fantasy.Team, len(teamRows))
	for i, t := range teamRows {
		fantasyTeams[i] = fantasy.Team{
			ID:        t.ID,
			GameID:    gameID,
			Name:      t.Name,
			ShortName: t.ShortName,
		}
	}

	computed = odds.MapTeams(computed, fantasyTeams, gameID)
	if len(computed) == 0 {
		slog.Warn("odds sync: no matches resolved after team mapping", "game", gameID)
		return nil
	}

	fixtureRows, err := s.store.QueryFixtures(ctx, gameID, 1, 999)
	if err != nil {
		return fmt.Errorf("QueryFixtures (%s): %w", gameID, err)
	}
	fantasyFixtures := make([]fantasy.Fixture, len(fixtureRows))
	for i, f := range fixtureRows {
		fantasyFixtures[i] = fantasy.Fixture{
			ID:         f.ID,
			GameID:     gameID,
			GW:         f.GW,
			HomeTeamID: f.HomeTeamID,
			AwayTeamID: f.AwayTeamID,
		}
	}

	computed = odds.LinkFixtures(computed, fantasyFixtures, cfg)

	storeRows := make([]store.MatchOddsRow, len(computed))
	for i, m := range computed {
		storeRows[i] = store.MatchOddsRow{
			OddsMatchID: m.OddsMatchID,
			GameID:      gameID,
			FixtureID:   m.FixtureID,
			HomeTeam:    m.HomeTeam,
			AwayTeam:    m.AwayTeam,
			LambdaHome:  m.LambdaHome,
			LambdaAway:  m.LambdaAway,
			HomeCSPct:   m.HomeCSPct,
			AwayCSPct:   m.AwayCSPct,
			KickoffTime: m.KickoffTime,
			FetchedAt:   m.FetchedAt,
		}
	}

	if err := s.store.DeleteMatchOdds(ctx, gameID); err != nil {
		return fmt.Errorf("DeleteMatchOdds (%s): %w", gameID, err)
	}
	if err := s.store.UpsertMatchOdds(ctx, storeRows, o.Cache, o.CacheTTL); err != nil {
		return fmt.Errorf("UpsertMatchOdds (%s): %w", gameID, err)
	}

	slog.Info("odds sync complete", "game", gameID, "matches", len(storeRows))
	return nil
}
