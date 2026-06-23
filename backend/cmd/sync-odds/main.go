// sync-odds is a one-shot CLI for manually triggering an odds sync for a
// single game. Usage: go run ./cmd/sync-odds -game wcf
package main

import (
	"context"
	"flag"
	"log"
	"log/slog"

	"fantasy-league/internal/config"
	"fantasy-league/internal/sources/odds"
	"fantasy-league/internal/store"
	syncsvc "fantasy-league/internal/sync"
)

func main() {
	game := flag.String("game", "", "game ID to sync odds for (wcf|fpl)")
	flag.Parse()

	if *game != "wcf" && *game != "fpl" {
		log.Fatalf("usage: sync-odds -game <wcf|fpl>")
	}

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

	oddsClient := odds.NewClient(cfg.OddsAPIKey, cfg.OddsCacheTTL, cache)
	syncer := syncsvc.New(nil, pg, cache, cfg.FormGWWindow).WithOdds(&syncsvc.OddsDeps{
		Client: oddsClient,
		Configs: map[string]odds.GameOddsConfig{
			"wcf": odds.WCFOddsConfig,
			"fpl": odds.FPLOddsConfig,
		},
		Enabled:  map[string]bool{*game: true},
		Cache:    cache,
		CacheTTL: cfg.OddsCacheTTL,
	})

	// Wrap in a minimal source so RunAll drives the odds-only path.
	if err := syncer.RunOdds(ctx, *game); err != nil {
		log.Fatalf("RunOdds: %v", err)
	}
	slog.Info("done", "game", *game)
}
