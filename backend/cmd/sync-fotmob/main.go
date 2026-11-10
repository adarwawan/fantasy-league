// sync-fotmob is a one-shot CLI for manually triggering the FotMob (v2)
// set-piece + shot-zone sync. Usage: go run ./cmd/sync-fotmob
package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"fantasy-league/internal/config"
	"fantasy-league/internal/fotmob"
	"fantasy-league/internal/setpiece"
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

	svc := fotmob.NewService(fotmob.Config{
		Enabled:         true, // force-run regardless of FM_ENABLED
		Season:          cfg.FMSeason,
		SPWindowMatches: cfg.FMSPWindowMatches,
		SZWindowMatches: cfg.FMSZWindowMatches,
		RecencyHalfLife: cfg.FMRecencyHalfLife,
		RequestDelay:    500 * time.Millisecond,
	}, fotmob.NewClient(cfg.OddsCacheTTL, cache), setpiece.NewStoreV2(pg.Pool()), shotzone.NewStoreV2(pg.Pool()))

	if err := svc.Sync(ctx); err != nil {
		log.Fatalf("fotmob sync: %v", err)
	}
	slog.Info("fotmob sync done", "season", cfg.FMSeason)
}
