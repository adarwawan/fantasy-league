CREATE TABLE player_top_n_ownerships (
    player_id   UUID        NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    top_n       INT         NOT NULL,
    ownership   NUMERIC(5,2) NOT NULL DEFAULT 0,
    PRIMARY KEY (player_id, top_n)
);

CREATE INDEX idx_player_topn_own ON player_top_n_ownerships (player_id, top_n);
