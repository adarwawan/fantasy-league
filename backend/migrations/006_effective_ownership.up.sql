-- Effective ownership weights each pick by its multiplier (bench = 0, starter = 1,
-- captain = 2, triple captain = 3), so it can exceed 100% when many managers
-- captain the same player. Plain top-N ownership counts squad membership only.
ALTER TABLE player_top_n_ownerships
    ADD COLUMN effective_ownership NUMERIC(6,2) NOT NULL DEFAULT 0;
