# forge

**Backend fundamentals, from the metal up.** A training repository that builds the operational layer *underneath* the framework — the database by hand, transactions and locking, indexing, the transactional outbox, Kafka, eventual consistency, and ops — one **deliberately-broken-then-fixed** step at a time.

## Why this exists

Most backend tutorials teach you the framework and let an ORM hide the database. `forge` does the opposite. Each step (`paso-NN`) starts by doing something the *tempting, plausible* way — the way that compiles, passes the demo, and ships a subtle bug — then reproduces the failure with a test and fixes it properly. The method is **do it wrong first**.

The goal is to turn backend *judgment* (recognizing a subtly-wrong answer) into backend *reps* (building the correct thing from scratch), at a mid/senior-in-a-team level.

## The spine: a concurrency-safe ledger

The domain is a double-entry ledger / wallet — the classic backend rite of passage — because a *correct* ledger forces every hard fundamental at once: money precision, the double-spend race, idempotency, indexed search, the transactional outbox, and eventual-consistency read models.

## Roadmap

| step | lang | Focus | Teaches |
|------|------|-------|---------|
| **00** | .NET | The database, by hand (no ORM) | psql, raw parameterized SQL, hand-written migrations, schema design |
| **00b** | .NET | The image, by hand | multi-stage builds, layer cache, non-root containers, compose wiring |
| 00c | *your call* | Two hats | authoring a contract someone else implements blind, then implementing one — the contract is written language-free first, and picking the language is part of the exercise |
| 01 | **both** | Money, correctly | `numeric` vs `float`; a decimal type where the language has one, integer minor units where it does not |
| 02 ⭐ | Go | Transactions, isolation & the double-spend race | `SELECT FOR UPDATE`, lock ordering, `SERIALIZABLE` + `40001` retry, concurrency tests |
| 03 | Go | Idempotency & uniqueness | idempotency keys, unique constraints, safe retries |
| ~~04~~ | — | ~~Indexes & query plans~~ | *skipped — already demonstrated in production (trigram GIN, ~500k rows)* |
| 04b ⚖️ | agnostic | Choosing the store | relational vs. document, argued from invariants and what you must count |
| ~~05~~ | — | ~~The ORM~~ | *skipped — EF Core is used professionally* |
| 06 | Go | Events & the transactional outbox | atomic write+publish, dual-write bugs |
| 06b ⚖️ | Go | Queues as buffers | work queue vs. fan-out, visibility timeout, backpressure, DLQ |
| 07 | Go | Kafka | producers/consumers, partitions, at-least-once, DLQ |
| 08 | Go | Eventual consistency & read models | projections, read-model lag |
| 09 | Go | Observability & ops | health/readiness/liveness, metrics, graceful shutdown |
| 10 | agnostic | The interview layer | each step mapped to a mid/senior interview question |
| 11 ⚖️ | agnostic | The compute seam | serverless vs. container vs. worker, and when *not* to go serverless |
| 11b | agnostic | The cloud seam | one managed cloud, then the primitive-mapping table across providers |
| 12 ⭐⚖️ | Go | The agent seam | an agent's choice, recorded idempotently and published through the outbox |

Full detail in [docs/ROADMAP.md](docs/ROADMAP.md). Each step ships a `docs/paso-NN-*.md` write-up, working code, tests, and a git tag `paso-NN` — so you can `git checkout paso-02` to see exactly that state.

## Stack

**PostgreSQL 17** throughout, driven with raw SQL and hand-written migrations — no ORM, no auto-migration on startup. Docker Compose for everything local. Kafka (Redpanda) arrives at paso-07.

**Two language lines, and the seam is deliberate:**

- **.NET 10** · Npgsql · xUnit — steps 00–01. The worked reference, then frozen.
- **Go** · pgx · the standard `testing` package and the race detector — step 02 onward, where concurrency, queues and delivery semantics live.

`paso-01` is the one step that exists on both sides, because the two languages solve the money problem differently and the difference is the lesson. The full reasoning, and the table of which step runs where, is in [`go/README.md`](go/README.md).

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
