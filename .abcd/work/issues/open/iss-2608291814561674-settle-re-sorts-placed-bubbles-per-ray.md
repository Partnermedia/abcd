---
schema_version: 1
id: "iss-2608291814561674"
slug: "settle-re-sorts-placed-bubbles-per-ray"
severity: "minor"
category: "tech-debt"
source: "impl-review"
found_during: "ultra-v0.6.8-followup"
found_at: "internal/core/site/layout.go"
---

ultra-v0.6.8 C10: clearOutward in internal/core/site/layout.go allocates a fresh forbidden slice and sort.SliceStable's it on every ray, and settle calls it up to 1+2*settleRays times per bubble, twice per build via byLinks. The fixed-point loop jumps rho to the hi of any interval covering it until none does, and every jump only crosses points that interval covers, so it converges to the least uncovered point at or beyond the start whatever the interval order — the sort is not needed for correctness. Fix: drop the sort and reuse one scratch slice across rays; benchmark before and after.
