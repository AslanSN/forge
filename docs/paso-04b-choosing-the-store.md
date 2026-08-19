# paso-04b · Choosing the store

**Goal:** decide, on the record, which shapes of data in this ledger belong in the relational model and which do not — and be able to hold that boundary under a follow-up question.

**Gap it closes:** "Storage-choice judgment — relational vs. document, argued rather than assumed." This is question 1 of the interview that produced [PLAN.md](PLAN.md), and the one that had no answer at all.

> **This step is ⚖️, so it has no failing spec.** Every step so far ends with a machine telling you whether you were right. Nothing goes red when you put an entity in the wrong store — the bill arrives eighteen months later, as a rewrite. So the loop is different: **decide → anchor the decision in a small piece of code → write down what would change your mind → then, and only then, get it read.** It is not done when something runs. It is done when someone who operates systems at a scale you haven't has read the decision and disagreed with at least one thing in it. The convention, and why *decide → then ask* is not optional, is in [ROADMAP.md](ROADMAP.md#️-steps-that-need-a-second-opinion).
>
> **It is also inverted** (see [COLOPHON.md](../COLOPHON.md)): the migration, the queries and the decision are yours to type. What follows states the contract and the questions, never the answer.

## Before you start: read, then probe

Every claim this step asks you to make is something the store's own documentation already says precisely. The skill is that the pages exist and that you can quote them back in your own words. Read these before you write the migration — each bullet points at the **primary source** and names what you should carry out of it. Half of what these probes measure is how fast each page's words become *your* reasons.

### PostgreSQL — the store you own

- **JSON Types** — [PostgreSQL docs: 8.14 JSON Types](https://www.postgresql.org/docs/current/datatype-json.html). `json` vs `jsonb`: how each is stored, the design guidance in *8.14.2 Designing JSON Documents* on when to prefer one, and *8.14.4 jsonb Indexing* — `GIN`, `jsonb_ops` vs `jsonb_path_ops`, and what that buys. This is the deep background of the whole step.
- **JSON Functions and Operators** — [PostgreSQL docs: 9.16 JSON Functions and Operators](https://www.postgresql.org/docs/current/functions-json.html). `->` vs `->>` vs `#>>` (value vs text vs path), containment `@>`, and `jsonb_populate_record`, so you can project a blob into a row shape without a new table.
- **GIN Indexes** — [PostgreSQL docs: 65.4 GIN Indexes](https://www.postgresql.org/docs/current/gin.html). What a GIN index actually indexes (elements, not whole values) and why that makes containment and `->` lookups something the index can answer.
- **Indexes on Expressions** — [PostgreSQL docs: 11.7 Indexes on Expressions](https://www.postgresql.org/docs/current/indexes-expressional.html). How to make a lookup fast on a value you don't store as its own column.
- **Generated Columns** — [PostgreSQL docs: 5.4 Generated Columns](https://www.postgresql.org/docs/current/ddl-generated-columns.html). `GENERATED ALWAYS AS (...) STORED`: the safe way to keep a denormalized copy honest.
- **Constraints** — [PostgreSQL docs: 5.5 Constraints](https://www.postgresql.org/docs/current/ddl-constraints.html). Read the `CHECK` sections: a constraint sees one row, never its neighbors — and that is exactly why Probe A behaves differently per shape.
- **TOAST** — [PostgreSQL docs: 66.2 TOAST](https://www.postgresql.org/docs/current/storage-toast.html). What physically happens when you `UPDATE` one key inside a large `jsonb` value, and what that implies for which fields deserve to live in the blob.
- **Size Functions** — [PostgreSQL docs: 9.28 System Administration Functions · Database Object Size Functions](https://www.postgresql.org/docs/current/functions-admin.html#FUNCTIONS-ADMIN-DBSIZE). `pg_column_size`, for measuring what each shape actually occupies.
- **Reading the probes** — [PostgreSQL docs: 14.1 Using EXPLAIN](https://www.postgresql.org/docs/current/using-explain.html) and the [EXPLAIN command reference](https://www.postgresql.org/docs/current/sql-explain.html). What `EXPLAIN (ANALYZE, BUFFERS)` is reporting, and why the buffers count matters more than the milliseconds.

### MongoDB — the store you're not fleeing to

- **Unique Indexes** — [MongoDB docs: Unique Indexes](https://www.mongodb.com/docs/manual/core/index-unique/). Enforceability lives there too; the question is never "can it?", it's "is it default-and-cheap?".
- **Transactions** — [MongoDB docs: Transactions](https://www.mongodb.com/docs/manual/core/transactions/). Multi-document ACID exists since 4.0 — do not start Probe A carrying a stale 2013 talking point.
- **Data Modeling** — [MongoDB docs: Data Modeling](https://www.mongodb.com/docs/manual/data-modeling/). How a document store models things that elsewhere get joined — and what the modeling *costs* the moment you need a sum across them.

## Do it wrong first — twice

This step has two wrong-firsts, because the mistake is symmetric and most people only recognize the half they don't commit.

### Wrong #1 — force every shape into the relational model

The tempting version, once a second LLM provider shows up and its response metadata doesn't look like the first one's:

```sql
-- ❌ The tempting one, half A: a column for every field anyone might ever send.
ALTER TABLE agent_calls ADD COLUMN openai_finish_reason   text;
ALTER TABLE agent_calls ADD COLUMN anthropic_stop_reason  text;
ALTER TABLE agent_calls ADD COLUMN gemini_safety_ratings  text;
```

You are now modelling *someone else's* schema, and they change it without asking you. Most rows are NULL, every provider release is a migration and a deploy, and the table's width grows monotonically because nothing is ever safe to drop. The usual escape — an entity-attribute-value table, one row per key — is the same defeat with extra joins, no types, and no constraint that can see across the rows it split apart.

### Wrong #2 — flee to a document store wholesale

"Relational is rigid. Let's put it in Mongo." Then `accounts` and `entries` go too, because splitting stores felt like the complicated option, and the first thing you lose is the thing this entire repo is built on: nothing refuses to let the balance go negative, Σ(entries) = 0 becomes a property your code has to remember rather than one the store guarantees, and paso-02's `SELECT FOR UPDATE` and paso-03's unique constraint turn into application logic you have to get right in every write path, forever, including the ones you write in a hurry.

**Do not reach for the stale talking points on this side.** MongoDB has unique indexes, and it has multi-document ACID transactions (since 4.0 on replica sets; 4.2 across shards). "Document stores don't do transactions" is a 2013 answer, and an interviewer who actually runs one will eat it in a single follow-up. The real distinction is what each store makes **default and cheap** versus **possible and expensive**, and — the part that decides it — **where the invariant ends up living** when the two shapes have to agree.

## The two questions that actually decide it

Not "rigid vs. flexible". That framing has never once decided a real case, because both words are compliments in the abstract.

1. **Which invariants do I need the store to enforce *for me*?**
   An invariant enforced in application code holds until the second writer. So: what must still be true after a backfill script, a support engineer in `psql`, a replay of the outbox, or the next service someone adds? Whatever survives that list is what belongs in a constraint.

2. **What am I going to have to count in order to bill for it?**
   The shape you can write fastest is not the shape you can aggregate. At month end somebody sends an invoice, and then reconciles it against the provider's own invoice, and the difference has to be explainable. Design for the write and you meet this question later, in production, under time pressure.

Question 1 sets the floor. Question 2 is the one people discover afterwards.

## The entity

The one genuinely document-shaped thing this ledger will own — chosen because paso-12 makes it real rather than hypothetical: **the agent call record.** One record per LLM invocation the platform makes on a user's behalf: the request, the model and version, the route the agent chose and the candidates it rejected with their scores, tool calls and their results, the provider's response metadata, timings, token counts, and the money it cost.

It is document-shaped for reasons that survive scrutiny, not merely because normalizing it is tedious:

- **Its schema belongs to someone else** and changes without a migration window — per provider, per model version, per week.
- **It is read whole, by id, mostly while debugging.** Nothing joins to its interior; no other table's correctness depends on the shape inside it.
- **It is evidence.** A record written in March must still read back exactly as it was recorded, in the vocabulary of the model that produced it. Migrating old evidence into a new shape destroys the thing that made it useful.

And now the part that is **not** document-shaped, which is the entire step: `account_id`, the idempotency key, the cost, the token counts, the timestamp. Those get joined, deduplicated, summed and invoiced. Question 2 owns them.

> So the boundary does not fall between two stores. It falls **inside one entity** — and drawing it, in a specific place, for stated reasons, is the exercise.

## The anchor

Deliberately small: one migration and a handful of queries you run by hand. No new service, no new container, no C#. **Keep it to one sitting.** If it is taking three, you have started building a product instead of a probe.

The contract:

1. **Both shapes exist, holding the same data.** Migration `db/migrations/002_agent_calls.sql`, hand-written and idempotent like `001`, applied by `scripts/migrate.sh` — one version that models the varying part relationally (child tables, or EAV — pick the one you'd actually be tempted by), and one where the varying part is a single `jsonb` column. Generate enough rows in SQL that `EXPLAIN ANALYZE` has something to say: tens of thousands, not twelve. You already know how to read a plan — that is paso-04, and the trigram-GIN work you did in production.

2. **Probe A — the invariant.** Choose one invariant that must hold for the billing to be honest ("the same provider call is never charged twice" is the sharp one; "cost is never negative" is the cheap one). Now try to violate it from `make psql`, directly, bypassing the application entirely, against both shapes. Write down which one said *no*, and what it would take to make the other one say *no* too.

3. **Probe B — the bill.** Write the month-end query: cost per account, per model, for a period. Run it against both shapes under `EXPLAIN (ANALYZE, BUFFERS)`. Then make the slow one fast, and record **what that cost you** — a new index, a new column, a write-path change, a rebuild. The milliseconds are not the finding. The price of the milliseconds is the finding.

4. **Probe C — the schema that moved under you.** A new provider starts sending a field nobody planned for, and rows written before today will never have it. Add it on both sides, then re-run A and B. What did each side make you do: a migration? a backfill? a deploy? a conversation with another team? How long is each shape unavailable, or wrong, while you do it?

Worth having looked up before you start — knowing that these exist is the point; which ones you use is your decision: `jsonb` vs `json`, the `->` / `->>` / `#>>` operators, containment `@>`, GIN with `jsonb_ops` vs `jsonb_path_ops`, expression indexes, `GENERATED ALWAYS AS (...) STORED` columns, `CHECK` over an expression, `jsonb_populate_record`, `pg_column_size`, and what TOAST does to a large value when you update one key inside it. All of them, linked and with what to extract named, are in [Before you start](#before-you-start-read-then-probe) above.

### Run it

```bash
make up && make migrate    # Postgres + your 002 migration
make psql                  # the whole step happens here
```

No API, no test project, no Docker image: this step is reachable even with paso-01's suite still red. That doesn't repeal [PLAN.md](PLAN.md)'s rule about closing the open steps first — it just means nothing here is blocked on them.

### Optional stretch — only if the four probes are done and the week isn't

Stand a real document store up as a compose service and repeat probes A and B on it. Then answer the question honestly: **what did the second store give you that `jsonb` in the database you already run did not?** If the answer is "nothing, at this volume" — that is a finding, and a strong one, provided you can name the volume, the access pattern or the operational fact at which it stops being true. That named threshold is worth more in an interview than the extra container.

## Write the decision down — before you talk to anyone

New file, `docs/paso-04b-decision.md`. It is an **ADR** — an architecture decision record; the term is worth knowing, because it is what this artifact is called everywhere you'll be asked to produce one. Sections:

- **Context** — what forge stores, at what volume, read by whom and when. Facts only, two or three sentences, no argument yet.
- **Decision** — where the line fell. Entity by entity, and column by column for the one that splits. Present tense and specific: a reader should be able to write the DDL from your paragraph.
- **Why** — the two questions, answered explicitly, citing your own probes with numbers. A reason that points at neither an invariant nor a measurement is a preference wearing a reason's clothes.
- **What it costs** — the price of *your* decision, stated by you before anyone states it for you. Every real decision has one. A write-up with no costs section is marketing.
- **What would have to be true for the opposite decision to be right** — the line most people skip, and the whole reason this step is ⚖️. Make it checkable: a volume, a read pattern, a latency budget, a team shape, a second writer. "If we scaled a lot" is not checkable. "Above ~X calls/day, or once a second team writes these records" is.
- **Revisit trigger** — the number or the event that means: go read this again.

One rule while writing: **no hedging in both directions.** "It depends" is allowed only if the next sentence says on *what*, and the one after gives the value at which it flips.

## Only then: get it read

Send the ADR, not a question. The brief is *"here's what I decided and why — where am I wrong?"*, and then three specific asks, because an open-ended request earns an open-ended answer:

1. Which of my reasons stops being true at 100× the volume?
2. What breaks first operationally — and would my probes have shown it, or does it only appear in production?
3. Where have you seen this same boundary drawn differently, and what made *that* the right call there?

**The step is done when they have disagreed with at least one thing and you have written the disagreement into the ADR** — either as a change, or as "considered and rejected, because…". An ADR that comes back unamended means you asked too late, asked too gently, or asked someone who doesn't run this.

## What to be able to explain afterwards

- The two questions, and why "rigid vs. flexible" is not one of them.
- Where each invariant in this ledger is enforced today, and what happens to it the moment a second writer exists.
- What `jsonb` in the database you already operate buys you that a separate document store buys you — and the specific point at which that stops being true. Name at least one thing on each side.
- Why updating one key inside a large `jsonb` document is not a cheap operation, and what that implies about which fields belong in the blob.
- How you would produce the month-end bill from your chosen shape without a nightly job that reads every row.
- The one thing your reviewer disagreed with, and where you landed.

## The interview question this answers

*"When do you reach for a relational store, and when for a document store?"* — verbatim, from the interview in [PLAN.md](PLAN.md).

Someone who has read about it says "relational for structured data, document for flexible schemas." That is a sentence about the data, and it dies to the first follow-up. Someone who has done this step says which invariants they needed the store to enforce, what they had to count and what making that cheap cost them, where they drew the boundary *inside a single entity*, and what would have made them choose the other way. That is a sentence about consequences — and consequences are what the question was always asking for.

## When it's decided, commit it yourself

```bash
git add -A && git commit -m "paso-04b: choosing the store (probes + decision record)"
git tag paso-04b
```
