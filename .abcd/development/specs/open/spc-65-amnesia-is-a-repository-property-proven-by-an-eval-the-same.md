---
id: spc-65
slug: amnesia-is-a-repository-property-proven-by-an-eval-the-same
intent: itd-187
---
# Amnesia is a repository property, proven by an eval

## Summary

spc-65 delivers [itd-187](../../intents/planned/itd-187-amnesia-is-a-repository-property-proven-by-an-eval-the-same.md):
a repository eval that assembles one reading definition twice over an unchanged
repository state and asserts the two assembled inputs are byte-identical, the
manifest excluded from the comparison. Amnesia becomes a property of what the
assembler passes, checkable by anyone with the repository, rather than an
instruction to an agent that only a case run could evidence. The decision log
entry of 2026-08-28, ruling (3), is explicit: amnesia is proven by a repository
eval, never by a case run, which is why the closing run of the cycle carries
only purpose durability and convergence.

The eval also pins the two determinism preconditions itd-187 places on the
assembler ([itd-183](../../intents/planned/itd-183-the-cold-reading-sees-exactly-what-the-assembler-passes-posi.md),
spec [spc-61](spc-61-the-cold-reading-sees-exactly-what-the-assembler-passes-posi.md)):
hash-only manifests with no timestamps, and a lexicographic walk order. It lands
in the existing smoke harness package ([`evals/`](../../../../evals/README.md))
behind the `smoke` build tag, so `make smoke` and CI's `smoke` job run it with
no change to [`Makefile`](../../../../Makefile) or
[`.github/workflows/ci.yml`](../../../../.github/workflows/ci.yml). Passing in
CI is a cycle exit criterion.

spc-61 is being written concurrently. Where it is still a stub, this spec
designs against itd-183's own text and the assembler row of the cycle build
plan, and inherits the four interface demands
[spc-64](spc-64-the-read-block-eval-falsifies-the-firewall-planted-warm-cont.md)
places on it rather than adding a fifth.

## Scope

In: the double-assembly comparison and the identity relation it asserts; the
exclusion of the manifest from that comparison and the separate, weaker
assertions made about the manifest itself; the independent lexicographic order
oracle; the order-adversarial fixture that gives the oracle something to catch;
the comparator meta-test that proves the comparison can fail.

Out: everything under [Out of scope](#out-of-scope). In particular this eval
asserts nothing about *what* the assembler included, which is spc-64's property.

## Approach

### The identity relation

Two assemblies of the same definition over the same repository state at the same
commit must produce byte-identical assembled input. The comparison is a byte
comparison of the whole artefact, not a semantic one: a semantic comparison
would tolerate exactly the reordering and re-serialisation this eval exists to
refuse.

The manifest sits outside the comparison, because it legitimately carries a run
identifier that differs between runs. It is not, however, unasserted. Two
weaker properties are checked on each run's manifest: no key or scalar value in
it is timestamp-shaped, and its item paths appear in lexicographic order. That
is the whole of itd-187's "determinism preconditions it enforces on the
assembler", made concrete.

### Assembling twice, from two paths

The eval materialises the shared fixture corpus
(`evals/testdata/cold-reading/baseline/`, owned by spc-64), initialises it as a
git repository, and makes one commit. It then copies the whole tree, `.git`
included, to a second temporary directory, so the second assembly runs at a
different absolute path over an identical commit.

Assembling from two different paths is a deliberate strengthening of itd-187's
criterion, decided here: a run-to-run comparison in one directory cannot see an
absolute path or a temporary directory name embedded in the output, and that
leak is both a determinism failure and a breach of the repository's rule that no
absolute local path enters an artefact. Comparing across paths catches it on the
first run.

### The order oracle

Byte-identity alone is satisfied by any order the assembler picks consistently,
including a filesystem-dependent one that changes on another machine. So the
eval carries its own oracle: it collects the item paths from the assembled
input, sorts a copy with the eval's own lexicographic sort, and asserts the two
agree. The oracle is the eval's sort, never the assembler's, which is what makes
this a check rather than a restatement.

`evals/testdata/cold-reading/order/` gives the oracle something to catch: a set
of records whose names sort differently under byte order, case-insensitive
order, and numeric-suffix order, and whose creation order is deliberately not
their sorted order. The eval owns this directory; spc-64 shares the corpus it
sits in.

### Decisions this spec makes that the record did not

- **The timestamp scan is confined to the manifest.** Projected record bodies
  legitimately quote dates in prose, so a timestamp scan over the assembled
  input would fire on the record's own text. The manifest carries paths, field
  names, and hashes only, so a timestamp-shaped token there is unambiguously a
  defect.
- **Identity is asserted across two paths**, per the reasoning above, rather
  than twice in one directory.
- **The comparison is over the assembled-input artefact as a whole**, taken as
  a file from the assembler's dry-run output directory rather than from stdout,
  so that interleaved diagnostics can never be mistaken for a difference.
- **The eval binds to the verb, not the package**, matching spc-64: it invokes
  the built binary through the harness's existing `TestMain` build and `run`
  helper in [`evals/smoke_test.go`](../../../../evals/smoke_test.go). Here the
  reason is not independence from the include table but that a repository
  property must be checkable by running the shipped binary.
- **One definition, not four.** itd-187's claim is about the assembler's
  determinism, which is position-independent, so the eval asserts identity at
  one position and, for cheapness rather than principle, repeats the order
  oracle at each of the others.

### Wiring

Go test files in package `evals` with `//go:build smoke`. `make smoke` runs
`go test -tags smoke ./evals/...`, and CI's `smoke` job, a required status check
on the protected branch, runs `make smoke`; neither needs an edit. The same
caveat spc-64 records applies: on a pull request confined to `docs/`,
`.abcd/development/`, `.abcd/work/`, and the root prose files, the diff
classifier stands the `smoke` job down, and the merge-queue entry, which runs
the full set on the would-be merge result, is what gates the merge.

## Acceptance criteria mapping

| itd-187 criterion | How spc-65 satisfies it | Pinned by |
|---|---|---|
| Given an unchanged repository state, when one definition is assembled twice, then the two assembled inputs are byte-identical | Two assemblies over one commit from two paths, compared as whole artefacts with the manifest excluded | `TestAssembledInputIsByteIdenticalAcrossRuns` |
| Given a nondeterminism introduced into the assembler (walk order, timestamps), when the eval runs, then it fails | Walk order is caught by the independent lexicographic oracle over the order-adversarial fixture; timestamps are caught by the manifest scan; the comparator itself is proven capable of failing by the meta-test | `TestWalkOrderIsLexicographic`, `TestManifestCarriesNoTimestamp`, `TestComparatorReportsADifference` |

The second criterion is the one that cannot be pinned by holing the shipped
assembler from inside the eval, since an eval must not patch the code under
test. It is pinned in three parts instead: two assertions that catch each named
nondeterminism through an oracle the assembler does not supply, and one
meta-test over the comparator. The watched-red run below closes the gap by
holing the assembler temporarily and by hand.

## Tests

Every test below is watched red before the change and green after.

- `TestAssembledInputIsByteIdenticalAcrossRuns` is watched red against an
  assembler whose walk sort is temporarily removed by a one-line local patch,
  which is the run that proves the comparison can fail against real
  nondeterminism. The run is recorded in the pull request body and the patch is
  reverted before the branch is pushed.
- `TestWalkOrderIsLexicographic` is watched red the same way, and stays the
  standing guard afterwards: it fails on any consistent-but-not-lexicographic
  order, which the byte comparison alone would accept.
- `TestManifestCarriesNoTimestamp` is watched red against a manifest with a
  generation time stamped into it.
- `TestComparatorReportsADifference` feeds the comparator two synthetic
  artefacts differing only in item order, and a second pair differing only in
  one scalar, and demands a reported difference naming the differing item. It is
  the anti-vacuity guard: without it, a comparator that silently compared
  nothing would pass every other test here.
- `TestFixtureOrderIsAdversarial` asserts the `order/` fixture's creation order
  is not its sorted order, so the order oracle cannot pass by coincidence.

## Out of scope

- The assembler, its include table, its projection, and its manifest format:
  spc-61.
- What the assembler included or excluded, and the sentinel corpus that tests
  it: spc-64.
- The reading definitions
  ([spc-62](spc-62-four-cold-reading-definitions-one-blindness-core-each-positi.md))
  and the output contract
  ([spc-63](spc-63-one-ingest-verb-validates-every-cold-reading-output-includin.md)).
- Determinism of anything downstream of the assembler: a reading's own output is
  not required to be reproducible, and nothing here suggests it is.
- Cross-commit stability. The claim is identity at one commit; a changed
  repository state is expected to change the assembled input, and that is what
  makes the manifest's commit reference meaningful.
