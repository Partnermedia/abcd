---
id: itd-74
slug: name-banlist
spec_id: null
kind: standalone
suggested_kind: null
reclassification_history: []
severity: minor
---

# abcd Keeps the Names You Ban Out of Everything You Publish

## Press Release

> **Name a thing once as off-limits; abcd keeps it out of every published surface — and keeps the truly private ones off the machine's commits entirely.** A repo abcd configures often must not name certain things in what it publishes: a specific agent harness (so the surface stays host-agnostic), a partner's product, or — most sensitively — a *private* project whose very name is confidential. abcd manages this as a two-layer banlist. The **public banlist** is enforced deterministically in CI: named tokens in user-facing content (README, `docs/`, the shipped artefact) fail the build, with a per-line escape hatch for the rare deliberate mention. The **private banlist** is enforced by a **local, untracked guard** — because a name that must never appear *anywhere public* cannot be written into public CI config to ban it there. Its patterns live only on the developer's machine; a pre-commit guard refuses to stage any content that matches, so the string never enters tracked history in the first place.
>
> "The names I can't afford to leak are exactly the ones I can't put in a public linter rule," said Kira, a maintainer. "abcd solved that by splitting it: public names get a CI gate, private names get a local guard whose list never leaves my machine. I stopped worrying that a stray paste would ship a name I'd promised to keep quiet."

## Why This Matters

Two failure modes share one root. First, a tool that claims to be *host-agnostic* undermines itself the moment its published docs name a specific harness — the naming dates the content and couples the surface. Second, and worse, a private collaborator's or project's name leaking into a public repo is a confidentiality breach that a history rewrite alone cannot fully undo (merged-PR diffs and cached views persist server-side). Both are cheap to prevent and expensive to remediate. The lesson learned the hard way on abcd's own repo: **prevent at authoring time, and never let the sensitive string reach a public artefact — including the linter config meant to catch it.**

The design tension is the interesting part: a deterministic CI gate is the right tool for *public* banned names, but it is the *wrong* place for a *private* one, because the rule would have to contain the very string it forbids. Splitting enforcement by sensitivity — public names in CI, private names in a local untracked guard — resolves it without compromise.

## What It Looks Like

- **`abcd` manages both layers as first-class config.** A public banlist (patterns + per-token severity + the allow-context escape) compiles into the deterministic docs-currency lint family; a private banlist is scaffolded as an untracked, per-machine file plus a committed guard hook that reads it. The literal private strings never enter tracked content or CI config.
- **Wired into install.** `abcd ahoy` scaffolds both surfaces for any repo it configures: the public family in the docs-lint config, the guard hook in the repo's committed hooks, and a gitignored banlist stub with instructions. A repo becomes name-safe by being abcd-managed, not by hand-rolling hooks.
- **Verbs to maintain it.** Add, list, and remove banned patterns; the public ones are visible and reviewable, the private ones addressed by reference (never printed into a shared artefact).
- **Reports what it cannot enforce.** The private guard is local by construction, so abcd states plainly that CI cannot enforce it — the guard protects the machine that opted in, and the public flip is gated on a from-scratch name scan.

## Dogfood (already running on this repo)

The concrete prototype this intent generalises is live in abcd-cli itself: a `harness/*` banned-token family in the docs-currency lint (public agent-harness names, blocker, README + `docs/` scanned), a committed `.githooks/pre-commit` guard that reads an untracked banlist for the private name, and the gitignored banlist that holds it. The feature is to lift that from a hand-wired arrangement into an abcd capability every managed repo inherits. Relates to the host-agnostic documentation principle and to the install surface (`ahoy`).

## Open Questions

- Config shape: one banlist file with a public/private split, or two files with different visibility and tooling?
- Whether the public family belongs in the docs-lint config verbatim or in a dedicated `names` rule with richer reporting.
- How `ahoy` seeds the private banlist without ever suggesting example private strings that could themselves leak.
