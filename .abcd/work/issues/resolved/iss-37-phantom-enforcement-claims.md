---
schema_version: 1
id: "iss-37"
slug: "phantom-enforcement-claims"
severity: "major"
category: "documentation"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "docs/reference/cli/README.md"
resolution: "Three doc-claim instances corrected: AGENTS.md names preflight's three lint gates and attributes the gofmt gate to CI's own step, the ci.yml record-lint comment reads as the blocking gate it is, and README and CONTRIBUTING describe the real gate suite. The 06-lint.md catalogue and the gate cross-check detector are re-filed as iss-181; original instance 1 (docs/reference/cli/README.md) was verified accurate in round 13."
impact: fix
---

Phantom enforcement claims: developer docs describe gates that do not run as described. Re-scoped by the 2026-08-03 maintainer disposition to three doc-claim instances, each a claim about the local pre-push gate or CI:

1. `AGENTS.md` — the `make preflight` comment in the build block names only build, vet, test and race, omitting the three lint gates the target actually runs (`lint-reviews`, `record-lint`, `docs-lint`); and the definition of done attributes the `gofmt -l .` gate to `make preflight`, which does not run gofmt (CI enforces it as its own format step).
2. `.github/workflows/ci.yml` — the record-lint step's comment calls the gate "visible but non-blocking for now", though the step carries no `continue-on-error` and `go run ./cmd/record-lint` exits non-zero on a blocker-severity finding, so it blocks the job.
3. `README.md` and `CONTRIBUTING.md` — both describe the gate suite incompletely: the README build block omits the lint gates, and CONTRIBUTING both omits them and claims `make preflight` runs "the same checks" as CI, which is not true of CI's gofmt step.

Of the five instances originally filed, instance 1 — `docs/reference/cli/README.md` describing CI-generated reference pages and a freshness check — was verified accurate in round 13 and carries no work.

The brief's 06-lint.md section 1 catalogue (the predecessor project's Python-era `IL001`–`RC007` families) and the proposed gate cross-check detector are re-filed as iss-181, which carries the maintainer's disposition on both: the catalogue is removed rather than archived and is not replaced by a Go-equivalent table, and the detector's two open scoping questions must be settled in that issue's body before any round implements it.
