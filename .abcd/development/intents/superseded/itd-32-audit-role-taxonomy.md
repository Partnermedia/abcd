---
id: itd-32
slug: audit-role-taxonomy
kind: bundle-member
kind_at_supersession: bundle-member
bundle_at_supersession: tier-0-audit-substrate
superseded_by: itd-31
bundle: null
spec_id: null
suggested_kind: null
reclassification_history:
  - { date: 2026-05-07, from: bundle-member, to: superseded, reason: "Premise dissolved when role-by-verb split landed (per command-structure-review round 2): the unified-/abcd:audit-surface assumption that this taxonomy intent rested on no longer holds. Each review/audit role now has its own verb (/abcd:intent review, /abcd:intent consistency, /abcd:intent shape, /abcd:audit). The 'register every audit role in one table' need is absorbed into itd-31's brief edits to 05-internals/01-agents.md (which already documents each role's inputs, outputs, triggers) — itd-32's separate taxonomy file (07-audits.md) is no longer load-bearing. The tier-0-audit-substrate bundle dissolves; itd-31 promotes to standalone." }
---

<!-- NOTE: `status` field intentionally omitted — supersession is encoded by directory location (`superseded/`) per `04-surfaces/05-intent.md § 1`. `kind` retains the value at retirement (matches `kind_at_supersession`); the schema's three valid kinds are `standalone | bundle-member | discipline`. -->


> **⚠️ Superseded by [itd-31](itd-31-cross-document-fidelity-reviewer.md).** This intent's premise — a unified `/abcd:audit` surface bundling all review/audit roles — was dissolved when the round-2 command-structure review split the three intent-fidelity-reviewer roles into three distinct verbs under `/abcd:intent` (review, consistency, shape). The single canonical audit-taxonomy file at `brief/05-internals/07-audits.md` is no longer needed because each role's inputs/outputs/triggers are documented inline in `05-internals/01-agents.md` against the relevant agent. Preserved as historical record per the supersession lifecycle in `04-surfaces/05-intent.md § 1`.

# Every Audit In Abcd Knows What It Audits, By What, And Against What

## Press Release

> **abcd ships a canonical audit-role taxonomy at `brief/05-internals/07-audits.md` — a single map enumerating every audit role in the system, what each audits, by whom (which agent), against what reference, on what trigger, and where its findings land.** Roles, not agents, are the unit of registration: a single agent (e.g., `intent-fidelity-auditor`) may own multiple roles. Adding a new audit role to abcd now requires registering it in this taxonomy; the documentation-auditor refuses to pass disembark if any existing audit invocation lacks a registered role. The map starts with the full audit cast: `intent-fidelity-auditor` in two roles — *single-document* (reality ↔ shipped intent acceptance) and *cross-document* (intent ↔ intent + intent ↔ brief, per itd-31); plus `lifeboat-oracle` (lifeboat content ↔ source repo state), `press-release-composer` (press release ↔ product-thinker scrutiny), `documentation-auditor` (lifeboat docs ↔ structural completeness), `oracle-prompt-audit` (agent prompts ↔ research files, per `05-prompt-quality.md`), and `/abcd:audit chain` (repo state ↔ hash-chain integrity, per itd-16). Agent count stays at 15; the taxonomy makes role count visible. Future audits land in the taxonomy first, code second.
>
> "Every time someone added a new auditor we'd lose track of what it was actually checking — and worse, two auditors would silently do the same job from different angles," said Kira, methodology lead. "The audit taxonomy made the cast visible. Now when someone proposes a new audit role, the first question is 'where does this fit in the taxonomy?' — and if the answer is 'I'm not sure', that's the first sign the role is duplicating an existing one."

## Why This Matters

The 2026-05-07 brief audit found seven distinct audit roles in abcd, owned by six agents: `intent-fidelity-auditor` (single-document role; a second cross-document role is added per itd-31), `lifeboat-oracle`, `press-release-composer`, `documentation-auditor`, `oracle-prompt-audit`, plus `/abcd:audit chain` (itd-16). Each is documented in a different place. None reference each other. Three problems result:

1. **Coverage gaps.** The 2026-05-07 audit found a class of inconsistency (intent ↔ intent + intent ↔ brief) that **no existing audit role covered** — the gap was discovered manually, not by any tool. itd-31 fills it. But the gap existed for months because no map of audit roles existed to reveal the missing cell.
2. **Role overlap.** `intent-fidelity-auditor` (per itd-1) and `lifeboat-oracle` (per disembark phase acceptance) both check "did the artefact match the source?" but at different layers. Without a taxonomy, future contributors will conflate them.
3. **Paper-only sprawl.** `intent-fidelity-auditor` was paper-only at first brief-write time; four intents have since extended its responsibilities (itd-1 verdict format, itd-25 cluster cross-reference, itd-27 term-drift, itd-31 sibling role). When implementation finally lands, the implementer inherits a constellation of obligations across 5+ intents and 5+ brief locations. A canonical taxonomy table makes the obligations visible in one place.

This intent is the **press-release-shaped commitment** behind the structural pattern observation from the 2026-05-07 audit ("there is no canonical map of what gets audited by whom against what"). It is intentionally lightweight — a brief section + an invariant — not a new tool.

## What's In Scope

### The taxonomy table

A new brief section `05-internals/07-audits.md` containing one canonical table. **Roles, not agents, are the unit of registration** — `intent-fidelity-auditor` appears as two adjacent rows because it owns two roles.

| Owning agent | Role | What's audited | Reference / against what | Trigger | Findings location | Severity model | Documented in |
|---|---|---|---|---|---|---|---|
| `intent-fidelity-auditor` | single-document fidelity | Delivered reality | Shipped intent's press release + acceptance criteria | Planned→shipped transition; manual `/abcd:intent audit <itd-N>` | Intent's `## Audit Notes` section | Per-criterion `MET/MET_WITH_CONCERNS/NOT_MET/INCONCLUSIVE` | itd-1; `04-surfaces/05-intent.md § 6` |
| `intent-fidelity-auditor` | cross-document fidelity | Brief + all intents | Each other + the brief | Disembark Phase 0; `/abcd:intent plan`; manual `/abcd:audit consistency` | `.abcd/logbook/audit/consistency-<ts>/report.{json,md}` | R1–R7 categorisation; per-finding blocker/warn/info | itd-31; `06-lint.md` (XD codes) |
| `intent-fidelity-auditor` | shape classification | Intent corpus | Cross-reference patterns, scope overlap, supersession candidates | Continuous (pre-commit + on `/abcd:intent` no-args invocation) | `.abcd/logbook/audit/shape-<ts>/report.{json,md}` + `/abcd:intent` status output | Suggestion-shaped (not blocker); declined suggestions cached so the auditor doesn't re-surface | itd-34; `04-surfaces/05-intent.md § 6 Role 3` |
| `lifeboat-oracle` | lifeboat-content fidelity | Lifeboat content | Source repo state at disembark time | Disembark Phase A/B/C acceptance gates | `.abcd/lifeboat/audit/oracle-<ts>.{json,md}` | "Sufficient" verdict + specific findings | `04-surfaces/02-disembark.md § 6` |
| `press-release-composer` | press-release product audit | Press release | Product-thinker scrutiny ("would a customer come away with a true mental model?") | Disembark Pass C; manual on intent draft | `.abcd/lifeboat/audit/press-release-oracle-<ts>.{json,md}` | Carmack-style review verdicts (`SHIP/NEEDS_WORK/MAJOR_RETHINK`) | `01-agents.md`; `04-surfaces/02-disembark.md` |
| `documentation-auditor` | lifeboat-docs structural audit | Lifeboat docs structure | Structural completeness (every dir has README, every link resolves, etc.) | Disembark/embark/launch sub-agent invocation | `.abcd/lifeboat/audit/documentation-audit-<ts>.{json,md}` | Pass / fail + specific gaps | `01-agents.md` (15th agent) |
| `oracle-prompt-audit` | agent-prompt SOTA audit | Agent `.md` prompts | Per-agent research file in `.abcd/development/research/prompting/agents/<name>.md` | Once per minor release (research-gated) | `.abcd/logbook/sota-audits/<date>.md` | RFC-style findings, not auto-applied | `05-prompt-quality.md` |
| (subverb only — no dedicated agent) | `/abcd:audit chain` | Repo state | Hash-chain / Merkle integrity (compliance-grade) | Manual; per itd-16 | TBD per itd-16's epic | TBD per itd-16's epic | itd-16 |

### Invariants

Three invariants registered in `02-constraints/03-invariants.md`:

1. **Every audit role registers in the taxonomy.** Adding a new audit role to abcd requires adding a row to `07-audits.md`. The `documentation-auditor` checks this at every disembark — if any agent / skill / command invokes an audit not in the taxonomy, disembark fails.
2. **Every existing audit invocation cites its taxonomy row.** Wherever an audit is invoked (in skill `.md`, agent `.md`, brief section), the invocation links to the taxonomy row. This makes orphaned audits visible.
3. **The taxonomy is read-mostly.** Adding rows is light; modifying a row's "what's audited" or "reference" columns triggers an ADR-level decision (because it changes the audit's contract).

### Cross-references back-fill

When this intent ships, the existing audit roles (intent-fidelity-auditor's single-document role, lifeboat-oracle, press-release-composer, documentation-auditor, oracle-prompt-audit, plus itd-16's `/abcd:audit chain`) gain a one-line "registered at `07-audits.md` row N" link in their canonical documentation. When itd-31 ships, intent-fidelity-auditor's *second* row (the cross-document role) is added in place; the agent's catalog entry in `01-agents.md` is extended to declare both roles in the same row (or two adjacent sub-rows under the agent's name). Agent count stays at 15.

## What's Out of Scope

- **Implementing any of the audit roles**: this intent is taxonomy + invariants only. `intent-fidelity-auditor` is paper-only after this intent ships, just as it is now.
- **A "meta-auditor" that audits the taxonomy itself**: the documentation-auditor's check (invariant 1) is sufficient. A separate meta-meta-audit role is over-engineering.
- **Cross-tool taxonomy** (e.g., flow-next's review verdicts): out of scope. abcd's taxonomy covers abcd-internal audits only.
- **Auto-running audits in some unified pipeline**: each audit role keeps its existing trigger mechanism; the taxonomy is observational, not orchestrational.
- **Numbered audit-role IDs** (e.g., `aud-1`, `aud-2`): YAGNI. Audit roles are named, not numbered.

## Acceptance Criteria

- **Given** the brief is updated, **when** a contributor reads `05-internals/07-audits.md`, **then** they find a complete table of every audit role in abcd with the seven columns specified above.
- **Given** an agent or skill invokes an audit role not registered in the taxonomy, **when** the documentation-auditor runs at disembark Phase 0, **then** the audit fails with a finding identifying the unregistered audit and the invocation site.
- **Given** an audit role's row is modified (column other than "Documented in"), **when** the change is committed, **then** the commit message references an ADR justifying the contract change OR the documentation-auditor flags the change as needing review.
- **Given** any new audit role lands in a future intent, **when** that intent's epic ships, **then** a new row is added to `07-audits.md` BEFORE the audit role is invoked anywhere in code.
- **Given** the initial audit cast (intent-fidelity-auditor single-document role, lifeboat-oracle, press-release-composer, documentation-auditor, oracle-prompt-audit) is registered, **when** itd-16 (`/abcd:audit chain`) and itd-31 (intent-fidelity-auditor's cross-document role) ship, **then** their rows are added without disrupting existing rows AND the agent count in `01-agents.md` remains at 15 (itd-31 extends an existing agent rather than adding a new one).
- **Given** the audit taxonomy is consulted during the next brief audit (post-2026-05-07 sweep), **when** an inconsistency surfaces that no audit role covers, **then** the gap is filled by adding either a new audit role (with new taxonomy row) or extending an existing one (with documented contract change).

## Open Questions

- **Where does the taxonomy live**: `05-internals/07-audits.md` (recommended — sibling of agents, adapters, configuration) vs `04-surfaces/07-audit.md` (treats audits as a surface, less accurate). Recommend internals.
- **Severity-model standardisation**: the audit cast has three different severity models (per-criterion verdicts, R1–R7 categorisation, "sufficient" verdict, RFC findings, pass/fail). Should the taxonomy push toward a single model, or accept that different audits warrant different shapes? Recommend: accept different shapes; document each in its row.
- **`oracle-prompt-audit` is the odd one out**: it's research-gated, RFC-style, not blocking. Does it deserve a different category in the taxonomy (e.g., a "review" type vs an "audit" type)? Probably yes; flag for follow-up.

## Audit Notes

_Empty. Populated by intent-fidelity-auditor when intent moves to shipped/._

## References

- 2026-05-07 manual brief audit (this session): produced the meta-finding that motivated this intent.
- Coordinates with: `itd-1` (acceptance gates), `itd-16` (hash-chain audit), `itd-31` (cross-document fidelity auditor — the most direct sibling).
- The taxonomy registers all audit roles documented in: `04-surfaces/05-intent.md § 6` (intent-fidelity-auditor), `04-surfaces/02-disembark.md § 6` (lifeboat-oracle, press-release-composer), `05-internals/01-agents.md` (documentation-auditor), `05-internals/05-prompt-quality.md` (oracle-prompt-audit).
