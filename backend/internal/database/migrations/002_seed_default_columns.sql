-- 002: Ensure the default board lanes exist.
--
-- Migration 001 also seeds them, but databases created before the seed was
-- added to 001 never got the rows (migrations are tracked by filename and are
-- forward-only, so editing 001 in place does not re-run it). This migration is
-- idempotent, so it is safe on both fresh and existing databases.

INSERT INTO columns (name, position) VALUES
    ('Backlog', 0),
    ('To Do',   1),
    ('In Dev',  2),
    ('Review',  3),
    ('Done',    4)
ON CONFLICT (name) DO NOTHING;
