#!/usr/bin/env bash
# Open an interactive psql shell against the local forge database.
set -euo pipefail
DB_URL="${FORGE_DB_URL:-postgres://forge:forge@localhost:5432/forge}"
exec psql "$DB_URL"
