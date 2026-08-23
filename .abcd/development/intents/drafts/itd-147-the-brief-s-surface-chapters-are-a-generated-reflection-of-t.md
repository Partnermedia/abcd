---
id: itd-147
slug: the-brief-s-surface-chapters-are-a-generated-reflection-of-t
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: []
severity: major
---

# The brief's surface chapters are a generated reflection of the shipped surface, so a shape claim cannot drift

## Press Release

> **The design record stops being able to lie about the shipped surface.** Each
> surface chapter in the brief carries a generated block — the verb's flags,
> sub-verbs, exit codes, schema fields and counts, derived from the command tree
> at build time and drift-tested like the CLI reference already is. The prose
> around it keeps doing what only prose can: saying why the surface exists, what
> it refuses to do, and which trade was made. A shape claim can no longer drift,
> because nobody writes one by hand.
>
> The measurement that prompted this is uncomfortable. A full-tier
> brief-to-surface crosscheck at the 0.6.2 release gate returned 147
> discrepancies across 23 chapters. Of those, exactly one touched a file the
> release bundle carries. The rest were the record describing a product that had
> moved: a PATH entry documented as a symlink when it has been an owned copy
> since spc-35, a user-scope directory tree missing two of the three directories
> that actually exist, a detection pass enumerating twelve steps where fourteen
> run.
>
> The same day, the same reviewers found **zero** discrepancies across the 775
> lines of `docs/reference/cli/commands.md`. That file is generated and
> drift-tested. That is the whole argument.
>
> "I read the brief to learn what the tool does before I touch it," said Maya, an
> autonomous-development practitioner whose agents read the record the same way.
> "When a chapter says the PATH entry is a symlink and the binary writes a copy,
> I do not discover a documentation bug. I write code against a surface that does
> not exist, and the mistake looks like mine."
>
> "We already knew how to fix this and had done it once," said Kira, who
> maintains the surface. "The CLI reference is generated, so it is right. The
> brief is typed, so it rots. We were keeping two copies of the same fact and
> hand-maintaining the one nobody could check."

## Why This Matters

The brief drifts steadily, not in bursts. Of 124 findings resolvable to a
blaming commit, 66 were last written in one month and 58 in the next — an even
rate, in every class. Roughly half the false claims survived four releases, each
of which recorded a PROMOTE receipt against this same detector at 26, 28 and 37
findings.

So the gate has been running and passing while the record got worse. That is the
proxy-gate class (iss-2608230847432286) with a demonstration attached, and the
deterministic half makes it sharper: `surface_coverage` blocked a commit during
this very release for a missing `history drain` **table row**, and passed green
while false prose claims sat beside those rows. Row presence stands in for
chapter correctness.

The diagnosis is that a chapter mixes two kinds of content with two different
failure modes in the same paragraphs:

- **Shape** — flags, sub-verbs, exit codes, schema fields, counts, file layouts.
  Derivable from the command tree, therefore checkable, therefore never worth
  hand-writing. `false-claim` and `stale-count` are shape by definition, and most
  `undocumented-surface` is the binary growing something a chapter's shape
  section never learned.
- **Intent** — why a surface exists, what it refuses to do, which trade was
  made. Not derivable, and structurally unable to drift against code, because it
  is not a claim about code shape.

Two copies of one fact, one of them authoritative: `one-canonical-primitive`
violated at two copies rather than three.

## What's In Scope

- **A generated shape block per surface chapter**, derived from the same command
  tree that already produces `docs/reference/cli/commands.md` and
  `.abcd/development/release/surface.json`, with a drift test that fails when the
  committed block and the tree disagree.
- **A seam that keeps prose hand-written.** The generated region is delimited so
  a chapter's rationale is never machine-authored and never clobbered.
- **The trust rule as an ADR plus a brief invariant:** shape claims are derived,
  never hand-authored.
- **Retiring the hand-written shape prose** the generated block replaces, chapter
  by chapter, so the 124 findings are closed by construction rather than by a
  sweep that starts rotting the next day.

## What's Out of Scope

- **A "you changed a surface, so touch its chapter" gate.** It forces an edit
  without forcing correctness, which is a phantom gate — worse than none, and the
  precise class this intent is an instance of.
- **A new principle.** `one-canonical-primitive` already says it; this is an
  application, and the confirmed decomposition records "none new" deliberately.
- **Generating the rationale.** Prose about why is the half that cannot drift and
  the half worth a human.
- **Fixing the 147 by hand ahead of the seam.** The corpus is evidence; a hand
  sweep destroys the dataset and buys a clean brief that drifts again at the
  measured rate.

## Mechanism

We expect this to work because it has already worked here, once, on the same
class of content. `docs/reference/cli/commands.md` is generated from the command
tree and pinned by `TestReferenceMatchesCommittedPage`; it returned zero
discrepancies on the day the hand-written brief returned 147. The mechanism is
not novel and needs no new infrastructure: the generator, the snapshot and the
drift test all exist and are exercised every CI run.

## Scope Conditions

This holds only where a claim is derivable from the command tree. A chapter's
claims about behaviour that no snapshot captures — ordering guarantees, failure
semantics, what a verb refuses and why — stay prose and stay hand-maintained,
and this intent does not make them safe. If the crosscheck's residue after the
seam lands is dominated by that kind of claim, the remedy is a different one.

## SOTA

Doc-generation from a command tree is the presumptive answer and is what cobra
ecosystems do by default. The adversary-filtered question is not whether to
generate but where to cut, because a fully generated chapter loses the rationale
that makes the brief worth reading. The cut proposed here — generated shape,
hand-written why — is the narrow version, chosen because the measurement says
shape is where the drift is.

## Acceptance Criteria

- **Given** a surface chapter with a generated shape block, **when** a verb gains
  a flag or a sub-verb and the block is not regenerated, **then** the drift test
  fails and names the chapter and the missing claim.
- **Given** a chapter's hand-written rationale, **when** the shape block is
  regenerated, **then** the prose is unchanged.
- **Given** a full-tier brief-to-surface crosscheck run after the seam lands,
  **when** its findings are classified, **then** no finding is a `false-claim` or
  `stale-count` about a claim the generated block covers.
- **Given** a chapter whose surface has no snapshot representation, **when** the
  generator runs, **then** it emits no block for that chapter rather than an
  empty or speculative one, and says so.
- **Given** the crosscheck's non-reproducibility (iss-2608231409595789),
  **when** progress is assessed, **then** it is assessed on which chapters are
  implicated rather than on a finding count, because the count is not a metric.

## Open Questions

- Where exactly is the seam? Per chapter, per section, or a single generated
  appendix per chapter? The per-section answer preserves the most rationale and
  costs the most machinery.
- Does the surface snapshot already carry enough to generate a useful block, or
  does it need fields it does not record today (exit codes and schema fields are
  the likely gaps)?
- What happens to a chapter documenting a *staged* surface — one the brief marks
  as designed but unbuilt? A generator that only knows the shipped tree cannot
  emit those, and they are a real and deliberate part of the record.
- Should `surface_coverage` be retired, folded into the drift test, or left as
  the row-level check it is? Leaving it is defensible; leaving it *unlabelled* is
  what let it read as a chapter-correctness gate.

- **Is the shape/intent split actually a two-way one?** This intent assumes every
  claim is either derivable from the command tree (generate it) or is rationale
  that cannot drift against code (leave it prose). A third category surfaced
  while this was being drafted and fits neither: a claim about **external** state
  that is checkable in principle but has no local source to check against.

  The worked example is `wrangler.jsonc`, which records a Cloudflare dashboard
  setting — automatic branch builds disabled — with no durable local home. Eight
  of eight pull requests on 2026-08-23 built through that integration, verified
  by two independent routes. So the recorded state describes no observed build at
  all (iss-2608231607594913, refining iss-2608220150157502).

  That is not stale drift. A stale record was true once and can be regenerated
  from its source; this one had no source, so its truth was never a property
  anything could establish. Prose does not make it safe and no generator can
  reach it.

  It is also the most dangerous shape of the three, for a reason worth stating
  plainly: the plan to change the external state was recorded, the record was
  written as though the change had landed, and nothing anywhere could have caught
  that it had not. The intent and the unverifiable assertion were written by the
  same hand in the same file. It does not read as an unchecked claim. It reads as
  a completed decision.

  If this category is admitted, the remedy is neither generation nor prose but a
  check that reads the real external state — and the scope of this intent widens.
  The two-way split was confirmed by the maintainer, so this is recorded as a
  question rather than folded in.

## Audit Notes

Filed from `iss-2608231346137587`, whose Routing section carries the four-piece
decomposition confirmed by the maintainer on 2026-08-23 (verdict SPLIT), graded
into the dated decomposition-calibration corpus. Capability here; trust rule to
an ADR plus a brief invariant; stance deliberately none new; plumbing to the
brief.

The evidence base is three full-tier crosscheck runs preserved out-of-tree with
their content commits, manifest hash and tier. Read it with
`iss-2608231409595789` in hand: the runs returned 125, 126 and 147, and runs 2
and 3 are a controlled comparison — the brief is byte-identical between their
commits, so the detector's whole subject was held constant and it still returned
17% more findings. Which chapters are implicated is broadly stable; how many and
of what class is not.
