# paso-01 · Money, correctly

**Goal:** money that never drifts, and a balance protected against overdraft.

**Your turn.** This step is an *exercise*: the tests are written and **red**; you make them green. I don't hand you the implementation — the reps are the point.

## Do it wrong first — and see it

`MoneyRulesTests.Float_drifts_but_decimal_is_exact` already passes. Run it and read it: adding `0.1` ten times in `double` is not `1.0`. That is why money is `decimal` (C#) / `numeric` (Postgres), never `float`.

## Your task

1. **`MoneyRules.NormalizeAmount`** — pure, no DB, **do this first (no Docker needed)**. Reject amounts `<= 0` and amounts with more than 2 decimal places. Make `MoneyRulesTests` green.
2. **`AccountStore.DepositAsync` / `WithdrawAsync`** — needs the DB. Keep money in `decimal` end to end. Validate with `NormalizeAmount`. Withdraw must reject overdrawing. Make `AccountMoneyTests` green.

Do the withdraw the **naive** way for now — read the balance, check it in C#, then write. It will pass these single-threaded tests. **paso-02 will run two concurrent withdrawals and make this leak money** — that is when you learn locking. Leaving the bug in on purpose is the method.

## Run the loop

```bash
make test                 # MoneyRules tests are red → implement → green (no Docker)
make up && make migrate   # then start the DB
make test                 # AccountMoneyTests now run for real → make them green
```

## What to be able to explain afterwards

- Why `decimal`/`numeric`, not `float`/`double`, for money.
- Where you enforced "no overdraft" — in C#, in the DB `CHECK`, or both — and why *both* is the honest answer (defense in depth).
- Why the naive read-check-write withdraw is safe now but a bug waiting for paso-02.

## When it's green, commit it yourself

This one is yours to land:

```bash
git add -A && git commit -m "paso-01: money (decimal end-to-end, overdraft rejected)"
git tag paso-01
```
