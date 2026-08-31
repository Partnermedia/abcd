---
schema_version: 1
id: "iss-2608311518056854"
slug: "the-supply-regime-registry-refuses-a-reading-for-quoting-the"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "itd-185 fidelity audit rcp-fe3450ca55ff"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/reading/ingest_regime.go"
resolution: "The four semantic signatures (RG-EVAL-ORDERING, RG-EVAL-RECOMMENDATION, RG-REG-FIXPROPOSAL, RG-EXPL-DISPOSITION) are degraded from enforce to flag: a hit raises a review flag on the run record naming the item and the signature id, and the item lands. The structural halves stay absolute. The mode of each entry is now pinned by name in a test, itd-185's ac-5 and ac-9 are rewritten to say flag, the ingested audit verdict for those two criteria is disclosed as superseded, and the residue on itd-185 and spc-63 names over-catching beside under-catching with the propose-versus-report distinction as the reason."
impact: internal
resolved_by:
  intent: "itd-185"
  spec: "spc-63"
---

The supply-regime registry refuses a reading for quoting the document it read, and the rate is high: fourteen of thirty-four realistic reading outputs were refused, all by one failure mode. The disposition detector matches the bare token followed by a colon or equals, and this repository's records carry that token everywhere, so any explicative claim that quotes a record refuses. So do a claim reporting that a clause settles a licensing question, a claim quoting a paper that closes by recommending further study, a claim reporting that a suite scores below a threshold, and a detection reading that one section says a fix is merged while another says it is pending, which is the canonical shape of a detection finding. This is not the NFKC normalisation, which was measured innocent: an A/B over thirty-four cases flipped five outcomes and every one was the registry's own phrase in compatibility code points, with no innocent case newly refused. The noise predates it and is structural, because the detectors cannot distinguish a reading that PROPOSES from a reading that REPORTS someone else proposing, and the second is most of what a reading legitimately does. The intent records signature noise as its open question, but the disclosed residue names only under-catching; on this evidence over-catching is the larger risk and the residue should say so.

## Grounds

- pursued: a detector that cannot separate a reading proposing from a reading reporting somebody else proposing belongs on the observed side of the line until it can, and the flags a first real reading raises are the calibration to reconsider it on; a real reading whose findings survive enforcement, or a flagged corpus that is mostly genuine breaches, would show this wrong
