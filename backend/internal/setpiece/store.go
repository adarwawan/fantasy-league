package setpiece

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store persists set-piece events and the materialised board. It owns only the
// sp_* tables and never touches players/teams/fixtures (docs §3).
type Store struct {
	db *pgxpool.Pool
	v2 bool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// NewStoreV2 targets the FotMob-sourced sp2_* tables (migration 014). The SQL is
// shared with v1; sql() swaps the table names so both stay in lockstep.
func NewStoreV2(db *pgxpool.Pool) *Store {
	return &Store{db: db, v2: true}
}

var v2Tables = strings.NewReplacer("sp_events", "sp2_events", "sp_board", "sp2_board")

func (s *Store) sql(q string) string {
	if s.v2 {
		return v2Tables.Replace(q)
	}
	return q
}

// UpsertEvents idempotently inserts qualifying set-piece shots. The unique key
// (match_id, role, duty, player_id, minute) makes re-syncing a match a no-op.
func (s *Store) UpsertEvents(ctx context.Context, events []Event) error {
	if s.v2 {
		return s.upsertEventsV2(ctx, events)
	}
	for _, e := range events {
		_, err := s.db.Exec(ctx, s.sql(`
			INSERT INTO sp_events
				(match_id, season, match_date, minute, understat_team, role, duty, player_id, player_name, is_header, is_goal, xg)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			ON CONFLICT (match_id, role, duty, player_id, minute) DO UPDATE SET
				understat_team = EXCLUDED.understat_team,
				player_name    = EXCLUDED.player_name,
				is_header      = EXCLUDED.is_header,
				is_goal        = EXCLUDED.is_goal,
				xg             = EXCLUDED.xg
		`), e.MatchID, e.Season, e.MatchDate, e.Minute, e.UnderstatTeam,
			string(e.Role), string(e.Duty), e.PlayerID, e.PlayerName, e.IsHeader, e.IsGoal, e.XG)
		if err != nil {
			return fmt.Errorf("upsert sp_event match=%s player=%s: %w", e.MatchID, e.PlayerID, err)
		}
	}
	return nil
}

// upsertEventsV2 keys on FotMob's per-shot id: unlike v1's
// (match, role, duty, player, minute), two shots in one minute stay distinct.
func (s *Store) upsertEventsV2(ctx context.Context, events []Event) error {
	for _, e := range events {
		_, err := s.db.Exec(ctx, `
			INSERT INTO sp2_events
				(match_id, shot_id, season, match_date, minute, understat_team, role, duty, player_id, player_name, is_header, is_goal, xg)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
			ON CONFLICT (match_id, shot_id) DO UPDATE SET
				understat_team = EXCLUDED.understat_team,
				role           = EXCLUDED.role,
				duty           = EXCLUDED.duty,
				player_name    = EXCLUDED.player_name,
				is_header      = EXCLUDED.is_header,
				is_goal        = EXCLUDED.is_goal,
				xg             = EXCLUDED.xg
		`, e.MatchID, e.ShotID, e.Season, e.MatchDate, e.Minute, e.UnderstatTeam,
			string(e.Role), string(e.Duty), e.PlayerID, e.PlayerName, e.IsHeader, e.IsGoal, e.XG)
		if err != nil {
			return fmt.Errorf("upsert sp2_event match=%s shot=%s: %w", e.MatchID, e.ShotID, err)
		}
	}
	return nil
}

// ExistingMatchIDs returns the set of match ids already parsed into sp_events
// for a season, so the syncer can skip re-fetching them.
func (s *Store) ExistingMatchIDs(ctx context.Context, season string) (map[string]bool, error) {
	rows, err := s.db.Query(ctx,
		s.sql(`SELECT DISTINCT match_id FROM sp_events WHERE season = $1`), season)
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

// ReadEvents returns all events for a season, ordered newest-first.
func (s *Store) ReadEvents(ctx context.Context, season string) ([]Event, error) {
	rows, err := s.db.Query(ctx, s.sql(`
		SELECT match_id, season, match_date, minute, understat_team, role, duty,
		       player_id, player_name, is_header, is_goal, xg
		FROM sp_events
		WHERE season = $1
		ORDER BY match_date DESC
	`), season)
	if err != nil {
		return nil, fmt.Errorf("read events: %w", err)
	}
	defer rows.Close()
	var out []Event
	for rows.Next() {
		var e Event
		var role, duty string
		if err := rows.Scan(&e.MatchID, &e.Season, &e.MatchDate, &e.Minute, &e.UnderstatTeam,
			&role, &duty, &e.PlayerID, &e.PlayerName, &e.IsHeader, &e.IsGoal, &e.XG); err != nil {
			return nil, err
		}
		e.Role, e.Duty = Role(role), Duty(duty)
		out = append(out, e)
	}
	return out, rows.Err()
}

// ReplaceBoard atomically swaps the entire sp_board for the freshly computed
// rows. A recompute is cheap and always full-set, so delete-all + insert inside
// a transaction is simplest and keeps the board consistent for readers.
func (s *Store) ReplaceBoard(ctx context.Context, rows []BoardRow) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, s.sql(`DELETE FROM sp_board`)); err != nil {
		return fmt.Errorf("clear board: %w", err)
	}
	now := time.Now().UTC()
	for _, r := range rows {
		var lastSeen *time.Time
		if !r.LastSeen.IsZero() {
			lastSeen = &r.LastSeen
		}
		_, err := tx.Exec(ctx, s.sql(`
			INSERT INTO sp_board
				(understat_team, role, duty, player_id, player_name, rank,
				 weighted_score, raw_count, goals, xg_sum, header_pct, last_seen, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		`), r.UnderstatTeam, string(r.Role), string(r.Duty), r.PlayerID, r.PlayerName,
			r.Rank, r.WeightedScore, r.RawCount, r.Goals, r.XGSum, r.HeaderPct, lastSeen, now)
		if err != nil {
			return fmt.Errorf("insert board row: %w", err)
		}
	}
	return tx.Commit(ctx)
}

// ReadBoard returns every board row, ordered for stable grouping in the handler.
func (s *Store) ReadBoard(ctx context.Context) ([]BoardRow, error) {
	rows, err := s.db.Query(ctx, s.sql(`
		SELECT understat_team, role, duty, player_id, player_name, rank,
		       weighted_score, raw_count, goals, xg_sum, header_pct, last_seen
		FROM sp_board
		ORDER BY understat_team, role, duty, rank
	`))
	if err != nil {
		return nil, fmt.Errorf("read board: %w", err)
	}
	defer rows.Close()
	return scanBoard(rows)
}

// BoardUpdatedAt returns when the board was last recomputed (all rows share the
// timestamp set by ReplaceBoard). Zero time when the board is empty.
func (s *Store) BoardUpdatedAt(ctx context.Context) (time.Time, error) {
	var t *time.Time
	if err := s.db.QueryRow(ctx, s.sql(`SELECT max(updated_at) FROM sp_board`)).Scan(&t); err != nil {
		return time.Time{}, fmt.Errorf("board updated_at: %w", err)
	}
	if t == nil {
		return time.Time{}, nil
	}
	return *t, nil
}

// ReadTeamBoard returns board rows for a single team.
func (s *Store) ReadTeamBoard(ctx context.Context, team string) ([]BoardRow, error) {
	rows, err := s.db.Query(ctx, s.sql(`
		SELECT understat_team, role, duty, player_id, player_name, rank,
		       weighted_score, raw_count, goals, xg_sum, header_pct, last_seen
		FROM sp_board
		WHERE understat_team = $1
		ORDER BY role, duty, rank
	`), team)
	if err != nil {
		return nil, fmt.Errorf("read team board: %w", err)
	}
	defer rows.Close()
	return scanBoard(rows)
}

func scanBoard(rows pgx.Rows) ([]BoardRow, error) {
	var out []BoardRow
	for rows.Next() {
		var r BoardRow
		var role, duty string
		var lastSeen *time.Time
		if err := rows.Scan(&r.UnderstatTeam, &role, &duty, &r.PlayerID, &r.PlayerName,
			&r.Rank, &r.WeightedScore, &r.RawCount, &r.Goals, &r.XGSum, &r.HeaderPct, &lastSeen); err != nil {
			return nil, err
		}
		r.Role, r.Duty = Role(role), Duty(duty)
		if lastSeen != nil {
			r.LastSeen = *lastSeen
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
