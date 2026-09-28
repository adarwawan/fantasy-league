package shotzone

import (
	"context"
	"fmt"
	"log/slog"
)

// Config holds the tunables for the shot-zone module, sourced from the app
// config (SZ_* env / yaml).
type Config struct {
	Enabled       bool
	Season        string // Understat season, starting year e.g. "2026"
	WindowMatches int
}

// Service orchestrates the sync: discover finished matches → fetch new ones →
// parse open-play shots → aggregate both rolling-window boards → replace
// them. It is the only entry point the scheduler and manual trigger call.
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
// boards are left intact if nothing new parses.
func (s *Service) Sync(ctx context.Context) error {
	if !s.cfg.Enabled {
		slog.Debug("shotzone sync disabled")
		return nil
	}
	season := s.cfg.Season
	slog.Info("shotzone sync start", "season", season)

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
			slog.Warn("shotzone: fetch match failed", "match", id, "err", err)
			skipped++
			continue
		}
		events := ParseShots(id, shots)
		if err := s.store.UpsertEvents(ctx, events); err != nil {
			slog.Warn("shotzone: upsert events failed", "match", id, "err", err)
			skipped++
			continue
		}
		fetched++
	}
	slog.Info("shotzone matches processed", "fetched", fetched, "skipped", skipped, "known", len(existing))

	allEvents, err := s.store.ReadEvents(ctx, season)
	if err != nil {
		return fmt.Errorf("ReadEvents: %w", err)
	}
	if len(allEvents) == 0 {
		slog.Warn("shotzone: no events, leaving boards intact", "season", season)
		return nil
	}

	taken := AggregateTaken(allEvents, s.cfg.WindowMatches)
	if err := s.store.ReplaceTakenBoard(ctx, taken); err != nil {
		return fmt.Errorf("ReplaceTakenBoard: %w", err)
	}

	conceded := AggregateConceded(allEvents, s.cfg.WindowMatches)
	if err := s.store.ReplaceConcededBoard(ctx, conceded); err != nil {
		return fmt.Errorf("ReplaceConcededBoard: %w", err)
	}

	slog.Info("shotzone sync complete", "events", len(allEvents), "taken_rows", len(taken), "conceded_rows", len(conceded))
	return nil
}
