#!/usr/bin/env bash
# Apply every db/migrations/*.sql in order. The SQL is idempotent (IF NOT EXISTS).
set -euo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

shopt -s nullglob
for f in db/migrations/*.sql; do
  echo "→ applying $f"
  forge_psql_stdin -v ON_ERROR_STOP=1 --quiet < "$f"
done
echo "✓ migrations applied"
