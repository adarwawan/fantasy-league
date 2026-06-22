package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// FixtureInfo is a single upcoming fixture for a player or team.
type FixtureInfo struct {
	GW         int       `json:"gw"`
	Opp        string    `json:"opp"`
	HA         string    `json:"ha"`
	Difficulty int       `json:"difficulty"`
	Kickoff    time.Time `json:"kickoff"`
}

// PlayerRow is the read model returned by QueryPlayers.
type PlayerRow struct {
	ID              string
	GameID          string
	Name            string
	TeamID          string
	TeamShortName   string
	TeamName        string
	Position        string
	Price           float64
	Form            float64
	GlobalOwnership float64
	TopNOwnership   float64
	Status          string
	News            string
	Fixtures        []FixtureInfo
}

// TeamRow is the read model returned by QueryTeams.
type TeamRow struct {
	ID        string
	GameID    string
	Name      string
	ShortName string
	AttForm   float64
	DefForm   float64
	OvrForm   float64
	Fixtures  []FixtureInfo
}

// FixtureRow is the read model returned by QueryFixtures.
type FixtureRow struct {
	ID             string
	GameID         string
	GW             int
	HomeTeamID     string
	HomeShortName  string
	AwayTeamID     string
	AwayShortName  string
	HomeDifficulty int
	AwayDifficulty int
	KickoffTime    time.Time
	Finished       bool
}

var validSortCols = map[string]string{
	"global_ownership": "p.global_ownership DESC",
	"top_n_ownership":  "COALESCE(o.ownership, 0) DESC",
	"form":             "p.form DESC",
	"price":            "p.price DESC",
	"name":             "p.name ASC",
}

// QueryPlayers returns players for a game with next-5-fixtures joined.
// sort must be one of the keys in validSortCols; defaults to global_ownership.
// pos filters by position ("" = all). maxPrice = 0 means no filter.
// topN selects which ownership tier to join from player_top_n_ownerships.
func (s *Store) QueryPlayers(ctx context.Context, gameID, pos string, maxPrice float64, sort string, topN int) ([]PlayerRow, error) {
	orderBy, ok := validSortCols[sort]
	if !ok {
		orderBy = validSortCols["global_ownership"]
	}

	q := fmt.Sprintf(`
		WITH next_gw AS (
			SELECT COALESCE(MIN(gw), 0) as gw FROM fixtures WHERE game_id = $1 AND NOT finished
		),
		team_fixtures AS (
			SELECT
				f.home_team_id AS team_id, f.gw,
				at.short_name AS opp, 'H' AS ha,
				f.home_difficulty AS difficulty, f.kickoff_time
			FROM fixtures f
			JOIN teams at ON at.id = f.away_team_id
			WHERE f.game_id = $1 AND NOT f.finished AND f.gw >= (SELECT gw FROM next_gw)
			UNION ALL
			SELECT
				f.away_team_id AS team_id, f.gw,
				ht.short_name AS opp, 'A' AS ha,
				f.away_difficulty AS difficulty, f.kickoff_time
			FROM fixtures f
			JOIN teams ht ON ht.id = f.home_team_id
			WHERE f.game_id = $1 AND NOT f.finished AND f.gw >= (SELECT gw FROM next_gw)
		),
		ranked AS (
			SELECT *, ROW_NUMBER() OVER (PARTITION BY team_id ORDER BY gw) AS rn FROM team_fixtures
		),
		top5 AS (SELECT * FROM ranked WHERE rn <= 5),
		player_fixtures AS (
			SELECT
				p.id AS player_id,
				COALESCE(
					json_agg(
						json_build_object(
							'gw', tf.gw, 'opp', tf.opp, 'ha', tf.ha,
							'difficulty', tf.difficulty, 'kickoff', tf.kickoff_time
						) ORDER BY tf.gw
					) FILTER (WHERE tf.team_id IS NOT NULL),
					'[]'::json
				) AS fixtures
			FROM players p
			LEFT JOIN top5 tf ON tf.team_id = p.team_id
			WHERE p.game_id = $1
			GROUP BY p.id
		)
		SELECT
			p.id, p.game_id, p.name,
			t.id, t.short_name, t.name,
			p.position, p.price, p.form,
			p.global_ownership, COALESCE(o.ownership, 0),
			p.status, COALESCE(p.news, ''),
			pf.fixtures
		FROM players p
		JOIN teams t ON t.id = p.team_id
		JOIN player_fixtures pf ON pf.player_id = p.id
		LEFT JOIN player_top_n_ownerships o ON o.player_id = p.id AND o.top_n = $4
		WHERE p.game_id = $1
		  AND ($2::text = '' OR p.position = $2)
		  AND ($3::numeric = 0 OR p.price <= $3)
		ORDER BY %s
	`, orderBy)

	rows, err := s.db.Query(ctx, q, gameID, pos, maxPrice, topN)
	if err != nil {
		return nil, fmt.Errorf("query players: %w", err)
	}
	defer rows.Close()

	var out []PlayerRow
	for rows.Next() {
		var r PlayerRow
		var fixturesJSON []byte
		if err := rows.Scan(
			&r.ID, &r.GameID, &r.Name,
			&r.TeamID, &r.TeamShortName, &r.TeamName,
			&r.Position, &r.Price, &r.Form,
			&r.GlobalOwnership, &r.TopNOwnership,
			&r.Status, &r.News, &fixturesJSON,
		); err != nil {
			return nil, fmt.Errorf("scan player: %w", err)
		}
		if err := json.Unmarshal(fixturesJSON, &r.Fixtures); err != nil {
			return nil, fmt.Errorf("unmarshal fixtures: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// QueryTeams returns all teams for a game with next-5-fixtures joined.
func (s *Store) QueryTeams(ctx context.Context, gameID string) ([]TeamRow, error) {
	q := `
		WITH next_gw AS (
			SELECT COALESCE(MIN(gw), 0) AS gw FROM fixtures WHERE game_id = $1 AND NOT finished
		),
		team_fixtures AS (
			SELECT
				f.home_team_id AS team_id, f.gw,
				at.short_name AS opp, 'H' AS ha,
				f.home_difficulty AS difficulty, f.kickoff_time
			FROM fixtures f
			JOIN teams at ON at.id = f.away_team_id
			WHERE f.game_id = $1 AND NOT f.finished AND f.gw >= (SELECT gw FROM next_gw)
			UNION ALL
			SELECT
				f.away_team_id AS team_id, f.gw,
				ht.short_name AS opp, 'A' AS ha,
				f.away_difficulty AS difficulty, f.kickoff_time
			FROM fixtures f
			JOIN teams ht ON ht.id = f.home_team_id
			WHERE f.game_id = $1 AND NOT f.finished AND f.gw >= (SELECT gw FROM next_gw)
		),
		ranked AS (
			SELECT *, ROW_NUMBER() OVER (PARTITION BY team_id ORDER BY gw) AS rn FROM team_fixtures
		),
		top5 AS (SELECT * FROM ranked WHERE rn <= 5),
		team_fix AS (
			SELECT
				t.id AS team_id,
				COALESCE(
					json_agg(
						json_build_object(
							'gw', tf.gw, 'opp', tf.opp, 'ha', tf.ha,
							'difficulty', tf.difficulty, 'kickoff', tf.kickoff_time
						) ORDER BY tf.gw
					) FILTER (WHERE tf.team_id IS NOT NULL),
					'[]'::json
				) AS fixtures
			FROM teams t
			LEFT JOIN top5 tf ON tf.team_id = t.id
			WHERE t.game_id = $1
			GROUP BY t.id
		)
		SELECT
			t.id, t.game_id, t.name, t.short_name,
			t.att_form, t.def_form, t.ovr_form,
			tf.fixtures
		FROM teams t
		JOIN team_fix tf ON tf.team_id = t.id
		WHERE t.game_id = $1
		ORDER BY t.ovr_form DESC
	`

	rows, err := s.db.Query(ctx, q, gameID)
	if err != nil {
		return nil, fmt.Errorf("query teams: %w", err)
	}
	defer rows.Close()

	var out []TeamRow
	for rows.Next() {
		var r TeamRow
		var fixturesJSON []byte
		if err := rows.Scan(
			&r.ID, &r.GameID, &r.Name, &r.ShortName,
			&r.AttForm, &r.DefForm, &r.OvrForm, &fixturesJSON,
		); err != nil {
			return nil, fmt.Errorf("scan team: %w", err)
		}
		if err := json.Unmarshal(fixturesJSON, &r.Fixtures); err != nil {
			return nil, fmt.Errorf("unmarshal fixtures: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// QueryFixtures returns fixtures for a game between fromGW and toGW (inclusive).
func (s *Store) QueryFixtures(ctx context.Context, gameID string, fromGW, toGW int) ([]FixtureRow, error) {
	q := `
		SELECT
			f.id, f.game_id, f.gw,
			ht.id, ht.short_name,
			at.id, at.short_name,
			f.home_difficulty, f.away_difficulty,
			f.kickoff_time, f.finished
		FROM fixtures f
		JOIN teams ht ON ht.id = f.home_team_id
		JOIN teams at ON at.id = f.away_team_id
		WHERE f.game_id = $1
		  AND ($2 = 0 OR f.gw >= $2)
		  AND ($3 = 0 OR f.gw <= $3)
		ORDER BY f.gw, f.kickoff_time
	`
	rows, err := s.db.Query(ctx, q, gameID, fromGW, toGW)
	if err != nil {
		return nil, fmt.Errorf("query fixtures: %w", err)
	}
	defer rows.Close()

	var out []FixtureRow
	for rows.Next() {
		var r FixtureRow
		if err := rows.Scan(
			&r.ID, &r.GameID, &r.GW,
			&r.HomeTeamID, &r.HomeShortName,
			&r.AwayTeamID, &r.AwayShortName,
			&r.HomeDifficulty, &r.AwayDifficulty,
			&r.KickoffTime, &r.Finished,
		); err != nil {
			return nil, fmt.Errorf("scan fixture: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// CurrentGW returns the minimum unfinished GW for a game, or 0 if none.
func (s *Store) CurrentGW(ctx context.Context, gameID string) (int, error) {
	var gw int
	err := s.db.QueryRow(ctx,
		`SELECT COALESCE(MIN(gw), 0) FROM fixtures WHERE game_id = $1 AND NOT finished`,
		gameID,
	).Scan(&gw)
	return gw, err
}
