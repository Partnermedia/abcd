---
schema_version: 1
id: "iss-200"
slug: "guard-env-split-string-glued-form-bypass"
severity: "critical"
category: "bug"
source: "agent-finding"
found_during: "bug-hunt loop round 9 (state issue #197), fix attempt for iss-196 blocked at pre-PR review by two independent reviewers"
found_at: "internal/core/guard/match.go:129-150 (skipWrapperArgs), match.go:43 (wrapperValueFlags)"
resolution: "Closed the execute-a-string wrapper family bypass: payloads expanded once in Check, env split-string owned by a raw-token pre-pass with a fail-closed whitelist, shell payloads inspected or loud-warned, depth capped at 2"
impact: fix
---

guard's env wrapper table listed -S/--split-string as a value-owning flag for the SEPARATE-TOKEN spelling only ('env -S git push --force origin main'), but GNU env -S/--split-string also accepts the value glued to the flag ('env -Sgit push --force origin main', 'env --split-string=git push --force origin main'), which re-splits identically and executes identically; a fix attempted in bug-hunt loop round 9 (bugfix/env-split-string-guard-bypass) closed only the separate-token form and was BLOCKed at pre-PR review by both an independent correctness reviewer and an independent security reviewer, who each independently reproduced that the glued forms remain a complete, silent bypass of every guard blocker on the live PreToolUse path; suggested remediation (from both reviewers, converging independently): in commandOf/skipWrapperArgs, treat -S/--split-string as a value-owning env flag in all three spellings — the exact separate-token form (-S as its own token, value in the following token), the glued form (-S<value>, len>2), and the --split-string=<value> form (value non-empty) — and in every case read the first word of <value> as command position, rather than either treating the whole glued token as a consumed wrapper flag or skipping the following token outright in the separate-token case — no nested re-tokenizing required, contained to the same two functions
## Attempt 2 (2026-08-15) — BLOCKED by both independent reviewers; the splice-and-restart approach is rejected

`bugfix/iss-200-env-split-string-guard-bypass-v2` (unpushed) implemented the
converged remediation above — split the `-S`/`--split-string` value and read
its first word as command position, contained to `commandOf`/`skipWrapperArgs`.
The three plain spellings block, but an independent correctness review and an
independent security review each BLOCKed the diff. The approach itself, not a
detail of it, is the problem, so it is recorded here rather than patched again:

1. **New quadratic cost regression (security, blocking).** The splice-and-
   restart re-allocates and re-walks the whole token slice once per `-S`
   encountered, so `commandOf` becomes O(n²) where the pre-fix walk was linear.
   Measured end-to-end through the front door: a 384 KB crafted payload took
   ~35 s wall (the same input is ~4 ms on `main`), scaling cleanly 4× per
   doubling, so the 1 MiB stdin cap the code already enforces admits roughly
   five minutes. The PreToolUse hook that is supposed to gate the command is
   the thing stalled. This defect is *introduced* by the fix; `main` has no
   such behaviour.
2. **The three in-scope spellings remain defeatable (correctness + security,
   blocking).** `strings.Fields` is not env's splitter and the restart drops
   env's option parser, so all of these still read the wrong command position
   and return `allow`: a value with a quoted flag (env strips the quotes,
   the guard keeps them); a value using env's `\_` argument separator (no
   whitespace, so `Fields` yields one word); and a value whose first word is
   an env option such as `-i`/`-u` (real env re-parses it as a flag, the walk
   reads it as the command). Verified against coreutils 9.10 and macOS BSD env.

**The complete env `-S` spelling set the next attempt must cover** (this is the
acceptance corpus, consolidated here rather than split across new issues):
separate-token, glued `-S<value>`, `--split-string=<value>`, GNU long-option
abbreviation (`--split=`, `--s=`), a bundled short cluster carrying `-S`
(`-iS…`, `-iSgit…`), a value with env quote-removal, a value using the `\_`
separator, and a value whose first word is itself an env option or `--`.

**Design question for the maintainer — this needs a decision before a third
attempt, per the pattern that shelved iss-189/190.** Faithfully modelling
env's parser (bounded, quote-aware, option-aware) is materially more than the
"no nested re-tokenizing" the converged remediation assumed. The alternative
is to treat `env -S<anything>` as an explicit, tested **non-match** the way
`tokenize.go` documents its `sh -c` v1 gap — but that means *allow*, which for
a denylist guard is a fail-open, the opposite of safe. A denylist guard cannot
simply "refuse when unsure" without false-blocking legitimate `env -S` use.
Resolving that tension (parse-faithfully-but-bounded vs. a narrowly-scoped
refusal for the force-push class vs. accepting a documented residual) is the
real work, and it is a design call, not a coding one.
