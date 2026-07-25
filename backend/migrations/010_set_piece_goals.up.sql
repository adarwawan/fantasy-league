-- Track goal conversion for observed set-piece shots.
ALTER TABLE sp_events ADD COLUMN is_goal BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE sp_board  ADD COLUMN goals   INT     NOT NULL DEFAULT 0;
