// sync-setpiece is a one-shot CLI for manually triggering a set-piece detector
// sync (Understat). Usage: go run ./cmd/sync-setpiece
package main

import (
	"context"
	"log"
	"log/slog"

	"fantasy-league/internal/config"
	"fantasy-league/internal/setpiece"
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

	client := setpiece.NewClient(cfg.OddsCacheTTL, cache)
	svc := setpiece.NewService(setpiece.Config{
		Enabled:         true, // force-run regardless of SP_ENABLED
		Season:          cfg.SPSeason,
		WindowMatches:   cfg.SPWindowMatches,
		RecencyHalfLife: cfg.SPRecencyHalfLife,
	}, client, setpiece.NewStore(pg.Pool()))

	if err := svc.Sync(ctx); err != nil {
		log.Fatalf("setpiece sync: %v", err)
	}
	slog.Info("setpiece sync done", "season", cfg.SPSeason)
}
