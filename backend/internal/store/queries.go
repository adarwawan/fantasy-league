package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// FixtureInfo is a single upcoming fixture for a player or team.
type FixtureInfo struct {
	GW         int       `json:"gw"`
	Opp        string    `json:"opp"`
	HA         string    `json:"ha"`
	Difficulty int       `json:"difficulty"`
	Kickoff    time.Time `json:"kickoff"`
	XG         *float64  `json:"xg"`
	CSPct      *float64  `json:"cs_pct"`
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
	EffectiveOwn    float64
	Status          string
	News            string
	// Set-piece taker ranks (1 = first choice), nil when the player has no duty.
	PenaltiesOrder       *int
	DirectFreekicksOrder *int
	CornersIndirectOrder *int
	Fixtures             []FixtureInfo
}

// PlayerOwnership is the minimal read model used for ownership ranking.
type PlayerOwnership struct {
	PlayerID        string
	Position        string
	GlobalOwnership float64
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
	XGSum     *float64
	CSAvg     *float64
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

var validTeamSortCols = map[string]string{
	"xg_sum":   "ta.xg_sum DESC NULLS LAST",
	"cs_avg":   "ta.cs_avg DESC NULLS LAST",
	"ovr_form": "t.ovr_form DESC",
}

var validSortCols = map[string]string{
	"global_ownership":    "p.global_ownership DESC",
	"top_n_ownership":     "COALESCE(o.ownership, 0) DESC",
	"effective_ownership": "COALESCE(o.effective_ownership, 0) DESC",
	"form":                "p.form DESC",
	"price":               "p.price DESC",
	"name":                "p.name ASC",
}

// QueryPlayers returns players for a game with fixtures over the next 5
// gameweeks joined (the fixed GW range [next_gw, next_gw+4]). Windowing is a
// calendar-GW range, not a fixture or per-team count: a double gameweek yields
// extra fixtures within the range, a blank gameweek yields fewer, and neither
// spills the window past next_gw+4.
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
				f.home_difficulty AS difficulty, f.kickoff_time,
				mo.lambda_home AS xg, mo.home_cs_pct AS cs_pct
			FROM fixtures f
			JOIN teams at ON at.id = f.away_team_id
			LEFT JOIN match_odds mo ON mo.fixture_id = f.id
			WHERE f.game_id = $1 AND NOT f.finished
			  AND f.gw >= (SELECT gw FROM next_gw)
			  AND f.gw <  (SELECT gw FROM next_gw) + 5
			UNION ALL
			SELECT
				f.away_team_id AS team_id, f.gw,
				ht.short_name AS opp, 'A' AS ha,
				f.away_difficulty AS difficulty, f.kickoff_time,
				mo.lambda_away AS xg, mo.away_cs_pct AS cs_pct
			FROM fixtures f
			JOIN teams ht ON ht.id = f.home_team_id
			LEFT JOIN match_odds mo ON mo.fixture_id = f.id
			WHERE f.game_id = $1 AND NOT f.finished
			  AND f.gw >= (SELECT gw FROM next_gw)
			  AND f.gw <  (SELECT gw FROM next_gw) + 5
		),
		player_fixtures AS (
			SELECT
				p.id AS player_id,
				COALESCE(
					json_agg(
						json_build_object(
							'gw', tf.gw, 'opp', tf.opp, 'ha', tf.ha,
							'difficulty', tf.difficulty, 'kickoff', tf.kickoff_time,
							'xg', tf.xg, 'cs_pct', tf.cs_pct
						) ORDER BY tf.gw, tf.kickoff_time
					) FILTER (WHERE tf.team_id IS NOT NULL),
					'[]'::json
				) AS fixtures
			FROM players p
			LEFT JOIN team_fixtures tf ON tf.team_id = p.team_id
			WHERE p.game_id = $1
			GROUP BY p.id
		)
		SELECT
			p.id, p.game_id, p.name,
			t.id, t.short_name, t.name,
			p.position, p.price, p.form,
			p.global_ownership, COALESCE(o.ownership, 0), COALESCE(o.effective_ownership, 0),
			p.status, COALESCE(p.news, ''),
			p.penalties_order, p.direct_freekicks_order, p.corners_indirect_freekicks_order,
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
			&r.GlobalOwnership, &r.TopNOwnership, &r.EffectiveOwn,
			&r.Status, &r.News,
			&r.PenaltiesOrder, &r.DirectFreekicksOrder, &r.CornersIndirectOrder,
			&fixturesJSON,
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

// QueryPlayerOwnerships returns every player's position and global ownership
// for a game, for ownership ranking at the service level.
func (s *Store) QueryPlayerOwnerships(ctx context.Context, gameID string) ([]PlayerOwnership, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, position, global_ownership FROM players WHERE game_id = $1`,
		gameID,
	)
	if err != nil {
		return nil, fmt.Errorf("query player ownerships: %w", err)
	}
	defer rows.Close()

	var out []PlayerOwnership
	for rows.Next() {
		var r PlayerOwnership
		if err := rows.Scan(&r.PlayerID, &r.Position, &r.GlobalOwnership); err != nil {
			return nil, fmt.Errorf("scan player ownership: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// QueryRecentGWPoints returns each player's points in the last `window`
// finished gameweeks of a game, plus how many gameweeks were inspected
// (fewer than window early in a season).
func (s *Store) QueryRecentGWPoints(ctx context.Context, gameID string, window int) (map[string][]int, int, error) {
	var gws []int
	gwRows, err := s.db.Query(ctx, `
		SELECT DISTINCT gw FROM fixtures
		WHERE game_id = $1 AND finished
		ORDER BY gw DESC
		LIMIT $2
	`, gameID, window)
	if err != nil {
		return nil, 0, fmt.Errorf("query recent gws: %w", err)
	}
	defer gwRows.Close()
	for gwRows.Next() {
		var gw int
		if err := gwRows.Scan(&gw); err != nil {
			return nil, 0, fmt.Errorf("scan recent gw: %w", err)
		}
		gws = append(gws, gw)
	}
	if err := gwRows.Err(); err != nil {
		return nil, 0, err
	}
	if len(gws) == 0 {
		return map[string][]int{}, 0, nil
	}

	rows, err := s.db.Query(ctx, `
		SELECT player_id, COALESCE(points, 0)
		FROM player_gw_stats
		WHERE game_id = $1 AND gw = ANY($2)
	`, gameID, gws)
	if err != nil {
		return nil, 0, fmt.Errorf("query recent gw points: %w", err)
	}
	defer rows.Close()

	points := make(map[string][]int)
	for rows.Next() {
		var playerID string
		var pts int
		if err := rows.Scan(&playerID, &pts); err != nil {
			return nil, 0, fmt.Errorf("scan recent gw points: %w", err)
		}
		points[playerID] = append(points[playerID], pts)
	}
	return points, len(gws), rows.Err()
}

// GWPoints is a single gameweek's points for a player.
type GWPoints struct {
	GW     int `json:"gw"`
	Points int `json:"points"`
}

// QueryRecentGWPointsByGW returns each player's points for the last `window`
// finished gameweeks of a game, ordered oldest → newest so callers can render
// them left-to-right. Only gameweeks a player has a stats row for are included.
func (s *Store) QueryRecentGWPointsByGW(ctx context.Context, gameID string, window int) (map[string][]GWPoints, error) {
	rows, err := s.db.Query(ctx, `
		WITH recent AS (
			SELECT DISTINCT gw FROM fixtures
			WHERE game_id = $1 AND finished
			ORDER BY gw DESC
			LIMIT $2
		)
		SELECT s.player_id, s.gw, COALESCE(s.points, 0)
		FROM player_gw_stats s
		JOIN recent r ON r.gw = s.gw
		WHERE s.game_id = $1
		ORDER BY s.player_id, s.gw
	`, gameID, window)
	if err != nil {
		return nil, fmt.Errorf("query recent gw points by gw: %w", err)
	}
	defer rows.Close()

	out := make(map[string][]GWPoints)
	for rows.Next() {
		var playerID string
		var gp GWPoints
		if err := rows.Scan(&playerID, &gp.GW, &gp.Points); err != nil {
			return nil, fmt.Errorf("scan recent gw points by gw: %w", err)
		}
		out[playerID] = append(out[playerID], gp)
	}
	return out, rows.Err()
}

// PlayerStatGW is one player's raw stat line for a single finished gameweek,
// with identity fields joined for presentation. The Stats service aggregates
// these over the window and derives per-position leaders; keeping the store
// output raw (no thresholds, no ranking) makes the scoring logic testable in Go.
type PlayerStatGW struct {
	PlayerID              string
	Position              string
	Name                  string
	TeamShortName         string
	Goals                 int
	Assists               int
	CleanSheets           int
	Bonus                 int
	DefensiveContribution int // defensive-contribution points for the GW (0/2/4), summed per fixture
	Influence             float64
	Creativity            float64
	Threat                float64
}

// QueryRecentPlayerStatLines returns every player's raw per-gameweek stat lines
// over the last `window` finished gameweeks of a game, joined to the player's
// position, name and team. One row per (player, gameweek).
func (s *Store) QueryRecentPlayerStatLines(ctx context.Context, gameID string, window int) ([]PlayerStatGW, error) {
	rows, err := s.db.Query(ctx, `
		WITH recent AS (
			SELECT DISTINCT gw FROM fixtures
			WHERE game_id = $1 AND finished
			ORDER BY gw DESC
			LIMIT $2
		)
		SELECT
			p.id::text, p.position, p.name, t.short_name,
			s.goals, s.assists, s.clean_sheets, s.bonus, s.defensive_contribution,
			s.influence, s.creativity, s.threat
		FROM player_gw_stats s
		JOIN recent r ON r.gw = s.gw
		JOIN players p ON p.id = s.player_id
		JOIN teams t ON t.id = p.team_id
		WHERE s.game_id = $1
	`, gameID, window)
	if err != nil {
		return nil, fmt.Errorf("query recent player stat lines: %w", err)
	}
	defer rows.Close()

	var out []PlayerStatGW
	for rows.Next() {
		var r PlayerStatGW
		if err := rows.Scan(
			&r.PlayerID, &r.Position, &r.Name, &r.TeamShortName,
			&r.Goals, &r.Assists, &r.CleanSheets, &r.Bonus, &r.DefensiveContribution,
			&r.Influence, &r.Creativity, &r.Threat,
		); err != nil {
			return nil, fmt.Errorf("scan player stat line: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// QueryTeams returns all teams for a game with fixtures over the next N
// gameweeks joined (the fixed GW range [next_gw, next_gw+window-1]), xG and CS%
// pulled from match_odds, and aggregate xg_sum / cs_avg over that range.
// Windowing is a calendar-GW range, not a fixture or per-team count: a double
// gameweek contributes all of its fixtures (xg_sum sums over them), a blank
// gameweek contributes none, and neither shifts the window's upper bound.
// window is clamped to [1, 10] and defaults to 5. sort must be one of
// "xg_sum", "cs_avg", or "ovr_form" (default).
func (s *Store) QueryTeams(ctx context.Context, gameID string, window int, sort string) ([]TeamRow, error) {
	if window < 1 {
		window = 1
	} else if window > 10 {
		window = 10
	}
	orderBy, ok := validTeamSortCols[sort]
	if !ok {
		orderBy = validTeamSortCols["ovr_form"]
	}

	q := fmt.Sprintf(`
		WITH next_gw AS (
			SELECT COALESCE(MIN(gw), 0) AS gw FROM fixtures WHERE game_id = $1 AND NOT finished
		),
		team_fixtures AS (
			SELECT
				f.home_team_id AS team_id, f.gw,
				at.short_name AS opp, 'H' AS ha,
				f.home_difficulty AS difficulty, f.kickoff_time,
				mo.lambda_home AS xg, mo.home_cs_pct AS cs_pct
			FROM fixtures f
			JOIN teams at ON at.id = f.away_team_id
			LEFT JOIN match_odds mo ON mo.fixture_id = f.id
			WHERE f.game_id = $1 AND NOT f.finished
			  AND f.gw >= (SELECT gw FROM next_gw)
			  AND f.gw <  (SELECT gw FROM next_gw) + $2
			UNION ALL
			SELECT
				f.away_team_id AS team_id, f.gw,
				ht.short_name AS opp, 'A' AS ha,
				f.away_difficulty AS difficulty, f.kickoff_time,
				mo.lambda_away AS xg, mo.away_cs_pct AS cs_pct
			FROM fixtures f
			JOIN teams ht ON ht.id = f.home_team_id
			LEFT JOIN match_odds mo ON mo.fixture_id = f.id
			WHERE f.game_id = $1 AND NOT f.finished
			  AND f.gw >= (SELECT gw FROM next_gw)
			  AND f.gw <  (SELECT gw FROM next_gw) + $2
		),
		team_agg AS (
			SELECT
				team_id,
				CASE WHEN COUNT(xg) > 0 THEN SUM(xg) ELSE NULL END AS xg_sum,
				CASE WHEN COUNT(cs_pct) > 0 THEN AVG(cs_pct) ELSE NULL END AS cs_avg
			FROM team_fixtures
			GROUP BY team_id
		),
		team_fix AS (
			SELECT
				t.id AS team_id,
				COALESCE(
					json_agg(
						json_build_object(
							'gw', tf.gw, 'opp', tf.opp, 'ha', tf.ha,
							'difficulty', tf.difficulty, 'kickoff', tf.kickoff_time,
							'xg', tf.xg, 'cs_pct', tf.cs_pct
						) ORDER BY tf.gw, tf.kickoff_time
					) FILTER (WHERE tf.team_id IS NOT NULL),
					'[]'::json
				) AS fixtures
			FROM teams t
			LEFT JOIN team_fixtures tf ON tf.team_id = t.id
			WHERE t.game_id = $1
			GROUP BY t.id
		)
		SELECT
			t.id, t.game_id, t.name, t.short_name,
			t.att_form, t.def_form, t.ovr_form,
			tf.fixtures,
			ta.xg_sum, ta.cs_avg
		FROM teams t
		JOIN team_fix tf ON tf.team_id = t.id
		LEFT JOIN team_agg ta ON ta.team_id = t.id
		WHERE t.game_id = $1
		ORDER BY %s
	`, orderBy)

	rows, err := s.db.Query(ctx, q, gameID, window)
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
			&r.XGSum, &r.CSAvg,
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

// MatchOddsRow is the read model for a single match_odds record.
type MatchOddsRow struct {
	OddsMatchID string    `json:"odds_match_id"`
	GameID      string    `json:"game_id"`
	FixtureID   string    `json:"fixture_id,omitempty"`
	GW          int       `json:"gw,omitempty"` // 0 when not linked to a fixture
	HomeTeam    string    `json:"home_team"`
	AwayTeam    string    `json:"away_team"`
	LambdaHome  float64   `json:"lambda_home"`
	LambdaAway  float64   `json:"lambda_away"`
	HomeCSPct   float64   `json:"home_cs_pct"`
	AwayCSPct   float64   `json:"away_cs_pct"`
	KickoffTime time.Time `json:"kickoff_time"`
	FetchedAt   time.Time `json:"fetched_at"`
}

// UpsertMatchOdds persists computed odds rows and refreshes the Redis cache
// under "{gameID}:odds:computed".
func (s *Store) UpsertMatchOdds(ctx context.Context, rows []MatchOddsRow, cache *Cache, ttl time.Duration) error {
	for _, r := range rows {
		fixtureID := (*string)(nil)
		if r.FixtureID != "" {
			fixtureID = &r.FixtureID
		}
		_, err := s.db.Exec(ctx, `
			INSERT INTO match_odds
				(odds_match_id, game_id, fixture_id, home_team, away_team,
				 lambda_home, lambda_away, home_cs_pct, away_cs_pct, kickoff_time, fetched_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (odds_match_id) DO UPDATE SET
				fixture_id   = EXCLUDED.fixture_id,
				home_team    = EXCLUDED.home_team,
				away_team    = EXCLUDED.away_team,
				lambda_home  = EXCLUDED.lambda_home,
				lambda_away  = EXCLUDED.lambda_away,
				home_cs_pct  = EXCLUDED.home_cs_pct,
				away_cs_pct  = EXCLUDED.away_cs_pct,
				kickoff_time = EXCLUDED.kickoff_time,
				fetched_at   = EXCLUDED.fetched_at
		`, r.OddsMatchID, r.GameID, fixtureID, r.HomeTeam, r.AwayTeam,
			r.LambdaHome, r.LambdaAway, r.HomeCSPct, r.AwayCSPct, r.KickoffTime, r.FetchedAt)
		if err != nil {
			return fmt.Errorf("upsert match_odds %s: %w", r.OddsMatchID, err)
		}
	}

	if cache != nil {
		b, err := json.Marshal(rows)
		if err == nil {
			_ = cache.Set(ctx, CacheKey(rows[0].GameID, "odds:computed"), b, ttl)
		}
	}
	return nil
}

// DeleteMatchOdds removes all odds rows for the given game.
func (s *Store) DeleteMatchOdds(ctx context.Context, gameID string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM match_odds WHERE game_id = $1`, gameID)
	if err != nil {
		return fmt.Errorf("delete match_odds (%s): %w", gameID, err)
	}
	return nil
}

// QueryMatchOdds returns match_odds rows for a game filtered to the given
// gameweeks. The full unfiltered set is cached in Redis under
// "{gameID}:odds:computed"; GW filtering is applied in Go after the cache hit.
// Pass gws=nil to return all gameweeks.
func (s *Store) QueryMatchOdds(ctx context.Context, gameID string, gws []int, cache *Cache) ([]MatchOddsRow, error) {
	var all []MatchOddsRow

	if cache != nil {
		if b, err := cache.Get(ctx, CacheKey(gameID, "odds:computed")); err == nil && b != nil {
			if err := json.Unmarshal(b, &all); err == nil {
				return filterByGW(all, gws), nil
			}
		}
	}

	pgRows, err := s.db.Query(ctx, `
		SELECT mo.odds_match_id, mo.game_id, COALESCE(mo.fixture_id::text, ''),
		       COALESCE(f.gw, 0),
		       mo.home_team, mo.away_team, mo.lambda_home, mo.lambda_away,
		       mo.home_cs_pct, mo.away_cs_pct, mo.kickoff_time, mo.fetched_at
		FROM match_odds mo
		LEFT JOIN fixtures f ON f.id = mo.fixture_id
		WHERE mo.game_id = $1
		ORDER BY mo.kickoff_time
	`, gameID)
	if err != nil {
		return nil, fmt.Errorf("query match_odds: %w", err)
	}
	defer pgRows.Close()

	for pgRows.Next() {
		var r MatchOddsRow
		if err := pgRows.Scan(
			&r.OddsMatchID, &r.GameID, &r.FixtureID, &r.GW,
			&r.HomeTeam, &r.AwayTeam, &r.LambdaHome, &r.LambdaAway,
			&r.HomeCSPct, &r.AwayCSPct, &r.KickoffTime, &r.FetchedAt,
		); err != nil {
			return nil, fmt.Errorf("scan match_odds: %w", err)
		}
		all = append(all, r)
	}
	if err := pgRows.Err(); err != nil {
		return nil, err
	}

	if cache != nil {
		if b, err := json.Marshal(all); err == nil {
			_ = cache.Set(ctx, CacheKey(gameID, "odds:computed"), b, 0)
		}
	}

	return filterByGW(all, gws), nil
}

// PlayerIDsByExternalIDs maps a game's external player IDs (as strings) to their
// internal UUIDs. External IDs with no matching player are omitted from the
// result. Used to resolve a manager's picks (which reference external IDs) to
// our player records.
func (s *Store) PlayerIDsByExternalIDs(ctx context.Context, gameID string, externalIDs []string) (map[string]string, error) {
	out := make(map[string]string, len(externalIDs))
	if len(externalIDs) == 0 {
		return out, nil
	}
	ids := make([]int, 0, len(externalIDs))
	for _, s := range externalIDs {
		n, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("PlayerIDsByExternalIDs: invalid external id %q: %w", s, err)
		}
		ids = append(ids, n)
	}
	rows, err := s.db.Query(ctx,
		`SELECT external_id, id FROM players WHERE game_id = $1 AND external_id = ANY($2)`,
		gameID, ids,
	)
	if err != nil {
		return nil, fmt.Errorf("PlayerIDsByExternalIDs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var extID int
		var id string
		if err := rows.Scan(&extID, &id); err != nil {
			return nil, fmt.Errorf("PlayerIDsByExternalIDs scan: %w", err)
		}
		out[strconv.Itoa(extID)] = id
	}
	return out, rows.Err()
}

func filterByGW(rows []MatchOddsRow, gws []int) []MatchOddsRow {
	if len(gws) == 0 {
		return rows
	}
	set := make(map[int]bool, len(gws))
	for _, gw := range gws {
		set[gw] = true
	}
	out := rows[:0:0]
	for _, r := range rows {
		if set[r.GW] {
			out = append(out, r)
		}
	}
	return out
}
