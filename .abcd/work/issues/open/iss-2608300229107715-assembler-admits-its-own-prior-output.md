---
schema_version: 1
id: "iss-2608300229107715"
slug: "assembler-admits-its-own-prior-output"
severity: "major"
category: "security"
source: "impl-review"
found_during: "itd-183 adversarial security review, 2026-08-30"
found_at: "internal/core/reading/include.go (.json root row), internal/core/reading/deny.go"
---

The assembler's own prior outputs re-enter as input: a prior run's bundle.json and manifest.json committed anywhere a root row's .json match reaches (the plugin page's own --out ./run-dir example, resolved against the repo root) are admitted whole as config items, so ruling (18) is breached and the prior manifest's repository paths enter the bundle text while that run's manifest still asserts the exclusion. The artefacts self-identify by _type; refuse any admitted .json whose top-level _type is either tag and refuse an --out that resolves to a path any row admits; extend the own-output test beyond .abcd/development/readings/.
