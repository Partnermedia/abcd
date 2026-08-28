---
schema_version: 1
id: "iss-2608260944055780"
slug: "brief-inverts-abcd-lint-exit-grammar"
severity: "major"
category: "inconsistency"
source: "user-observation"
found_during: "README review session, cross-checking record claims about abcd lint"
found_at: ".abcd/development/brief/05-internals/06-lint.md"
resolution: "06-lint.md now states the code's exit grammar (blocker 2, warn 1, clean 0)"
impact: internal
resolved_by:
  commit: "b760a67c"
---

Brief self-contradiction on abcd lint's exit grammar: 05-internals/06-lint.md:174 states "Exit code follows the standalone grammar (blocker -> 1, warn -> 2, clean -> 0)", which inverts the two non-zero codes. internal/core/repolint/repolint.go:147 exitCode() returns 2 on any blocker, 1 on warnings only, 0 clean, and its own comment names the Conftest tri-state; 04-surfaces/16-lint.md:41 agrees with the code. Observed: this tree at 19 errors and 3 warnings exits 2. The consequence is worse than count drift, because an exit code is a number a CI job branches on: a job written against 06-lint.md would treat 1 as its failure condition and pass green on a tree full of blockers, a gate that runs and never fires. Same class as iss-142 (the iss-37 phantom-enforcement class) and the same file, different line; iss-181 and iss-251 also sit on this file and cover neither. Brief-tier, so under adr-28 it ships in no release artefact and reaches no installed surface, which is the standing disposition for brief drift and the argument for a lower severity. Refines iss-2608261029074763, which records the wider defect this is one symptom of: 06-lint.md section 3 documents an architecture that was never built, verified as internal/core/lint having no package main so the documented invocation cannot run, --codes-filter, --promote-check, --touched-line and --display-path having zero hits across internal/ and cmd/, no .pre-commit-config.yaml, and IntentLinter, _classify_and_filter, _parse_hunk_lines and _detect_location present in no source file. That record cites this one for the exit grammar rather than restating it. Adjacent and distinct: iss-2608231000561060 covers the gates-CI claim and gate-readiness. Detector: enforcement-claims-are-facts sweep over the brief's exit-code and gating claims. Acceptance corpus: the three lines above.