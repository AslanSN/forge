-- forge · paso-00 · migration 001
-- The accounts table, written by hand. No ORM, no auto-migration on startup.
-- Applied in order by scripts/migrate.sh. SQL is idempotent (IF NOT EXISTS).
--
-- Two deliberate choices to notice:
--   * balance is numeric(18,2), never float — paso-01 proves why.
--   * balance is a MUTABLE column — deliberately naive; paso-02 (the double-spend
--     race) will break it and force an append-only `entries` table.

CREATE EXTENSION IF NOT EXISTS pgcrypto;   -- provides gen_random_uuid()

CREATE TABLE IF NOT EXISTS accounts (
    id         uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
    name       text          NOT NULL,
    balance    numeric(18,2) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at timestamptz   NOT NULL DEFAULT now()
);
