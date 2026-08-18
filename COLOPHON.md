# Colophon

`forge` is a learning repository. The point is the reps, so it matters who did what.

**Human (Alan Staub Negro)** owns the learning and the decisions: which fundamental each step targets, doing it wrong first and feeling the failure, reading the query plans, and being able to defend every choice out loud. The competence this repo builds is the human's, not the tooling's. A step is not "done" until the human can explain the failure and the fix *without help*.

**AI (Claude Code)** drafts scaffolding, boilerplate, prose, and test-harness structure under direction, and pairs on reproducing and fixing each gotcha. It accelerates the typing, not the understanding.

This division is deliberate, and it is itself part of the point: the AI-native way to get stronger at a fundamental is to let the AI carry the *scaffolding* while you keep the *judgment*.

## Attribution, per step

Blanket credit for a repository is worthless, so the split is recorded step by step.

- **`paso-00`** is the **worked reference example**: the AI wrote the implementation so there would be one solved step to read for shape and conventions.
- **`paso-00b` onward are inverted.** The AI writes the write-up, the contract, and an executable spec that fails; the *implementation is the human's*, typed by hand. The AI reviews and answers with questions, and does not supply the fix. Where a step is inverted, its write-up says so.

The honest sentence about this repo is therefore "I designed and directed it, and from `paso-00b` I implement it by hand against specs I asked for" — not "I hand-built a .NET backend."

Related work by the same author: [gotcha](https://github.com/AslanSN/gotcha) — a catalog of subtly-wrong backend code, with an MCP server and evals. `forge` is its constructive twin: where `gotcha` catalogs the traps, `forge` builds the way out of each one by hand.
