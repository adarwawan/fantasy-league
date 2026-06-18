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
			INSERT INTO fixtures (game_id, external_id, gw, home_team_id, away_team_id, home_difficulty, away_difficulty, kickoff_time, finished)
			VALUES (
				$1, $2, $3,
				(SELECT id FROM teams WHERE game_id = $1 AND external_id = $4),
				(SELECT id FROM teams WHERE game_id = $1 AND external_id = $5),
				$6, $7, $8, $9
			)
			ON CONFLICT (game_id, external_id) DO UPDATE SET
				gw              = EXCLUDED.gw,
				home_team_id    = EXCLUDED.home_team_id,
				away_team_id    = EXCLUDED.away_team_id,
				home_difficulty = EXCLUDED.home_difficulty,
				away_difficulty = EXCLUDED.away_difficulty,
				kickoff_time    = EXCLUDED.kickoff_time,
				finished        = EXCLUDED.finished
		`, f.GameID, f.ExternalID, f.GW, homeExtID, awayExtID, f.HomeDifficulty, f.AwayDifficulty, f.KickoffTime, f.Finished)
		if err != nil {
			return fmt.Errorf("upsert fixture %d: %w", f.ExternalID, err)
		}
	}
	return nil
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
