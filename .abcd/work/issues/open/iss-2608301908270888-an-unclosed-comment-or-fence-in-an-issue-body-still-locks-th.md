---
schema_version: 1
id: "iss-2608301908270888"
slug: "an-unclosed-comment-or-fence-in-an-issue-body-still-locks-th"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "itd-179-fix-delta-ruthless"
found_at: "internal/core/grounds/record.go"
---

an unclosed comment or fence in an issue body still locks the record out of every triage verb and no open record tracks it

**This record is the marker the residual did not have.** It exists because
`iss-2608301803423101` -- a critical, self-declared branch-introduced defect --
was moved to `resolved/` with this consequence still live, and nothing in
`open/` mentioned it. The repository's own rule is that a fixed-but-open issue
leaves no marker to find it by; a not-fixed-but-closed one is the same fault
mirrored, and it is worse, because the record reads as settled.

What survives, at HEAD, reachable with ordinary prose and no malice:

```
abcd capture "the loader drops rules when the <!-- abcd-review marker is stale"
```

That text is accepted without a warning and written verbatim into the body.
Then all three triage verbs refuse, permanently:

```
resolve => grounds refused: the record's body leaves an HTML comment
           (`<!--` with no `-->`) open: the opener is body line 2, ...
wontfix => ... same
promote => ... same
```

There is no `edit`, `amend` or `reopen` verb. The only exit is a hand edit of a
committed file.

The delta made it SURVIVABLE and that was the agreed scope: the refusal now
names the construct and its line instead of blaming the grounds operand, and
`Promote` no longer mints a draft it will orphan. Body scope closed the
frontmatter half only. None of that closes this half, and the fix's own DECISIONS
entry concedes as much.

Two asymmetries worth recording for whoever takes it. `AppendToRecord` already
refuses an unclosed `<!--` in GROUNDS text, so the write side is inconsistent
with itself: it guards the operand and not the body it appends into. And
`Capture` validates nothing about a body it will later require to be parseable
-- the check belongs where the text enters, not where it is next appended to.
