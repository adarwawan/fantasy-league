package fotmob

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"fantasy-league/internal/setpiece"
	"fantasy-league/internal/shotzone"
)

// Config holds the tunables for the FotMob module (FM_* env / yaml `fotmob:`).
type Config struct {
	Enabled         bool
	Season          string // starting year, e.g. "2026"
	SPWindowMatches int
	SZWindowMatches int
	RecencyHalfLife time.Duration
	// RequestDelay spaces match fetches; FotMob is unofficial, so stay polite.
	RequestDelay time.Duration
}

// Service runs one sync: finished matches -> fetch new -> parse -> store ->
// recompute both boards. Each match is fetched once for both features.
type Service struct {
	cfg    Config
	client *Client
	sp     *setpiece.Store // NewStoreV2
	sz     *shotzone.Store // NewStoreV2
}

func NewService(cfg Config, client *Client, sp *setpiece.Store, sz *shotzone.Store) *Service {
	return &Service{cfg: cfg, client: client, sp: sp, sz: sz}
}

func (s *Service) Sync(ctx context.Context) error {
	if !s.cfg.Enabled {
		slog.Debug("fotmob sync disabled")
		return nil
	}
	season := s.cfg.Season
	slog.Info("fotmob sync start", "season", season)

	ids, err := s.client.FinishedMatches(ctx, season)
	if err != nil {
		return fmt.Errorf("FinishedMatches: %w", err)
	}
	// Every processed match has open-play shots, so sz2_events is the marker.
	// Set-piece events are written first: if the sz write fails the match is
	// retried and the idempotent upsert makes the repeat harmless.
	existing, err := s.sz.ExistingMatchIDs(ctx, season)
	if err != nil {
		return fmt.Errorf("ExistingMatchIDs: %w", err)
	}

	var fetched, skipped int
	for _, id := range ids {
		if existing[id] {
			continue
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		m, err := s.client.Match(ctx, id)
		if err != nil {
			slog.Warn("fotmob: fetch match failed", "match", id, "err", err)
			skipped++
			continue
		}
		spEvents, szEvents := Parse(id, season, m)
		if err := s.sp.UpsertEvents(ctx, spEvents); err != nil {
			slog.Warn("fotmob: upsert setpiece events failed", "match", id, "err", err)
			skipped++
			continue
		}
		if err := s.sz.UpsertEvents(ctx, szEvents); err != nil {
			slog.Warn("fotmob: upsert shotzone events failed", "match", id, "err", err)
			skipped++
			continue
		}
		fetched++
		time.Sleep(s.cfg.RequestDelay)
	}
	slog.Info("fotmob matches processed", "fetched", fetched, "skipped", skipped, "known", len(existing))

	szAll, err := s.sz.ReadEvents(ctx, season)
	if err != nil {
		return fmt.Errorf("ReadEvents shotzone: %w", err)
	}
	if len(szAll) == 0 {
		slog.Warn("fotmob: no shot events, leaving boards intact", "season", season)
		return nil
	}
	spAll, err := s.sp.ReadEvents(ctx, season)
	if err != nil {
		return fmt.Errorf("ReadEvents setpiece: %w", err)
	}
	if len(spAll) > 0 {
		board := setpiece.Aggregate(spAll, setpiece.AggregateConfig{
			WindowMatches: s.cfg.SPWindowMatches, RecencyHalfLife: s.cfg.RecencyHalfLife,
			// Real per-team match lists (every match has open-play shots for both
			// sides), so a match with no set-piece shots still counts toward the window.
			TeamMatches: shotzone.TeamWindows(szAll, s.cfg.SPWindowMatches),
		})
		if err := s.sp.ReplaceBoard(ctx, board); err != nil {
			return fmt.Errorf("ReplaceBoard: %w", err)
		}
	}

	if err := s.sz.ReplaceTakenBoard(ctx, shotzone.AggregateTaken(szAll, s.cfg.SZWindowMatches)); err != nil {
		return fmt.Errorf("ReplaceTakenBoard: %w", err)
	}
	if err := s.sz.ReplaceConcededBoard(ctx, shotzone.AggregateConceded(szAll, s.cfg.SZWindowMatches)); err != nil {
		return fmt.Errorf("ReplaceConcededBoard: %w", err)
	}
	slog.Info("fotmob sync complete", "setpiece_events", len(spAll), "shot_events", len(szAll))
	return nil
}
