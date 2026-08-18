#!/usr/bin/env bash
# paso-00b · executable spec for the container image.
#
# This is the "test that fails first". It asserts PROPERTIES of the image you
# build by hand — it never tells you the Dockerfile. Every failure prints the
# question you should be able to answer, not the answer.
#
#   ./scripts/verify-image.sh          # or: make verify-image
#
# Exit 0 only when every property holds.

set -uo pipefail

IMAGE="${FORGE_IMAGE:-forge-api:paso-00b}"
CONTAINER_PORT=8080          # the contract: the app listens here INSIDE the container
HOST_PORT="${HOST_PORT:-15000}"
MAX_SIZE_BYTES=$((400 * 1024 * 1024))   # 400 MB — an SDK-in-final-stage image blows past this
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

pass=0
fail=0
ok()   { printf '  \033[32m✓\033[0m %s\n' "$1"; pass=$((pass + 1)); }
no()   { printf '  \033[31m✗\033[0m %s\n' "$1"; fail=$((fail + 1)); [ $# -gt 1 ] && printf '      \033[33m?\033[0m %s\n' "$2"; return 0; }
head_() { printf '\n\033[1m%s\033[0m\n' "$1"; }

command -v docker >/dev/null 2>&1 || { echo "docker not found — start it first (systemctl enable --now docker)"; exit 2; }
docker info >/dev/null 2>&1     || { echo "the docker daemon is not responding — 'sudo systemctl start docker'"; exit 2; }

# ── A · The build context ────────────────────────────────────────────────────
head_ "A · the build context"

if [ -f Dockerfile ]; then
  ok "Dockerfile exists"
else
  no "Dockerfile exists" "Nothing to verify yet — this whole spec is the enunciado. Write it."
  printf '\n%d passed, %d failed\n' "$pass" "$fail"; exit 1
fi

if [ -f .dockerignore ]; then
  ok ".dockerignore exists"
  missing=""
  for p in bin obj .git; do
    grep -qE "(^|/)\**${p}\**/?$" .dockerignore || missing="${missing} ${p}"
  done
  [ -z "$missing" ] && ok ".dockerignore excludes bin/ obj/ .git" \
                    || no ".dockerignore excludes bin/ obj/ .git (missing:${missing})" \
                          "Your local bin/obj were compiled for THIS machine. What happens when they land in the image and shadow the build?"
else
  no ".dockerignore exists" "Without it, 'COPY . .' ships your .git history and your local build output. What is in your build context right now?"
fi

# Layer ordering: the project files must be restored BEFORE the full source is copied.
first_proj_copy=$(grep -nEi '^[[:space:]]*COPY .*\.(csproj|slnx|sln)' Dockerfile | head -1 | cut -d: -f1)
restore_line=$(grep -nEi '^[[:space:]]*RUN .*dotnet restore' Dockerfile | head -1 | cut -d: -f1)
full_copy=$(grep -nEi '^[[:space:]]*COPY (\.|src|\./)' Dockerfile | grep -viE '\.(csproj|slnx|sln)' | head -1 | cut -d: -f1)
if [ -n "$first_proj_copy" ] && [ -n "$restore_line" ] && [ -n "$full_copy" ] \
   && [ "$first_proj_copy" -lt "$restore_line" ] && [ "$restore_line" -lt "$full_copy" ]; then
  ok "layers ordered: copy project files → restore → copy source"
else
  no "layers ordered: copy project files → restore → copy source" \
     "You changed one .cs file. Why did Docker re-download every NuGet package? Which layer got invalidated, and what was above it?"
fi

# ── B · The image ───────────────────────────────────────────────────────────
head_ "B · the image"

printf '  … building %s\n' "$IMAGE"
if build_log=$(docker build -t "$IMAGE" . 2>&1); then
  ok "image builds"
else
  no "image builds"
  printf '%s\n' "$build_log" | tail -25 | sed 's/^/      /'
  printf '\n%d passed, %d failed\n' "$pass" "$fail"; exit 1
fi

stages=$(grep -ciE '^[[:space:]]*FROM ' Dockerfile)
last_from=$(grep -iE '^[[:space:]]*FROM ' Dockerfile | tail -1)
if [ "$stages" -ge 2 ]; then
  ok "multi-stage build ($stages stages)"
else
  no "multi-stage build (found $stages stage)" "One stage means the thing that COMPILES your code is also the thing you deploy. Why is a compiler in production a problem — beyond size?"
fi
if printf '%s' "$last_from" | grep -qi 'sdk'; then
  no "final stage is a runtime image, not the SDK" "Your last FROM is an SDK image. What can an attacker with shell access do in a container that ships a compiler and your source?"
else
  ok "final stage is a runtime image, not the SDK"
fi

size=$(docker inspect -f '{{.Size}}' "$IMAGE" 2>/dev/null || echo 0)
size_mb=$((size / 1024 / 1024))
if [ "$size" -gt 0 ] && [ "$size" -lt "$MAX_SIZE_BYTES" ]; then
  ok "image size ${size_mb} MB (< $((MAX_SIZE_BYTES / 1024 / 1024)) MB)"
else
  no "image size ${size_mb} M B (< $((MAX_SIZE_BYTES / 1024 / 1024)) MB)" "What exactly is taking up those megabytes? 'docker history $IMAGE' answers it layer by layer."
fi

user=$(docker inspect -f '{{.Config.User}}' "$IMAGE" 2>/dev/null)
if [ -n "$user" ] && [ "$user" != "root" ] && [ "$user" != "0" ]; then
  ok "runs as non-root (USER=${user})"
else
  no "runs as non-root (USER is empty → root)" "A container escape starts with the identity inside. Why is root-by-default the wrong one, and what does USER actually change?"
fi

if docker inspect -f '{{json .Config.Env}}' "$IMAGE" 2>/dev/null | grep -qi 'password='; then
  no "no credentials baked into the image" "An image layer is public once you push it. Where should a connection string come from instead — and at which moment?"
else
  ok "no credentials baked into the image"
fi

# ── C · It actually serves ──────────────────────────────────────────────────
head_ "C · it actually serves"

docker rm -f forge-api-spec >/dev/null 2>&1
if cid=$(docker run -d --rm --name forge-api-spec -p "${HOST_PORT}:${CONTAINER_PORT}" "$IMAGE" 2>&1); then
  served=""
  for _ in $(seq 1 20); do
    if body=$(curl -fsS --max-time 2 "http://localhost:${HOST_PORT}/" 2>/dev/null); then served="$body"; break; fi
    sleep 0.5
  done
  if [ -n "$served" ]; then
    ok "GET / answers through the published port (${HOST_PORT}→${CONTAINER_PORT})"
  else
    no "GET / answers through the published port (${HOST_PORT}→${CONTAINER_PORT})" \
       "The process is up but nothing answers. Look at src/Forge.Api/Program.cs: which interface is Kestrel bound to? What does 'localhost' mean inside a network namespace, and who is on the other side of -p?"
    docker logs forge-api-spec 2>&1 | tail -12 | sed 's/^/      /'
  fi
  docker rm -f forge-api-spec >/dev/null 2>&1
else
  no "container starts" "$cid"
fi

# ── D · Wired into compose ──────────────────────────────────────────────────
head_ "D · wired into compose next to Postgres"

if compose_cfg=$(docker compose config 2>/dev/null); then
  if printf '%s' "$compose_cfg" | grep -qE '^[[:space:]]{2}api:'; then
    ok "compose defines an 'api' service"

    printf '%s' "$compose_cfg" | grep -q 'service_healthy' \
      && ok "api waits for the db healthcheck (condition: service_healthy)" \
      || no "api waits for the db healthcheck (condition: service_healthy)" \
            "'depends_on' alone only waits for the container to START. What is the difference between a started Postgres and a Postgres accepting connections?"

    api_db=$(printf '%s' "$compose_cfg" | grep -iE 'FORGE_DB' | head -1)
    if printf '%s' "$api_db" | grep -qi 'localhost'; then
      no "api's FORGE_DB points at the db service, not localhost" \
         "Inside the api container, what is 'localhost'? Which name resolves to Postgres on a compose network?"
    elif [ -n "$api_db" ]; then
      ok "api's FORGE_DB points at the db service, not localhost"
    else
      no "api gets FORGE_DB from the environment" "The image must not know the password. Who supplies it, and when?"
    fi
  else
    no "compose defines an 'api' service" "paso-00 ran the API on your host against a containerized DB. Now both are containers — what has to change for them to find each other?"
  fi
else
  no "docker compose config parses" "Fix the YAML first."
fi

# ── Verdict ─────────────────────────────────────────────────────────────────
printf '\n\033[1m%d passed, %d failed\033[0m\n' "$pass" "$fail"
if [ "$fail" -eq 0 ]; then
  printf '\033[32mpaso-00b green — tag it: git tag paso-00b\033[0m\n'
  exit 0
fi
printf '\033[33mRead the questions, not the answers. Ask me "why" on any of them.\033[0m\n'
exit 1
