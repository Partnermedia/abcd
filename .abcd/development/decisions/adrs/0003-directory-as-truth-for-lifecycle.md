---
id: adr-3
slug: directory-as-truth-for-lifecycle
status: accepted
date: 2026-05-07
supersedes: null
superseded_by: null
related_intents: [itd-34]
related_rfcs: []
related_adrs: [adr-2]
---

# ADR-3: Directory location is the source of truth for lifecycle state

## Context

After ADR-2 introduced three intent kinds, two field-vs-directory inconsistencies surfaced:

1. **Disciplines initially carried `status: active`** in frontmatter. But the `disciplines/` directory already encoded that — there is no `inactive` discipline state. Field and directory said the same thing twice.
2. **For standalone/bundle-member intents**, `status` was historically a positional state in a sequence (`draft` / `planned` / `shipped`). For disciplines, `status: active` was a type predicate doubling for one of two directories (`disciplines/` vs `superseded/`). Different jobs, same field name — a smell.
3. **Superseded intents** initially recorded only `superseded_by: <itd-M>`. But "superseded" means different things depending on what the intent *was*: a superseded standalone is a retired capability; a superseded discipline is a retired rule. Without recording the original kind, future readers (and the shape-classification auditor) had to infer it from context.

Two failure modes followed: field-vs-directory drift (an intent could in principle have `status: active` in the wrong directory and `internal/core/lint` had no clear contract to assert) and lossy supersession (kind context erased on retirement).

## Decision

**Directory location is the single source of truth for lifecycle state across all intent kinds.**

- Standalone and bundle-member intents derive lifecycle state from `drafts/` / `planned/` / `shipped/`.
- Disciplines derive state from `disciplines/` / `superseded/`.
- **No intent carries a `status:` frontmatter field.** Disciplines drop it entirely (it was redundant); standalone/bundle never had a `status:` field that disagreed with directory.
- Superseded intents preserve `kind_at_supersession: <original-kind>` so the shape the intent had when retired stays legible.

`internal/core/lint` enforces:
- No file in `drafts/` / `planned/` / `shipped/` carries a `kind: discipline` value.
- No file in `disciplines/` / `superseded/` carries a `status:` field.
- Every file in `superseded/` carries both `superseded_by:` and `kind_at_supersession:`.

## Alternatives Considered

1. **Keep `status:` field as the source of truth; make directory advisory.** Rejected: filesystem moves are cheap, atomic, and self-documenting in `git log`; field edits are easy to forget and create silent drift. Directory-first matches the existing intent lifecycle (`/abcd:intent plan` already moved files between directories) — making the field authoritative would have required a second mechanism.
2. **Keep both `status:` and directory; require lint to verify they agree.** Rejected: redundancy is a smell; the lint rule would only ever catch discrepancies that a single source of truth prevents structurally. The two-source design encourages each writer to update one and forget the other.
3. **Drop `status:` only from disciplines; keep it for standalone/bundle.** Rejected: the principle is now non-uniform. Two kinds use directory-as-truth, one kind uses field-as-truth. Future readers and tooling must know which kind they're looking at to know which to trust. Uniformity is the gain; partial application loses it.

## Consequences

**Gains:**
- One source of truth per concern. `git log` shows lifecycle transitions as file moves; no parallel field history to reconcile.
- The principle is uniform across kinds, applies consistently to future kinds, and generalises beyond intents (disembark snapshots, lifeboat artefacts, etc.).
- `internal/core/lint` contract is simpler (assert directory matches kind, not directory matches field matches kind).

**Costs / obligations:**
- Manual `mv` operations outside `/abcd:intent` verbs can put a file in the wrong directory; lint catches at pre-commit but the moment-of-error is the moment of move, not the moment of verb. Mitigated by `/abcd:intent reclassify` being the only sanctioned move path, with `reclassification_history` written in frontmatter.
- The `status:` field is reserved vocabulary for non-intent surfaces (e.g., RFCs still use `status: open|resolved-yes|...`). Naming registry must distinguish — clarified in `02-constraints/04-naming.md`.

**Downstream consequences:**
- ADR-5 (brief is current state) extends the same principle to the brief itself: the live `brief/` directory IS the brief; no parallel `version:` field, no parallel `archive/<NN>/` directories duplicating git history.
- The issue ledger (spc-20, itd-4) applies the same pattern: the `.abcd/work/issues/{open,resolved,wontfix}/iss-N-<slug>.md` ledger carries no `status:` field; the sub-directory IS the lifecycle state. A `wontfix_reason:` or `resolution:` field is required in the corresponding directory (enforced by `_issue_lib._validate_invariants`, not by schema location alone), keeping per-state context legible.

## Worked examples

| Lifecycle | Directories | Required-by-location field(s) | Enforced by |
|---|---|---|---|
| Intents (standalone / bundle-member) | `drafts/` / `planned/` / `shipped/` | — | `internal/core/lint` |
| Disciplines | `disciplines/` / `superseded/` | `superseded_by:` + `kind_at_supersession:` in `superseded/` | `internal/core/lint` |
| Issue ledger | `issues/{open,resolved,wontfix}/` | `resolution:` in `resolved/`; `wontfix_reason:` in `wontfix/` | `_issue_lib._validate_invariants` |
