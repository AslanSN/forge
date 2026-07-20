#!/usr/bin/env bash
# Apply every db/migrations/*.sql in order. The SQL is idempotent (IF NOT EXISTS).
set -euo pipefail

DB_URL="${FORGE_DB_URL:-postgres://forge:forge@localhost:5432/forge}"
cd "$(dirname "$0")/.."

shopt -s nullglob
for f in db/migrations/*.sql; do
  echo "→ applying $f"
  psql "$DB_URL" -v ON_ERROR_STOP=1 -f "$f"
done
echo "✓ migrations applied"
