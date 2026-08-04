package trends

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// retainGWs is how many gameweek partitions to keep. Arming GW N drops GW N-2
// so at most this many windows of snapshots ever exist on disk.
const retainGWs = 2

// Store owns the trends_* tables. It shares the pool with the main app but
// never touches its tables except a read-only join to players/teams for display.
type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store { return &Store{db: db} }

// ArmSession starts (or re-arms) a capture window for a gameweek. In one
// transaction it creates the GW's snapshot partition, drops the now-expired
// GW N-retainGWs partition (retention), and upserts the session row active.
func (s *Store) ArmSession(ctx context.Context, gw int, deadline time.Time) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create this GW's partition (idempotent).
	if _, err := tx.Exec(ctx, fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS trends_snapshot_gw%d
		   PARTITION OF trends_snapshot FOR VALUES IN (%d)`, gw, gw)); err != nil {
		return fmt.Errorf("create partition gw%d: %w", gw, err)
	}

	// Drop the expired partition — instant reclaim, no vacuum needed.
	if old := gw - retainGWs; old > 0 {
		if _, err := tx.Exec(ctx, fmt.Sprintf(
			`DROP TABLE IF EXISTS trends_snapshot_gw%d`, old)); err != nil {
			return fmt.Errorf("drop partition gw%d: %w", old, err)
		}
	}

	// Deactivate any other session, then upsert this one active.
	if _, err := tx.Exec(ctx, `UPDATE trends_session SET active = false WHERE gameweek <> $1`, gw); err != nil {
		return fmt.Errorf("deactivate others: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO trends_session (gameweek, ends_at, active, poll_count)
		VALUES ($1, $2, true, 0)
		ON CONFLICT (gameweek) DO UPDATE SET
			ends_at = EXCLUDED.ends_at,
			active  = true
	`, gw, deadline); err != nil {
		return fmt.Errorf("upsert session: %w", err)
	}
	return tx.Commit(ctx)
}

// ActiveSession returns the currently active session, or nil if none. A session
// past its deadline is auto-deactivated and treated as inactive.
func (s *Store) ActiveSession(ctx context.Context) (*Session, error) {
	var sess Session
	err := s.db.QueryRow(ctx, `
		SELECT gameweek, started_at, ends_at, active, poll_count
		FROM trends_session
		WHERE active = true
		ORDER BY started_at DESC
		LIMIT 1
	`).Scan(&sess.Gameweek, &sess.StartedAt, &sess.EndsAt, &sess.Active, &sess.PollCount)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("active session: %w", err)
	}
	if time.Now().After(sess.EndsAt) {
		_, _ = s.db.Exec(ctx, `UPDATE trends_session SET active = false WHERE gameweek = $1`, sess.Gameweek)
		return nil, nil
	}
	return &sess, nil
}

// InsertSnapshots appends changed snapshots for a gameweek at captured_at and
// bumps the session poll counter. Dedup (skipping unchanged players) happens in
// the poller; this just writes what it's given.
func (s *Store) InsertSnapshots(ctx context.Context, gw int, capturedAt time.Time, snaps []Snapshot) error {
	if len(snaps) == 0 {
		_, err := s.db.Exec(ctx, `UPDATE trends_session SET poll_count = poll_count + 1 WHERE gameweek = $1`, gw)
		return err
	}
	batch := &pgx.Batch{}
	for _, sn := range snaps {
		batch.Queue(`
			INSERT INTO trends_snapshot
				(captured_at, gameweek, player_ext_id, selected_bp, transfers_in, transfers_out, now_cost)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
		`, capturedAt, gw, sn.PlayerExtID, sn.SelectedBP, sn.TransfersIn, sn.TransfersOut, sn.NowCost)
	}
	batch.Queue(`UPDATE trends_session SET poll_count = poll_count + 1 WHERE gameweek = $1`, gw)

	br := s.db.SendBatch(ctx, batch)
	defer br.Close()
	for range snaps {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("insert snapshot: %w", err)
		}
	}
	if _, err := br.Exec(); err != nil { // poll_count update
		return fmt.Errorf("bump poll_count: %w", err)
	}
	return nil
}

// Leaders returns the top-limit players ranked by the change over the trailing
// window ending now, for the given gameweek. The ranking metric is net transfers
// (in-minus-out delta) or ownership (selected_by_percent delta) per `metric` —
// GW1 uses ownership since transfers don't exist before the first deadline.
// Direction "out" ranks by the most negative delta (mass exodus); "in" by the
// largest gain. Player display fields are joined read-only from players/teams.
//
// LeaderRow.RankDelta carries the metric's delta in its native unit: an integer
// transfer count for transfers, or basis points of ownership for ownership.
func (s *Store) Leaders(ctx context.Context, gw, limit int, window time.Duration, dir Direction, metric Metric) ([]LeaderRow, error) {
	cutoff := time.Now().Add(-window)

	// Metric-specific delta expression, ranked by the same column.
	deltaExpr := "(l.transfers_in - l.transfers_out) - COALESCE(b.net, l.transfers_in - l.transfers_out)"
	if metric == MetricOwnership {
		deltaExpr = "l.selected_bp - COALESCE(b.selected_bp, l.selected_bp)"
	}
	order := "rank_delta DESC"
	if dir == DirectionOut {
		order = "rank_delta ASC"
	}

	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		WITH latest AS (
			SELECT DISTINCT ON (player_ext_id)
				player_ext_id, selected_bp, transfers_in, transfers_out, now_cost
			FROM trends_snapshot
			WHERE gameweek = $1
			ORDER BY player_ext_id, captured_at DESC
		),
		baseline AS (
			SELECT DISTINCT ON (player_ext_id)
				player_ext_id, selected_bp, (transfers_in - transfers_out) AS net
			FROM trends_snapshot
			WHERE gameweek = $1 AND captured_at <= $2
			ORDER BY player_ext_id, captured_at DESC
		)
		SELECT l.player_ext_id,
		       COALESCE(p.name, 'Player ' || l.player_ext_id),
		       COALESCE(t.short_name, ''),
		       COALESCE(p.position, ''),
		       l.selected_bp,
		       l.now_cost,
		       (l.transfers_in - l.transfers_out) AS net,
		       %s AS rank_delta
		FROM latest l
		LEFT JOIN baseline b ON b.player_ext_id = l.player_ext_id
		LEFT JOIN players p ON p.game_id = 'fpl' AND p.external_id = l.player_ext_id
		LEFT JOIN teams   t ON t.id = p.team_id
		ORDER BY %s
		LIMIT $3
	`, deltaExpr, order), gw, cutoff, limit)
	if err != nil {
		return nil, fmt.Errorf("leaders: %w", err)
	}
	defer rows.Close()

	var out []LeaderRow
	for rows.Next() {
		var r LeaderRow
		var selectedBP, nowCost int
		if err := rows.Scan(&r.PlayerExtID, &r.Name, &r.Team, &r.Position,
			&selectedBP, &nowCost, &r.NetTransfers, &r.RankDelta); err != nil {
			return nil, err
		}
		r.SelectedPct = float64(selectedBP) / 100
		r.NowCost = float64(nowCost) / 10
		out = append(out, r)
	}
	return out, rows.Err()
}

// Series returns the full snapshot series for one player in a gameweek.
func (s *Store) Series(ctx context.Context, gw, playerExtID int) ([]SeriesPoint, error) {
	rows, err := s.db.Query(ctx, `
		SELECT captured_at, selected_bp, transfers_in, transfers_out, now_cost
		FROM trends_snapshot
		WHERE gameweek = $1 AND player_ext_id = $2
		ORDER BY captured_at
	`, gw, playerExtID)
	if err != nil {
		return nil, fmt.Errorf("series: %w", err)
	}
	defer rows.Close()

	var out []SeriesPoint
	for rows.Next() {
		var p SeriesPoint
		var selectedBP, nowCost int
		if err := rows.Scan(&p.CapturedAt, &selectedBP, &p.TransfersIn, &p.TransfersOut, &nowCost); err != nil {
			return nil, err
		}
		p.SelectedPct = float64(selectedBP) / 100
		p.NowCost = float64(nowCost) / 10
		p.NetTransfers = p.TransfersIn - p.TransfersOut
		out = append(out, p)
	}
	return out, rows.Err()
}
