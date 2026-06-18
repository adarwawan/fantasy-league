CREATE TABLE teams (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id       TEXT NOT NULL,
    external_id   INT  NOT NULL,
    name          TEXT,
    short_name    TEXT,
    att_form      NUMERIC(5,2),
    def_form      NUMERIC(5,2),
    ovr_form      NUMERIC(5,2),
    updated_at    TIMESTAMPTZ,
    UNIQUE (game_id, external_id)
);

CREATE TABLE players (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id           TEXT        NOT NULL,
    external_id       INT         NOT NULL,
    name              TEXT        NOT NULL,
    team_id           UUID        REFERENCES teams(id),
    position          TEXT        NOT NULL,
    price             NUMERIC(5,1),
    form              NUMERIC(5,2),
    global_ownership  NUMERIC(5,2),
    top_n_ownership   NUMERIC(5,2),
    top_n_size        INT,
    status            TEXT,
    news              TEXT,
    updated_at        TIMESTAMPTZ,
    UNIQUE (game_id, external_id)
);

CREATE INDEX idx_players_game_pos        ON players (game_id, position);
CREATE INDEX idx_players_game_global_own ON players (game_id, global_ownership DESC);
CREATE INDEX idx_players_game_topn_own   ON players (game_id, top_n_ownership DESC);
CREATE INDEX idx_players_game_form       ON players (game_id, form DESC);

CREATE TABLE fixtures (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id         TEXT NOT NULL,
    external_id     INT  NOT NULL,
    gw              INT  NOT NULL,
    home_team_id    UUID REFERENCES teams(id),
    away_team_id    UUID REFERENCES teams(id),
    home_difficulty INT  CHECK (home_difficulty BETWEEN 1 AND 5),
    away_difficulty INT  CHECK (away_difficulty BETWEEN 1 AND 5),
    kickoff_time    TIMESTAMPTZ,
    finished        BOOL DEFAULT false,
    UNIQUE (game_id, external_id)
);

CREATE INDEX idx_fixtures_game_gw ON fixtures (game_id, gw);

CREATE TABLE managers (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id       TEXT NOT NULL,
    external_id   INT  NOT NULL,
    name          TEXT,
    overall_rank  INT,
    team_value    NUMERIC(5,1),
    updated_at    TIMESTAMPTZ,
    UNIQUE (game_id, external_id)
);

CREATE INDEX idx_managers_game_rank ON managers (game_id, overall_rank ASC);

CREATE TABLE manager_picks (
    manager_id       UUID REFERENCES managers(id),
    player_id        UUID REFERENCES players(id),
    game_id          TEXT NOT NULL,
    gw               INT  NOT NULL,
    is_captain       BOOL DEFAULT false,
    is_vice_captain  BOOL DEFAULT false,
    multiplier       INT  DEFAULT 1,
    PRIMARY KEY (manager_id, player_id, gw)
);

CREATE INDEX idx_picks_game_gw_player ON manager_picks (game_id, gw, player_id);

CREATE TABLE player_gw_stats (
    player_id   UUID REFERENCES players(id),
    game_id     TEXT NOT NULL,
    gw          INT  NOT NULL,
    minutes     INT,
    points      INT,
    goals       INT,
    assists     INT,
    bonus       INT,
    PRIMARY KEY (player_id, gw)
);
