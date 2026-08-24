package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-co-op/gocron/v2"

	"fantasy-league/internal/config"
	"fantasy-league/internal/fantasy"
	"fantasy-league/internal/handler"
	"fantasy-league/internal/musthave"
	"fantasy-league/internal/setpiece"
	fplsrc "fantasy-league/internal/sources/fpl"
	"fantasy-league/internal/sources/odds"
	wcfsrc "fantasy-league/internal/sources/wcf"
	"fantasy-league/internal/store"
	syncsvc "fantasy-league/internal/sync"
)

var validGames = map[string]bool{"fpl": true, "wcf": true, "uclf": true}

func main() {
	cfg := config.Load()
	ctx := context.Background()

	// Store
	pg, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pg.Close()

	cache, err := store.NewCache(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}

	// Sources
	var startupSources, scheduledSources []fantasy.Source
	if cfg.FPLSyncEnabled {
		fpl := fplsrc.NewSource(cfg.FPLLeagueID)
		startupSources = append(startupSources, fpl)
		if !cfg.FPLSyncOnce {
			scheduledSources = append(scheduledSources, fpl)
		}
	}
	if cfg.WCFSyncEnabled {
		wcf := wcfsrc.NewSource(cfg.WCFAuthToken)
		startupSources = append(startupSources, wcf)
		scheduledSources = append(scheduledSources, wcf)
	}

	// Odds client (shared across games).
	oddsClient := odds.NewClient(cfg.OddsAPIKey, cfg.OddsCacheTTL, cache)
	oddsDeps := &syncsvc.OddsDeps{
		Client: oddsClient,
		Configs: map[string]odds.GameOddsConfig{
			"wcf": odds.WCFOddsConfig,
			"fpl": odds.FPLOddsConfig,
		},
		Enabled: map[string]bool{
			"wcf": cfg.WCFOddsEnabled,
			"fpl": cfg.FPLOddsEnabled,
		},
		Cache:    cache,
		CacheTTL: cfg.OddsCacheTTL,
	}

	startupSyncer := syncsvc.New(startupSources, pg, cache, cfg.FormGWWindow).
		WithOdds(oddsDeps).WithGWStatsWindows(gwStatsWindows(cfg)).WithMustHave(mustHaveConfigs(cfg)).WithPicksWorkers(cfg.PicksWorkers)
	scheduledSyncer := syncsvc.New(scheduledSources, pg, cache, cfg.FormGWWindow).
		WithOdds(oddsDeps).WithGWStatsWindows(gwStatsWindows(cfg)).WithMustHave(mustHaveConfigs(cfg)).WithPicksWorkers(cfg.PicksWorkers)

	// Scheduler — daily at 09:15 UTC (recurring sources only)
	// Aligned to sit after the GW lockdown at 09:00 UK time (scores final).
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Fatalf("scheduler: %v", err)
	}
	_, err = scheduler.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(9, 15, 0))),
		gocron.NewTask(func() {
			scheduledSyncer.RunAll(context.Background())
		}),
	)
	if err != nil {
		log.Fatalf("schedule job: %v", err)
	}

	// Set-piece detector — isolated Understat module with its own store, sync,
	// handler and schedule. Never wired through fantasy.Source.
	var spHandler *setpiece.Handler
	if cfg.SPEnabled {
		spClient := setpiece.NewClient(cfg.OddsCacheTTL, cache)
		spStore := setpiece.NewStore(pg.Pool())
		spService := setpiece.NewService(setpiece.Config{
			Enabled:         cfg.SPEnabled,
			Season:          cfg.SPSeason,
			WindowMatches:   cfg.SPWindowMatches,
			RecencyHalfLife: cfg.SPRecencyHalfLife,
		}, spClient, spStore)
		spHandler = setpiece.NewHandler(spStore, cache, cfg.SPWindowMatches)

		if _, err := scheduler.NewJob(
			gocron.CronJob(cfg.SPSyncCron, false),
			gocron.NewTask(func() {
				if err := spService.Sync(context.Background()); err != nil {
					slog.Error("setpiece sync failed", "err", err)
				}
			}),
		); err != nil {
			log.Fatalf("schedule setpiece job: %v", err)
		}

		go func() {
			if err := spService.Sync(context.Background()); err != nil {
				slog.Error("setpiece startup sync failed", "err", err)
			}
		}()
	}

	scheduler.Start()

	// Sync on startup (all sources, including once-only)
	go startupSyncer.RunAll(ctx)

	// Handlers
	playersH := handler.NewPlayersHandler(pg, cache)
	teamsH := handler.NewTeamsHandler(pg, cache)
	statsH := handler.NewStatsHandler(pg, cache)
	oddsH := handler.NewOddsHandler(pg, cache)

	deadlineH := handler.NewDeadlineHandler(cache)

	// Entry handler enables team-loading for any source implementing
	// fantasy.EntryLoader (FPL today; WCF/UCLF automatically once they add it).
	entryH := handler.NewEntryHandler(startupSources, pg)

	// Monitor freshness for every game that syncs at least once (startup set is
	// the superset of scheduled). Disabled games (e.g. sunset WCF) are excluded
	// so their stale status never trips the probe.
	monitoredGames := make([]string, 0, len(startupSources))
	for _, src := range startupSources {
		monitoredGames = append(monitoredGames, src.GameID())
	}
	syncHealthH := handler.NewSyncHealthHandler(cache, monitoredGames, cfg.SyncFreshnessMaxAge)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-Sync-Secret"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	r.Get("/health/sync", syncHealthH.Sync)

	// Set-piece detector routes — isolated namespace, not under /api/{game}.
	if spHandler != nil {
		r.Route("/api/setpiece", func(r chi.Router) {
			r.Get("/teams", spHandler.Teams)
			r.Get("/teams/{understat_team}", spHandler.Team)
		})
	}

	r.Route("/api/{game}", func(r chi.Router) {
		r.Use(validateGame)

		r.Get("/sync", func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Sync-Secret") != cfg.SyncEndpointSecret {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			go startupSyncer.RunAll(context.Background())
			slog.Info("manual sync triggered", "game", chi.URLParam(r, "game"))
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("sync triggered"))
		})

		r.Get("/players/scatter", playersH.Scatter)
		r.Get("/players", playersH.List)
		r.Get("/stats", statsH.List)
		r.Get("/stats/teams", statsH.Teams)
		r.Get("/teams", teamsH.List)
		r.Get("/fixtures", teamsH.Fixtures)
		r.Get("/fixtures/odds", oddsH.List)
		r.Get("/deadline", deadlineH.Deadline)
		r.Get("/entry/{id}", entryH.Load)
	})

	log.Printf("server listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}

// mustHaveConfigs converts per-game config thresholds into musthave configs.
func mustHaveConfigs(cfg config.Config) map[string]musthave.Config {
	out := make(map[string]musthave.Config, len(cfg.MustHave))
	for game, mh := range cfg.MustHave {
		out[game] = musthave.Config{
			FormWindow:    mh.FormWindow,
			FormPointsMin: mh.FormPointsMin,
			FormRatio:     mh.FormRatio,
			MaxNextFDR:    mh.MaxNextFDR,
			TopGK:         mh.TopGK,
			TopDEF:        mh.TopDEF,
			TopMID:        mh.TopMID,
			TopFWD:        mh.TopFWD,
		}
	}
	return out
}

// gwStatsWindows extracts each game's form window for the GW-stats sync step.
func gwStatsWindows(cfg config.Config) map[string]int {
	out := make(map[string]int, len(cfg.MustHave))
	for game, mh := range cfg.MustHave {
		out[game] = mh.FormWindow
	}
	return out
}

func validateGame(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		game := chi.URLParam(r, "game")
		if !validGames[game] {
			http.Error(w, `{"error":"invalid game"}`, http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
