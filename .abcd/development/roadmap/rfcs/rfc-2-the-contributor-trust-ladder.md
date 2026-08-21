---
id: rfc-2
slug: the-contributor-trust-ladder
status: open
discussion_opened: 2026-08-19
discussion_closes: TBD
spawned_from: adr-43
spawned_intents: []
related_intents: []
related_adrs: [adr-43]
authors: [project]
---

# RFC-2: The contributor trust ladder — what earns a rung, and when

## The question

adr-43 settles the entry rung: a new collaborator gets Triage, and anything
above it is a maintainer decision about a person. This RFC holds the part that
is *not* settled: what the rungs above Triage are, what earns one, and whether
any of it should ever be numeric.

## Where the reasoning starts

Trust is a function of **duration, not volume** — the xz-utils lesson, where
fourteen months of good patches preceded the payload. Whatever ladder this
project adopts must therefore price time-in-community, not patch count. What
scales down honestly from the big projects (Kubernetes, Apache) is only:

- the **shape** — a named rung between stranger and publisher, so declining to
  grant push is policy rather than rudeness;
- **sponsorship** — a rung is proposed by someone already on it;
- a **time floor** — no rung inside the first N months, whatever the volume.

Every *numeric* threshold from those projects (reviews-per-quarter, member
counts) is governance cosplay at this size and is not proposed.

## Sketch, to react to

| Rung | Holds | Earned by |
|---|---|---|
| 0 — stranger | fork + PR | nothing; the fence does the work |
| 1 — Triage | issue/PR management, no push | an accepted contribution and a maintainer's nod |
| 2 — write/maintain | push, merge through the queue | sponsorship + a time floor (months, not weeks) |
| 3 — publisher | release approval (the environment's reviewer list) | individual decision, never automatic |

Current reality: two trusted partners hold maintain by direct decision — the
ladder describes *future* strangers, not them.

## Where it is weak — react here

1. Is rung 1 worth having at all with a two-maintainer team, or is it ceremony
   until the first stranger PR actually arrives?
2. What is the right time floor for rung 2 — six months feels defensible and
   arbitrary in equal measure.
3. Should rung 3 ever hold more than one person, given the release environment
   currently names a single reviewer?
