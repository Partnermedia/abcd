# Field research: building the collaborator fence by hand (2026-08-19)

One day of applying the full collaborator-readiness fence to this repository by
hand — org migration, rulesets, environments, merge queue, custom required
checks, release gates — recorded as evidence for the itd-92 capability-ladder
extension. Every lesson cites the record or change that carries it.

## What the day proved

1. **The tiers are real, and the boundaries are sharp.** On a personal
   repository there is no rung between "no access" and "can publish" — two
   write invitations were accepted before the gap closed, and only the org
   transfer created the Triage rung (adr-43). On the earlier free-tier private
   repo, rulesets were structurally unavailable (the retired deferral in
   `.githooks/pre-push`). Plan, visibility, and ownership each change which
   fence pieces *exist*, not merely which are convenient.

2. **Every server-side lever lives in a web console until it is mirrored.**
   The applied rulesets are mirrored under `.abcd/work/rulesets/` so gate
   claims have an in-tree source of truth; the live-vs-mirror drift check is
   unbuilt (iss-277). A doctor that cannot diff live against committed state
   cannot tell the user their fence quietly moved.

3. **Bootstrap ordering bites, twice in one day.** A required check must exist
   on the default branch before the ruleset requires it, or every pull request
   wedges on "Expected": the merge queue needed the `merge_group` triggers
   merged first (PR 344, then the flip), and the two-reviewer gate needed its
   workflow merged first (PR 355, then the flip). A PR already queued when the
   flip landed wedged until a fresh event re-fired the new check (PR 356,
   close/reopen). An applying verb must own this ordering, not document it.

4. **A gate a repo has not configured must not be scaffolded to look like
   one.** The release-environment gate is two lines of workflow plus one
   settings call (PR 340) — but the scaffold-parity test forced it behind the
   `.Abcd` conditional, because a scaffolded repo without the environment
   would run ungated while *looking* gated. Looking-protected-while-open is
   the false green `principles/loud-staging.md` forbids.

5. **Identity assumptions break on org-owned remotes.** The launch scanner's
   `real_name` suppression keyed the caller's public handle to the remote
   URL's owner; the transfer turned the maintainer's own login into 13
   hard-fail findings (iss-283, fixed by deriving the login from the caller's
   noreply address). Probes must derive identity from caller-local facts,
   never from repository ownership.

6. **Author-conditional rules do not exist natively.** "Two invited reviewers
   for external PRs" (iss-281) required a custom required check, and it must
   run on `pull_request_target` — on plain `pull_request` a fork runs the
   workflow from its own tree and can neuter the logic while keeping the
   check's name green.

7. **The loud-degradation idiom already ships and works.** The private banlist
   prints its `reach:` lines; a scaffolded store with zero entries warned on
   every commit until populated, and the first staged match was blocked naming
   only the key. "Armed but empty, and saying so" is the exact posture the
   doctor needs per fence piece.

8. **Cost advice is tier-dependent too.** Actions minutes are free on public
   repositories (macOS included); the same fence on a private repo bills at a
   10× macOS multiplier. What the doctor recommends should say what the
   user's tier makes free.

9. **A rule without an executor needs saying.** Release retention is fully
   specified and computed every cut, but nothing prunes (iss-282). "The rule
   exists; nothing enforces it" is a distinct verdict from "applied" and from
   "impossible", and the doctor vocabulary needs all three.

## Where this feeds

The itd-92 draft carries the capability-ladder extension these findings
ground; iss-277, iss-281, iss-282 and iss-283 hold the individual gaps.
