---
schema_version: 1
id: "iss-2608231322321751"
slug: "overlap-gate-measures-only-the-coil"
severity: "minor"
category: "bug"
source: "agent-observation"
found_during: "agent-verification"
found_at: "internal/core/site/layout.go"
resolution: "the overlap gate counts both arrangements in the renderer's own space (point x coil_radius against the reference-pixel radii)"
impact: fix
resolved_by:
  intent: "itd-157"
  spec: "spc-50"
---

The site check's overlapping-bubbles gate measures the COIL arrangement only: the counter lives inside the coil packer and never sees the by-links layout, so a links arrangement can ship with overlapping bubbles while the gate reports zero. Widening it is not a one-line change — the published positions are normalised to the unit disk while the radii are quoted in reference pixels, so a correct check has to compare both in one space (an attempt that compared them across spaces reported 159,600 overlaps on a 400-node fixture). Found while diagnosing why the by-links arrangement never settles (iss-2608231243286557). The renderer-side half of that defect IS fixed: the collision resolver no longer adds padding the layout never promised.