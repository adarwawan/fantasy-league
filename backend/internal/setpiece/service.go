package setpiece

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Config holds the tunables for the set-piece module, sourced from the app
// config (SP_* env / yaml).
type Config struct {
	Enabled         bool
	Season          string // Understat season, starting year e.g. "2025"
	WindowMatches   int
	RecencyHalfLife time.Duration
}

// Service orchestrates the sync: discover finished matches → fetch new ones →
// parse both roles → aggregate rolling window → replace board. It is the only
// entry point the scheduler and manual trigger call.
type Service struct {
	cfg    Config
	client *Client
	store  *Store
}

func NewService(cfg Config, client *Client, store *Store) *Service {
	return &Service{cfg: cfg, client: client, store: store}
}

// Sync runs one full detection pass. A scrape failure for a single match is
// logged and skipped so one bad match doesn't abort the run; the last-good
// board is left intact if nothing new parses.
func (s *Service) Sync(ctx context.Context) error {
	if !s.cfg.Enabled {
		slog.Debug("setpiece sync disabled")
		return nil
	}
	season := s.cfg.Season
	slog.Info("setpiece sync start", "season", season)

	ids, err := s.client.FinishedMatches(ctx, season)
	if err != nil {
		return fmt.Errorf("FinishedMatches: %w", err)
	}

	existing, err := s.store.ExistingMatchIDs(ctx, season)
	if err != nil {
		return fmt.Errorf("ExistingMatchIDs: %w", err)
	}

	var fetched, skipped int
	for _, id := range ids {
		if existing[id] {
			continue
		}
		shots, err := s.client.MatchShots(ctx, id)
		if err != nil {
			slog.Warn("setpiece: fetch match failed", "match", id, "err", err)
			skipped++
			continue
		}
		events := ParseShots(id, shots)
		if err := s.store.UpsertEvents(ctx, events); err != nil {
			slog.Warn("setpiece: upsert events failed", "match", id, "err", err)
			skipped++
			continue
		}
		fetched++
	}
	slog.Info("setpiece matches processed", "fetched", fetched, "skipped", skipped, "known", len(existing))

	// Recompute the board over all events for the season.
	allEvents, err := s.store.ReadEvents(ctx, season)
	if err != nil {
		return fmt.Errorf("ReadEvents: %w", err)
	}
	if len(allEvents) == 0 {
		slog.Warn("setpiece: no events, leaving board intact", "season", season)
		return nil
	}

	board := Aggregate(allEvents, AggregateConfig{
		WindowMatches:   s.cfg.WindowMatches,
		RecencyHalfLife: s.cfg.RecencyHalfLife,
	})
	if err := s.store.ReplaceBoard(ctx, board); err != nil {
		return fmt.Errorf("ReplaceBoard: %w", err)
	}

	slog.Info("setpiece sync complete", "events", len(allEvents), "board_rows", len(board))
	return nil
}
