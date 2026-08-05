---
schema_version: 1
id: "iss-43"
slug: "readme-capability-currency"
severity: "major"
category: "documentation"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "README.md"
resolution: "Claims 1-2 (the Phase 0 status claim; the native-capability surface list) were overtaken by the manual README revision 73428b6 (2026-07-17) and are closed as filed. Claim 3's surviving half is the Layout line's blanket never-shipped claim, fixed here: the line states that every repository checkout carries .abcd/, marketplace installs and release source archives included, and the released binaries do not. The detector is re-pointed at iss-181 as a candidate extension; the same claim's residue across the record is iss-183."
impact: fix
---

README capability currency. Re-scoped by the 2026-08-05 maintainer disposition to one surviving instance; the original three-claim corpus is closed as filed.

Claims 1 and 2 are overtaken, no work: the Status section's Phase 0 claim and the surface list's native review oracle, spec/task engine and autonomous run are gone from the README. Both were removed by `73428b6` (2026-07-17, "manual: Revise README with new badges and project details"), a manual revision predating the v0.5.0 plan; the iss-143 tagline commit `48a3524` (2026-07-27) touched only the strapline. `git log -S` over README.md for `"## Status"`, `"review oracle"`, `"spec/task"` and `"autonomous run"` places each of them in `73428b6` and the scaffold commit `8624854` alone. (`"spec/task engine"` returns nothing: the phrase was line-wrapped in the old README, so the whole-string probe misses it — the shorter substring is the one that reproduces.) The claims those two were written against no longer exist to correct.

Claim 3's surviving half: the Layout section's `.abcd/` entry claims the directory is "never shipped". `73428b6` never touched this line — `git log -S "never shipped" -- README.md` places it in the scaffold commit `8624854` and in this branch's fix alone — so it stood untouched while the other two claims were being overtaken. It holds for one thing the project distributes and fails for every other. The released binaries carry nothing but themselves — `release.yml` uploads the four binaries and `checksums.txt` — but the plugin channel installs from `.claude-plugin/marketplace.json`, whose plugin `source` is `"./"`, so an install takes the whole repository; and GitHub attaches auto-generated source archives to each release, which carry `.abcd/` as well, since `.gitattributes` declares no `export-ignore`. Fix: the line divides on the boundary that is real — every repository checkout against the released binaries — rather than asserting a blanket exclusion.

The fix is doc-side because the alternative is unwiring work, not doc repair. An exclusion mechanism does exist: `internal/core/launch/bundle.go:29-31` denies the `.abcd` namespace structurally and `bundle_test.go` pins it. It is simply not on the path any release takes — `Ship` stops at `WouldPublish` (`internal/core/launch/ship.go:36-41`) with no network call, and `release.yml` uploads the binaries directly without invoking the verb. Making the blanket claim true therefore means wiring that path and finding an answer for the two channels it does not govern (a marketplace install of the repository root, and GitHub's auto-generated source archives). Out of scope here.

Residue recorded as iss-183: the same claim survives across the record — CONTEXT.md's sharp-edges list, AGENTS.md, both `.abcd/` READMEs, two brief documents and a phase doc — several of them attributing the exclusion to packaging that is implemented but has never run. All are outside this issue's one-line scope.

Detector dropped. The proposed extension of the docs-currency gate to README capability and status claims is re-pointed at iss-181 as a candidate extension of the gate cross-check detector's scope, to be settled with that detector's own scoping questions.
