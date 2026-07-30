---
schema_version: 1
id: "iss-176"
slug: "public-banlist-family-unenforceable-under-public-visibility"
severity: "major"
category: "inconsistency"
source: "impl-review"
found_during: "itd-74-increment-2-review"
found_at: ".abcd/docs-lint.json"
---

The public banned-names family lives at .abcd/docs-lint.json, which the installed .gitignore fence ignores under visibility public, so the layer that claims to be committed and CI-enforced is untracked exactly where public exposure is the risk

The two sides of the conflict are each defensible and both are in the record.
spc-20 places the public banlist "committed beside the docs-lint config"; the
iss-169 fence (brief §1) ignores the anchored `/.abcd/` wholesale under public
visibility, with no per-subdirectory exceptions, so one switch decides the whole
namespace. Under that fence a repo abcd configures as public carries a
`.abcd/docs-lint.json` git never tracks, and the family it holds reaches no CI
run anywhere.

itd-74 increment 2 does not resolve it: moving the file amends the iss-169
design record, and carving an exception into the fence gives up the
one-switch property that record chose deliberately. What it does instead is stop
claiming otherwise — detection reports `banlist.public_family_ignored`, and the
status board reads "public family NOT ENFORCEABLE (git ignores …, so CI never
sees it)".

Candidate reconciliations for a maintainer to pick between:

- Move the public banlist (and the docs-lint config generally) to a committed
  path outside `.abcd/`, accepting a second config location.
- Add the single un-ignore the public fence would need, accepting one
  per-subdirectory exception in a table that has none today.
- Keep the placement and declare the public layer private-visibility-only,
  making the private layer the answer for a public repo.
