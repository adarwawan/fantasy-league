-- Isolated observed shot-zone detector, the open-play companion to the
-- set-piece board (docs: "Shot zone rankings — requirements & solution").
-- No FK to players/teams; player identity is keyed by Understat id.

-- Raw qualifying open-play shots, one row per shot.
CREATE TABLE sz_events (
    id           BIGSERIAL PRIMARY KEY,
    match_id     TEXT NOT NULL,
    season       TEXT NOT NULL,
    match_date   DATE NOT NULL,
    minute       INT  NOT NULL,
    team_for     TEXT NOT NULL,       -- canonical PL team of the shooter
    team_against TEXT NOT NULL,       -- canonical PL team defending
    zone         TEXT NOT NULL,       -- 'center' | 'left' | 'right' | 'outside'
    player_id    TEXT NOT NULL,       -- Understat player_id (the shooter)
    player_name  TEXT NOT NULL,
    is_goal      BOOLEAN NOT NULL DEFAULT false,
    xg           NUMERIC NOT NULL DEFAULT 0,
    UNIQUE (match_id, player_id, minute, zone)
);
CREATE INDEX ON sz_events (team_for, match_date DESC);
CREATE INDEX ON sz_events (team_against, match_date DESC);

-- Materialised "shots taken" ranking: a team's own players by zone over the
-- rolling window. zone='overall' carries the cross-zone aggregate per player.
CREATE TABLE sz_taken_board (
    team        TEXT NOT NULL,
    zone        TEXT NOT NULL,       -- 'center'|'left'|'right'|'outside'|'overall'
    player_id   TEXT NOT NULL,
    player_name TEXT NOT NULL,
    rank        INT  NOT NULL,       -- 1 = most shots (within team/zone)
    shots       INT  NOT NULL,
    goals       INT  NOT NULL,
    xg          NUMERIC NOT NULL DEFAULT 0,
    last_seen   DATE,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (team, zone, player_id)
);

-- Materialised "shots conceded" totals: a team's own aggregate per zone over
-- its own rolling window — a weakness signal, no player breakdown.
CREATE TABLE sz_conceded_board (
    team       TEXT NOT NULL,
    zone       TEXT NOT NULL,        -- 'center'|'left'|'right'|'outside'|'overall'
    shots      INT NOT NULL,
    goals      INT NOT NULL,
    xg         NUMERIC NOT NULL DEFAULT 0,
    last_seen  DATE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (team, zone)
);
