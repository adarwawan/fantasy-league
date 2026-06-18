package main

import (
	"context"
	"log"

	"fantasy-league/internal/config"
	"fantasy-league/internal/fantasy"
	fplsrc "fantasy-league/internal/sources/fpl"
	"fantasy-league/internal/store"
	syncsvc "fantasy-league/internal/sync"
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

	var sources []fantasy.Source
	if cfg.FPLSyncEnabled {
		sources = append(sources, fplsrc.NewSource(cfg.FPLLeagueID, cfg.FPLTopNDefault))
	}

	syncer := syncsvc.New(sources, pg, cache, cfg.FPLTopNDefault)
	syncer.RunAll(ctx)
}
