# Roadmap

The method is **do it wrong first**: each step ships the tempting-but-wrong version, reproduces the failure with a test, then fixes it. Each maps to a real backend fundamental and, where relevant, to an entry in the [gotcha](https://github.com/AslanSN/gotcha) catalog.

Each step is identified as `paso-NN` — the write-up filename, and the git tag.

| step | Focus | The wrong-first | The fix / lesson | gotcha |
|------|-------|-----------------|------------------|--------|
| 00 | The database, by hand | string-interpolated SQL; ad-hoc DDL | parameterized queries; versioned migrations | — |
| 00b | The image, by hand | single-stage SDK image, root user, `COPY . .` before restore | multi-stage build, non-root, cache-ordered layers, compose wiring | — |
| 00c | Two hats | write the interface already picturing the implementation you're about to type, so the "contract" just narrates code you'd have written anyway | a real wall between the client hat (contract + acceptance tests, zero implementation) and the engineer hat (implementation against the contract alone, no more context) — the same rep paso-12 later asks of an agent | — |
| 01 · both | Money | `float`/`double` balance → drift | `numeric(18,2)`, CHECKs, Σ(entries)=0 — and the two standard escapes from float: a **decimal** type where the language has one (C#), **integer minor units** where it does not (Go has no decimal at all) | #06 |
| 02 · Go | Concurrency | read-balance-then-write → double-spend; "a transaction will fix it" | the four rungs: transaction ≠ lock → `FOR UPDATE` + lock ordering → `SERIALIZABLE` + `40001` retry → the guarded single-statement `UPDATE`; then the balance stops being a column | — |
| 03 | Idempotency | check-then-insert race → duplicate rows | unique constraint + idempotency key → idempotent 409 | #14 |
| 04 | Indexing | sequential scan on a hot query | `EXPLAIN ANALYZE` → the right index; connection pooling | #15 |
| 04b ⚖️ | Choosing the store | force every shape into the relational model (or flee to a document store wholesale) | where relational saves you vs. taxes you; what a document store buys and charges; one genuinely document-shaped entity, and the write-up of why | — |
| 05 | ORM | EF Core hides the transaction/lock | map the same ops, log the generated SQL, keep control | #07 |
| 06 | Outbox | publish-before-commit / dual-write | transactional outbox (atomic write + enqueue) | #04 |
| 06b ⚖️ | Queues as buffers | call the downstream service inline; treat a queue as "async Kafka" | the queue as a *shock absorber*: work queue vs. fan-out, at-least-once → idempotent consumers, visibility timeout, retry with backoff, DLQ, backpressure, priority lanes — and why this is a different question from streaming | — |
| 07 | Kafka | fire-and-forget, wrong partition key | keys, offsets, at-least-once, idempotent consumers, DLQ | #13 |
| 08 | Eventual consistency | read the read-model right after a write | projections + reconciling read-model lag | — |
| 09 | Ops | no health checks, abrupt shutdown | readiness/liveness probes, metrics, graceful shutdown | #11 |
| 10 | Interview layer | — | each step ↔ the mid/senior question it answers | — |
| 11 ⚖️ | The compute seam | run the heavy, CPU-bound job inside a serverless function and hit the hard timeout | serverless function vs. container vs. always-on worker: cold starts, execution ceilings, memory, concurrency per instance — and the criterion for *when not* to go serverless | — |
| 11b | The cloud seam | "it runs on my machine with Compose"; learn one provider's console instead of the primitives | stand the seam up on **one** managed cloud, then write the primitive-mapping table across GCP / AWS / Azure — managed identity instead of secrets, connection limits against a pooled database, and what the cloud silently changes about steps 02–07 | — |
| 12 ⭐⚖️ | The agent seam | let the agent call the service directly and hope | an agent chooses the route, but the choice is recorded idempotently, published through the outbox, and a downstream failure cannot corrupt state — *architecture for AI systems* | — |

## The language line

**Steps 00–01 are .NET; from paso-02 the code is Go.** The reasoning, and the table of which step
runs in which language, is in [`go/README.md`](../go/README.md). In one line: the fundamentals here
are properties of Postgres and of distributed systems, so they transfer between languages, while the
*evidence* does not — .NET is already the paid production language, Go is the gap. The .NET projects
stay frozen at paso-01 as the worked reference for paso-00/00b, and `paso-01` is re-done in Go, small,
as the precondition of paso-02: the naive read-check-write has to exist before it can be broken.

**A note on paso-00c:** every gap below comes from the systems-architecture interview in
[PLAN.md](PLAN.md) — a fixed list, one question per row. `paso-00c` doesn't answer one of those
questions; it hardens the method the other steps already run on. From `paso-00b` onward, forge
*is* a client/engineer split — the AI writes the contract, you implement it blind to nothing more
than what the contract states. `paso-00c` is the one step where you write both halves yourself,
in sequence, with a real wall in between: the rep of *authoring* a contract someone else has to
implement without your help is different from the rep of *receiving* one, and `paso-12` will
eventually ask an agent to work from exactly that kind of contract with no more context than you
give it.

## Gaps this is built to close

Self-assessed, honestly:

- **Operating a database directly** — raw SQL, `psql`, transactions, isolation, locking, index maintenance, connection pooling. (steps 00–05)
- **Docker beyond the basics** — *authoring* images, not just consuming them: stage boundaries, layer cache, the runtime user, and wiring services together. (step 00b, then 07, 09)
- **Message brokers / Kafka** — from zero to producers, consumers, partitions, and delivery semantics. (step 07)
- **Storage-choice judgment** — relational vs. document, argued rather than assumed. (step 04b)
- **Queueing as a load-shaping tool**, distinct from streaming. (step 06b)
- **Compute-model judgment** — serverless function vs. container vs. worker, and the cost of
  getting it wrong for long or CPU-bound work. (step 11)
- **Running the thing on a managed cloud**, not just on Compose — and holding the primitives
  provider-independently rather than learning one console. (step 11b)
- **Putting an agent's decision on top of all of it** without letting a non-deterministic choice
  corrupt state. (step 12)

The order I actually walk these — and the two steps I skip because the evidence already exists —
is in [PLAN.md](PLAN.md).

## ⚖️ Steps that need a second opinion

Most steps here have a **verifiable** answer: a test fails, you fix it, the test passes, and the
machine tells you whether you were right. A few do not. Choosing a store, choosing a compute
model, shaping a queue and putting an agent's decision on top of the whole thing are **judgment**
— and judgment is where being confidently wrong is invisible, because nothing goes red.

Those steps are marked **⚖️**. The rule for them:

1. **Decide first, alone.** Write down the decision, the reasons, and — the line most people skip —
   *what would have to be true for the opposite decision to be right*.
2. **Only then, get it read** by someone who runs systems like this at a scale you haven't.
   Not "teach me this", but *"here's what I decided and why — where am I wrong?"*

The order matters. Ask first and you inherit an answer you can recite but not defend under a
follow-up question, which is worse than not knowing: it fails at exactly the moment it counts.
Decide first and the same conversation returns their criterion applied to *your* reasoning.

A ⚖️ step is not finished when the code runs. It is finished when someone qualified has read the
write-up and disagreed with at least one thing in it.

## Conventions

- One migration file per change under `db/migrations/`, numbered, idempotent, applied by `scripts/migrate.sh` — never auto-applied on startup.
- Every step has: a `docs/paso-NN-*.md` write-up (goal · do-it-wrong-first · the fix · what to be able to explain), working code, tests, and a `paso-NN` git tag.
- Tests: pure unit tests never need Docker; integration tests skip cleanly when the DB is down.
- **Inverted steps** ship the spec *before* the implementation: the write-up states the contract, an executable spec fails until you satisfy it, and the answers are questions. `paso-00b` is the first (`make verify-image`). Its spec is bash rather than xUnit because it asserts properties of an *image*, and must run without a local .NET SDK. `paso-00c` is the second, and the only step where the human writes the contract *and* the spec, not just the implementation — see its write-up and [COLOPHON.md](../COLOPHON.md).
