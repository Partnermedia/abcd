---
schema_version: 1
id: "iss-270"
slug: "the-attribution-gate-still-cannot-be-documented-inside-a-lis"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "manual-capture"
---
The attribution gate's fence handling reads a fence's indent in ABSOLUTE columns, but CommonMark measures the three-space opener limit relative to the enclosing container. That diverges in BOTH directions inside a list item, and the dangerous direction is the one this record originally stated backwards.

UNDER-REJECTION (the hole, live today). A first-level `- ` item's content sits at column 2 and a `1. ` item's at column 3 — both UNDER the absolute limit. So a fence opened inside such an item is read as a document-level fence, the span it "closes" over is stripped, and a footer inside that span never reaches the rules although the forge renders it as an ordinary paragraph. Reproduced and pinned in the corpus as a known-accept (`KNOWN: a fence inside a list item is read as document-level`).

OVER-REJECTION (the usability complaint). Nested deeper — a fence under a second-level bullet, or inside a blockquote — the indent exceeds the limit, the fence is not recognised, and the gate refuses the example exactly as iss-268 described. This repo's pull-request bodies are heavily bulleted, so both shapes are common.

DO NOT relax the opener indent limit. An earlier draft of this record proposed exactly that, on the ground that deeply-indented content is itself a code block. It is the wrong move: relaxing the limit widens the under-rejection hole above while appearing to fix the complaint, and it was grounded in the "stripped region is a subset of what markdown renders as code" invariant that iss-268 retracted as false.

What iss-268 settled on instead: the property the gate actually defends is that a footer appended LAST is caught whatever precedes it — an accident at scale (itd-91), not an adversary — and a body constructed around the footer was never defended. Any fix here must keep that property and must not be justified by the retracted subset claim.

Options: (a) track container context properly (list-item content indent, blockquote markers), which is a markdown block parser and probably too much for a bash gate; (b) refuse to strip anything from a document containing a list marker, as is already done for HTML blocks — safe by construction since striking less can only over-reject, but it would disable the concession for most of this repo's bodies; (c) accept and document, which is the current state, with the residual pinned in the corpus so a future change is visible.

Found by the ruthless and security reviews of the iss-268 fix.
