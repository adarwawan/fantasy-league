package store

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"fantasy-league/internal/fantasy"
)

type Store struct {
	db *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return &Store{db: pool}, nil
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) UpsertTeams(ctx context.Context, teams []fantasy.Team) error {
	for _, t := range teams {
		_, err := s.db.Exec(ctx, `
			INSERT INTO teams (game_id, external_id, name, short_name, att_form, def_form, ovr_form, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (game_id, external_id) DO UPDATE SET
				name       = EXCLUDED.name,
				short_name = EXCLUDED.short_name,
				att_form   = EXCLUDED.att_form,
				def_form   = EXCLUDED.def_form,
				ovr_form   = EXCLUDED.ovr_form,
				updated_at = EXCLUDED.updated_at
		`, t.GameID, t.ExternalID, t.Name, t.ShortName, t.AttForm, t.DefForm, t.OvrForm, t.UpdatedAt)
		if err != nil {
			return fmt.Errorf("upsert team %d: %w", t.ExternalID, err)
		}
	}
	return nil
}

// UpsertPlayers upserts players. Player.TeamID is expected to be the external team ID as a string;
// this method resolves it to the internal UUID via a subquery.
func (s *Store) UpsertPlayers(ctx context.Context, players []fantasy.Player) error {
	for _, p := range players {
		extTeamID, err := strconv.Atoi(p.TeamID)
		if err != nil {
			return fmt.Errorf("player %d: invalid team id %q: %w", p.ExternalID, p.TeamID, err)
		}
		_, err = s.db.Exec(ctx, `
			INSERT INTO players (game_id, external_id, name, team_id, position, price, form, global_ownership, top_n_ownership, top_n_size, status, news, updated_at)
			VALUES (
				$1, $2, $3,
				(SELECT id FROM teams WHERE game_id = $1 AND external_id = $4),
				$5, $6, $7, $8, $9, $10, $11, $12, $13
			)
			ON CONFLICT (game_id, external_id) DO UPDATE SET
				name             = EXCLUDED.name,
				team_id          = EXCLUDED.team_id,
				position         = EXCLUDED.position,
				price            = EXCLUDED.price,
				form             = EXCLUDED.form,
				global_ownership = EXCLUDED.global_ownership,
				status           = EXCLUDED.status,
				news             = EXCLUDED.news,
				updated_at       = EXCLUDED.updated_at
		`, p.GameID, p.ExternalID, p.Name, extTeamID, p.Position, p.Price, p.Form, p.GlobalOwnership, p.TopNOwnership, p.TopNSize, p.Status, p.News, p.UpdatedAt)
		if err != nil {
			return fmt.Errorf("upsert player %d: %w", p.ExternalID, err)
		}
	}
	return nil
}

// UpsertFixtures upserts fixtures. HomeTeamID/AwayTeamID are external team IDs as strings.
func (s *Store) UpsertFixtures(ctx context.Context, fixtures []fantasy.Fixture) error {
	for _, f := range fixtures {
		homeExtID, err := strconv.Atoi(f.HomeTeamID)
		if err != nil {
			return fmt.Errorf("fixture %d: invalid home team id: %w", f.ExternalID, err)
		}
		awayExtID, err := strconv.Atoi(f.AwayTeamID)
		if err != nil {
			return fmt.Errorf("fixture %d: invalid away team id: %w", f.ExternalID, err)
		}
		_, err = s.db.Exec(ctx, `
			INSERT INTO fixtures (game_id, external_id, gw, home_team_id, away_team_id, home_difficulty, away_difficulty, kickoff_time, finished, home_score, away_score)
			VALUES (
				$1, $2, $3,
				(SELECT id FROM teams WHERE game_id = $1 AND external_id = $4),
				(SELECT id FROM teams WHERE game_id = $1 AND external_id = $5),
				$6, $7, $8, $9, $10, $11
			)
			ON CONFLICT (game_id, external_id) DO UPDATE SET
				gw              = EXCLUDED.gw,
				home_team_id    = EXCLUDED.home_team_id,
				away_team_id    = EXCLUDED.away_team_id,
				home_difficulty = EXCLUDED.home_difficulty,
				away_difficulty = EXCLUDED.away_difficulty,
				kickoff_time    = EXCLUDED.kickoff_time,
				finished        = EXCLUDED.finished,
				home_score      = EXCLUDED.home_score,
				away_score      = EXCLUDED.away_score
		`, f.GameID, f.ExternalID, f.GW, homeExtID, awayExtID, f.HomeDifficulty, f.AwayDifficulty, f.KickoffTime, f.Finished, f.HomeScore, f.AwayScore)
		if err != nil {
			return fmt.Errorf("upsert fixture %d: %w", f.ExternalID, err)
		}
	}
	return nil
}

// ResetManagerRanks nullifies overall_rank for all managers in a game so that
// managers who fell out of the top-N in the current sync are excluded from
// RecomputeTopNOwnership (NULL <= N is false in SQL).
func (s *Store) ResetManagerRanks(ctx context.Context, gameID string) error {
	_, err := s.db.Exec(ctx, `UPDATE managers SET overall_rank = NULL WHERE game_id = $1`, gameID)
	return err
}

func (s *Store) UpsertManagers(ctx context.Context, managers []fantasy.Manager) error {
	for _, m := range managers {
		_, err := s.db.Exec(ctx, `
			INSERT INTO managers (game_id, external_id, name, overall_rank, team_value, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (game_id, external_id) DO UPDATE SET
				name         = EXCLUDED.name,
				overall_rank = EXCLUDED.overall_rank,
				team_value   = EXCLUDED.team_value,
				updated_at   = EXCLUDED.updated_at
		`, m.GameID, m.ExternalID, m.Name, m.OverallRank, m.TeamValue, m.UpdatedAt)
		if err != nil {
			return fmt.Errorf("upsert manager %d: %w", m.ExternalID, err)
		}
	}
	return nil
}

// UpsertPicks upserts manager picks. ManagerID and PlayerID are external IDs as strings;
// this method resolves them to internal UUIDs.
func (s *Store) UpsertPicks(ctx context.Context, picks []fantasy.ManagerPick) error {
	for _, p := range picks {
		managerExtID, err := strconv.Atoi(p.ManagerID)
		if err != nil {
			return fmt.Errorf("pick: invalid manager id %q: %w", p.ManagerID, err)
		}
		playerExtID, err := strconv.Atoi(p.PlayerID)
		if err != nil {
			return fmt.Errorf("pick: invalid player id %q: %w", p.PlayerID, err)
		}
		_, err = s.db.Exec(ctx, `
			INSERT INTO manager_picks (manager_id, player_id, game_id, gw, is_captain, is_vice_captain, multiplier)
			VALUES (
				(SELECT id FROM managers WHERE game_id = $1 AND external_id = $2),
				(SELECT id FROM players  WHERE game_id = $1 AND external_id = $3),
				$1, $4, $5, $6, $7
			)
			ON CONFLICT (manager_id, player_id, gw) DO UPDATE SET
				is_captain      = EXCLUDED.is_captain,
				is_vice_captain = EXCLUDED.is_vice_captain,
				multiplier      = EXCLUDED.multiplier
		`, p.GameID, managerExtID, playerExtID, p.GW, p.IsCaptain, p.IsViceCaptain, p.Multiplier)
		if err != nil {
			return fmt.Errorf("upsert pick manager=%s player=%s: %w", p.ManagerID, p.PlayerID, err)
		}
	}
	return nil
}

// DeleteTestGame removes all rows for a game_id. Only for use in tests.
func (s *Store) DeleteTestGame(ctx context.Context, gameID string) {
	s.db.Exec(ctx, `DELETE FROM manager_picks WHERE game_id = $1`, gameID)
	s.db.Exec(ctx, `DELETE FROM players WHERE game_id = $1`, gameID)
	s.db.Exec(ctx, `DELETE FROM fixtures WHERE game_id = $1`, gameID)
	s.db.Exec(ctx, `DELETE FROM managers WHERE game_id = $1`, gameID)
	s.db.Exec(ctx, `DELETE FROM teams WHERE game_id = $1`, gameID)
}

// RecomputeTeamForm recalculates att_form, def_form, and ovr_form for every team
// in a game by averaging results across the last gwWindow finished gameweeks.
// The divisor is the count of finished fixtures (not gwWindow), so unplayed GWs
// don't dilute the average.
func (s *Store) RecomputeTeamForm(ctx context.Context, gameID string, gwWindow int) error {
	_, err := s.db.Exec(ctx, `
		WITH last_gws AS (
			SELECT DISTINCT gw
			FROM fixtures
			WHERE game_id = $1
			  AND finished = true
			ORDER BY gw DESC
			LIMIT $2
		),
		home_results AS (
			SELECT
				home_team_id                          AS team_id,
				SUM(home_score)::float                AS goals_for,
				SUM(away_score)::float                AS goals_against,
				SUM(CASE
					WHEN home_score > away_score THEN 3
					WHEN home_score = away_score THEN 1
					ELSE 0
				END)::float                           AS points,
				COUNT(*)::float                       AS played
			FROM fixtures
			WHERE game_id = $1
			  AND finished = true
			  AND home_score IS NOT NULL
			  AND away_score IS NOT NULL
			  AND gw IN (SELECT gw FROM last_gws)
			GROUP BY home_team_id
		),
		away_results AS (
			SELECT
				away_team_id                          AS team_id,
				SUM(away_score)::float                AS goals_for,
				SUM(home_score)::float                AS goals_against,
				SUM(CASE
					WHEN away_score > home_score THEN 3
					WHEN away_score = home_score THEN 1
					ELSE 0
				END)::float                           AS points,
				COUNT(*)::float                       AS played
			FROM fixtures
			WHERE game_id = $1
			  AND finished = true
			  AND home_score IS NOT NULL
			  AND away_score IS NOT NULL
			  AND gw IN (SELECT gw FROM last_gws)
			GROUP BY away_team_id
		),
		combined AS (
			SELECT
				team_id,
				SUM(goals_for)     AS total_goals_for,
				SUM(goals_against) AS total_goals_against,
				SUM(points)        AS total_points,
				SUM(played)        AS total_played
			FROM (
				SELECT * FROM home_results
				UNION ALL
				SELECT * FROM away_results
			) r
			GROUP BY team_id
		)
		UPDATE teams t
		SET
			att_form = ROUND((c.total_goals_for     / NULLIF(c.total_played, 0))::numeric, 2),
			def_form = ROUND((c.total_goals_against / NULLIF(c.total_played, 0))::numeric, 2),
			ovr_form = ROUND((c.total_points        / NULLIF(c.total_played, 0))::numeric, 2)
		FROM combined c
		WHERE t.id      = c.team_id
		  AND t.game_id = $1
	`, gameID, gwWindow)
	return err
}

func (s *Store) RecomputeTopNOwnership(ctx context.Context, gameID string, topN int, gw int) error {
	_, err := s.db.Exec(ctx, `
		WITH top_managers AS (
			SELECT id FROM managers
			WHERE game_id    = $1
			  AND overall_rank <= $2
		),
		pick_counts AS (
			SELECT player_id, COUNT(*) AS owned_by
			FROM manager_picks
			WHERE game_id    = $1
			  AND gw         = $3
			  AND manager_id IN (SELECT id FROM top_managers)
			GROUP BY player_id
		)
		UPDATE players p
		SET    top_n_ownership = ROUND(pc.owned_by::numeric / $2 * 100, 2),
		       top_n_size      = $2
		FROM   pick_counts pc
		WHERE  p.id      = pc.player_id
		  AND  p.game_id = $1
	`, gameID, topN, gw)
	return err
}
