# `/abcd` — Top-Level Where-Am-I Status Board

The bare top-level `/abcd` command (no sub-verb) is the cross-command
re-orientation surface (itd-20): type `/abcd` and see, at a glance, where you
left off across the whole abcd setup. It is STRICTLY read-only — the
status-render module opens files for reading only and never writes.

**Design target (itd-20 — the intent is in `intents/planned/`, `spec_id: null`;
no such board is on any shipped surface).** The shipped bare `abcd` renders a
four-field read-only snapshot — the directory, `is_git_repo`, `has_record` (is
`.abcd/development` present), and the `.abcd/` `work_tiers` that exist
(`internal/core/core.go`, `StatusInfo`); the plugin invokes it as `abcd --json`
(`commands/abcd.md`). The six-section board, its per-source known-state lines,
staleness thresholds, and performance bounds described below are itd-20's
design, unbuilt.

This is the top-level command (`commands/abcd.md`), distinct from the
per-verb bare renders (`/abcd:ahoy`, `/abcd:capture`, …). Each per-verb bare
render is scoped to its own command's surface; this board is the *cross-verb*
answer to "what's the state of my abcd project right now?".

The `/abcd` surface routes every verb through the transport-agnostic core (the
Cobra CLI is the front door today; an MCP server follows later, per
[adr-23](../../decisions/adrs/0023-transport-agnostic-core.md)) — it no longer
hides any bundled dependency behind the wrapper.

## Sub-verbs and aliases

Bare `/abcd` renders the six-section board. `status` and `help` are
POSITIONAL ALIAS tokens that route to the SAME render — their output is
byte-identical to bare `/abcd`. **Design target (itd-20):** no such alias
routing is on any shipped surface — the plugin command (`commands/abcd.md`)
documents only `status` as a positional alias, while the CLI itself has no
alias: `abcd status` is refused as an unknown command and `abcd help` prints
Cobra usage text, not the render.

| Token | Behaviour |
|-------|-----------|
| *(bare)* | Render the six-section where-am-i board |
| `status` | Alias — byte-identical to bare |
| `help` | Alias — byte-identical to bare (bare-as-render IS the help) |

Any other token is refused; the alias surface is a closed set. **Design target
(itd-20):** the shipped CLI refuses an unknown token with **exit 1** (Cobra's
`unknown command`), not exit 2, and applies that same refusal to `status` and
`help` (neither is a CLI alias today); the exit-2 closed-set is itd-20's design.

## SD001 alias rationale (investigation-gated)

The [SD001 rule](../02-constraints/04-naming.md) (`02-constraints/04-naming.md`,
the "Bare-command-as-render discipline" + "Forbidden sub-verbs" sections) was
read for this task. Its explicit forbidden set is `show` / `stats` / plain
`list` / `view` — sub-verbs that *name what bare already renders*. `status`
and `help` are NOT in that forbidden set, and the rule's spirit ("a sub-verb
must earn its existence by doing something bare cannot") is satisfied by
treating `status` and `help` as **aliases, not sub-verbs**: they route to the
identical bare render and add no distinct behaviour. They exist only so the
intuitive `/abcd status` / `/abcd help` land on the same board — improving
discoverability, which is exactly what SD001 protects. The alias tokens
therefore satisfy SD001 rather than tripping it. (Had they instead introduced
divergent behaviour, they would be forbidden sub-verbs and the design would be
reconsidered — the gate is real, not a rubber stamp.)

## spc-17 stub replacement (investigation-gated)

itd-20 was planned assuming a spc-17 probe stub for the bare top-level command
would be *replaced*. Investigation found NO such stub: the spc-17 work is
credited with probe/bare renders for the *sub-verb* surfaces — of which only
`/abcd:launch` is on a shipped surface today. `/abcd:disembark`, `/abcd:embark`,
and an `/abcd:intent` surface are design targets with no command file
(`commands/abcd/` ships `ahoy`, `capture`, `docs`, `history`, `launch`,
`memory`, `version`) and no CLI verb — and spc-17 never shipped a top-level
`commands/abcd.md`. `git log` confirms `commands/abcd.md`
has never existed. This task therefore CREATES the top-level command fresh
rather than replacing a stub; the "stub replacement" premise is recorded here
as not-applicable so a later reader does not hunt for a stub that was never
shipped.

## The six sections (fixed order)

The render emits exactly these six sections, always in this order. Each
source is a local-filesystem read; each source's absence maps to a NAMED
known-state line (never an exception, never a silent omission).

1. **Project + visibility** — project name (the cwd's directory name) and
   `repo.visibility` from `.abcd/config.json`.
2. **Lifeboat presence / age** — the `.abcd/lifeboat/` directory's mtime,
   rendered as an age in days with a staleness flag.
3. **Dev-sync staleness** — v1 terminal known-state line (see below).
4. **Recent logbook** — the last five logbook ENTRIES by mtime. The logbook
   is organised `<category>/<entry>` (e.g. `grill/20260701T000000Z-itd-67`),
   so the render descends exactly ONE bounded level into each category
   directory and lists entries labelled `<category>/<name>`; loose files at
   the logbook root are entries in their own right. Two levels total — never
   a recursive walk. Fewer than five → render what exists with a count.
5. **Active intents** — intents in `intents/planned/` and `intents/shipped/`
   that carry a linked `spec_id` whose spec is NOT `done`, rendered with the
   spec status read from the native spec store.
6. **Suggested next actions** — a short bullet list keyed off the state above.

## Per-source known-state table

**Design target (itd-20; unbuilt).** None of these known-state lines exist in
the shipped render — `core.Status` reads only `.git`, `.abcd/development`, and
the `.abcd/` work tiers (`internal/core/core.go`); it does not read
`repo.visibility`, `.abcd/lifeboat/`, a dev-sync record, `.abcd/logbook/`,
linked intents, or any spec store.

| Source | Absent / unreadable → known-state line |
|--------|----------------------------------------|
| `.abcd/` directory | outside-abcd guidance message (single line, replaces the whole board) |
| `repo.visibility` in `.abcd/config.json` | `visibility: unknown (no repo.visibility in config)` |
| `.abcd/lifeboat/` | `lifeboat: never packed — run /abcd:disembark when ready` |
| dev-sync last-run artifact | `dev-sync: no dev-sync record …` (v1 terminal — see below) |
| `.abcd/logbook/` (empty / unreadable) | `logbook: no entries yet` |
| linked intents (none with a live spec) | `intents: no planned or active intents with a linked spec` |
| spec status via the native spec store (missing / fail / timeout / unparseable) | `unknown` |

## Staleness thresholds (decided here)

**Design target (itd-20; unbuilt).** No lifeboat-age, staleness, capping, or
timeout logic exists in the shipped binary — the status path is four `isDir`
checks (`internal/core/core.go`) with no directory walks or thresholds.

| Signal | Threshold | Rationale |
|--------|-----------|-----------|
| Lifeboat age → `(stale)` | 7 days | A point-in-time rescue snapshot a week old warrants a re-pack cue without nagging on daily work. |
| Dev-sync staleness | n/a (v1) | No durable last-run artifact exists — see below. |

## Dev-sync source (probed, recorded — v1 terminal stub)

The `abcd dev-sync work` migration surface
(into the structured `.abcd/work/issues/` `iss-N` ledger) is a **design target —
no `dev-sync` verb is on any shipped surface** (`abcd --help` lists no such
command, and no dev-sync code exists under `internal/` or `cmd/`). Even in that
design it is **migration logic, not a durable last-run timestamp**: no config field and no history-store
record captures "when dev-sync last ran". No such artifact exists anywhere in
the repo.

Therefore section (3) is PERMANENTLY the `no dev-sync record` known-state line
in v1 — the dev-sync-staleness signal does **not** function until a state
substrate exists. This is an ALLOWED terminal state, recorded here and in the
itd-20 intent so it is not read as a shipped capability. Adding a dev-sync
state substrate is explicitly OUT of scope for spc-83.2 (and named in the .5
out-of-scope update).

## Performance bounds (NFR)

**Design target (itd-20; unbuilt).** These bounds constrain the unbuilt render:
the shipped status path does no directory walks, last-N sorting, capping, or
timed spec-store reads (`internal/core/core.go`).

- Bounded directory reads only — `.abcd/logbook/` is read to a depth of two
  (category → entry) and the two named intent directories to a depth of one;
  NO recursive full-tree scans.
- Last-N sorting only (logbook capped at 5, active intents capped at 10) — no
  full-history loads.
- The native spec-store status read runs under a 5-second timeout; expiry maps
  to the `unknown` status line so a slow read never wedges the render.

## Zero-writes guarantee

The command markdown performs zero writes; the render module performs zero
writes. **Design target (itd-20):** the two tests that would prove this — a
static zero-mutation lint over the status-render module and an fs-snapshot test
that asserts a render over a populated fixture repo mutates nothing at run
time — are unbuilt. The shipped status path (`internal/core/core.go`, `Status`)
has only `TestStatusBareDir` and `TestStatusWithRecordAndGit`
(`internal/core/core_test.go`), which assert field values on temp dirs; there is
no distinct status-render module yet.

## Related documentation

- Naming / SD001 discipline: [`../02-constraints/04-naming.md`](../02-constraints/04-naming.md)
- Intent: `itd-20` (`../intents/…/itd-20-top-level-abcd-dispatcher.md`)
- The per-verb bare renders this board complements: [`05-intent.md`](05-intent.md), [`01-ahoy.md`](01-ahoy.md)
