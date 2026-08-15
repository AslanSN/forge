#!/usr/bin/env bash
# Shared helpers. Sourced by the other scripts — not meant to be run directly.
#
# Why this exists: paso-00 talks to Postgres with psql, but the Postgres it talks
# to lives in a container. Rather than require a host-side client just to get a
# binary, prefer the psql that already ships *inside* the db image — it is always
# version-matched to the server. A host psql, if present, wins (it's faster and
# gives you readline history).

cd "$(dirname "${BASH_SOURCE[0]}")/.."   # repo root: docker compose needs it

FORGE_DB_URL="${FORGE_DB_URL:-postgres://forge:forge@localhost:5432/forge}"

_forge_host_psql() { command -v psql >/dev/null 2>&1; }

_forge_db_running() {
  [ -n "$(docker compose ps -q db 2>/dev/null)" ] &&
    [ "$(docker inspect -f '{{.State.Running}}' forge-db 2>/dev/null)" = "true" ]
}

_forge_require_psql() {
  _forge_host_psql && return 0
  if ! command -v docker >/dev/null 2>&1; then
    echo "Need either a host psql or Docker. Neither found." >&2
    return 1
  fi
  if ! _forge_db_running; then
    echo "No host psql, and the db container isn't running — run 'make up' first." >&2
    return 1
  fi
}

# Non-interactive psql. SQL arrives on **stdin**, never as a file path, because
# inside the container the host's paths don't exist.
#   forge_psql_stdin [psql args...] < some.sql
forge_psql_stdin() {
  _forge_require_psql || return 1
  if _forge_host_psql; then
    psql "$FORGE_DB_URL" "$@"
  else
    docker compose exec -T db psql "$FORGE_DB_URL" "$@"
  fi
}

# Interactive psql shell (keeps the TTY).
forge_psql_shell() {
  _forge_require_psql || return 1
  if _forge_host_psql; then
    exec psql "$FORGE_DB_URL" "$@"
  else
    exec docker compose exec db psql "$FORGE_DB_URL" "$@"
  fi
}
