# The Go line

From **paso-02 onward, forge's code is Go**. The `src/` and `tests/` .NET projects stay where they are, frozen at paso-01, as the worked reference for paso-00/00b — they are not dead weight and they are not a migration in progress.

## Why the switch

The fundamentals in this repo — isolation levels, locking, idempotency, the outbox, delivery semantics — are properties of Postgres, of brokers and of distributed systems. They transfer between languages essentially intact. What does *not* transfer is the evidence.

- **.NET is already the paid production language.** Rehearsing fundamentals in C# deepens the one thing nobody disputes on a CV that already carries shipped .NET work.
- **Go is where the evidence is missing.** There is a Go web service in production (a personal site: Echo, templ, unit-tested i18n, CI) — that proves *Go*, not *backend in Go*. Concurrency, transactions and queues under load are exactly the gap, and exactly what paso-02 onward is about.
- **Go's defaults suit the method.** This repo's whole premise is working underneath the framework — raw SQL, no ORM, hand-written migrations. In .NET that means swimming against the ecosystem; in Go with pgx it is simply how it is done. And the double-spend race is Go's signature demo: goroutines make the failure trivial to provoke and the race detector's silence teaches a lesson C# never sets up as cleanly.

Nothing is dropped by doing this. .NET keeps its place on the CV where it is strongest — production experience — and the study repo stops duplicating it.

## Which steps run where

| Steps | Language | Why |
|---|---|---|
| 00, 00b | .NET (done / frozen) | the worked reference: schema by hand, image by hand |
| 01 | **both** | the seam. The two languages escape `float` differently — `decimal` where the runtime has one, integer minor units where it does not — and that difference is the lesson |
| 00c | **decided as part of the step** | the contract is written language-free, then landed in one; choosing which is the exercise, not an oversight. See the write-up |
| **02, 03, 06, 06b, 07, 08, 09** | **Go** | the code steps — concurrency, idempotency, outbox, queues, Kafka, read models, ops. The language earns its keep here |
| 04b, 10, 11, 11b | language-agnostic | criterion and prose; no implementation to write |
| 04 | skipped | index/`EXPLAIN` work is already demonstrated in production (trigram GIN, ~500k rows, seq scan → ~3 ms) |
| 05 | skipped / optional | EF Core is used professionally. The Go analogue (sqlc vs. hand-rolled) is a footnote, not a step |
| 12 | Go | the agent seam sits on top of the outbox and queue built in Go |

`paso-01` gets re-done in Go as the *precondition* of paso-02 — the naive read-check-write has to exist before it can be broken. It is deliberately small.

Because it runs on both lines back to back, `paso-01` doubles as a controlled experiment: same task, same scope, two languages. Time each side and sort what you look up into *lexical* (a method name — heals on its own) and *design* (how to model an error here — the real gap). **Documentation yes, model no**, or the reading means nothing. That result is what decides `paso-00c`'s language.

## Layout

```
go/
  internal/money/      Minor (int64 cents), the numeric↔Go boundary, and the paso-01 rules stub
  internal/db/         DSN from the environment, pgxpool, and a test pool that skips when the DB is down
  internal/accounts/   the store: paso-00 reference (Create/Get) + the paso-01 and paso-02 stubs
                       store_test.go  — single-threaded spec (paso-01)
                       race_test.go   — the concurrency spec (paso-02): double spend, lost update,
                                        transfer atomicity, transfer conservation
```

## Running it

Every target sources `.env`, so `FORGE_DB_PORT` points at forge's Postgres and not at whatever else holds 5432 on this machine.

```bash
make go-test     # unit tests always; integration tests skip if the DB is down
make up          # start Postgres
make migrate     # apply db/migrations/*.sql
make go-test     # now the real thing
make go-race     # under the race detector — read paso-02 on what it does NOT catch
make go-vet      # vet + gofmt
```

Unimplemented stubs `panic("paso-NN: …")`, the Go equivalent of the `NotImplementedException`s on the .NET side. The test helpers turn that panic into a skip that names the step to do next, so a red run is always readable.

## Authorship

Same split as everywhere else in this repo ([COLOPHON.md](../COLOPHON.md)): the contracts, the write-ups and the failing specs are drafted by the AI; **every implementation under a `YOUR TURN` marker is typed by hand**, and a step is not done until the failure and the fix can be explained without help. `money.ParseMinor`/`String` and `accounts.Create`/`Get` are the paso-00-level reference, ported from the C# line so the Go line starts where that one did.
