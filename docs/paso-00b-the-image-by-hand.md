# paso-00b · The image, by hand

**Goal:** package the service as a container image *you wrote*, and run it next to Postgres on a compose network — so both sides of the app are infrastructure you can explain.

**Gap it closes:** "Docker beyond the basics." paso-00 *consumed* a container (`make up` starts someone else's Postgres image). Running `docker compose up` teaches you Docker the way an ORM teaches you SQL: it works, and you learn nothing about the layer underneath. This step makes you the **author** of the image, not its consumer.

> **This step is inverted** (see [COLOPHON.md](../COLOPHON.md)). The spec is written for you; the `Dockerfile`, the `.dockerignore` and the compose `api` service are **yours to type**. `scripts/verify-image.sh` is the failing test — it asserts properties and answers every failure with a *question*, never with the fix.

## Do it wrong first

Here is the Dockerfile almost every tutorial hands you. It builds. It runs. It would be a bad thing to deploy:

```dockerfile
# ❌ The tempting one.
FROM mcr.microsoft.com/dotnet/sdk:10.0
WORKDIR /app
COPY . .
RUN dotnet publish src/Forge.Api -o /app/out
ENTRYPOINT ["dotnet", "/app/out/Forge.Api.dll"]
```

Four separate defects, none of which the demo reveals:

1. **The final image is the SDK.** You are shipping a *compiler*, a package manager, and your complete source tree to production — roughly 1 GB where ~100 MB would do. Size is the symptom; the attack surface is the disease.
2. **It runs as root.** Nothing here says otherwise, and the default is uid 0. The first thing a container escape wants is the identity it starts from.
3. **`COPY . .` before `restore`.** Docker caches per layer and invalidates everything below the first change. Touch one `.cs` file and you re-download every NuGet package, forever.
4. **No `.dockerignore`.** So `COPY . .` sweeps in your `.git` history and your local `bin/`/`obj/` — output compiled for *your* machine, now shadowing the build inside the image. This is the one that produces the truly baffling bug: it works on your laptop and fails in CI, or vice versa, for reasons invisible in the Dockerfile.

There is a fifth defect that only shows up when you try to reach the thing. Run the image, publish the port, `curl` it — and get nothing, with a perfectly healthy process in the logs. `src/Forge.Api/Program.cs` binds Kestrel to `http://localhost:5000`. Work out what `localhost` means inside a network namespace and you have the answer; the spec's group C will walk you into it.

## The contract

`scripts/verify-image.sh` enforces exactly this. Nothing here is a hint about *how*:

**The image**
- A multi-stage build: the stage that compiles is not the stage you ship, and the final `FROM` is not an SDK image.
- Under 400 MB.
- A non-root `USER`.
- Project files copied and restored **before** the rest of the source, so a code edit doesn't invalidate the package cache.
- A `.dockerignore` that excludes at least `bin/`, `obj/` and `.git`.
- No credentials in any layer or in `ENV`. The connection string arrives at **run** time, not build time.
- The app listens on **port 8080 inside the container** and answers `GET /` through a published port.

**The wiring**
- `docker-compose.yml` gains an `api` service built from your Dockerfile.
- It waits for the database's *healthcheck*, not merely for the container to start.
- Its `FORGE_DB` resolves Postgres by its **compose service name** — the API is no longer on your host, so `localhost` is now the wrong answer for the same reason it was in group C.

## Run it

```bash
make verify-image     # red now: there is no Dockerfile. That is the starting line.
```

Iterate against it. Useful while you work:

```bash
docker build -t forge-api:paso-00b .
docker history forge-api:paso-00b        # which layer is fat, and why
docker run --rm -p 15000:8080 forge-api:paso-00b
docker compose up -d --build             # once the api service exists
docker compose logs -f api
```

Base images you'll want to know about (all three exist for .NET 10): `mcr.microsoft.com/dotnet/sdk:10.0` to build, `mcr.microsoft.com/dotnet/aspnet:10.0` to run, and `mcr.microsoft.com/dotnet/aspnet:10.0-noble-chiseled` — no shell, no package manager, ~100 MB — when you want to see how small correct can get. Chiseled changes how you create a user; that difference is the lesson, not an obstacle.

**No local .NET SDK required.** The compile happens inside the build stage, which is why this is the one step you can finish on a machine with only Docker installed.

## What to be able to explain afterwards

- Why a multi-stage build is a *security* argument before it is a size argument.
- How Docker's layer cache invalidates, and why `COPY` order is a build-time performance decision.
- What `USER` actually changes, and why root-by-default is a real risk rather than a lint warning.
- What `localhost` resolves to inside a container, and why `-p 15000:8080` doesn't save you from binding to the loopback interface.
- Why `depends_on` without `condition: service_healthy` still gives you a connection-refused on startup — the same started-vs-ready distinction that paso-09 turns into readiness probes.
- Where a secret is allowed to live: not in a layer, not in `ENV` at build time. Say where instead, and when it arrives.

## The interview question this answers

*"Walk me through how you'd containerize this service for production."* A candidate who has consumed Docker says "I'd write a Dockerfile and run docker build." A candidate who has authored one talks about stage boundaries, cache ordering, the runtime user, and where the config comes from — and can say why each choice is the way it is. That gap is this step.
