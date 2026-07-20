# Roadmap

The method is **do it wrong first**: each paso ships the tempting-but-wrong version, reproduces the failure with a test, then fixes it. Each maps to a real backend fundamental and, where relevant, to an entry in the [gotcha](https://github.com/AslanSN/gotcha) catalog.

| paso | Focus | The wrong-first | The fix / lesson | gotcha |
|------|-------|-----------------|------------------|--------|
| 00 | The database, by hand | string-interpolated SQL; ad-hoc DDL | parameterized queries; versioned migrations | — |
| 01 | Money | `float`/`double` balance → drift | `numeric(18,2)`, CHECKs, Σ(entries)=0 | #06 |
| 02 | Concurrency | read-balance-then-write → double-spend | transaction + `SELECT FOR UPDATE`; isolation levels | — |
| 03 | Idempotency | check-then-insert race → duplicate rows | unique constraint + idempotency key → idempotent 409 | #14 |
| 04 | Indexing | sequential scan on a hot query | `EXPLAIN ANALYZE` → the right index; connection pooling | #15 |
| 05 | ORM | EF Core hides the transaction/lock | map the same ops, log the generated SQL, keep control | #07 |
| 06 | Outbox | publish-before-commit / dual-write | transactional outbox (atomic write + enqueue) | #04 |
| 07 | Kafka | fire-and-forget, wrong partition key | keys, offsets, at-least-once, idempotent consumers, DLQ | #13 |
| 08 | Eventual consistency | read the read-model right after a write | projections + reconciling read-model lag | — |
| 09 | Ops | no health checks, abrupt shutdown | readiness/liveness probes, metrics, graceful shutdown | #11 |
| 10 | Interview layer | — | each paso ↔ the mid/senior question it answers | — |

## Gaps this is built to close

Self-assessed, honestly:

- **Operating a database directly** — raw SQL, `psql`, transactions, isolation, locking, index maintenance, connection pooling. (paso 00–05)
- **Docker beyond the basics** — standing up and wiring real infrastructure. (paso 00, 07, 09)
- **Message brokers / Kafka** — from zero to producers, consumers, partitions, and delivery semantics. (paso 07)

## Conventions

- One migration file per change under `db/migrations/`, numbered, idempotent, applied by `scripts/migrate.sh` — never auto-applied on startup.
- Every paso has: a `docs/paso-NN-*.md` write-up (goal · do-it-wrong-first · the fix · what to be able to explain), working code, tests, and a `paso-NN` git tag.
- Tests: pure unit tests never need Docker; integration tests skip cleanly when the DB is down.
