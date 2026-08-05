---
schema_version: 1
id: "iss-43"
slug: "readme-capability-currency"
severity: "major"
category: "documentation"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "README.md"
resolution: "README's Layout line states what each distribution channel contains — every repository checkout carries .abcd/, marketplace installs and release source archives included, and the released binaries do not — instead of a blanket never-shipped claim. The original three claims were overtaken by the manual README revision 73428b6 (2026-07-17); the detector is re-pointed at iss-181 as a candidate extension."
impact: fix
---

README capability currency. Re-scoped by the 2026-08-05 maintainer disposition to one surviving instance; the original three-claim corpus is closed as filed.

Overtaken, no work: the Status section's Phase 0 claim and the surface list's native review oracle, spec/task engine and autonomous run are gone from the README. All three were removed by `73428b6` (2026-07-17, "manual: Revise README with new badges and project details"), a manual revision predating the v0.5.0 plan; the iss-143 tagline commit `48a3524` (2026-07-27) touched only the strapline. `git log -S` over README.md places every one of the three strings in `73428b6` and the scaffold commit `8624854` alone. The claims the corpus was written against no longer exist to correct.

Surviving instance: the Layout section's `.abcd/` entry claims the directory is "never shipped". That holds for one thing the project distributes and fails for every other. The released binaries carry nothing but themselves — `release.yml` uploads the four binaries and `checksums.txt` — but the plugin channel installs from `.claude-plugin/marketplace.json`, whose plugin `source` is `"./"`, so an install takes the whole repository; and GitHub attaches auto-generated source archives to each release, which carry `.abcd/` as well, since `.gitattributes` declares no `export-ignore`. Fix: the line divides on the boundary that is real — every repository checkout against the released binaries — rather than asserting a blanket exclusion.

The fix is doc-side because the alternative is feature work, not doc repair: the marketplace manifest schema offers no per-path exclusion, so making "never shipped" true across every channel means building a publish path that produces a curated plugin tree. Packaging changes are out of scope here.

Residue recorded as iss-183: the same blanket claim survives in `.abcd/work/CONTEXT.md` and `AGENTS.md`, the latter also naming a packaging mechanism that does not exist. Both are outside this issue's one-line scope.

Detector dropped. The proposed extension of the docs-currency gate to README capability and status claims is re-pointed at iss-181 as a candidate extension of the gate cross-check detector's scope, to be settled with that detector's own scoping questions.
