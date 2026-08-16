# Design: the guard and the execute-a-string wrapper family (iss-200)

**Status:** design, revision 4 (2026-08-16). Rev 1, 2, and 3 each had
independent review; the rev-3 security re-review found the approach **sound**
(all five rev-2 findings closed) with one text-precision must-fix plus two
tightenings, folded in here. The remaining items were wording gaps an
implementer could get wrong, not approach flaws. Design-first by maintainer decision, because two
prior code patches to iss-200 were BLOCKed — the approach, not a detail, was
wrong each time. Maintainer decisions of record: the **split failure posture**
and **folding iss-231**.

## Problem

`internal/core/guard/tokenize.go:26` records a deliberate v1 gap: a command
carried as a **data argument** — `sh -c '<payload>'`, `bash -lc "<payload>"`,
`eval '<payload>'` — is one opaque token, never tokenized, so a hazard inside
it is never matched. `env -S<value>` (GNU env re-splits the value and executes
it) is the **same class**; iss-200 is the specimen: `env -S 'gh repo delete
o/r'` sails past the `gh` delete blocker, and `sh -c 'git push --force'` does
the identical thing, currently accepted as out-of-scope-for-v1.

So iss-200 is "where does the whole execute-a-string family sit", not "parse
env". Two prior env-only patches failed: v1 closed one spelling; v2
(splice-and-restart in `commandOf`) went quadratic — a DoS on the very hook
that gates the command — and still let the in-scope spellings through because
`strings.Fields` is not env's splitter and a command-position restart drops
env's option parser.

## Prerequisite: the warn tier is silent today (iss-231) — fold the fix, hook only

`internal/surface/cli/guard.go`'s **hook** maps `VerdictWarn` to
`Fprintln(stderr, message); return nil` — exit 0. The adjacent `failOpen`
comment documents that a PreToolUse hook exiting 0 has its stderr **discarded**;
that is why fail-open uses **exit 1** (loud, non-blocking) and block uses **exit
2** (deny, replayed to the agent). So every warn today writes a message nobody
sees and runs as if allowed.

**Decision (maintainer, 2026-08-16): fold iss-231.** In the **hook** switch
(guard.go ~205), change the `VerdictWarn` case from `return nil` to exit 1 — the
same loud-but-non-blocking status fail-open uses. This changes behaviour for
*all* warn entries (e.g. `git-reset-hard`): they become visible. The existing
`TestGuardHookWarnAllowsAndSurfaces` asserts exit 0 and must invert to exit 1.

**Scope it to the hook only.** The `check` verb's warn path writes to **stdout**
(not discarded) at exit 0 and documents `block=1 / warn=0 / fault=2` as a
scriptable contract — do **not** change it. iss-231's discard rationale is
hook-specific.

## Decision: option D, whole family, split posture

Map the guard's three verdicts to severity, for the whole family
(`sh`/`bash`/`dash`/`eval`/`env -S`):

- **Block** the serious/irreversible — a payload that inspects to a **blocker**.
- **Allow** the harmless — a payload that inspects cleanly and matches nothing.
- **Warn (loud)** the visible middle — a payload we *choose* not to block.

The hard case is a payload we **cannot inspect**. Posture splits by wrapper:

- **`env -S` uninspectable → BLOCK (fail-closed).** `env -S` at a command line
  is rare (its common use, shebangs, does not reach the Bash tool) and its
  purpose is to disguise execution, so "cannot read it → refuse it" honours
  "absolutely block the serious/irreversible". No per-invocation override exists
  on a hook; the block message names the remedy (rewrite without `-S`; the
  committed registry-disable is the reviewable escape).
- **`sh`/`bash`/`dash`/`eval` uninspectable → WARN (loud).** Common and honest;
  blocking every `sh -c "$(…)"` would be a false-positive storm.

Inspectable payloads in both families match normally (block if they hit a
blocker, allow if clean).

## Mechanism

### Where it lives (the v2 trap, designed out)

`Check` (guard.go:348) tokenizes the whole line **once**, then loops entries
over the shared `segs`. Expand `segs` **once, in `Check`, immediately after
`segs, err := tokenize(command)` and before the entry loop** — never inside the
per-entry `commandOf`/`matchesAny` path. Every entry then sees the payload
segments for free, the matchers are unchanged, and cost is `O(line × depth)`.
Positive placement in `Check` is load-bearing: a per-entry callee is what made
v2 quadratic. The original opaque wrapper segment stays in the slice and simply
matches nothing (no entry has command `sh`/`env`), so there is no double count.

### Recognising a family member — two different traversals

**Shell `-c` family reads `commandOf` output.** `sh`/`bash`/`dash`/`eval` are
not wrappers whose flags swallow the payload, so after the ordinary wrapper walk
(`sudo sh -c …`, `env sh -c …` still reach the `sh`), `commandOf` yields command
`sh` with the `-c <payload>` in its args. The payload is the argument to `-c`
(`-lc`/`-ic` etc.); `eval` joins its args with spaces first.

**`env -S` reads RAW segment tokens, NOT `commandOf` output.** This is the fix
both rev-2 reviews caught. `env` is a registered wrapper (`match.go:17`) and
`-S`/`--split-string` are in `wrapperValueFlags["env"]` (`match.go:43`), so the
ordinary `commandOf`/`skipWrapperArgs` walk **consumes and discards** the `-S`
value before any recognizer built on `commandOf` output could see it — a literal
"recognize on `commandOf` output" re-opens iss-200. Instead, env-S extraction is
a dedicated pre-pass over the raw tokens. Walk the leading wrapper/assignment
chain (the same wrapper set, so `sudo env -S …`, `timeout 5 env -S …` reach the
`env`), and **at every token whose basename is `env`** — not just the first —
scan its following raw tokens for the split-string spellings (separate `-S <v>`,
glued `-S<v>`, the `--split-string=<v>` / `--s…--split-string` prefix range, and
a short cluster carrying `S`, `-iS…`) *before* the value-flag skip runs. An
`env` carrying no split-string flag is stepped over as an ordinary wrapper and
the scan re-enters at the next command-position token. This is load-bearing:
`env -i env -S 'gh repo delete o/r'` uses two *known* wrappers, and a pre-pass
that scanned only the first `env` would let `commandOf` consume the inner `-S`
value and silent-allow the deletion — so `env -i env -S '<blocker>'` is a
required known_bad fixture. Once the pre-pass owns `-S`/`--split-string` they
need no longer be double-handled by `skipWrapperArgs` (removing them from
`wrapperValueFlags["env"]` is a safe implementation choice).

### Chain namespacing (prevents false blocks)

`matchesAny → precededByCD` keys on `segment.chain`. A fresh `tokenize(payload)`
numbers chains from 0, so a naive append collides indices. Rule: append each
payload's segments after the top-level segments, **offsetting its chain numbers
by a running maximum**, so every payload occupies a disjoint chain range while
preserving its own internal chain structure. Then `cd foo && git fetch ; sh -c
'rm -rf x'` puts the payload `rm` in its own `cd`-less chain (no false block),
while `sh -c 'cd x && rm -rf *'` keeps the `cd`+`rm` adjacency (real block).

### Shell `-c` family payload

1. **Uninspectable → loud warn.** If the payload contains command substitution
   `$(…)` or backticks, expansion `${…}` / `$((…))` that could resolve to a flag
   or command, or a pipe into an interpreter (`… | sh`, `… | bash`) — the
   tokenizer cannot see the real command; emit a synthetic loud warn. (This is
   an explicit pre-check; `tokenize` does not treat `$(` as inspection-defeating
   on its own.)
2. Otherwise tokenize the payload once and match its segments. Chained hazards
   are caught: `sh -c 'cd x && rm -rf *'` → block; `sh -c 'git push --force'` →
   block; `sh -c 'make build'` → allow. No new false-*block* surface: a payload
   blocks only if it already contains a top-level blocker.

### `env -S` payload

The payload obeys **env's** splitter, not the shell's.

1. Recognise every spelling (pre-pass, above).
2. **Decode env's fixed escape table** (`\_`→space, `\t \n \r \f \v \# \\ \$`,
   GNU env's finite set) and, if the decoded value is a plain command with no
   residual env-special construct, split on whitespace and match normally: `env
   -S 'gh\_repo\_delete\_o/r'` **blocks** like `env -S 'gh repo delete o/r'`.
3. **Env-special → BLOCK (fail-closed).** If, after decoding, the value still
   holds `${…}`/`$VAR` expansion, a leading env option the guard would have to
   re-parse, or a quote env removes in a tokenisation-changing way, block. Never
   guess-and-allow (v2), never guess-and-warn-silently (iss-231). This is
   fail-closed by construction: the only risk is over-blocking, bounded by the
   STOP condition, never under-blocking. (BSD env — this repo's dev runtime —
   errors on any backslash escape outside its table, so an escape the guard
   leaves literal is one env itself refuses to run: fail-safe either way.)

### The synthetic (entry-less) Decision — severity pools

An uninspectable verdict has no registry entry (the `Pattern` language matches
command-position tokens; it cannot express "value contains a backslash"), so
`Check` constructs its `Decision` directly. Merge with registry matches by
**severity pool**, not by a single "registry outranks synthetic" rule:

- The **block pool** = registry blocker entries + any synthetic block. The
  **warn pool** = registry warn entries + any synthetic warn.
- Verdict = block if the block pool is non-empty, else warn if the warn pool is
  non-empty, else allow. So a synthetic block never hides behind a registry
  warn, and a registry blocker still outranks a synthetic warn.
- Among equal severity, a concrete registry entry supplies `Message`/`Successor`
  (its lesson is more specific); a synthetic verdict supplies them only when it
  is the sole member of the winning pool. `Matches` lists every fired id,
  including the reserved synthetic id (`execute-string-uninspectable`,
  documented as reserved so no entry may claim it).
- **Winner construction branches on synthetic:** a synthetic id must not index
  `r.Entries` (that yields a zero `Entry` and an empty `"Blocked … (): . Run
  instead:"` message) — build its `Decision` directly.
- **Fields carried:** `Family` and `Reason` on the `Decision` (needed to build
  the synthetic `Message`). The synthetic `Decision` must **also populate the
  `Why`/`Successor` fields the check-verb renderer (`writeGuardDecision`) reads**
  — or that renderer must fall back to `Message` — so the human report is not
  blank; the block still fires either way, but a thin report is a defect. The
  **raw payload is NOT carried** — it has no reader in this change; the later
  host-delegated interpretation intent adds the raw-payload field when it adds
  the reader (wired-or-it-isn't-done).

### env-special is a whitelist gate, not a denylist

To keep the env fail-closed guarantee airtight, split-and-match runs **only**
when the escape-decoded value passes a strict plain-command whitelist: no
residual backslash, no `$`, no quote, no leading `-`, no `#`, no other env
metacharacter. **Anything failing the whitelist blocks** — the gate is not "block
if it contains one of these known-special constructs" (a denylist a novel env
form could slip through) but "allow inspection only if it is provably plain,
else block". Over-blocking is the fail-safe direction, bounded by the storm STOP.

### Cost bound and depth

Each payload is tokenized **at most once per `Check`**, in `Check`, never
re-spliced into the per-entry loop — `O(line × maxPayloadDepth)`, no quadratic
restart. `maxPayloadDepth` (proposed 2) bounds nesting. **Within** the budget
the posture applies. **Past** the budget the guard has lost the thread — it
cannot see whether an `env -S` hazard hides deeper — so **depth-exceeded is
fail-closed: BLOCK**, regardless of the outermost wrapper's family. This closes
the rev-2 hole where a shallow `sh -c` layer laundered a buried `env -S` block
into a warn (`sh -c 'sh -c "… env -S <hazard>"'`), and treats execute-a-string
nesting past two levels as the red flag it is (deep-nested pure shell is rare
and suspicious; the registry-disable remains the reviewable escape).

### Inner-tokenize errors

If `tokenize(payload)` errors (unterminated quote, trailing backslash), treat as
uninspectable and apply the posture (env → block, shell → loud warn) — never
propagate as a whole-command parse failure that fails the hook open.

## Verdict mapping (against the real registry)

Registry facts (`defaults/guard.json`): the only `rm` blocker is
`rm-rf-after-cd-chain` (`after_cd:true`); `git-reset-hard` is **warn**;
`git-push-force`, `git-push-force-refspec`, the `*-no-verify`, and the `gh repo
delete` / `gh api -X DELETE repos` entries are **ungated blockers**; `rm -rf
./build` is **known-good**. Bare `rm -rf` does not block and must not be made to.

| Payload | Family | Outcome |
|---|---|---|
| `env -S 'gh repo delete o/r'`, `sh -c 'git push --force'` | ungated blocker | **block** |
| `sh -c 'cd x && rm -rf *'` | payload cd-chain fires `rm-rf-after-cd-chain` | **block** |
| `env -S 'gh\_repo\_delete\_o/r'` | escapes decoded → blocker | **block** |
| `sh -c 'git reset --hard'` | warn entry | **warn (loud)** |
| `env -S ls`, `sh -c 'make build'` | clean, no match | **allow** |
| `env -S '-i gh repo delete o/r'`, `env -S 'x=${Y} …'` | env-special | **block** |
| `sh -c "git push $(printf -- --force)"`, `sh -c "…|sh"` | shell uninspectable | **warn (loud)** |
| nesting past `maxPayloadDepth` | lost visibility | **block** |

## Documented residuals (do not widen — the warn-storm STOP)

- **Bare `$VAR` / ANSI-C `$'…'` in a shell payload** are not in the uninspectable
  pre-check, so `eval 'git push $F'` / `sh -c $'…'` tokenize to a non-matching
  form and **allow**. This loses no prevention — uninspectable shell never blocks
  (a block needs a literal inspectable hazard) — it is a visibility gap only.
  Warning on every `$VAR` would trip the storm STOP.
- **`python -c` / `perl -e`** payloads are not shell and not shell-tokenizable;
  treat as the shell-family posture (**loud warn**), not allow.

## Fixtures and STOP conditions

- **Corpus:** bad fixtures use real ungated blockers (`gh repo delete`, the
  git-push spellings) or the cd-chained `rm` shape *inside the payload*; confirm
  each fires its entry under `TestBundledEntriesPassAdmissionGate` before listing
  it. Good: `env -S ls`, `sh -c 'echo hi'`, `bash -lc 'make build'`. The
  synthetic verdicts have **no entry**, so they are not covered by the entry gate
  — they need their own test asserting block/warn + the `Family`/`Reason` fields.
- **STOP — false-positive storm.** Before shipping, measure block+warn rate on a
  corpus of real agent commands. A warn rate that trains users to ignore warns,
  env-special blocks hitting real workflows, or depth-exceeded blocks on
  legitimate deep nesting is a STOP: re-scope and report.
- **STOP — env-special is a construct denylist.** If enumerating "env-special"
  proves open-ended, the default stays block (fail-closed), never a silent allow.

## Open questions (resolve at implementation)

- `maxPayloadDepth` value (2 vs 3) — empirical.
- Whether this graduates to an ADR ("the guard treats the execute-a-string
  family as first-class") plus a `guard-tiers-map-to-reversibility` principle —
  recommended once it ships, as a follow-on record, not this note.
