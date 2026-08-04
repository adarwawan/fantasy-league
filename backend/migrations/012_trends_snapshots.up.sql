-- Trends: FPL transfer-velocity snapshots captured near a deadline.
--
-- trends_snapshot is LIST-partitioned by gameweek so retention ("keep last 2
-- GW") is a partition DROP — instant, no DELETE bloat, reclaims disk on Neon
-- immediately. Child partitions are created/dropped dynamically by cmd/trends
-- on session arm (CREATE gw N, DROP gw N-2); this migration only defines the
-- parent + indexes, which partitions inherit automatically.
CREATE TABLE trends_snapshot (
    id            BIGSERIAL,
    captured_at   TIMESTAMPTZ NOT NULL,
    gameweek      INT         NOT NULL,
    player_ext_id INT         NOT NULL,           -- FPL element id
    selected_bp   INT         NOT NULL,           -- selected_by_percent in basis points (1234 = 12.34%)
    transfers_in  INT         NOT NULL,           -- transfers_in_event  (cumulative this GW)
    transfers_out INT         NOT NULL,           -- transfers_out_event (cumulative this GW)
    now_cost      INT         NOT NULL,           -- price in tenths (55 = £5.5m)
    PRIMARY KEY (gameweek, id)                     -- partition key must be in the PK
) PARTITION BY LIST (gameweek);

-- Defined on the parent → inherited by every partition. The per-player series
-- and the leaders "closest snapshot to cutoff" lookups both scan by this.
CREATE INDEX trends_snapshot_player_time_idx
    ON trends_snapshot (player_ext_id, captured_at);

-- The active-window control. One armed session per gameweek; the poller no-ops
-- unless a row here is active and inside [started_at, ends_at].
CREATE TABLE trends_session (
    id         BIGSERIAL PRIMARY KEY,
    gameweek   INT         NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ends_at    TIMESTAMPTZ NOT NULL,               -- the FPL deadline
    active     BOOLEAN     NOT NULL DEFAULT true,
    poll_count INT         NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX trends_session_gw_idx ON trends_session (gameweek);
