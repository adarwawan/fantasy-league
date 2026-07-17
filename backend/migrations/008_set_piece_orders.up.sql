-- Team-internal taker ranks for set-piece duties, from FPL bootstrap-static.
-- NULL means the player holds no duty of that kind.
ALTER TABLE players
    ADD COLUMN penalties_order        INT,
    ADD COLUMN direct_freekicks_order INT,
    ADD COLUMN corners_indirect_freekicks_order INT;
