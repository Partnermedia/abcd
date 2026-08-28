---
id: itd-154
slug: s4-gate-failed-bootstrap-provisioned-no-binary
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
promoted_from: iss-253
---

# The Cut A §4 manual gate FAILED on assertions 3 and 4 (transcript: .abcd/.work.local/scratch/2026-08-16-s4-transcript.txt, machine: a no-Go Mac with a fresh harness install). After /plugin marketplace add + /plugin install, NO bootstrap line ever appeared at any of three session starts, the plugin root (…/plugins/cache/abcd-marketplace/abcd/<sha>/) contained a full source checkout but no provisioned binary, and every UserPromptSubmit and PreToolUse hook errored 'No such file or directory' for the entire evening — the exact failure family the gate checks for zero of. The command surface functioned only via rung two (the PATH binary the separate one-liner installed), so the no-Go promise is unproven; ahoy install itself reports plugin.root_missing as resolvable:false. Assertions 1, 2 and 5 passed (checksum-verified one-liner with no sudo, abcd v0.5.0; transcript captured; non-interactive install with a clean, path-scrubbed receipt — including a correct abort in a non-repo directory). Per the gate's own rule this BLOCKS Cut B. Environment notes for reproduction: the harness never visibly ran the plugin's SessionStart chain after install (bootstrap.sh emitted nothing — neither success nor refusal, despite loud-staging), and the machine's home dirs are cloud-synced (anomalous stat results, 65535 link counts, multi-minute tree walks in the clone). Detector: the §4 checklist; acceptance: a fresh-machine plugin install yields one bootstrap success line, a provisioned plugin-root binary, and /abcd answering in about a second with no Go.

## Press Release

> _Seeded by promotion from iss-253. Expand into the full press-release narrative before planning._

## Why This Matters

Graduated from `iss-253`: The Cut A §4 manual gate FAILED on assertions 3 and 4 (transcript: .abcd/.work.local/scratch/2026-08-16-s4-transcript.txt, machine: a no-Go Mac with a fresh harness install). After /plugin marketplace add + /plugin install, NO bootstrap line ever appeared at any of three session starts, the plugin root (…/plugins/cache/abcd-marketplace/abcd/<sha>/) contained a full source checkout but no provisioned binary, and every UserPromptSubmit and PreToolUse hook errored 'No such file or directory' for the entire evening — the exact failure family the gate checks for zero of. The command surface functioned only via rung two (the PATH binary the separate one-liner installed), so the no-Go promise is unproven; ahoy install itself reports plugin.root_missing as resolvable:false. Assertions 1, 2 and 5 passed (checksum-verified one-liner with no sudo, abcd v0.5.0; transcript captured; non-interactive install with a clean, path-scrubbed receipt — including a correct abort in a non-repo directory). Per the gate's own rule this BLOCKS Cut B. Environment notes for reproduction: the harness never visibly ran the plugin's SessionStart chain after install (bootstrap.sh emitted nothing — neither success nor refusal, despite loud-staging), and the machine's home dirs are cloud-synced (anomalous stat results, 65535 link counts, multi-minute tree walks in the clone). Detector: the §4 checklist; acceptance: a fresh-machine plugin install yields one bootstrap success line, a provisioned plugin-root binary, and /abcd answering in about a second with no Go.. Read that issue record for the source observation.

## Acceptance Criteria

> _Required (the itd-1 discipline): add at least one Given-When-Then bullet describing the verifiable bar for "shipped" before this draft can be planned._

## Open Questions

_None recorded yet._

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._
