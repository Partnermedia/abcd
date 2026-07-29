---
id: itd-74
slug: name-banlist
spec_id: spc-20
kind: standalone
suggested_kind: null
reclassification_history: []
severity: minor
impact: additive
---

# abcd Keeps the Names You Ban Out of Everything You Publish

## Press Release

> **Name a thing once as off-limits; abcd keeps it out of every published surface — and keeps the truly private ones off the machine's commits entirely.** A repo abcd configures often must not name certain things in what it publishes: a specific agent harness (so the surface stays host-agnostic), a partner's product, a *private* project whose very name is confidential, or — most sensitively — the user's own machine identifiers: hostnames, device names, and the IPs or address prefixes of their private network. abcd manages this as a two-layer banlist. The **public banlist** is enforced deterministically in CI: named tokens in user-facing content (README, `docs/`, the shipped artefact) fail the build, with a per-line escape hatch for the rare deliberate mention. The **private banlist** is enforced by a **local, untracked guard** — because a name that must never appear *anywhere public* cannot be written into public CI config to ban it there. Its patterns live only on the developer's machine; a pre-commit guard refuses to stage any content that matches, so the string never enters tracked history in the first place.
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

## Acceptance Criteria

> _BDD format, per the itd-1 discipline. Confirmed by the maintainer in the 2026-07-29 planning interview._

- **Given** a repo whose public banlist bans token X, **when** docs-lint runs over user-facing content (README, `docs/`, the shipped artefact) containing X without the allow escape, **then** a blocker finding names the file and line; **given** the line carries the escape, **then** no finding is raised.
- **Given** a private banlist entry with key K and pattern P, **when** a commit stages content matching P, **then** the commit is refused and the refusal message names only K — the matched string and the pattern value are never echoed to any output.
- **Given** a private banlist containing a hostname, an IP or CIDR prefix, or a device name, **when** staged content matches one, **then** the commit is refused exactly as for a name entry — machine identifiers are first-class private entries.
- **Given** a fresh clone where the private banlist file does not exist, **when** the guard hook runs, **then** it prints a loud warning that the private layer is inactive on this machine and does not block the commit — never a silent pass that looks like protection.
- **Given** a repo configured by `abcd ahoy`, **when** scaffolding completes, **then** the public family is present in the docs-lint config, the guard hook is committed, the private banlist stub exists and is gitignored, and the stub's seeded examples use only reserved documentation values (RFC 5737/3849/2606/7042 ranges, persona-derived hostnames) per the `examples-use-reserved-identifiers` principle.
- **Given** the banlist maintenance verbs (add, list, remove), **when** entries are rendered, **then** public entries render in full and private entries render by key only.
- **Given** any status or report surface describing the private layer, **when** it renders, **then** it states plainly that CI cannot enforce the private list — it protects only machines that have opted in.

## Resolved Questions (2026-07-29 planning interview)

- **Config shape: two files.** The public banlist is committed beside the docs-lint config; the private banlist is a separate gitignored file with its own tooling. A single split file makes the whole file's visibility load-bearing — accidental commit of the private half is the exact failure mode this intent exists to prevent.
- **Public-layer home: generalise the docs-lint family.** The existing `harness/*` banned-token family is lifted into a general banned-names family — one canonical primitive, ratchet not big-bang. A dedicated rule kind with richer reporting is future work, not this intent.
- **Private-stub seeding: reserved documentation values only.** The `examples-use-reserved-identifiers` principle answers the seeding question — the stub's examples come from RFC-reserved ranges and persona-derived hostnames, so the scaffold can never suggest a plausible real value.
- **Scope: machine identifiers included.** Hostnames, device names, and IP/CIDR entries are first-class private-banlist entries (evidence: iss-158, the 2026-07-29 managed-repo NEXT.md privacy-leak investigation).

## Dogfood (already running on this repo)

The concrete prototype this intent generalises is live in abcd-cli itself: a `harness/*` banned-token family in the docs-currency lint (public agent-harness names, blocker, README + `docs/` scanned), a committed `.githooks/pre-commit` guard that reads an untracked banlist for the private name, and the gitignored banlist that holds it. The feature is to lift that from a hand-wired arrangement into an abcd capability every managed repo inherits. Relates to the host-agnostic documentation principle and to the install surface (`ahoy`).

## Open Questions

_None — all three original questions resolved in the 2026-07-29 planning interview (see Resolved Questions)._
