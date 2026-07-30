---
id: spc-20
slug: name-banlist
intent: itd-74
---
# name-banlist

## Summary

Two-layer banned-names enforcement, generalised from the hand-wired prototype
already live on abcd-cli, delivered as a capability every abcd-managed repo
inherits through `ahoy`. The **public layer** is a banned-token family in the
deterministic docs-lint (committed config, CI-enforced, per-line escape). The
**private layer** is a gitignored per-machine banlist read by a committed
pre-commit guard, so a private string never appears in tracked content — not
even in the config that bans it. Private entries include machine identifiers
(hostnames, device names, IP/CIDR prefixes) as first-class citizens.

## Scope

In: the two banlist stores and their formats; generalising the existing
`harness/*` docs-lint banned-token family into a banned-names family; the
committed guard hook and its refuse-by-key behaviour; `ahoy` scaffolding of
all three surfaces (lint config entry, hook, gitignored stub); banlist
maintenance verbs (add / list / remove) on both layers; honest reporting of
the private layer's local-only reach.

Out: richer dedicated reporting for the public rule (future work per the
planning interview); CI enforcement of the private layer (impossible by
construction — the point); generic network-identifier pattern detection
(iss-154/157, a separate detector); the from-scratch name scan gating a
public flip.

## Approach

- **Two files, per the planning decision.** The public banlist lives in the
  committed docs-lint config (extending the existing banned-token family
  shape: pattern, severity, allow-context escape). The private banlist is a
  gitignored file scaffolded with instructions and reserved-value examples
  only (`examples-use-reserved-identifiers`: RFC 5737/3849/2606/7042 values,
  persona-derived hostnames). Entry format is key + pattern; the key is the
  only thing tooling ever prints.
- **One canonical primitive.** The public family generalises the existing
  `harness/*` family in place — same lint engine, same escape hatch — rather
  than adding a second banned-token mechanism. The guard hook generalises the
  existing `.githooks/pre-commit` prototype.
- **Refuse by key.** The guard greps staged content against private patterns;
  on match it exits non-zero naming the entry key only. The matched string
  and pattern never reach stdout/stderr/logs. Absent banlist ⇒ loud warning,
  exit zero (the layer protects machines that opted in; silence must never
  impersonate protection).
- **`ahoy` wiring.** Scaffolding writes: the public family into the repo's
  docs-lint config, the committed hook, the gitignore entry, and the stub.
  A repo becomes name-safe by being abcd-managed.
- **Verbs.** `abcd`-level add/list/remove operate on both layers; visibility
  follows the layer (public entries render fully, private by key).

## Acceptance-criteria mapping

1. Public-gate criterion → the generalised banned-names docs-lint family
   (blocker severity, file+line findings, allow escape).
2. Refuse-by-key criterion → the guard's output contract; tested with a
   fixture banlist asserting the pattern value is absent from all output.
3. Machine-identifier criterion → entry format treats hostname/IP/CIDR
   patterns identically to name patterns; fixtures use reserved values.
4. Absent-list criterion → guard's no-file branch: warn loudly, never block,
   never silent-pass.
5. Scaffold criterion → `ahoy` writes all four artefacts; scaffold test
   asserts stub content contains only reserved-range values.
6. Verb-visibility criterion → list rendering redacts private entries to
   keys; asserted in surface tests.
7. Honest-reach criterion → status/report text states CI cannot enforce the
   private layer; asserted against the rendered surface.

## Evidence and lineage

Prototype: abcd-cli's own docs-lint `harness/*` family, `.githooks/pre-commit`
guard, and gitignored banlist. Motivating incident: the 2026-07-29
managed-repo NEXT.md privacy-leak investigation (iss-154..iss-158); scope
extension to machine identifiers recorded in iss-158. Related principle:
`examples-use-reserved-identifiers` (seeding). Related discipline path:
itd-79 (registry + lint promotion pattern).

## Acceptance-criteria satisfaction

AC as numbered in itd-74 → covering evidence.

1. **Public gate** — the generalised `banned_tokens` family in
   `.abcd/docs-lint.json` with the per-line allow escape; verb-written entries
   under the `names/` prefix (`internal/core/banlist/public_test.go`), enforced by
   `abcd docs lint`. The layer's reach is bounded by what git tracks, and a config
   git ignores is now reported as unenforceable rather than claimed
   (`banlist.public_family_ignored`; the placement question is iss-176).
2. **Refuse by key** — the committed guard's output contract, asserted against a
   fixture store in `internal/core/banlist/hook_test.go`: the refusal names the key,
   and the pattern and matched text appear nowhere in any output.
3. **Machine identifiers** — hostname, IP, CIDR and MAC entries are ordinary
   entries with no special case, on both readers; the shared corpora under
   `testdata/` carry them.
4. **Absent list** — the guard's INACTIVE branch (absent store) and NO ENTRIES
   branch (present, nothing usable) each warn loudly and exit zero, pinned by
   `TestPreCommitHook_AbsentBanlistWarnsLoudly` and its entryless sibling.
5. **Scaffold** — `abcd ahoy` writes the guard hooks, the merge half beside abcd's
   own guard only, the public family, the EOL pin, and the gitignored stub
   (`internal/core/ahoy/banlist_scaffold_test.go`). Every write is create-if-absent
   and contained; the stub lands only where `git check-ignore` reports the path as
   ignored. `TestBanlistStubSeedsOnlyReservedIdentifiers` asserts the stub's
   examples against the repo's own network-identifier detector, with a control
   value proving the detector is armed.
6. **Verb visibility** — private entries render by key only; the redaction is
   structural (the exported entry type carries no pattern field), asserted in
   `internal/surface/cli/banlist_surface_test.go`.
7. **Honest reach** — `banlist.PrivateReachNote` is stated by the verb, the ahoy
   status board and the JSON envelope, and it names both limits: opt-in machines
   only, and only the commits git runs a hook for.

Residual, recorded rather than closed: the private layer cannot cover a rebase,
`git am`, `git revert`, a cherry-pick or `--no-verify` — a property of where git
runs hooks, which is why the reach sentence names them.

