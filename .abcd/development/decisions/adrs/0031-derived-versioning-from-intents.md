---
id: adr-31
slug: derived-versioning-from-intents
status: accepted
date: 2026-07-07
supersedes: null
superseded_by: null
related_intents: [itd-73]
related_rfcs: []
related_adrs: [adr-19, adr-20, adr-28]
---

# ADR-31: The release version is derived from the intents in it, never authored

## Context

abcd's working tree is unversioned by decision (ADR-19): the plugin version lives only in the released artefact, and the two release manifests are held in lockstep (ADR-20). Those decisions fix *where* the version lives, not *how the number is chosen* — and today that choice is implicitly manual, someone picking the next `vX.Y.Z`.

abcd's product model makes the intent the unit of the *why*, and the launch flow already cuts a curated release from a `v*` tag (ADR-28). A manually-chosen version is release-thinking imposed on a tool whose whole grain is product-thinking in intents. It is also unsafe: a hand-picked SemVer can silently disagree with what actually changed — a "minor" that removed a command breaks the consumers who trusted the contract.

## Decision

The release version is **derived, not authored**. Two signals feed it, and the working tree never carries a version (extending ADR-19):

1. **Intent-declared impact (primary).** Every intent carries `impact: additive | breaking | fix`, set once when the intent is shaped. At release, `/abcd:launch` gathers the intents shipped since the previous release and takes the highest-severity impact: any `breaking` → major, else any `additive` → minor, else patch. `internal/core/lint` enforces that every intent carries a valid `impact`.

2. **Surface diff (guardrail).** `launch` snapshots the `/abcd:*` command, flag, and manifest surface and compares it to the previous release. A removed or altered surface with no `breaking` intent in the release **fails the launch** — a mislabel cannot ship a compatibility lie.

The computed number is written only to the release artefact — the tag, the GitHub Release, and the marketplace manifest (ADR-20 lockstep) — never to the working tree.

## Alternatives Considered

- **Conventional-commit derivation (semantic-release style).** Compute the bump from `feat:` / `fix:` / `feat!:` commit prefixes since the last tag. Rejected as the *primary* signal because the commit is the wrong unit for abcd — the intent is where product judgement lives — but retained as a fallback for changes not tied to an intent.
- **Manual SemVer.** The status quo. Rejected: it is release-thinking, and it lets the number disagree with reality.
- **Pure surface-diff, mechanical only.** Derive the whole version from the surface snapshot with no human input. Rejected: it cannot see *behavioural* breaks behind an unchanged surface, so the human `breaking` judgement is still required; surface-diff is the guardrail, not the sole source.

## Consequences

- Contributors think in intents and mark each `additive`, `breaking`, or `fix`; nobody chooses a version number.
- The derived SemVer keeps its promise to consumers, because the guardrail blocks a minor or patch that structurally broke the surface.
- The intent schema gains an `impact` field, `internal/core/lint` gains an impact-presence check, and `launch` gains version-derivation and surface-diff stages. This implementation is tracked by itd-73.
- Pre-release channels, deprecation windows, and changelog prose are out of scope here and left to later decisions.
