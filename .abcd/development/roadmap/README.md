# abcd Roadmap

Status dashboard for the abcd plugin.

---

## Phases

abcd organises work into ordered **phases**, each ending in a **milestone** — a
concrete, checkable end condition. Phases are the project's sequencing axis;
they replace plugin-version language (`v1`, `v2`). See
[adr-9](../decisions/adrs/0009-phase-as-product-layer.md) for why.

- **Brief**: lives at `.abcd/development/brief/README.md` (canonical, current).
  It reflects the project's *current* state — not versioned, not archived;
  history lives in `git log`, inflection-point rationale in
  [`../decisions/adrs/`](../decisions/adrs) (see
  [adr-5](../decisions/adrs/0005-brief-is-current-state.md)).
- **Phases**: `phases/phase-N-<slug>.md` — the ordered build plan. Each phase
  doc opens with the product Expectation, then its milestone, scope, and
  traceability. See [phases/README.md](phases/README.md).
- **Intents** (`itd-N`): forward-looking user-facing capability in press-release
  format under `intents/{drafts,planned,shipped,disciplines,superseded}/`. An
  intent's phase membership is recorded editorially in the owning phase doc's
  `## Scope`, not in intent frontmatter.
- **Plugin releases**: tracked by `.claude-plugin/plugin.json` `version`. A
  release version is an *output* of completing a phase, never an input that
  organises work.

The brief is not re-versioned per release; it stays the canonical current-state
record. What has shipped is defined by which phases are complete and which
intents are in `shipped/`.

---

## Current State

**Plugin v1** — in active design and implementation. The brief at [`../brief/README.md`](../brief/README.md) is the source of truth.

This dashboard does **not** hand-maintain spec or intent counts — they drift the
moment work ships. Status is read live from `.flow/` (specs) and the filesystem
(intent buckets) via the commands below, the same stale-proof pattern
[`../brief/06-delivery/03-out-of-scope.md`](../brief/06-delivery/03-out-of-scope.md)
uses. Per-phase progress is coarse and editorial; per-spec truth is whatever
`flowctl specs` prints right now.

**Phase progress.** Phases 0–3 are complete (their entry/closeout/cleanup specs
— fn-6, fn-18, fn-24, fn-31/32/33 — are closed). Phase 4 (the lifeboat pipeline)
is in progress, opening with the entry-verification spec fn-49. Phases run in
spec-dependency order, not lockstep; the spec-level cut is whatever `flowctl
specs` reports. See [phases/README.md](phases/README.md) for each phase's scope
and milestone — that file plus the live spec list are jointly authoritative, and
this dashboard keeps no static per-spec status table of its own.

**Flow-next specs** — read the live list (state + per-task done counts) straight
from the harness; never transcribe it here:

```sh
# Full spec roster with state and N/M task progress.
scripts/ralph/flowctl specs

# Open specs only (what is in flight right now).
scripts/ralph/flowctl specs | grep -E '^\s*\[(open|in_progress)\]'
```

**Intent buckets** — counts derived from the filesystem, never hand-kept:

```sh
# Live count per lifecycle bucket.
for b in drafts planned shipped disciplines superseded; do
  printf '%-12s %s\n' "$b" \
    "$(ls .abcd/development/roadmap/intents/$b/itd-*.md 2>/dev/null | wc -l)"
done
```

| Bucket | Location | Lifecycle stage |
|---|---|---|
| Drafts | [intents/drafts/](../intents/drafts) | Captured (press release written), no flow-next plan yet |
| Planned | [intents/planned/](../intents/planned) | `/abcd:intent plan` has linked an in-flight flow-next spec |
| Shipped | [intents/shipped/](../intents/shipped) | Linked spec closed; fidelity audit queued/appended (fn-48 backfilled the lifecycle state across closed specs; the RC006/RC007 spec-side guard keeps it from re-drifting) |
| Disciplines | [intents/disciplines/](../intents/disciplines) | Active cross-cutting rules (itd-1, itd-5, itd-37) |
| Superseded | [intents/superseded/](../intents/superseded) | Killed by reclassification |

**Phase membership** is editorial, not counted here — each phase doc's `## Scope`
is the single source for which intents it bundles (per
[adr-9](../decisions/adrs/0009-phase-as-product-layer.md)). The later-phase
(not-yet-scoped) set is enumerated stale-proof in
[`../brief/06-delivery/03-out-of-scope.md`](../brief/06-delivery/03-out-of-scope.md).
See [intents/README.md](../intents/README.md) for the full intent index.

**Phase audit.** A phase's delivered reality is reviewed against its structured
`## Phase Acceptance` by the **phase-audit reviewer**
(`scripts/abcd/phase_audit_reviewer/`, run via `python -m
scripts.abcd.phase_audit_reviewer <phase-id>`) — a sibling of
`intent-fidelity-reviewer` that resolves member specs through the editorial
`## Scope` membership chain (intents → implementing specs), emits per-acceptance
verdicts, and writes a receipt to `.abcd/logbook/audit/phase-<ts>/` without
mutating the phase doc. The companion `PA001` lint verifies any `phase:` anchor
names a real phase. (fn-66; see [adr-9](../decisions/adrs/0009-phase-as-product-layer.md).)

There is no `intents/active/` directory — "active" is implicit (a planned intent's linked spec is currently in flight under `.flow/specs/`). See [intents/README.md](../intents/README.md) for full lifecycle details.

The first public release is cut when Phase 5 completes (see [phases/phase-5-roundtrip.md](phases/phase-5-roundtrip.md)). Each major capability defined in the brief gets a corresponding shipped intent as its phase closes, so the intent registry remains the canonical "what abcd does" record.

---

## Intent Capture

abcd uses the **press release format** (Amazon working-backwards inspired) for capturing intent — both for its own roadmap items AND for the artefacts produced by `press-release-composer` during disembark.

**Why press releases instead of feature specs:** Feature specs are engineering-shaped from the start (Problem → Design → Tasks). Press releases are user-experience-shaped (what *exists for the user* once shipped). Forcing intent capture in user-facing language disciplines product clarity before engineering scope.

See [intents/README.md](../intents/README.md) for the format guide and the planned-intent index.

---

## Related Documentation

- [Brief](../brief/README.md) — canonical plugin v1 design specification
- [intents/](../intents) — intents (drafts / planned / shipped / disciplines / superseded)
- [phases/](phases) — the ordered build plan; one doc per phase, each ending in a milestone
- [research/](../research) — SOTA research baseline + per-agent prompting research
- [activity/](../activity/) — curated-from-volatile-sources artefacts (reviews, issues, notes)
