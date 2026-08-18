---
id: adr-42
slug: guard-parse-layer-is-a-mistake-filter
status: accepted
date: 2026-08-18
supersedes: null
superseded_by: null
related_intents: [itd-103]
related_rfcs: []
related_adrs: [adr-25]
---

# ADR-42: The guard's parse layer is a mistake filter, not a security boundary — matching is two-tier, and completeness is abandoned as a goal

## Context

`abcd guard` reads a proposed shell command, walks past "wrappers" to find the
real command, and matches it against a registry of hazard patterns
([itd-103](../../intents/shipped/itd-103-abcd-teaches-repo-agents-the-shell-commands-they-must-never.md),
spc-16). Three separate defects have now been filed against the same shape:

| defect | the enumeration that was incomplete |
|---|---|
| gh-297 | the interpreter set (`sh bash dash …`) |
| gh-299 | git's global value flags (`-c`, `--config-env`, `--shallow-file`, …) |
| iss-272 | the wrapper set (`sudo doas command env nohup time xargs timeout exec`) |

Each was fixed by extending a list. A third instance of one defect predicts a
fourth, so iss-272 was investigated design-first rather than patched: a plan
drafted from the issue text went to adversarial review before any code and came
back with three blockers and four factual errors. The evidence is in
[`2026-08-18-guard-wrapper-family-and-parse-layer-limits.md`](../../research/notes/2026-08-18-guard-wrapper-family-and-parse-layer-limits.md);
this record holds the decision it informed.

Three findings forced the framing.

**The defect is an asymmetry, not a short list.** The interpreter path already
fails loud on what it cannot resolve — `shellUnresolved` yields a warn, so a
payload the guard cannot read is never mistaken for clearance. The wrapper path
has no equivalent: an unrecognised wrapper simply *becomes* the command, nothing
matches, and the verdict is a confident `allow`. Ten wrapper bypasses were
verified live on the merged tree, `nice git push --force` among them, each
exiting 0.

**The set is unbounded in principle.** Any program that execs another is a
wrapper. `-S` is a required-argument flag in `unshare` and an optional-argument
flag in `nsenter` — same package, same version, same letter — so even the
*tractable* subset carries a standing per-flag maintenance liability, on a
grammar that differs again on macOS.

**Every vendor shipping this control documents it as steering, not enforcement.**
[Claude Code][cc-permissions] calls argument-constraining patterns "fragile" and
refuses to implement `Bash(command:rm *)` because a compound command would bypass
it; [Cursor][cursor-security] shipped a command denylist and it was bypassed
(CVE-2026-22708). The posture is [CWE-184 (Incomplete List of Disallowed
Inputs)][cwe-184] and the literature behind it is settled —
[Ranum's "Enumerating Badness"][ranum-dumbest],
[Garfinkel (NDSS 2003)][garfinkel-traps].
Meanwhile the repo rates this class two ways at once: the GitHub-filed siblings
gh-297 and gh-299 both carry `severity:critical` — labelled "silent
security/data-loss at a trust boundary" — with adversarial framing
("agent-proposed commands are steerable by untrusted content the agent has
read"), while the ledger entry for the same defect, iss-272, is `minor`. Both
cannot be right, and the critical framing is a claim no denylist can honour.

## Decision

**1. The guard's parse layer is a mistake filter.** It catches accidents and
casual evasion by a cooperating agent. It does not withstand a determined author,
and the guard's own documentation says so in those terms. Anything needing an
actual boundary needs an execution-layer control behind it; the parse layer never
claims to be one. The guard's own documentation states this scope (part D below),
and the split rating is resolved with it: a missing wrapper or interpreter name is
a real defect in a mistake filter, not a silent failure of a trust boundary, so
`severity:critical` is the wrong label for the class.

**2. Enumeration-completeness is abandoned as the matching strategy, and
replaced by two tiers:**

- **Tier 1 — position-anchored, blocks.** A registry entry matched at command
  position after stepping known wrappers. Precise, and the only tier that blocks.
- **Tier 2 — position-agnostic, warns.** When *no entry matched at all*, the
  matcher re-runs speculatively from each later command-position token, and each
  speculative suffix is run through `expandPayloads`. A hit is a loud **warn**,
  never a block, because an unknown argv[0] is not proof that it execs its
  arguments.

Tier 2 converts the entire unbounded wrapper class from *silent allow* to *loud
warn* without enumerating anything, and gives the wrapper path the fail-safe the
interpreter path already had.

**3. Tier 2 is bounded against the denial of service it would otherwise
reintroduce.** The obvious form of speculation is N starts × E entries — exactly
the per-entry splice-and-restart `payload.go` forbids by name. Prototyped, it
is strictly quadratic: 14.9 s at 6,000 tokens, and roughly **eight hours of CPU
inside a `PreToolUse` hook** extrapolated to the 1 MiB stdin cap. The required
shape is therefore binding: command-position starts computed once in an O(N)
pass, a hard cap of ~64 starts past which the warn is emitted unconditionally
rather than skipped, short-circuit on first hit, pinned by a benchmark test.

**4. The Tier 2 gate is "no entry matched", never "argv[0] is unknown."** An
argv[0] gate switches speculation off for `git`, and git launches things —
`git bisect run`, `git submodule foreach`, `git rebase --exec` are silent allows
today and would stay silent under it.

**5. Enumeration continues, demoted to an upgrade.** Adding wrapper names lifts a
case from "loud warn" to "precise block" — better UX and a stronger verdict —
and it is safe to be incomplete precisely because Tier 2 backs it. It is no
longer the fix.

**6. The exec-string family gets a small table, not a generalisation of
`shellCPayload`.** `su -c`, `runuser -c`, `script -c`, `flock -c` take a command
string a shell executes, but none is a shell. Generalising the existing helper
resolves every *short* spelling free and returns `shellNone` for every long one
(`su --command`, `--session-command`, …) — six new **silent** allows, not even
the `shellUnresolved` fail-safe. A command → payload-flag table is the correct
shape and leaves that contract untouched.

**7. Sequencing is D → A → B → C:** scope statement first, then the Tier 2
fail-safe, then wrapper names, then the exec-string table. The scope statement
leads because shipping enumeration without it manufactures exactly the false
confidence that makes the next gap dangerous.

**8. The warn-storm STOP binds this work.** Per
[`2026-08-15-guard-execute-string-family-design.md`](../../research/notes/2026-08-15-guard-execute-string-family-design.md):
a warn rate that trains users to ignore warns is a STOP, measured on real agent
commands before merge.

## Alternatives Considered

**Add the missing wrapper names (the issue's own proposal).** Cheapest, and it
closes the ten verified bypasses. Rejected as the *primary* fix: it is the third
instance of the identical defect, the class is unbounded, and it leaves the
silent-allow verdict intact for everything not listed. Retained as decision 5,
below the fail-safe.

**Speculative re-match in its drafted form.** Rejected as drafted for the
quadratic DoS above, and because the coverage claim behind it was wrong —
speculation alone catches 5 of the 9 bypasses in the note's coverage table, not
all of them, since `matchSegment` never descends into a payload. Adopted only in the bounded,
`expandPayloads`-composed form of decisions 2–4, which recovers `busybox sh -c`
at zero additional false positives.

**Descend into any argument token containing a space.** Catches everything,
including `su -c` and `su bob -c`, at +2 false positives. Rejected with evidence:
it breaks the shipped known-good gate — `abcd capture "git clean -fd wiped my
scratch notes"` fires, and that fixture is an incident capture, the case spc-16
says the design turns on. The true-negative floor is 100%.

**Speculate only when every token before the suffix is flag-shaped.** Silences
the `rg -F git push --force docs/` false positive. Rejected: it also silences
`find . -exec`, `chroot /jail`, `flock /tmp/l`, `chrt -f 99`, `taskset -c 0`,
`docker run`, `ssh host` — most of the true positives.

**Invert to an allowlist.** The only shape that is complete by construction, and
what Claude Code's read-only fast path does. Rejected for this surface: the guard
sits on `PreToolUse` for arbitrary developer work, where an allowlist either
blocks ordinary commands or grows into the same unbounded list from the other
side. Recorded rather than dismissed — if the guard ever gains a fail-closed
mode, this is the shape it takes.

**Demote the parse layer to advisory and put the boundary elsewhere.** Partly
adopted: decision 1 *is* the demotion of the claim. Not adopted in full, because
removing the block tier would discard precise, correct verdicts on the cases the
registry does match, and abcd is host-delegated by default ([adr-25](0025-host-delegated-llm-default.md))
— the execution-layer boundary belongs to the host, not to abcd.

## Consequences

- **The guard stops lying about the cases it cannot see.** Every unrecognised
  wrapper becomes a loud warn instead of a confident allow, and the fix is not
  hostage to a list staying complete.
- **False positives become the accepted cost, with a known shape.** Tier 2 fires
  on unquoted hazard-shaped text under an unrecognised argv[0]: 0 fires across
  1,144 repo-mined lines, 21 of 79 on an adversarial corpus (2 of them true
  positives). Quoting silences it; a `git`/`gh`/`rm` argv[0] silences it — so
  `git grep git push --force` is quiet while `rg …` warns, an arbitrary asymmetry
  that must be documented rather than explained away. This repo's subject *is*
  the hazard registry, so expect interactive hits here specifically.
- **A new standing obligation: the DoS bound is load-bearing, not an
  optimisation.** The start cap and the O(N) pass carry a benchmark test, and any
  future change to matching re-runs it. The guard gates the hook; a slow guard is
  an outage.
- **Wrapper and payload-flag tables are documented as one platform's grammar.**
  `nsenter unshare chrt taskset ionice setsid flock stdbuf runuser` do not exist
  on macOS, and `su`, `script`, `chroot`, `nice` carry BSD grammars there. CI runs
  both platforms, so the table says which platform it encodes.
- **`su -c` is not guaranteed to be shell grammar** — it runs the target user's
  login shell, overridable with `-s`. This costs false negatives only (a mis-parse
  yields a non-match, never a false block), and it belongs in the scope statement
  rather than in a claim that the family parses uniformly.
- **Residuals stay open and named**, so no reader mistakes the fix for a boundary:
  top-level `curl … | bash` is a silent allow today (`pipesIntoInterpreter` is
  consulted only inside a payload — separate defect, found while testing), and
  `echo <b64> | base64 -d | sh`, `make push-force`, `npm run deploy`, shell
  functions, aliases and variables are permanently invisible to any parser.
- **The next enumeration gap is no longer an incident.** A missing wrapper name
  degrades a block to a warn instead of to silence, which is the whole point of
  spending the false positives.

## References

[cc-permissions]: https://code.claude.com/docs/en/iam "Claude Code — permissions and command-pattern matching (Anthropic docs)"
[cursor-security]: https://cursor.com/security "Cursor — run modes, terminal allowlisting, and security advisories (GHSA-82wg-qcm4-fp2w / CVE-2026-22708)"
[cwe-184]: https://cwe.mitre.org/data/definitions/184.html "CWE-184: Incomplete List of Disallowed Inputs (under CWE-693 Protection Mechanism Failure)"
[ranum-dumbest]: https://www.ranum.com/security/computer_security/editorials/dumb/ "The Six Dumbest Ideas in Computer Security (Ranum, 2005) — 'Enumerating Badness'"
[garfinkel-traps]: https://www.ndss-symposium.org/ndss2003/ "Traps and Pitfalls: Practical Problems in System Call Interposition Based Security Tools (Garfinkel, NDSS 2003)"
