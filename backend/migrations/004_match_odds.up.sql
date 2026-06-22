CREATE TABLE match_odds (
    odds_match_id  TEXT        PRIMARY KEY,
    game_id        TEXT        NOT NULL,
    fixture_id     UUID        REFERENCES fixtures(id) ON DELETE SET NULL,
    home_team      TEXT        NOT NULL,
    away_team      TEXT        NOT NULL,
    lambda_home    NUMERIC(6,4) NOT NULL,
    lambda_away    NUMERIC(6,4) NOT NULL,
    home_cs_pct    NUMERIC(5,2) NOT NULL,
    away_cs_pct    NUMERIC(5,2) NOT NULL,
    kickoff_time   TIMESTAMPTZ,
    fetched_at     TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_match_odds_game_id ON match_odds (game_id);
