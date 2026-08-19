# paso-00c · Two hats

**Goal:** write a contract that a stranger could implement without asking you anything, then be that stranger — and feel exactly where your own contract fails you.

**Gap it closes:** forge already runs a client/engineer split on every step from `paso-00b` onward — the AI writes the contract, you implement it blind to nothing more than the contract states. But you've only ever worn the engineer hat. This step came out of a conversation with a staff backend engineer about an exercise almost nobody runs deliberately: define the interface and the acceptance tests wearing a *client* hat — no implementation, nothing else — then implement against them wearing an *engineer* hat, with no more context than the contract gives. The comparison that stuck: it's exactly how a secretive engineering org — the kind that can't hand a contractor the full picture — has to scope work. The wall isn't bureaucracy; it's what makes the contract carry all the weight instead of the conversation around it.

> **This step is inverted like `paso-00b`, but the AI's role is smaller here.** I write this write-up — the exercise's rules and mechanics — and nothing else. I do not write the interface, the acceptance tests, or the spec they enforce. That's yours to produce in Phase 1, wearing the client hat. See [COLOPHON.md](../COLOPHON.md).

## Do it wrong first — and notice it in yourself

The tempting version isn't bad code, it's a bad *sequence*: you sit down already picturing roughly how you'll implement the thing, and the "interface" you write is really just a narration of that implementation with the bodies stripped out. It compiles, it looks like a contract, and it teaches you nothing — because nothing was ever actually hidden from you. The tell is retroactive: if, once you put the engineer hat on, implementing feels like *transcribing* rather than *deciding*, the wall wasn't real. Notice that, and don't fix it by redoing the exercise "more strictly" — write down where it happened. That observation is worth more than a clean run.

## Phase 1 — client hat

Pick one small piece of behavior forge doesn't already implement. It does not need to touch the ledger's actual domain — a good candidate is deliberately disposable and self-contained: a rate limiter, a retry-with-backoff policy, an in-memory cache with eviction, an idempotency-key store. Anything with a real behavioral contract *and* at least one edge case that isn't obvious from the happy path. Keep it small enough for one sitting — the same rule as `paso-04b`: if it's taking three sittings, you've started building a product instead of running a probe. Put it somewhere disposable, e.g. `exercises/00c-<name>/`, its own tiny project — nothing here needs Postgres, Docker, or wiring into `src/Forge.Api`.

Rules, and they're the whole point, so don't relax them:

1. **Interface only.** Method signatures and doc comments describing observable behavior — inputs, outputs, error modes, pre/post-conditions. Every method body is `throw new NotImplementedException()`. If you catch yourself typing logic, stop.
2. **Start simple, then complicate it.** One interface, a handful of methods, for the happy path. Once that's stated, add the harder capability on top — a second interface, a concurrency constraint, an error case you'd been avoiding. Both stages stay contract-only.
3. **Write every acceptance test.** All of them, against the interface, before any implementation exists. They will not compile until Phase 2 creates a class — that's expected; a factory method or a `NotImplementedException` stub is enough to make the suite red for the right reason.
4. **A one-paragraph spec**, prose, stating only *what* the contract guarantees — never *how*. If a sentence describes a data structure or an algorithm, it's leaked implementation; rewrite it as a guarantee instead.

Commit Phase 1 on its own: `git commit -m "paso-00c: client hat — contract for <name>"`.

## The wall

Let it go cold before Phase 2 — close the editor, let a day pass if you can, genuinely try to lose the mental image of "how I was going to build this." A contractor who receives your interface tomorrow has zero access to what you were picturing today; the exercise only produces its lesson if you approximate that. If you can't wait a day, at minimum switch to unrelated work for an hour first.

## Phase 2 — engineer hat

Implement strictly against the interface, the acceptance tests, and the one-paragraph spec — nothing else. Two rules:

- **No amending the contract.** Not the signatures, not the spec. If a test seems impossible to satisfy as written, that's a finding, not a bug in the exercise — write down the exact ambiguity instead of silently resolving it in the interface's favor. This is the same discipline as *"ante un rojo, preguntar por qué, nunca pegar el arreglo"* — except now you're both sides of that conversation, so the "question" goes into your notes instead of to me.
- **No re-reading anything from Phase 1 beyond the contract, the tests, and the spec.** No design notes, no half-formed intentions you remember having. If you don't remember why you wrote a method the way you did, that's data, not a problem to route around.

Get to green using only what the contract gives you.

## Phase 3 — blue: harden it

Once green, put the client hat back on and raise the bar: a test on execution time, a concurrency case you deliberately excluded in Phase 1, a boundary condition you now realize the first spec left open. Then put the engineer hat back on and satisfy those too, under the same two rules as Phase 2. This is the loop your friend meant by "red-green-blue" — not a synonym for TDD, a description of what TDD does on repeat.

### Run it

```bash
dotnet test exercises/00c-<name>   # red at the end of Phase 1, green at the end of Phase 2/3
```

## What to be able to explain afterwards

- The one place your own interface leaked an implementation detail — there is always at least one. Name it and say what a cleaner version would have looked like.
- The exact sentence in your spec that was ambiguous when Phase 2 needed it to be precise. What would you have had to know on day one to close that gap?
- What changes about how you read a contract the AI hands you in every other step of forge, now that you've had to write one that binds *you*.
- Where the "need-to-know" framing predicted the hardest part of the exercise, and whether it actually was.

## The interview question this answers

*"How do you write a spec that someone else — or an agent — can implement without more context than the spec gives them?"* Not asked verbatim anywhere, but underneath every question about API contracts, service boundaries, and — directly — `paso-12`: an agent that chooses a route has to work from exactly this kind of contract, with no more context than it's given, and the same rule about the wall applies to it that applied to you here.

## When it's decided, commit it yourself

```bash
git add -A && git commit -m "paso-00c: two hats — <name>, contract + implementation"
git tag paso-00c
```
