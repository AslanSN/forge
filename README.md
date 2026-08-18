# forge

**Backend fundamentals, from the metal up.** A training repository that builds the operational layer *underneath* the framework — the database by hand, transactions and locking, indexing, the transactional outbox, Kafka, eventual consistency, and ops — one **deliberately-broken-then-fixed** step at a time.

## Why this exists

Most backend tutorials teach you the framework and let an ORM hide the database. `forge` does the opposite. Each step (`paso-NN`) starts by doing something the *tempting, plausible* way — the way that compiles, passes the demo, and ships a subtle bug — then reproduces the failure with a test and fixes it properly. The method is **do it wrong first**.

The goal is to turn backend *judgment* (recognizing a subtly-wrong answer) into backend *reps* (building the correct thing from scratch), at a mid/senior-in-a-team level.

## The spine: a concurrency-safe ledger

The domain is a double-entry ledger / wallet — the classic backend rite of passage — because a *correct* ledger forces every hard fundamental at once: money precision, the double-spend race, idempotency, indexed search, the transactional outbox, and eventual-consistency read models.

## Roadmap

| step | Focus | Teaches |
|------|-------|---------|
| **00** | The database, by hand (no ORM) | psql, raw parameterized SQL, hand-written migrations, schema design |
| **00b** | The image, by hand | multi-stage builds, layer cache, non-root containers, compose wiring |
| 01 | Money, correctly | `numeric` vs `float`, constraints, the double-entry invariant |
| 02 ⭐ | Transactions, isolation & the double-spend race | `SELECT FOR UPDATE`, isolation levels, concurrency tests |
| 03 | Idempotency & uniqueness | idempotency keys, unique constraints, safe retries |
| 04 | Indexes & query plans | `EXPLAIN ANALYZE`, index scans, connection pooling |
| 05 | Now the ORM (EF Core) | mapping the same ops, seeing the SQL it generates |
| 06 | Events & the transactional outbox | atomic write+publish, dual-write bugs |
| 07 | Kafka | producers/consumers, partitions, at-least-once, DLQ |
| 08 | Eventual consistency & read models | projections, read-model lag |
| 09 | Observability & ops | health/readiness/liveness, metrics, graceful shutdown |
| 10 | The interview layer | each step mapped to a mid/senior interview question |

Full detail in [docs/ROADMAP.md](docs/ROADMAP.md). Each step ships a `docs/paso-NN-*.md` write-up, working code, tests, and a git tag `paso-NN` — so you can `git checkout paso-02` to see exactly that state.

## Stack

.NET 10 · PostgreSQL 17 · Npgsql (raw, *before* EF) · Docker Compose · xUnit. Kafka (Redpanda) arrives at paso-07.

## Try it in 60 seconds

```bash
make up          # start Postgres (needs Docker/Colima running)
make migrate     # apply db/migrations/*.sql by hand
make run         # start the API on http://localhost:5000

# in another shell:
curl -X POST localhost:5000/accounts -H 'content-type: application/json' -d '{"name":"Alice"}'
curl localhost:5000/accounts/<id-from-above>
```

Tests: `make test`. The unit tests always run; the integration tests **skip cleanly** if Postgres isn't up, so the suite is green either way.

`make verify-image` runs paso-00b's executable spec for the container image — red until you write the Dockerfile, and it needs only Docker (no local .NET SDK).

## Repository layout

```
db/migrations/      hand-written SQL migrations (applied by scripts/migrate.sh)
src/Forge.Api/      the service (raw Npgsql at paso-00; EF arrives at paso-05)
tests/Forge.Tests/  unit (no Docker) + integration (skips if the DB is down)
docs/               the step write-ups + roadmap
```

## Colophon

Who did what (human vs AI): [COLOPHON.md](COLOPHON.md).

MIT © Alan Staub Negro
