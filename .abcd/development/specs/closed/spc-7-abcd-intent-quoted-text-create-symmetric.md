---
id: spc-7
slug: abcd-intent-quoted-text-create-symmetric
intent: itd-46
---
# abcd-intent-quoted-text-create-symmetric

## Summary

spc-7 delivers itd-46: `abcd intent "<text>"` becomes the sub-verb-free create
path, symmetric with `abcd capture "<text>"`. The three invocation shapes are
now crisp on both ledger commands — bare → read-only status, quoted text →
create, sub-verb + id → act on existing. A seeded draft lands in
`intents/drafts/`, `abcd intent new` survives as a deprecation-warning alias
for one transition period, the skill-surface promote flow hands an issue's
text to the create verb (engine-backed at last), and both bare renders carry
the one-line capture-vs-intent decision rule.

## Approach

Mirrors the shipped capture engine where it applies and the spec store where
it is stronger:

- **Engine:** `internal/core/intent/create.go` — `CreateFromText` derives a
  kebab slug (non-alphanumeric runs → `-`, trimmed, capped 60 chars,
  validated against `slugRe` before it becomes a filename), refuses
  empty/whitespace text before any write, mints `itd-N` under an exclusive
  flock on the `intents/` directory fd (`O_NOFOLLOW`, 5s timeout, no lock
  artifact in the record tree — the spec store's mint-lock pattern, not
  capture's), scanning max N across all five buckets inside the lock, and
  writes atomically via `fsutil.WriteFileAtomic`.
- **CLI wiring:** `newIntentCommand` accepts a positional; bare invocation
  keeps rendering read-only status (the create branch is guarded on
  `len(args) > 0`); `new` is a registered alias that routes to the same
  create call and prints the deprecation warning on stderr only, keeping
  stdout byte-identical to the quoted-text form.
- **Canonical seeded-draft shape** (the record this spec defines):
  frontmatter `id`, `slug`, `spec_id: null`, `kind: null`,
  `suggested_kind: null`, `reclassification_history: []`, `builds_on: []`,
  `severity: minor`; body `# <title from whitespace-collapsed seed>`,
  `## Press Release` (placeholder blockquote), `## Why This Matters` (the raw
  seed text), `## Acceptance Criteria` (itd-1 discipline placeholder, no
  bullets — the human fills them before `intent plan`), `## Open Questions`,
  `## Audit Notes`. The shape passes `intent.Validate` and the
  `intent_lifecycle` record-lint drafts rule by construction (verified by a
  test that runs the real lint engine).

### Adjudications (deviations from the intent letter, recorded not hidden)

1. **"Byte-identical to `intent new`" is historical.** No `intent new` verb
   ever existed in the Go CLI (it was a pre-rebuild markdown-surface shape);
   the seeded shape defined above IS the canonical artefact, and the alias
   produces it identically.
2. **Two scope bullets name files that do not exist in this tree** —
   `commands/abcd/intent.md` and `docs/reference/commands.md` are
   old-system references. The intent verb family has no plugin markdown
   surface at all (pre-existing boundary gap, now on the ledger as iss-105);
   the routing/table updates those bullets describe have no native
   counterpart to edit. The AC-level requirements they served (create path
   reachable, decision rule visible) are met at the surfaces that do exist.
3. **Typo-guard asymmetry accepted for now:** unlike capture's
   `suspectedTypoedSubcommand` heuristic, any non-sub-verb first token is
   create text, so a mistyped sub-verb files a draft instead of erroring.
   The ACs do not require the guard; recorded as iss-104 for a symmetry
   follow-up.

## Milestones as delivered

1. Create engine + tests (`a9329f8`), watched-fail first.
2. CLI wiring: quoted-text route, `new` alias + stderr warning, bare
   no-mutation pin, decision rule in both bare renders (`c3d4d0a`).
3. Skill-surface re-plumb: `commands/abcd/capture.md` promote now hands the
   issue body to `abcd intent "<text>"` and honestly states the
   `related_intents` back-link is still verb-less (spc-6 AC3 stays open on
   it); CHANGELOG entry (`7289aa5`).

## Acceptance-criteria satisfaction

AC as ordered in itd-46 → evidence (tests in
`internal/surface/cli/intent_cli_test.go` unless noted):

1. **Quoted text creates a seeded draft** — `TestIntentQuotedTextCreates`;
   engine: `internal/core/intent/create_test.go`
   (`TestCreateFromTextSeedsDraft`, `TestCreateFromTextAllocatesNextID`,
   `TestCreateFromTextPassesRecordLint` — the real lint engine over the
   created file). Watched-fail captured before implementation (undefined
   `CreateFromText`; CLI `unknown command`).
2. **`new` routes to create with a clear transition signal** — lean (a) as
   pre-adjudicated: `TestIntentNewAliasWarnsAndCreates` (stdout identical,
   deprecation warning on stderr only). Watched-fail: `unknown command "new"`.
3. **Bare invocation stays status-only** — `TestIntentBareCreatesNothing`
   (status render asserted, zero files under `drafts/`).
4. **Promote hands the issue text to the create path** —
   `commands/abcd/capture.md` promote section now routes through
   `abcd intent "<text>"`, which the AC1 tests cover end-to-end; the
   markdown states exactly what promote does and does not do today.
5. **The decision rule is visible in both bare helps** —
   `TestBareHelpsCarryDecisionRule` (both bare renders contain the
   nitpick→capture / user-facing-change→intent rule). Watched-fail: bare
   intent render lacked the rule.

Out-of-scope confirmations: no ID-scheme rename, no command merge, mutating
sub-verbs keep requiring ids, `iss-101`/`iss-102` code untouched.
