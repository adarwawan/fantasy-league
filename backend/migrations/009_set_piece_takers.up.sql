-- Isolated observed set-piece detector (docs/set-piece-detector.md).
-- No FK to players/teams; player identity is keyed by Understat id.

-- Raw observed set-piece shots, one row per qualifying shot.
-- `role` distinguishes the two signals; the shooter is the subject of both.
CREATE TABLE sp_events (
    id             BIGSERIAL PRIMARY KEY,
    match_id       TEXT NOT NULL,
    season         TEXT NOT NULL,
    match_date     DATE NOT NULL,
    minute         INT  NOT NULL,
    understat_team TEXT NOT NULL,
    role           TEXT NOT NULL,        -- 'taker' | 'target'
    duty           TEXT NOT NULL,        -- taker: 'penalty'|'dfk'  ·  target: 'corner'|'setpiece'
    player_id      TEXT NOT NULL,        -- Understat player_id (the shooter)
    player_name    TEXT NOT NULL,
    is_header      BOOLEAN NOT NULL DEFAULT false,
    xg             NUMERIC NOT NULL DEFAULT 0,
    UNIQUE (match_id, role, duty, player_id, minute)
);
CREATE INDEX ON sp_events (understat_team, role, match_date DESC);

-- Materialised ranking, one row per team/role/duty/player over the window.
-- Takers rank by weighted_score (weighted count); target men rank by
-- weighted_score (volume × xG blend). Target men carry duty='all' for the
-- league-wide aggregate plus per-duty rows ('corner'/'setpiece').
CREATE TABLE sp_board (
    understat_team TEXT NOT NULL,
    role           TEXT NOT NULL,        -- 'taker' | 'target'
    duty           TEXT NOT NULL,        -- takers: 'penalty'|'dfk'  ·  target: 'all'|'corner'|'setpiece'
    player_id      TEXT NOT NULL,
    player_name    TEXT NOT NULL,
    rank           INT  NOT NULL,        -- 1 = primary observed (within team/role/duty)
    weighted_score NUMERIC NOT NULL,     -- taker: weighted count · target: volume×xG blend
    raw_count      INT NOT NULL,         -- shots in window
    xg_sum         NUMERIC NOT NULL DEFAULT 0,
    header_pct     NUMERIC,              -- target-man context; NULL for takers
    last_seen      DATE,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (understat_team, role, duty, player_id)
);
