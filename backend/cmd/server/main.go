package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-co-op/gocron/v2"

	"fantasy-league/internal/config"
	"fantasy-league/internal/fantasy"
	fplsrc "fantasy-league/internal/sources/fpl"
	"fantasy-league/internal/store"
	syncsvc "fantasy-league/internal/sync"
)

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

	syncer := syncsvc.New(sources, pg, cache, cfg.FPLTopNDefault)

	// Scheduler
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Fatalf("scheduler: %v", err)
	}
	interval := time.Duration(cfg.FPLSyncIntervalMin) * time.Minute
	_, err = scheduler.NewJob(
		gocron.DurationJob(interval),
		gocron.NewTask(func() { syncer.RunAll(context.Background()) }),
	)
	if err != nil {
		log.Fatalf("schedule job: %v", err)
	}
	scheduler.Start()

	// Sync on startup
	go syncer.RunAll(ctx)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Get("/api/{game}/sync", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Sync-Secret") != cfg.SyncEndpointSecret {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		go syncer.RunAll(context.Background())
		slog.Info("manual sync triggered", "game", chi.URLParam(r, "game"))
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("sync triggered"))
	})

	log.Printf("server listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}
