#!/usr/bin/env bash
# Open an interactive psql shell against the local forge database.
set -euo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

forge_psql_shell "$@"
