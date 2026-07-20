# paso-00 · The database, by hand

**Goal:** talk to Postgres directly — no ORM, no magic. Stand up the DB, connect with `psql`, write the schema as versioned SQL, and read/write from the app with raw, *parameterized* queries.

**Gap it closes:** "I've never touched a database directly."

## Do it wrong first

The tempting way to build a query from user input:

```csharp
// ❌ NEVER. String interpolation into SQL is SQL injection.
var sql = $"INSERT INTO accounts (name) VALUES ('{name}')";
```

It works in the demo. Then `name` is `Robert'); DROP TABLE accounts;--` and your table is gone (xkcd 327). The bug isn't exotic — it's the *default* if you concatenate.

## The right way

Parameters. The value is sent to the server separately and never becomes SQL:

```csharp
const string sql = "INSERT INTO accounts (name) VALUES (@name) RETURNING id, name, balance;";
cmd.Parameters.AddWithValue("name", name);
```

See [`AccountStore.cs`](../src/Forge.Api/Accounts/AccountStore.cs). The test `Parameterized_query_neutralizes_sql_injection` fires the injection payload and asserts the table survives.

## The schema

[`db/migrations/001_accounts.sql`](../db/migrations/001_accounts.sql), written by hand and applied by [`scripts/migrate.sh`](../scripts/migrate.sh) — not by an ORM, not on startup. Two deliberate choices:

- **`balance numeric(18,2)`**, never `float`. Money in binary floating point drifts — paso-01 proves it.
- **`balance` is a mutable column.** Deliberately naive. It works now, but paso-02 will show two concurrent transfers racing on it and overdrawing the account — which forces the redesign to an append-only `entries` table where the balance is *derived*, not stored.

## Run it

```bash
make up && make migrate      # start Postgres + apply the schema
make psql                    # then:  \d accounts    to inspect the table by hand
make run                     # start the API, then curl it (see README)
make test                    # unit tests pass; integration tests run if the DB is up
```

## What to be able to explain afterwards

- Why parameterized queries stop injection — and why an ORM is *not required* to be safe.
- What `numeric` buys you over `float` for money.
- Why a stored, mutable `balance` is a time bomb under concurrency (the setup for paso-02).
- What a migration is, and why applying it explicitly (not on app startup) matters.
