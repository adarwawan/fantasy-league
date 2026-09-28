// sync-shotzone is a one-shot CLI for manually triggering a shot-zone
// detector sync (Understat). Usage: go run ./cmd/sync-shotzone
package main

import (
	"context"
	"log"
	"log/slog"

	"fantasy-league/internal/config"
	"fantasy-league/internal/shotzone"
	"fantasy-league/internal/store"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pg, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pg.Close()

	cache, err := store.NewCache(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}

	client := shotzone.NewClient(cfg.OddsCacheTTL, cache)
	svc := shotzone.NewService(shotzone.Config{
		Enabled:       true, // force-run regardless of SZ_ENABLED
		Season:        cfg.SZSeason,
		WindowMatches: cfg.SZWindowMatches,
	}, client, shotzone.NewStore(pg.Pool()))

	if err := svc.Sync(ctx); err != nil {
		log.Fatalf("shotzone sync: %v", err)
	}
	slog.Info("shotzone sync done", "season", cfg.SZSeason)
}
