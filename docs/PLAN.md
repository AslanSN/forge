# Working plan — August 2026

This is the *personal* execution plan for `forge`: which pasos to build first, which to skip,
and how much time this is allowed to take. [ROADMAP.md](ROADMAP.md) is the curriculum;
this file is the order I actually walk it, and why.

## Where this came from

August 2026, a systems-architecture interview. The role was a full-stack AI-native position whose
**seniority axis was backend architecture**, and the questions were:

1. When do you reach for a relational store, and when for a document store?
2. What is a serverless function actually for, and when do you *not* reach for one?
3. Queues — the different kinds, and how you manage them.
4. Kafka.
5. All of it on a managed cloud.
6. Orchestrating which service an agent invokes (image rendering vs. schema vs. text).

The product behind the questions: a document-processing platform where an LLM sits **server-side**
and decides which microservice to invoke for a user's request. So the questions were not a
checklist — they were one question asked five ways: *where does the agent's decision go, and what
keeps it from corrupting state when the downstream work is slow, bursty and fallible?*

I said honestly that this isn't my ground today, and the process ended there. Everything *else*
in the conversation landed. So the gap is single-axis, bounded and learnable — which makes it
worth a plan rather than a shrug.

## The finding that reorders the roadmap

Cross-referencing those questions against the existing roadmap:

| Asked about | Covered by `forge` today? |
|---|---|
| Relational vs. document store | ❌ Not in the roadmap at all |
| Serverless vs. container vs. worker | ❌ Not in the roadmap at all |
| Queues as buffers | ⚠️ Partial — paso-07 is Kafka, which answers a different question |
| Kafka | ✅ paso-07 |
| Running it on a managed cloud | ❌ Everything is local Docker Compose |
| Orchestrating an agent's service calls | ❌ No AI layer exists |

And in the other direction, `forge` goes *deep* on transactions, isolation, idempotency, the
outbox and eventual consistency.

> **`forge` is built on the axis of correctness. The interview probed the axis of distribution
> and platform.**

Same family, different question. Both matter — but following the roadmap in numerical order
would produce someone who can defend a ledger against a race condition and still can't answer
question 1.

## The principle that sets the order

**Judgment pasos pay out in hours. Construction pasos pay out in weeks.**

Choosing a store, choosing a compute model and knowing which queue semantics apply are things I
can hold and defend after a few hours of deliberate work with one small piece of code to anchor
them. Kafka, the cloud seam and the projections are builds. In an interview both get asked, but
only the first kind can be *closed* on a short horizon — so the judgment pasos go first, even
though they are less satisfying to build.

## Priority order

Pasos marked **⚖️** are the judgment ones — they do not get to count as done on a green test
alone, they need a second opinion first. The convention, and why the order of *decide → then ask*
is not optional, is in [ROADMAP.md](ROADMAP.md#️-pasos-that-need-a-second-opinion). Of the four,
**04b and 11 are the ones to spend a reviewer's time on first**: they are pure criterion, they are
cheap to review, and they are the two questions I could not answer cold. 06b can wait until there
is code to point at; 12 is the capstone and gets read last.

**1. paso-04b ⚖️ — choosing the store** ⭐
Highest value per hour in the whole list, because it is *judgment*, not construction: where the
relational model saves you, where it taxes you, what a document store actually buys and what it
charges. Put one genuinely document-shaped entity into the ledger and write down why. The real
criterion to land: not "rigid vs. flexible" but **which invariants do I need the store to enforce
for me**, and **what am I going to have to count in order to bill for it**.

**2. paso-02 + paso-03 — concurrency and idempotency, as one block**
Not a detour: at-least-once delivery is the whole reason queue consumers must be idempotent, so
these are the prerequisite for paso-06b rather than a parallel interest. Build them together and
keep moving — `SELECT FOR UPDATE`, isolation levels, a concurrency test that genuinely fails
before it passes, then the unique constraint + idempotency key.

**3. paso-06 + paso-06b ⚖️ — outbox, then queues as buffers**
The single biggest gap. Build the **queue as a buffer** as its own paso, not folded into Kafka:
work queue vs. fan-out, at-least-once, visibility timeout, retry with backoff, DLQ, backpressure,
priority lanes. They answer a different question from streaming and the distinction is exactly
what gets probed.

**4. paso-12 ⚖️ vertical slice** ⭐ — *thin, and early*
See "The endgame" below. Once 03 + 06 + 06b exist, the agent seam is buildable as a thin slice
with no Kafka and no cloud. This is the first point where there is something to *show* rather than
describe, so it does not wait for the foundations to be complete.

**5. paso-11 ⚖️ — the compute seam**
Judgment again, so it is cheap: serverless function vs. container vs. always-on worker. Hard
timeouts, cold starts, CPU-bound work, concurrency per instance. The wrong-first writes itself —
run the heavy job in a serverless function and watch it hit the wall.

**6. paso-07 — Kafka**, then **paso-11b — the cloud seam**, then the rest of the roadmap.

### On paso-11b: build on one cloud, learn all three

The interview was on GCP, but tying the paso to one provider makes the knowledge rot the moment
the next posting says AWS. Build the seam **once**, on whichever provider is cheapest and fastest
to stand up — and write the **primitive mapping table** as part of the paso, because the mapping
is the part that actually transfers:

| Primitive | GCP | AWS | Azure |
|---|---|---|---|
| Managed PostgreSQL | Cloud SQL | RDS / Aurora | Azure Database for PostgreSQL |
| Container, scale-to-zero | Cloud Run | App Runner / Fargate | Container Apps |
| Serverless function | Cloud Functions | Lambda | Azure Functions |
| Pub/sub messaging | Pub/Sub | SNS + SQS | Service Bus / Event Grid |
| Work queue | Cloud Tasks | SQS | Storage Queues / Service Bus |
| Log-structured stream | Managed Kafka / Pub/Sub | MSK / Kinesis | Event Hubs |
| Object storage | GCS | S3 | Blob Storage |
| Workload identity | Workload Identity | IAM roles | Managed Identity |

The paso's real content is provider-independent anyway: managed identity instead of secrets,
connection limits against a pooled database, cold starts, and **what the cloud silently changes
about pasos 02–07**.

### Deliberately skipped

- **paso-04 (indexing / `EXPLAIN`)** — already demonstrated **in production**: PostgreSQL trigram
  GIN indexes verified with `EXPLAIN` on ~500k rows, sequential scan → ~3 ms. This is the one
  item on the whole list I can already defend. No investment where I already win.
- **paso-05 (EF Core)** — used professionally. Zero learning delta.

Both stay in the roadmap for anyone else walking it; they're just not on my path.

## The endgame

Question 6 — *which service should the agent invoke* — is not a classical architecture question.
It is the intersection of `forge` and [gotcha](https://github.com/AslanSN/gotcha), and it is the
one place where I am already ahead rather than behind: I have already built an LLM router that
classifies a request's intent and dispatches it to a specialised model. What I have not built is
the substrate underneath it.

**paso-12, the agent seam:** a router where an agent makes the routing decision, but the decision
is recorded idempotently, published through the outbox, and a failure in the downstream service
cannot corrupt state. That is *systems architecture for AI systems* — and far fewer people hold
both halves than hold Kafka alone.

It used to sit at the end of the road. It now has a **thin slice pulled forward to position 4**,
because it only needs idempotency, an outbox and a queue — not Kafka, not a cloud account:

> agent picks a route → the choice is written with an idempotency key → published through the
> outbox → a worker consumes it → the downstream call fails on purpose → nothing is lost, nothing
> is double-charged, and a replay of the same message changes nothing.

That is the demo. The full paso-12 still lands at the end, once Kafka and the cloud seam are real.

## Budget

**4–6 h/week. One paso every 2–3 weeks.**

This does not compete with the job search. `forge` is not what lands the next job — the next job
comes from the axis that already works: product, full-stack, AI-native, judgment. This is for the
job *after*, and for retiring one specific recurring rejection.

**Honest horizon:** the thin slice at position 4 is roughly **8–10 weeks** at this budget, or
**3–4** if front-loaded. Anything promising a working router in two weeks is promising the demo
without the foundations, which is the same mistake in a new costume.

The short-term payoff arrives much earlier than that, though: after 04b alone, "how are you on
architecture?" stops having a naked answer and becomes *"not my ground yet — I'm building it by
hand; here's the repo, and here's why I put this entity in a document store and kept that one
relational."* A gap becomes a trajectory.

## Immediate next actions

- [ ] Commit paso-00b — finished, sitting uncommitted in the working tree since 22 July.
- [ ] Start paso-04b.
