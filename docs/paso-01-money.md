# paso-01 · Money, correctly

**Goal:** money that never drifts, and a balance protected against overdraft.

**Your turn.** This step is an *exercise*: the tests are written and **red**; you make them green. The implementation is not handed to you — the reps are the point.

## This step runs twice, in two languages, on purpose

`paso-01` is the seam where forge changes language ([`go/README.md`](../go/README.md)), so it is the one step that exists on both lines. That is not duplicated work: **the two languages answer the same question differently**, and the difference is the most interview-useful thing in this step.

| line | what you implement | the money type | why do it |
|---|---|---|---|
| **Go** — `go/internal/money`, `go/internal/accounts` | `NormalizeAmount`, `Deposit`, `Withdraw` | `Minor` — an `int64` count of cents | **Required.** The naive `Withdraw` has to exist before `paso-02` can break it. `race_test.go` is already written and waiting |
| **.NET** — `src/Forge.Api/Accounts` | `MoneyRules.NormalizeAmount`, `DepositAsync`, `WithdrawAsync` | `decimal` | **Housekeeping, not study.** `go/README.md` calls the .NET projects "the worked reference for paso-00/00b", and a reference that throws `NotImplementedException` is not a reference |

**Order: Go first, .NET second.** Go is on the critical path to `paso-02`; .NET is a short session that puts the suite back to green and makes the frozen reference honest. Do not treat the second one as a learning step — it is the same lesson in another syntax, and once the Go version exists, the C# version is typing.

> Not spending a session on the .NET side at all is a defensible call — but then say so in `go/README.md` explicitly ("the .NET line stops mid-`paso-01`") rather than leaving it ambiguous. Ambiguous is the only wrong answer here.

## Before you start: read, then implement

The whole step rests on one property of binary floating point, and it is documented precisely by every store and language involved. Read the primary sources — this is specified behaviour, not folklore.

- **PostgreSQL, the store's own answer** — [8.1.2 Arbitrary Precision Numbers](https://www.postgresql.org/docs/current/datatype-numeric.html). Extract: `numeric` is *"especially recommended for storing monetary amounts"*, and what **precision** and **scale** mean exactly — `numeric(18,2)` in `001_accounts.sql` is that vocabulary.
- **PostgreSQL, the warning** — [8.1.3 Floating-Point Types](https://www.postgresql.org/docs/current/datatype-numeric.html#DATATYPE-FLOAT), same page. Extract: the sentence *"If you require exact storage and calculations (such as for monetary amounts), use the `numeric` type instead"*, and **why comparing two floats for equality can fail**.
- **.NET's answer** — [System.Decimal](https://learn.microsoft.com/en-us/dotnet/api/system.decimal). Extract: `decimal` is base-10 with 28–29 significant digits — *narrower in range than `double`, exact in the decimal fractions money actually uses*. The opposite trade.
- **Go's answer — and the interesting part** — [Go spec: Numeric types](https://go.dev/ref/spec#Numeric_types). Extract: the complete list of built-in numeric types. **There is no decimal.** Go gives you integers, IEEE-754 floats and complex numbers, and nothing else — which is precisely why `go/internal/money` counts *cents in an `int64`* instead of reaching for a decimal type that does not exist.
- **The boundary between them** — [pgx `pgtype.Numeric`](https://pkg.go.dev/github.com/jackc/pgx/v5/pgtype#Numeric). Extract: a PostgreSQL `numeric` can be scanned into `Numeric`, `int64` *or* `float64`. That last option is offered and is a trap: the store held the value exactly, and the scan is where you would throw that away. `store.go` deliberately scans into a `string` and parses it — read `ParseMinor` and work out why.
- **The underlying reason, once and properly** — [What Every Computer Scientist Should Know About Floating-Point Arithmetic (Goldberg)](https://docs.oracle.com/cd/E19957-01/806-3568/ncg_goldberg.html). Long; the first two sections are enough. Optional, but this is the source everything else paraphrases.

## Do it wrong first — and see it

`MoneyRulesTests.Float_drifts_but_decimal_is_exact` already passes. Run it and read it: adding `0.1` ten times in `double` is not `1.0`.

That is the "wrong" this step exists to make visible. Note that **the fix differs per line, and both are correct**:

- **.NET** reaches for a decimal type the runtime provides — base-10 arithmetic, exact for the fractions money uses.
- **Go** has no such type, so it removes the fraction entirely: store cents as an integer and there is nothing left to round. This is the older and more portable answer, and it is what most payment systems do regardless of language.

Being able to say *"the trap is the same, and there are two standard escapes — a decimal type where the language has one, integer minor units where it does not"* is a materially better answer than knowing only that money is not a float.

## Your task — the Go line (do this first)

1. **`money.NormalizeAmount`** (`go/internal/money/money.go`) — pure, no DB, **start here**. Reuse `ParseMinor`. Reject amounts `<= 0` and anything with more than two decimal places — *an error, never a rounding*. Make `money_test.go` green.
2. **`accounts.Deposit` / `accounts.Withdraw`** (`go/internal/accounts/store.go`) — needs the DB. Validate with `NormalizeAmount`; `Withdraw` must refuse to overdraw and return `ErrInsufficientFunds`. Make `store_test.go` green.

Do the withdraw the **naive** way on purpose — read the balance, check it in Go, write it back. It passes every single-threaded test in `store_test.go`. **`race_test.go` exists to prove that is not enough**, and `paso-02` is where you find out why. Leaving the bug in deliberately is the method (see [AGENTS.md](../AGENTS.md)).

```bash
make go-test              # money tests are red → implement → green (no Docker)
make up && make migrate   # then start the DB
make go-test              # store_test.go now runs for real → make it green
make go-race              # optional preview: watch it fail. Do NOT fix it here.
```

## Your task — the .NET line (housekeeping)

The same two moves, three stubs: `MoneyRules.NormalizeAmount` (`Models.cs:38`), then `AccountStore.DepositAsync` / `WithdrawAsync` (`AccountStore.cs:57,60`). Keep money in `decimal` end to end. Same naive withdraw — the .NET line stops at `paso-01`, so its bug is never exploited, but keeping it consistent with the Go line is what makes it a usable reference.

```bash
make test                 # MoneyRules red → implement → green (no Docker)
make up && make migrate
make test                 # AccountMoneyTests run for real → green
```

## What to be able to explain afterwards

- Why `numeric` in the database, and the exact sentence PostgreSQL's own docs use about it.
- **The two escapes from float, and when each applies**: a decimal type where the language has one (C#), integer minor units where it does not (Go). Why Go has no decimal, and what that forces.
- Where the value could silently lose precision **at the boundary** — scanning a `numeric` into a `float64` is offered by pgx and is exactly the mistake; what `store.go` does instead, and why.
- Where you enforced "no overdraft" — in the application, in the database `CHECK`, or both — and why *both* is the honest answer (defense in depth).
- Why the naive read-check-write withdraw is safe now and a bug waiting for `paso-02`.
- Why more than two decimal places is rejected rather than rounded. (Rounding is a *decision about someone's money*, made silently, in a layer with no authority to make it.)

## When it's green, commit it yourself

Two commits, one tag — the tag goes on the second, once both lines are green:

```bash
git add -A && git commit -m "paso-01(go): money (minor units end-to-end, overdraft rejected)"
git add -A && git commit -m "paso-01(dotnet): money (decimal end-to-end, overdraft rejected)"
git tag paso-01
```
