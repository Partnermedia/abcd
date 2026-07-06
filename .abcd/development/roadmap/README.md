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
moment work ships. Status is read live from the **native spec store** (specs, via
the Go CLI; and the companion harness `ccpm` when that backend is attached) and the filesystem
(intent buckets), the same stale-proof pattern
[`../brief/06-delivery/03-out-of-scope.md`](../brief/06-delivery/03-out-of-scope.md)
uses. Per-phase progress is coarse and editorial; per-spec truth is whatever the
native spec store reports right now.

**First milestone.** **Install and launch (Phase 1)** is the first milestone —
`/abcd:ahoy` installs abcd and `/abcd:launch` cuts a curated single-repo release.
The delivery order is **MVP → the companion harness → Claude Code** (per
[adr-28](../decisions/adrs/0028-single-repo-curated-release.md) and
[adr-26](../decisions/adrs/0026-native-spec-layer-ccpm-backend.md)): the native
MVP surface first, the the companion harness-backed deepening next, the Claude Code surface
last.

**Phase progress.** Phases run in spec-dependency order, not lockstep; the
spec-level cut is whatever the native spec store reports. See
[phases/README.md](phases/README.md) for each phase's scope and milestone — that
file plus the live spec list are jointly authoritative, and this dashboard keeps
no static per-spec status table of its own.

**Specs** — read the live list (state + per-task done counts) straight from the
native spec store via the Go CLI; never transcribe it here.

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
| Drafts | [intents/drafts/](../intents/drafts) | Captured (press release written), no plan yet |
| Planned | [intents/planned/](../intents/planned) | `/abcd:intent plan` has linked an in-flight native spec |
| Shipped | [intents/shipped/](../intents/shipped) | Linked spec closed; fidelity audit queued/appended (a spec-side guard keeps the lifecycle state from re-drifting) |
| Disciplines | [intents/disciplines/](../intents/disciplines) | Active cross-cutting rules (itd-1, itd-5, itd-37) |
| Superseded | [intents/superseded/](../intents/superseded) | Killed by reclassification |

**Phase membership** is editorial, not counted here — each phase doc's `## Scope`
is the single source for which intents it bundles (per
[adr-9](../decisions/adrs/0009-phase-as-product-layer.md)). The later-phase
(not-yet-scoped) set is enumerated stale-proof in
[`../brief/06-delivery/03-out-of-scope.md`](../brief/06-delivery/03-out-of-scope.md).
See [intents/README.md](../intents/README.md) for the full intent index.

**Phase audit.** A phase's delivered reality is reviewed against its structured
`## Phase Acceptance` by the **phase-audit reviewer** (Go tooling) — a sibling of
`intent-fidelity-reviewer` that resolves member specs through the editorial
`## Scope` membership chain (intents → implementing specs), emits per-acceptance
verdicts, and writes a receipt under `.abcd/logbook/audit/` without mutating the
phase doc. The companion `PA001` lint verifies any `phase:` anchor names a real
phase. (See [adr-9](../decisions/adrs/0009-phase-as-product-layer.md).)

There is no `intents/active/` directory — "active" is implicit (a planned
intent's linked spec is currently in flight in the native spec store). See
[intents/README.md](../intents/README.md) for full lifecycle details.

Release is a **curated single-repo artifact** — the repo itself is the
marketplace, and `/abcd:launch` packages a distributable with `.abcd/**` excluded
(per [adr-28](../decisions/adrs/0028-single-repo-curated-release.md)); there is no
separate public repository to promote into. Each major capability defined in the
brief gets a corresponding shipped intent as its phase closes, so the intent
registry remains the canonical "what abcd does" record.

---

## Intent Capture

abcd uses the **press release format** (Amazon working-backwards inspired) for capturing intent — both for its own roadmap items AND for the artefacts produced by `press-release-composer` during disembark.

**Why press releases instead of feature specs:** Feature specs are engineering-shaped from the start (Problem → Design → Tasks). Press releases are user-experience-shaped (what *exists for the user* once shipped). Forcing intent capture in user-facing language disciplines product clarity before engineering scope.

See [intents/README.md](../intents/README.md) for the format guide and the planned-intent index.

---

## Related Documentation

- [Brief](../brief/README.md) — the canonical design specification
- [intents/](../intents) — intents (drafts / planned / shipped / disciplines / superseded)
- [decisions/](../decisions) — ADRs (the ratified architecture decisions)
- [phases/](phases) — the ordered build plan; one doc per phase, each ending in a milestone
- [research/](../research) — SOTA research baseline + per-agent prompting research
