# abcd RFCs

Request-for-comment artefacts for abcd plugin design questions where the answer isn't obvious and reasonable people might disagree.

---

## What's an RFC?

An **RFC** captures a *question* — not a feature, not a decision. It exists to gather input before a decision is made, and the discussion itself is the deliverable. RFCs are used when:

- The question has a real values-tension (not a clear right answer).
- Outside input would change the design, not just ratify it.
- The cost of getting it wrong is real but the cost of waiting for input is acceptable.
- The tone wants to be inviting ("here's our reasoning, where's it weak?") rather than declarative.

RFCs are *not* used for:

- Feature specifications — those are intents (`itd-N`), in press-release format.
- Architectural decisions already made — those are ADRs (`adr-N`), at [`../../decisions/adrs/`](../../decisions/adrs/). ADR format is for retrospective decision records, not forward-looking discussion.
- Internal-only nitpicks — those go in `.work/issues.md` and surface via `/abcd:dredge` later.

---

## RFCs vs ADRs vs intents

Three decision-record surfaces, each with a distinct job:

| Class | Direction | Surface | Use when |
|---|---|---|---|
| **intent** (`itd-N`) | Forward-looking, user-facing | [`../intents/`](../intents/) | Capturing capability the project will ship. |
| **RFC** (`rfc-N`) | Forward-looking, contested | this directory | A decision is contested; discussion is the deliverable. |
| **ADR** (`adr-N`) | Retrospective | [`../../decisions/adrs/`](../../decisions/adrs/) | A decision is settled; record the *why* + alternatives rejected. |

For an intent-driven configuration like abcd, the natural flow is:

```
itd-N (proposed work) ──► rfc-N (community discussion) ──► itd-N (refined or new) ──► fn-N (spec) ──► ship
                                                                                        │
                                                                                        ▼
                                                                              adr-N (retrospective record,
                                                                              when the decision was hard
                                                                              to reverse + the trade-off
                                                                              was real)
```

RFCs are where contested intents go to be *sharpened*. The output of an RFC is either a refined version of the original intent, a different intent, or an explicit "no, and here's why" entry that future RFC-stubs can reference. ADRs are written after settlement, when an inflection-point decision warrants permanent rationale.

RFCs live here under `.abcd/development/roadmap/` because they are roadmap artefacts (forward-looking, decision-shaping). ADRs live at `.abcd/development/decisions/adrs/` because they are settled-decision records (retrospective, decision-preserving). Intents live at `.abcd/development/roadmap/intents/` for the same reason as RFCs.

---

## RFC IDs

RFC IDs follow the pattern `rfc-N` (unpadded — mirrors `itd-N` and `fn-N`). Filenames: `rfc-N-<slug>.md`. Lexical-vs-numeric sort handled at tool layer.

---

## RFC Lifecycle (Status Field)

| Status | Meaning |
|---|---|
| `open` | Discussion period is live. Comments welcome. Resolution section is empty. |
| `resolved-yes` | Discussion closed; the proposed direction is endorsed. Spawned an intent (`spawned_intents` field) or refined an existing one. |
| `resolved-no` | Discussion closed; the proposed direction is rejected. The "no, and here's why" itself is a deliverable — future similar proposals should reference this RFC. |
| `resolved-modified` | Discussion closed; outcome differs from the original question's framing. New direction documented in Resolution section. |
| `withdrawn` | Discussion closed without resolution. Author or maintainers chose not to proceed; reason recorded in Resolution. Distinct from `resolved-no` (no community decision; the question itself was reframed away). |

Status transitions are deliberate. The `discussion_closes` frontmatter field sets a soft deadline; resolution is recorded by editing the RFC's frontmatter and Resolution section.

---

## Format

Every RFC has frontmatter (machine-readable) plus a Markdown body following this structure:

```markdown
---
id: rfc-N
slug: <kebab-case-slug>
status: open                          # see Lifecycle table above
discussion_opened: YYYY-MM-DD
discussion_closes: YYYY-MM-DD-or-TBD
spawned_from: null                    # itd-N if RFC originated from a contested intent
spawned_intents: []                   # [itd-N, ...] populated when status = resolved-yes/modified
related_intents: []                   # [itd-N, ...] cross-references
authors: [project]                    # "project" for maintainer-authored; usernames for community-authored
---

# RFC-N: <Title — phrased as a question or a tension>

## The Question

Single, concrete question being asked. Not a feature spec.

## Why We're Asking

The pull and the tension. What's the impulse, what's the worry?

## What We've Already Decided

Constraints that are NOT up for discussion. This section is load-bearing — it
tells the community what's in scope vs out of scope for the discussion, so
people don't waste energy debating settled points.

## Considered Alternatives

2–4 options laid out fairly, including "do nothing".

## What We're Hoping to Learn

The values-question underneath the feature-question. Not "should we ship this?"
but "what does this discussion reveal about how we balance X and Y?"

## How to Engage

Where to comment. What kind of input is most useful. What we'll do with it.

## Resolution

Empty until status changes. When resolved, this section gets filled with the
discussion summary, the outcome, and any spawned intents (with bidirectional
frontmatter links).
```

---

## Bidirectional Linking

| File | Frontmatter field |
|---|---|
| `rfcs/rfc-N-<slug>.md` | `spawned_from: itd-N` (if the RFC originated from a contested intent) |
| `rfcs/rfc-N-<slug>.md` | `spawned_intents: [itd-N, ...]` (if the RFC's resolution produced new intents) |
| `intents/{drafts,planned,shipped}/itd-N-<slug>.md` | `related_rfcs: [rfc-N, ...]` (when an intent references an RFC) |

`intent_lint.py` (per itd-4 + brief § 11) extends to verify these reciprocally.

---

## Related Documentation

- [Intents](../intents/) — the press-release-format roadmap surface
- [Phases](../phases/) — the ordered build plan
- [Brief § 5](../../brief/README.md) — command and naming conventions (RFC-1 is referenced from `/abcd:loot`'s licence-check non-circumventability rule)
