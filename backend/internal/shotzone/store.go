package shotzone

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store persists open-play shot events and the two materialised boards. It
// owns only the sz_* tables and never touches players/teams/fixtures.
type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// UpsertEvents idempotently inserts qualifying open-play shots. The unique
// key (match_id, player_id, minute, zone) makes re-syncing a match a no-op.
func (s *Store) UpsertEvents(ctx context.Context, events []Event) error {
	for _, e := range events {
		_, err := s.db.Exec(ctx, `
			INSERT INTO sz_events
				(match_id, season, match_date, minute, team_for, team_against, zone, player_id, player_name, is_goal, xg)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
			ON CONFLICT (match_id, player_id, minute, zone) DO UPDATE SET
				team_for     = EXCLUDED.team_for,
				team_against = EXCLUDED.team_against,
				player_name  = EXCLUDED.player_name,
				is_goal      = EXCLUDED.is_goal,
				xg           = EXCLUDED.xg
		`, e.MatchID, e.Season, e.MatchDate, e.Minute, e.TeamFor, e.TeamAgainst,
			string(e.Zone), e.PlayerID, e.PlayerName, e.IsGoal, e.XG)
		if err != nil {
			return fmt.Errorf("upsert sz_event match=%s player=%s: %w", e.MatchID, e.PlayerID, err)
		}
	}
	return nil
}

// ExistingMatchIDs returns the set of match ids already parsed into sz_events
// for a season, so the syncer can skip re-fetching them.
func (s *Store) ExistingMatchIDs(ctx context.Context, season string) (map[string]bool, error) {
	rows, err := s.db.Query(ctx,
		`SELECT DISTINCT match_id FROM sz_events WHERE season = $1`, season)
	if err != nil {
		return nil, fmt.Errorf("query existing match ids: %w", err)
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

// ReadEvents returns all events for a season.
func (s *Store) ReadEvents(ctx context.Context, season string) ([]Event, error) {
	rows, err := s.db.Query(ctx, `
		SELECT match_id, season, match_date, minute, team_for, team_against, zone,
		       player_id, player_name, is_goal, xg
		FROM sz_events
		WHERE season = $1
	`, season)
	if err != nil {
		return nil, fmt.Errorf("read events: %w", err)
	}
	defer rows.Close()
	var out []Event
	for rows.Next() {
		var e Event
		var zone string
		if err := rows.Scan(&e.MatchID, &e.Season, &e.MatchDate, &e.Minute, &e.TeamFor, &e.TeamAgainst,
			&zone, &e.PlayerID, &e.PlayerName, &e.IsGoal, &e.XG); err != nil {
			return nil, err
		}
		e.Zone = Zone(zone)
		out = append(out, e)
	}
	return out, rows.Err()
}

// ReplaceTakenBoard atomically swaps sz_taken_board for the freshly computed
// rows. A recompute is cheap and always full-set, so delete-all + insert
// inside a transaction is simplest and keeps the board consistent for readers.
func (s *Store) ReplaceTakenBoard(ctx context.Context, rows []TakenRow) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM sz_taken_board`); err != nil {
		return fmt.Errorf("clear taken board: %w", err)
	}
	now := time.Now().UTC()
	for _, r := range rows {
		var lastSeen *time.Time
		if !r.LastSeen.IsZero() {
			lastSeen = &r.LastSeen
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO sz_taken_board
				(team, zone, player_id, player_name, rank, shots, goals, xg, last_seen, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		`, r.Team, string(r.Zone), r.PlayerID, r.PlayerName, r.Rank, r.Shots, r.Goals, r.XG, lastSeen, now)
		if err != nil {
			return fmt.Errorf("insert taken board row: %w", err)
		}
	}
	return tx.Commit(ctx)
}

// ReplaceConcededBoard atomically swaps sz_conceded_board for the freshly
// computed rows.
func (s *Store) ReplaceConcededBoard(ctx context.Context, rows []ConcededRow) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM sz_conceded_board`); err != nil {
		return fmt.Errorf("clear conceded board: %w", err)
	}
	now := time.Now().UTC()
	for _, r := range rows {
		var lastSeen *time.Time
		if !r.LastSeen.IsZero() {
			lastSeen = &r.LastSeen
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO sz_conceded_board
				(team, zone, shots, goals, xg, last_seen, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
		`, r.Team, string(r.Zone), r.Shots, r.Goals, r.XG, lastSeen, now)
		if err != nil {
			return fmt.Errorf("insert conceded board row: %w", err)
		}
	}
	return tx.Commit(ctx)
}

// ListTeams returns the distinct teams observed (as shooters), sorted, for
// the frontend's team selector.
func (s *Store) ListTeams(ctx context.Context) ([]string, error) {
	rows, err := s.db.Query(ctx, `SELECT DISTINCT team FROM sz_taken_board ORDER BY team`)
	if err != nil {
		return nil, fmt.Errorf("list teams: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// BoardUpdatedAt returns when the boards were last recomputed. Zero time when
// empty.
func (s *Store) BoardUpdatedAt(ctx context.Context) (time.Time, error) {
	var t *time.Time
	if err := s.db.QueryRow(ctx, `SELECT max(updated_at) FROM sz_taken_board`).Scan(&t); err != nil {
		return time.Time{}, fmt.Errorf("board updated_at: %w", err)
	}
	if t == nil {
		return time.Time{}, nil
	}
	return *t, nil
}

// ReadTakenBoard returns every taken-board row for a team (all zones incl.
// overall).
func (s *Store) ReadTakenBoard(ctx context.Context, team string) ([]TakenRow, error) {
	rows, err := s.db.Query(ctx, `
		SELECT team, zone, player_id, player_name, rank, shots, goals, xg, last_seen
		FROM sz_taken_board
		WHERE team = $1
		ORDER BY zone, rank
	`, team)
	if err != nil {
		return nil, fmt.Errorf("read taken board: %w", err)
	}
	defer rows.Close()
	return scanTaken(rows)
}

// ReadConcededBoard returns every conceded-board row for a team (all zones
// incl. overall).
func (s *Store) ReadConcededBoard(ctx context.Context, team string) ([]ConcededRow, error) {
	rows, err := s.db.Query(ctx, `
		SELECT team, zone, shots, goals, xg, last_seen
		FROM sz_conceded_board
		WHERE team = $1
		ORDER BY zone
	`, team)
	if err != nil {
		return nil, fmt.Errorf("read conceded board: %w", err)
	}
	defer rows.Close()
	var out []ConcededRow
	for rows.Next() {
		var r ConcededRow
		var zone string
		var lastSeen *time.Time
		if err := rows.Scan(&r.Team, &zone, &r.Shots, &r.Goals, &r.XG, &lastSeen); err != nil {
			return nil, err
		}
		r.Zone = Zone(zone)
		if lastSeen != nil {
			r.LastSeen = *lastSeen
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func scanTaken(rows pgx.Rows) ([]TakenRow, error) {
	var out []TakenRow
	for rows.Next() {
		var r TakenRow
		var zone string
		var lastSeen *time.Time
		if err := rows.Scan(&r.Team, &zone, &r.PlayerID, &r.PlayerName, &r.Rank, &r.Shots, &r.Goals, &r.XG, &lastSeen); err != nil {
			return nil, err
		}
		r.Zone = Zone(zone)
		if lastSeen != nil {
			r.LastSeen = *lastSeen
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
