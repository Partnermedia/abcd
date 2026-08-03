# itd-88 lifeboat coverage experiment — the cross-repo readout

itd-88 inverts the lifeboat build: **probe before pack**. Rather than assume the
brief's 23-section structure fits reality, the experiment runs the same read-only
`disembark probe` over repositories of mixed record quality and reads the coverage
delta as *what keeping a record is worth*. This note records the first run of that
experiment and the section list the packer inherits from it.

## Method

Three repositories, probed read-only with `disembark probe <repo> --json`, then
aggregated with `disembark coverage <report.json>...`:

- **this repository** — the record-rich case (Tier 0 git + Tier 1 conventions +
  Tier 2 abcd-native), 489 commits.
- **[spf13/cobra](https://github.com/spf13/cobra)** — a well-known Go CLI library,
  1106 commits, no abcd record (Tier 0 + Tier 1 only).
- **[psf/requests](https://github.com/psf/requests)** — a well-known Python HTTP
  library, 6486 commits, no abcd record (Tier 0 + Tier 1 only).

The two foreign repositories are the record-less case the experiment targets: a
popular project of a different ecosystem each, cloned full-history so Tier 0 git
archaeology has commits to read, and carrying no `.abcd/` directory. Both were
byte-identical after probing — the working trees stayed clean and no `.abcd/` was
written — so the read-only, out-of-tree contract holds across ecosystems. Probing
the same repository twice yields byte-identical JSON, so a delta in the aggregate
is a delta in the repositories, never in the tool.

## The aggregate

23 brief sections × 3 repositories. Per-repository status counts:

| repository      | grounded | partial | blank |
|-----------------|----------|---------|-------|
| this repository | 21       | 2       | 0     |
| cobra           | 4        | 7       | 12    |
| requests        | 3        | 11      | 9     |

The record delta is legible as a number: the record-rich repository grounds **21
of 23** sections; the two git-plus-conventions repositories ground **4** and **3**.
Keeping an abcd-native record is worth roughly **17–18 grounded sections** over an
otherwise comparable repository that keeps only a README, docs, and git history.

The coverage verdict names **0 of 23 sections blank in every probed repository** —
no section is universally ungroundable, because this repository authored a brief
page for nearly all of them. The always-blank cut therefore removes nothing.

## What survives, and the one section that does not

The `Kind` field (schema v2) partitions the 23 sections identically in all three
probes: **19 extractable** and **4 human-owned** (`product/mental-model`,
`product/personas`, `delivery/verification-matrix`, `delivery/out-of-scope`). A
human-owned blank is reported not as a failure but as *"yours to write, not an
extraction"* — the probe declines to pretend a repository could ground it.

Reading the section × repo table for the sections that no repository grounds by
extraction (best status below `grounded` across all three) yields exactly one:

- **`product/personas`** — `partial` here (abcd-native, low confidence),
  `blank` on cobra, `blank` on requests. Its blank on both foreign repos searched
  *"personas registry, press-release quote attributions"* and found nothing. This
  is the section itd-88's Open Question predicted: *"`product/personas` is
  predicted blank below abcd-native and only partial there; if that holds across
  the corpus, the section is not derivable from a repository at all."* **The
  experiment confirms the prediction.** Personas survives the strict always-blank
  cut only because this repository hand-authored a personas page; on the
  extraction test it is never grounded on any repository, and only ever partial.

The remaining three human-owned sections (`mental-model`,
`verification-matrix`, `out-of-scope`) ground on this repository — someone wrote
the brief page — but stay blank on both foreign repos: they are authored, not
extracted. They survive because the record carries them, not because a repository
yields them.

## The surviving section list

Read as the packer's input, the aggregate's own always-blank verdict removes no
section, so **the packer is built to all 23 brief sections**:

```
product/press-release, product/context, product/mental-model, product/scope,
product/personas, constraints/platform, constraints/dependencies,
constraints/invariants, constraints/naming, evidence/what-worked,
evidence/what-didnt, evidence/open-questions, evidence/tradeoffs, surfaces,
internals, delivery/build-sequence, delivery/verification-matrix,
delivery/out-of-scope, glossary, graveyard, rescue/spine, docs/adrs,
activity/issues
```

The one section the experiment flags for review is `product/personas`: it clears
the always-blank cut only on the strength of this repository's authored page and
is never groundable by extraction. It stays in the section list as a **human-owned
blank** — the packer emits it with its searched-list and its question rather than
dropping it — which is the honest home for a section a repository cannot yield.

## Load-bearing evidence lines

The three probe lines that carry the finding:

1. **cobra, `graveyard` — grounded (git, high):** *"4 reverted commits, 48 files
   deleted (e.g. .circleci/config.yml)"*. A repository whose richest tier is Tier 0
   still grounds `graveyard` from git history alone, satisfying that acceptance
   criterion on a real foreign repository.
2. **requests, `constraints/naming` — blank:** searched *"GLOSSARY.md,
   docs/glossary\*, docs/naming\*, naming document"*, question *"What names and
   reserved vocabulary are fixed? No naming document and no glossary found."* The
   blank is a first-class result: it names what it searched and the question a
   human must answer, rather than inventing a plausible section.
3. **cobra, `product/personas` — blank:** *"human-owned — yours to write, not an
   extraction"*, searched *"personas registry, press-release quote
   attributions"*. This is the evidence behind the one section the experiment
   flags as never groundable by extraction.

## Observed reality vs spc-3's assumptions

- **The rich-vs-git-only framing holds, but the graveyard inverts it.** spc-3
  frames the delta as record-rich beating git-only everywhere. `graveyard` is the
  exception: it is `grounded (git, high)` on cobra (4 reverts, 48 deletions) and
  requests (39 reverts, 272 deletions) but only `partial (git, medium)` on this
  repository (40 deletions, no reverts detected). `graveyard` grounds from Tier 0
  archaeology, which is orthogonal to record richness — a busy history with
  reverts grounds it better than a curated one, so the record-rich repository
  scores *lower* here than the poor ones. This confirms the tiers table's claim
  that `graveyard` grounds from git alone, and refines the delta story: the record
  premium is 17–18 sections *elsewhere*, and roughly zero on `graveyard`.
- **The always-blank cut is uninformative on a corpus that includes the author's
  own repository.** Because this repository grounds nearly everything, no section
  is blank-everywhere, so the aggregate's headline verdict prunes nothing. The
  section that actually fails the derivability test (`personas`) is visible only by
  reading the section × repo table for *never-grounded-by-extraction*, not from the
  always-blank count. A future run over several record-less repositories *without*
  the authoring repository in the corpus would let the always-blank cut do the
  pruning the spec expects of it.
- **Corpus size (Open Question 2).** Two foreign repositories agree closely
  (grounded 4 vs 3; the same human-owned sections blank), which is suggestive but
  not yet a trustworthy population — the finding here is a first reading, not a
  settled number.

## Conclusion

The premise survives contact with reality: an abcd-native record is worth roughly
17–18 grounded brief sections over a git-plus-conventions repository, and the probe
holds its honesty discipline — every grounded section cites a file, every blank
names what it searched and the question it raises, and it never fails merely
because a repository is poor. The section structure survives too: 22 of 23 sections
ground somewhere, and the packer is built to the full 23-section list, with
`product/personas` carried as the human-owned blank the experiment singles out as
never derivable from a repository.
