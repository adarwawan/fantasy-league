package main

import (
	"context"
	"log"

	"fantasy-league/internal/config"
	"fantasy-league/internal/fantasy"
	"fantasy-league/internal/musthave"
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
		sources = append(sources, fplsrc.NewSource(cfg.FPLLeagueID))
	}

	windows := make(map[string]int, len(cfg.MustHave))
	mustHave := make(map[string]musthave.Config, len(cfg.MustHave))
	for game, mh := range cfg.MustHave {
		windows[game] = mh.FormWindow
		mustHave[game] = musthave.Config{
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
	syncer := syncsvc.New(sources, pg, cache, cfg.FormGWWindow).WithGWStatsWindows(windows).WithMustHave(mustHave).WithPicksWorkers(cfg.PicksWorkers)
	syncer.RunAll(ctx)
}
