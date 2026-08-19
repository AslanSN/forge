# paso-02 · Transactions, isolation & the double-spend race

**Goal:** a ledger that cannot be made to create money, no matter how many callers push at it at once — and the ability to say *which* mechanism you used and what it costs.

**Your turn.** This step is inverted (see [COLOPHON.md](../COLOPHON.md)): the contract and the failing spec are written; the implementation is yours, typed by hand. Nobody hands you the fix.

> **This step is where the code line switches to Go.** The reasoning is in [`go/README.md`](../go/README.md). Short version: .NET is already your paid production language, so the fundamentals rehearsed in C# add depth to something nobody disputes; the same fundamentals in Go build evidence you do not have yet. What is *not* language-specific — isolation levels, locking, the outbox — transfers either way, so nothing is lost by moving.

## Read this before you start

Primary sources only. Read them first; the step is much shorter afterwards.

| Read | What you'll extract |
|---|---|
| [Concurrency Control — overview](https://www.postgresql.org/docs/17/mvcc.html) | Why readers never block writers in Postgres, and why that is exactly what makes your read-then-write unsafe |
| [Transaction Isolation](https://www.postgresql.org/docs/17/transaction-iso.html) | What READ COMMITTED actually promises (and that it is the default), what REPEATABLE READ and SERIALIZABLE add, and the named anomalies: lost update, non-repeatable read, write skew |
| [Explicit Locking](https://www.postgresql.org/docs/17/explicit-locking.html) | Row-level locks, `FOR UPDATE` vs `FOR NO KEY UPDATE`, and the paragraph on deadlocks and lock ordering |
| [`SELECT … FOR UPDATE`](https://www.postgresql.org/docs/17/sql-select.html#SQL-FOR-UPDATE-SHARE) | The exact syntax and what it locks — rows, not tables, and only the ones actually returned |
| [Error codes](https://www.postgresql.org/docs/17/errcodes-appendix.html) | `40001` serialization_failure and `40P01` deadlock_detected — you will meet both, and both are *retryable*, not bugs |
| [pgx `Tx`](https://pkg.go.dev/github.com/jackc/pgx/v5#Tx) and [`pgxpool`](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool) | `Begin`/`Commit`/`Rollback`, `BeginTx` with `pgx.TxOptions{IsoLevel: …}`, and that a pool hands out *connections*, so a transaction lives on one |
| [`pgconn.PgError`](https://pkg.go.dev/github.com/jackc/pgx/v5/pgconn#PgError) | How to read the SQLSTATE out of an error with `errors.As` instead of matching on strings |
| [The Go race detector](https://go.dev/doc/articles/race_detector) | What it instruments — Go memory accesses — which is precisely why it will stay silent here |
| [`sync.WaitGroup`](https://pkg.go.dev/sync#WaitGroup) | `wg.Go` (Go 1.25+), used by the spec to release every goroutine at once |

## Before you can fail properly: paso-01, in Go

The race needs something naive to break, and the Go line does not have it yet. Two stubs, both small, both marked `YOUR TURN`:

- `go/internal/money/money.go` → `NormalizeAmount` — pure, no database, do it first (`make go-test` runs it without Docker).
- `go/internal/accounts/store.go` → `Deposit` and `Withdraw` — **the tempting way on purpose**: read the balance, check it in Go, write the new one back. `store_test.go` goes green with that, single-threaded.

A note the C# line did not need: **Go has no `decimal`**. `money.Minor` (int64 cents) is the scaffolding's assumption, not a verdict — the alternative is a decimal library. If you change it, change it now, before the concurrency work, and write the reason down here. What is not up for debate is that `float64` never touches money; `TestFloatDrifts` is there to be read.

```bash
make go-test    # money rules are red → implement → green (no Docker needed)
make up && make migrate
make go-test    # the store tests now run for real → green
```

Commit that before going on: `git commit -m "paso-01 (go): money rules and the naive read-check-write"`.

## Do it wrong first — and watch it happen

Now run the spec that was written to break what you just wrote:

```bash
make go-test    # race_test.go stops skipping the moment Withdraw exists
```

Four tests, and they are four different shapes of the same bug:

- **`TestConcurrentWithdrawals_NeverDoubleSpend`** — 100.00 in the account, 20 goroutines each taking 60.00. Arithmetic says exactly one can win. Count how many actually did.
- **`TestConcurrentDeposits_LoseNoUpdates`** — 50 deposits of 1.00, none of them refused, and the balance still comes out short. This is the *lost update*: no error anywhere, money simply overwritten.
- **`TestConcurrentTransfers_ConserveTotal`** — money moving both ways at once; the total across both accounts is the invariant.
- **`TestTransferIsAtomic`** — a transfer that must fail has to leave *both* sides untouched.

**Write down the numbers you actually saw**, here, before you fix anything — how many withdrawals succeeded, what the final balance was. A step whose failure you did not record is a step you will half-remember.

Then run it again with the race detector:

```bash
make go-race
```

**It says nothing.** Sit with that for a minute: your Go program has no data race — every goroutine touches its own memory. The race is in the *database*, between a `SELECT` and an `UPDATE` that other transactions are free to interleave with. The race detector cannot see it, no linter can, and no amount of Go-level care fixes it. That is the single most transferable thing in this step.

## The ladder — four rungs, and you climb all of them

Do **not** skip to the one that works. Each rung is a belief worth killing.

### Rung A · "I'll wrap it in a transaction"

Put the read and the write inside one `BEGIN … COMMIT` with pgx (`pool.Begin`, `defer tx.Rollback(ctx)`, `tx.Commit(ctx)`). Run the spec again.

**It still double-spends.** A transaction gives you atomicity and rollback, not mutual exclusion: under the default READ COMMITTED, two transactions can both read 100.00 and both write. Re-read the isolation doc with that failure in front of you. *A transaction is not a lock* is the sentence most candidates get wrong in interviews, and now you have a test that proves it.

### Rung B · `SELECT … FOR UPDATE`

Take a row lock on the account inside the transaction, then check, then write. Run it.

Green — and now the questions that matter: what is locked (the row? the table?), for how long, and what happens to the other nineteen goroutines while one holds it. Watch what the throughput does when you raise the goroutine count. Then transfer money **in both directions at once** and meet `40P01 deadlock_detected` — two transactions locking the same two rows in opposite order. The fix is not a retry, it is **lock ordering**: always lock the two accounts in a deterministic order (by id), and say why that removes the cycle.

### Rung C · `SERIALIZABLE`

Drop the explicit lock; open the transaction with `pgx.TxOptions{IsoLevel: pgx.Serializable}` instead. Run it.

It is correct too — and it fails differently: Postgres aborts one of the transactions with `40001 serialization_failure` and expects **you** to retry. Write that retry loop (bounded, with the SQLSTATE read via `errors.As` on `*pgconn.PgError`, never by matching the error text). The lesson: a stricter isolation level does not remove the concurrency problem, it *relocates* it into your application as a retry policy.

### Rung D · Make the invariant part of the statement

No explicit lock, no isolation escalation, one statement:

```sql
UPDATE accounts
   SET balance = balance - $1
 WHERE id = $2::uuid
   AND balance >= $1;
```

Zero rows affected *is* the insufficient-funds answer. Run the spec. Then explain to yourself why this one is safe: the read and the write are the same statement, so nothing can interleave between them, and the row lock Postgres takes for the update is held for the shortest possible time.

**Then choose.** Write, in this file, which rung you would ship for a real ledger and what would have to be true for you to pick a different one. That paragraph is the actual deliverable of paso-02 — the code is just the evidence you earned it.

## Go-specific things this step will teach you if you let it

- **`defer tx.Rollback(ctx)` right after `Begin`** — it is a no-op after a successful commit, and it is the only thing that saves you on an early `return err`. Every leaked transaction you ever meet in production is a missing one of these.
- **`context` is not decoration.** Pass the test's context into every query, then cancel it mid-transaction on purpose and watch what the pool does with the connection.
- **The pool is finite.** `db.Pool` takes `maxConns` explicitly. Set it to 2, run 20 goroutines, and watch them queue. Correctness must not depend on the pool being big enough — that is the difference between a system that is slow under load and one that is wrong under load.
- **Errors are values.** `errors.Is(err, accounts.ErrInsufficientFunds)` for your own sentinels, `errors.As` into `*pgconn.PgError` for the database's. No string matching, ever.

## The turn: the balance stops being a column

`db/migrations/001_accounts.sql` says it out loud: `balance` is a mutable column, deliberately naive, and this step is what breaks it. Once the race is fixed, the deeper problem is still there — a ledger that only stores the *current* balance cannot answer "how did we get here", and any bug that does slip through is unrecoverable because the history was overwritten.

Write the entries migration yourself — the next free number in `db/migrations/` (paso-04b has a claim on `002`) (the database, by hand, is your part — same as paso-00). What it needs to express:

- an append-only `entries` table: id, account_id, a signed amount in `numeric(18,2)`, created_at, and whatever you need to tie the two legs of a transfer together;
- **no `UPDATE` and no `DELETE` on it, ever** — decide how you enforce that, and whether the database should enforce it rather than your good intentions;
- the balance becomes derived: `SELECT sum(amount) FROM entries WHERE account_id = …`.

And then the question that makes this step worth doing, which you should answer *before* writing the migration: **when the balance is a `SUM`, where does "never go negative" live?** There is no column left to put a `CHECK` on. Your options are a lock on the account row, `SERIALIZABLE` with retries, or a maintained balance column that the entries have to agree with — and each buys correctness with a different currency. Pick one, write down what it costs, and make the same spec green against it.

## What to be able to explain afterwards, without notes

1. Why READ COMMITTED does not stop a read-then-write, in one sentence, without saying "race condition".
2. The difference between a lost update and a double spend — and why the second one is the first one wearing a uniform.
3. What `FOR UPDATE` locks, for how long, and how two of them deadlock.
4. Why `SERIALIZABLE` makes `40001` *your* problem, and what a correct retry loop looks like.
5. Why the guarded single-statement `UPDATE` needs neither, and when it stops being enough (hint: two rows).
6. Why the Go race detector was silent the whole time.
7. Where the no-overdraft invariant lives once the balance is a `SUM`, and what you gave up to put it there.

## When it's green, land it

```bash
make go-race            # green, and now you know what that does and does not prove
git add -A && git commit -m "paso-02 (go): transactions, isolation and the double-spend race"
git tag paso-02
```
