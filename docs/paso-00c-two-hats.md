# paso-00c · Two hats

**Goal:** write a contract that a stranger could implement without asking you anything, then be that stranger — and feel exactly where your own contract fails you.

**Gap it closes:** forge already runs a client/engineer split on every step from `paso-00b` onward — the AI writes the contract, you implement it blind to nothing more than the contract states. But you've only ever worn the engineer hat. This step came out of a conversation with a staff backend engineer about an exercise almost nobody runs deliberately: define the interface and the acceptance tests wearing a *client* hat — no implementation, nothing else — then implement against them wearing an *engineer* hat, with no more context than the contract gives. The comparison that stuck: it's exactly how a secretive engineering org — the kind that can't hand a contractor the full picture — has to scope work. The wall isn't bureaucracy; it's what makes the contract carry all the weight instead of the conversation around it.

> **This step is inverted like `paso-00b`, but the AI's role is smaller here.** I write this write-up — the exercise's rules and mechanics — and nothing else. I do not write the interface, the acceptance tests, or the spec they enforce. That's yours to produce in Phase 1, wearing the client hat. See [COLOPHON.md](../COLOPHON.md).

## Do it wrong first — and notice it in yourself

The tempting version isn't bad code, it's a bad *sequence*: you sit down already picturing roughly how you'll implement the thing, and the "interface" you write is really just a narration of that implementation with the bodies stripped out. It compiles, it looks like a contract, and it teaches you nothing — because nothing was ever actually hidden from you. The tell is retroactive: if, once you put the engineer hat on, implementing feels like *transcribing* rather than *deciding*, the wall wasn't real. Notice that, and don't fix it by redoing the exercise "more strictly" — write down where it happened. That observation is worth more than a clean run.

**That tell is introspective, which makes it weak** — it asks you to catch yourself, and the whole reason the failure is worth studying is that it does not feel like failure while it happens. So this step adds a *hard* test alongside it: **write the contract without committing to a language, then land it in one.** A contract you wrote while picturing goroutines, or `IAsyncEnumerable`, or a particular collection type, will resist the moment you try to state it in something else — and that resistance is visible in the diff rather than in your memory of your own intentions. Phase 1 is split in two for exactly this reason.

## Phase 1 — client hat

Pick one small piece of behavior forge doesn't already implement. It does not need to touch the ledger's actual domain — a good candidate is deliberately disposable and self-contained: a rate limiter, a retry-with-backoff policy, an in-memory cache with eviction, an idempotency-key store. Anything with a real behavioral contract *and* at least one edge case that isn't obvious from the happy path. Keep it small enough for one sitting — the same rule as `paso-04b`: if it's taking three sittings, you've started building a product instead of running a probe. Put it somewhere disposable, e.g. `exercises/00c-<name>/`, its own tiny project — nothing here needs Postgres, Docker, or wiring into `src/Forge.Api`.

### Phase 1a — the contract, language-free

Two artifacts, both in prose, in `exercises/00c-<name>/CONTRACT.md`. **No code, no types, no language chosen yet.**

1. **A one-paragraph spec** stating only *what* the contract guarantees — never *how*. If a sentence describes a data structure or an algorithm, that is leaked implementation; rewrite it as a guarantee instead.
2. **Every acceptance case, as behavior.** Given / when / then, in words. All of them, including the edge case that isn't obvious from the happy path. If you cannot state a case without naming a type or a collection, that case is not yet a case — it is a design note wearing one.

**Start simple, then complicate it.** State the happy path first. Once it holds, add the harder capability on top — a concurrency constraint, an error mode you'd been avoiding, a bound on time or memory. Both stages stay in prose.

This is the artifact that would survive being handed to a team that uses neither of your languages. That is the bar.

### Phase 1b — landing it

Now, and only now, pick a language (see [Choosing the language](#choosing-the-language) below) and translate 1a into it:

3. **Interface only.** Method signatures and doc comments describing observable behavior — inputs, outputs, error modes, pre/post-conditions. Every method body is the language's not-implemented escape hatch: `throw new NotImplementedException()` in C#, `panic("00c: not implemented")` in Go. If you catch yourself typing logic, stop.
4. **Write every acceptance test**, one per case in 1a, against the interface, before any implementation exists. They will not compile or will panic until Phase 2 supplies a real type — that's expected; the suite must be red for the right reason.

**The measurement is in this translation.** Every time landing 1a forces you to *change* 1a — a guarantee that turns out to be unstateable, a case that splits in two, an error mode that has no home — that is a leak, and it is the finding this step exists to produce. Note each one where it happens, with the sentence before and after. Do not quietly fix 1a and move on.

Commit the two sub-phases separately, so the diff between them *is* the evidence:

```bash
git commit -m "paso-00c: client hat 1a — language-free contract for <name>"
git commit -m "paso-00c: client hat 1b — <lang> interface + acceptance tests"
```

## Choosing the language

`paso-00c` is the one step in forge with no language assigned by the roadmap — `00`–`01` are .NET, `02` onward is Go, and this sits in the seam. That is not an oversight to be tidied up; it is the only step where the choice is *yours* and therefore the only one where making it deliberately teaches anything.

The exercise here is disposable and tiny — a rate limiter, a cache — so the language is a vehicle, not the content. Which frees the choice to be made on a different basis: **where is the diagnostic worth more?**

Two candidates, two different kinds of discomfort:

- **Go — rusty.** Learned and practiced by hand, before assistance was an option. That competence doesn't vanish, it oxidizes, and it comes back fast.
- **.NET — assisted.** Read fluently (the similarity to TypeScript does a lot of work here) but never written cold. **Reading fluency is passive comprehension and it does not predict writing from a blank file.**

**Use `paso-01` to decide, because it already runs the experiment.** That step is implemented on *both* lines — same lesson, same scope, back to back. Time each side roughly, and record what you had to look up, sorting it into two buckets: *lexical* ("what's this method called") heals on its own; *design* ("how do I model this error here") is the real gap.

Then:

- **Times and stalls comparable** → the choice doesn't matter; take whichever is convenient.
- **Go clearly more fluent** → do `00c` in **.NET**. It is the language on the CV as paid production work, so it is where an unassisted gap costs the most — and a two-hour disposable exercise is the cheapest possible place to find that out, rather than in an interview.

One rule that makes the measurement mean anything, in `paso-01` and here: **documentation yes, model no.** Consulting the .NET or Go docs is what you would do at work. Asking a model for the body of the method destroys the reading — and it is precisely the habit the assisted line was built on.

Whichever you pick, **write the reason in one line at the top of `CONTRACT.md`.** In Phase 3 you will want to know whether it was the language or the contract that hurt.

## The wall

Let it go cold before Phase 2 — close the editor, let a day pass if you can, genuinely try to lose the mental image of "how I was going to build this." A contractor who receives your interface tomorrow has zero access to what you were picturing today; the exercise only produces its lesson if you approximate that. If you can't wait a day, at minimum switch to unrelated work for an hour first.

## Phase 2 — engineer hat

Implement strictly against the interface, the acceptance tests, and the one-paragraph spec — nothing else. Two rules:

- **No amending the contract.** Not the signatures, not the spec. If a test seems impossible to satisfy as written, that's a finding, not a bug in the exercise — write down the exact ambiguity instead of silently resolving it in the interface's favor. This is the same discipline as *"ante un rojo, preguntar por qué, nunca pegar el arreglo"* — except now you're both sides of that conversation, so the "question" goes into your notes instead of to me.
- **No re-reading anything from Phase 1 beyond the contract, the tests, and the spec.** No design notes, no half-formed intentions you remember having. If you don't remember why you wrote a method the way you did, that's data, not a problem to route around. The leak notes from 1b count as design notes — do not open them until Phase 3.

Get to green using only what the contract gives you.

## Phase 3 — blue: harden it

Once green, put the client hat back on and raise the bar: a test on execution time, a concurrency case you deliberately excluded in Phase 1, a boundary condition you now realize the first spec left open. Then put the engineer hat back on and satisfy those too, under the same two rules as Phase 2. This is the loop your friend meant by "red-green-blue" — not a synonym for TDD, a description of what TDD does on repeat.

### Run it

```bash
# .NET
dotnet test exercises/00c-<name>          # red at the end of Phase 1, green at the end of Phase 2/3

# Go
go test ./exercises/00c-<name>/...        # same contract, same red→green
```

Deliberately outside `Forge.slnx` and outside the `go/` module either way: this is a probe, not part of the ledger, and it should be deletable without touching anything that matters.

## What to be able to explain afterwards

- The one place your own interface leaked an implementation detail — there is always at least one. Name it and say what a cleaner version would have looked like.
- **Where 1a resisted being landed in 1b** — the guarantee that turned out to be unstateable, or the case that split in two. That is the same leak as the point above, caught by measurement instead of by introspection; say which of the two found it first.
- The exact sentence in your spec that was ambiguous when Phase 2 needed it to be precise. What would you have had to know on day one to close that gap?
- **Whether the language was ever actually the problem.** If Phase 2 was hard, separate the two causes: a contract that didn't carry enough, versus a language you can read but were not fluent writing. They feel identical from the inside and they have completely different fixes.
- What changes about how you read a contract the AI hands you in every other step of forge, now that you've had to write one that binds *you*.
- Where the "need-to-know" framing predicted the hardest part of the exercise, and whether it actually was.

## The interview question this answers

*"How do you write a spec that someone else — or an agent — can implement without more context than the spec gives them?"* Not asked verbatim anywhere, but underneath every question about API contracts, service boundaries, and — directly — `paso-12`: an agent that chooses a route has to work from exactly this kind of contract, with no more context than it's given, and the same rule about the wall applies to it that applied to you here.

## When it's decided, commit it yourself

```bash
git add -A && git commit -m "paso-00c: two hats — <name>, contract + implementation"
git tag paso-00c
```
