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
	fplsrc "fantasy-league/internal/sources/fpl"
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
	var sources []fantasy.Source
	if cfg.FPLSyncEnabled {
		sources = append(sources, fplsrc.NewSource(cfg.FPLLeagueID, cfg.FPLTopNDefault))
	}
	if cfg.WCFSyncEnabled {
		sources = append(sources, wcfsrc.NewSource(cfg.FPLTopNDefault, cfg.WCFAuthToken))
	}

	syncer := syncsvc.New(sources, pg, cache, cfg.FPLTopNDefault, cfg.FormGWWindow)

	// Scheduler — daily at 08:00 UTC
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Fatalf("scheduler: %v", err)
	}
	_, err = scheduler.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(8, 0, 0))),
		gocron.NewTask(func() { syncer.RunAll(context.Background()) }),
	)
	if err != nil {
		log.Fatalf("schedule job: %v", err)
	}
	scheduler.Start()

	// Sync on startup
	go syncer.RunAll(ctx)

	// Handlers
	playersH := handler.NewPlayersHandler(pg, cache)
	teamsH := handler.NewTeamsHandler(pg, cache)

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

	r.Route("/api/{game}", func(r chi.Router) {
		r.Use(validateGame)

		r.Get("/sync", func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Sync-Secret") != cfg.SyncEndpointSecret {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			go syncer.RunAll(context.Background())
			slog.Info("manual sync triggered", "game", chi.URLParam(r, "game"))
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("sync triggered"))
		})

		r.Get("/players/scatter", playersH.Scatter)
		r.Get("/players", playersH.List)
		r.Get("/teams", teamsH.List)
		r.Get("/fixtures", teamsH.Fixtures)
	})

	log.Printf("server listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
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
