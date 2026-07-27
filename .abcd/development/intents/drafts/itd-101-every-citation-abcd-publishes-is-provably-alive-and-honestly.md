---
id: itd-101
slug: every-citation-abcd-publishes-is-provably-alive-and-honestly
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
impact: additive
---

# **Every citation abcd publishes is provably alive and honestly labelled — and the gate that enforces it never flakes.** A cited reference page is only as trustworthy as its links, and links rot silently: pages retitle, URLs redirect, whole platforms announce their own shutdown. abcd splits citation validation across a deterministic gate and an explicit refresh. An offline lint family checks structure and source policy on every commit — footnote markers and definitions in bijection, URLs and DOIs well-formed, aggregator domains refused — with zero network in the gate. `abcd docs cite refresh` does the live fetching on demand and writes a committed baseline: per URL, the final resolved address, when it was last checked, and whether verification was automatic or manual — never how. Sources that block automated fetchers join a manual queue the maintainer clears link by link: printed as a checklist first, later as a generated, disposable checklist page that hands back a receipt file the verb ingests. The lint then enforces the baseline offline — no broken entries, no stale entries, no cited URL without a receipt, no citation whose recorded final address has drifted. The engine is native and dependency-free first; a specialist link checker can slot in later behind the same seam. "My release gate stays deterministic and my citations stay alive," said Kira, an open-source maintainer. "When a source needs human eyes, abcd hands me the exact list to click through and records that I did — not how I did it."

## Press Release

> _Seeded from a quoted-text intent capture. Expand into the full press-release narrative before planning._

## Why This Matters

**Every citation abcd publishes is provably alive and honestly labelled — and the gate that enforces it never flakes.** A cited reference page is only as trustworthy as its links, and links rot silently: pages retitle, URLs redirect, whole platforms announce their own shutdown. abcd splits citation validation across a deterministic gate and an explicit refresh. An offline lint family checks structure and source policy on every commit — footnote markers and definitions in bijection, URLs and DOIs well-formed, aggregator domains refused — with zero network in the gate. `abcd docs cite refresh` does the live fetching on demand and writes a committed baseline: per URL, the final resolved address, when it was last checked, and whether verification was automatic or manual — never how. Sources that block automated fetchers join a manual queue the maintainer clears link by link: printed as a checklist first, later as a generated, disposable checklist page that hands back a receipt file the verb ingests. The lint then enforces the baseline offline — no broken entries, no stale entries, no cited URL without a receipt, no citation whose recorded final address has drifted. The engine is native and dependency-free first; a specialist link checker can slot in later behind the same seam. "My release gate stays deterministic and my citations stay alive," said Kira, an open-source maintainer. "When a source needs human eyes, abcd hands me the exact list to click through and records that I did — not how I did it."

## Acceptance Criteria

> _Required (the itd-1 discipline): add at least one Given-When-Then bullet describing the verifiable bar for "shipped" before this draft can be planned._

## Open Questions

_None recorded yet._

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
