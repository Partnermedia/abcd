# PR #294 — uppercase YAML nulls: commissioned review summary

- **Subject**: PR #294 ("fix: IsNull resolves the uppercase YAML nulls"), an
  outside contribution widening `frontmatter.IsNull` and lint's `isNull` to
  the uppercase null spellings after GitHub #290.
- **Provenance**: commissioned host-delegated multi-agent code review,
  2026-08-19 — 8 finder angles (7 reported; the efficiency finder stalled on a
  two-comparison diff), candidates deduped, recall-biased adversarial
  verification with primary-source reads. Conducted outside abcd's own
  command machinery, hence recorded here per this folder's charter.

## Verdict

Sound in its stated scope: the widening itself is correct, tests pass,
gofmt/vet clean, attribution trailer correct (the `Signed-off-by` is
defensible as the repo's first outside contribution). The findings cluster
around one theme: the PR widens two **hand-synced copies** of the null
predicate while the real divergence axis — **quoting** — is untouched, and
the two most severe findings show the PR's own "the gates can never disagree"
guarantee failing for `"NULL"` (accepted by capture, blocked by record-lint),
with the PR's own tests pinning both sides of the contradiction.

Two candidate claims were refuted during verification and are **not**
findings: a lifeboat quoted-`"NULL"` "silent data loss" claim (the memory
dumper and the ADR scanner read disjoint trees, and the `status: superseded`
clause fires independently), and a DCO-trailer conventions claim (an outside
contribution is exactly the deferral endpoint the convention names).

## Ranked actions

Findings are detailed in `01-findings.md` (F1–F10, most severe first).

| Action | Findings | Route |
| --- | --- | --- |
| Request changes on PR #294 (in scope): delegate lint's `isNull` to `frontmatter.IsNull` (kills the drift class and F4's mirrored-test gap), fix the doc-comment miscount, `t.Errorf` in the spelling matrix, widen the capture agreement test, cite the ledger handle in the CHANGELOG entry | F3, F4, F7, F9, F5, F10 (cite) | PR review comment |
| Capture: quoted-null verdict split between capture and record-lint (the PR's contradiction pair F2 is the pinned evidence) | F1, F2 | issues ledger |
| Capture: release-derivation vs record-lint diagnosis skew for null impact | F8 | issues ledger |
| Capture retroactively: the GitHub #290 bug itself, so the CHANGELOG cites a ledger handle | F10 | issues ledger |
| Intent: one canonical exported scalar resolver; the four independent decoders delegate | F6 | intent draft |

Decision context recorded the same day in `../../DECISIONS.md`
(ledger-canonical issue store, one-way forge mirror), grounded in
`../../../development/research/notes/2026-08-19-issue-ledger-forge-sync-sota.md`.
