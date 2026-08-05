---
schema_version: 1
id: "iss-43"
slug: "readme-capability-currency"
severity: "major"
category: "documentation"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "README.md"
resolution: "README's Layout line states what each distribution channel contains — the release artifact excludes .abcd/, a marketplace checkout carries it — instead of a blanket never-shipped claim. The original three claims were overtaken by the iss-143 rewrite; the detector is re-pointed at iss-181 as a candidate extension."
impact: fix
---

README capability currency. Re-scoped by the 2026-08-05 maintainer disposition to one surviving instance; the original three-claim corpus is closed as filed.

Overtaken, no work: the Status section's Phase 0 claim and the surface list's native review oracle, spec/task engine and autonomous run are gone from the README, which the iss-143 rewrite (`48a3524`, 2026-07-27) replaced wholesale. The claims the corpus was written against no longer exist to correct.

Surviving instance: the Layout section's `.abcd/` entry claims the directory is "never shipped". That is true of one distribution channel and false of the other. The release artifact carries the four binaries and `checksums.txt` only (`.github/workflows/release.yml`), so `.abcd/` is genuinely excluded there; the plugin channel installs from `.claude-plugin/marketplace.json`, whose plugin `source` is `"./"` — the whole repository, `.abcd/` included. Fix: the line names what each channel contains rather than asserting a blanket exclusion.

The fix is doc-side because the alternative is feature work, not doc repair: the marketplace manifest schema offers no per-path exclusion, so making "never shipped" true across both channels means building a publish path that produces a curated plugin tree. Packaging changes are out of scope here.

Detector dropped. The proposed extension of the docs-currency gate to README capability and status claims is re-pointed at iss-181 as a candidate extension of the gate cross-check detector's scope, to be settled with that detector's own scoping questions.
