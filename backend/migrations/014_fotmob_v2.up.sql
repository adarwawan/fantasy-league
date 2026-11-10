-- v2 of the set-piece and shot-zone stores, sourced from FotMob (Understat is in
-- maintenance). Same shape as sp_* / sz_* so the aggregation and read handlers
-- are shared; kept as separate tables so v1 keeps serving until cutover and
-- Understat ids never mix with FotMob ids.
-- Differences from v1: events are keyed by FotMob's per-shot id (v1's
-- (player, minute, zone) key silently merged two shots in the same minute), and
-- player_id / match_id are FotMob ids. `understat_team` keeps its v1 name so the
-- shared queries work; it holds the canonical PL team name.

CREATE TABLE sp2_events (
    id             BIGSERIAL PRIMARY KEY,
    match_id       TEXT NOT NULL,
    shot_id        TEXT NOT NULL,
    season         TEXT NOT NULL,
    match_date     DATE NOT NULL,
    minute         INT  NOT NULL,
    understat_team TEXT NOT NULL,
    role           TEXT NOT NULL,        -- 'taker' | 'target'
    duty           TEXT NOT NULL,        -- taker: 'penalty'|'dfk'  ·  target: 'corner'|'setpiece'
    player_id      TEXT NOT NULL,
    player_name    TEXT NOT NULL,
    is_header      BOOLEAN NOT NULL DEFAULT false,
    is_goal        BOOLEAN NOT NULL DEFAULT false,
    xg             NUMERIC NOT NULL DEFAULT 0,
    UNIQUE (match_id, shot_id)
);
CREATE INDEX ON sp2_events (understat_team, role, match_date DESC);

CREATE TABLE sp2_board (
    understat_team TEXT NOT NULL,
    role           TEXT NOT NULL,
    duty           TEXT NOT NULL,        -- takers: 'penalty'|'dfk'  ·  target: 'all'|'corner'|'setpiece'
    player_id      TEXT NOT NULL,
    player_name    TEXT NOT NULL,
    rank           INT  NOT NULL,
    weighted_score NUMERIC NOT NULL,
    raw_count      INT NOT NULL,
    goals          INT NOT NULL DEFAULT 0,
    xg_sum         NUMERIC NOT NULL DEFAULT 0,
    header_pct     NUMERIC,
    last_seen      DATE,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (understat_team, role, duty, player_id)
);

CREATE TABLE sz2_events (
    id           BIGSERIAL PRIMARY KEY,
    match_id     TEXT NOT NULL,
    shot_id      TEXT NOT NULL,
    season       TEXT NOT NULL,
    match_date   DATE NOT NULL,
    minute       INT  NOT NULL,
    team_for     TEXT NOT NULL,
    team_against TEXT NOT NULL,
    zone         TEXT NOT NULL,          -- 'center' | 'left' | 'right' | 'outside'
    player_id    TEXT NOT NULL,
    player_name  TEXT NOT NULL,
    is_goal      BOOLEAN NOT NULL DEFAULT false,
    xg           NUMERIC NOT NULL DEFAULT 0,
    UNIQUE (match_id, shot_id)
);
CREATE INDEX ON sz2_events (team_for, match_date DESC);
CREATE INDEX ON sz2_events (team_against, match_date DESC);

CREATE TABLE sz2_taken_board (
    team        TEXT NOT NULL,
    zone        TEXT NOT NULL,
    player_id   TEXT NOT NULL,
    player_name TEXT NOT NULL,
    rank        INT  NOT NULL,
    shots       INT  NOT NULL,
    goals       INT  NOT NULL,
    xg          NUMERIC NOT NULL DEFAULT 0,
    last_seen   DATE,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (team, zone, player_id)
);

CREATE TABLE sz2_conceded_board (
    team       TEXT NOT NULL,
    zone       TEXT NOT NULL,
    shots      INT NOT NULL,
    goals      INT NOT NULL,
    xg         NUMERIC NOT NULL DEFAULT 0,
    last_seen  DATE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (team, zone)
);
