---
schema_version: 1
id: "iss-250"
slug: "brief-surface-drift-v050-release-gate"
severity: "minor"
category: "drift"
source: "agent-finding"
found_during: "v0.5.0 release gate crosscheck"
found_at: ".abcd/development/brief/04-surfaces"
---

Brief-surface drift found by the v0.5.0 release gate: the full-tier crosscheck (22 checkers, both directions) returned 28 unique discrepancies between the brief's surface prose and the shipped binary — 14 false-claim, 8 undocumented-surface, 2 criterion-violation, 2 stale-count, 2 fictional-layout. Dominant clusters: the vintage/staleness surfaces (bare 'abcd version' three-line render, version --json four fields, version --check, --allow-stale-binary, uninstall --bin-dir) landed without their brief rows; the citation-baseline dry-run gate and render line are undocumented in 04-launch.md; 01-ahoy.md's envelope shape omits the banlist key and guard.detail; 05-intent.md:335 claims gap_audit is deferred (shipped) and an MG004 check ships (does not exist); 07-memory.md's spec-count claim is stale (spc-2..spc-27 exist); the agents/ plugin surface has no registry row and surface_coverage does not check it; skills/-era phrasing survives in 02-constraints/04-naming.md, 06-delivery/01-build-sequence.md and 02-verification-matrix.md. Full list with file:line, claim, reality and class: the iss35-brief-surface-crosscheck receipt for content commit d1afd2c under .abcd/work/reviews/. Not release-blocking (design-record drift, no user-facing doc affected) — same class and disposition as iss-192 at v0.4.2.

Gate re-confirmation: the v0.5.1 full-tier crosscheck (content commit 264d9b1) returned 25 unique discrepancies — the same corpus, unrepaired between cuts, with counts shifted by the cycle's changes (the spec-count staleness now reads spc-30, the principle-distiller shipped-status claim joins the set). Full list: the iss35-brief-surface-crosscheck receipt under .abcd/work/reviews/ for that commit.
