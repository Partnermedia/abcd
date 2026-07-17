---
id: spc-4
slug: start-the-transcript-clock
intent: itd-89
---
# start-the-transcript-clock

## Summary

spc-4 delivers itd-89's trigger: the session-transcript store
(`internal/core/history`, adr-29) existed correct and unused, and this spec
wires the missing caller so the corpus starts accruing. It adds one hidden,
operator-internal verb — `abcd hook session-end` — plus the `SessionEnd` entry
in `hooks/hooks.json` that invokes it. When a session ends, the hook reads the
harness payload from stdin, resolves the repo's root-commit SHA, reads the
transcript defensively, and hands it to the already-shipped redacting
`history.Capture`. No user-facing surface, no flag, no command to remember.

## Approach

Follows the shipped `hook prompt-router` pattern exactly (itd-3): a hidden Go
subcommand under the `hook` sub-tree, fail-closed, always exit 0, diagnostics
out of band on stderr, stdout kept empty. A new verb was required because
`history capture` reading stdin *requires* `--session <id>`, while the hook
delivers its session id inside the JSON payload.

- **Verb:** `internal/surface/cli/cli.go` (`hook session-end` block; RunE
  parses the payload, resolves cwd → `ahoy.Detect` root SHA, calls
  `history.Capture(captureRoot(cwd), rootSHA, sessionID, raw, "native")`).
- **Defensive read:** `readTranscript` opens `O_RDONLY|O_NONBLOCK` (a FIFO or
  device node cannot hang the hook), accepts only regular files, and enforces
  the size cap before reading.
- **Wiring:** `hooks/hooks.json` gains a `SessionEnd` entry running
  `"$CLAUDE_PLUGIN_ROOT/abcd" hook session-end`. Without it the verb is dead
  code; with it the corpus accrues with zero user action.
- **Redaction is not touched:** the two-stage, fail-closed pass lives in
  `history.Capture` and this spec only calls it, per the intent's scope.

### Deviation: SessionEnd, not Stop

The plan's literal wording said `Stop`. `Stop` fires once per assistant
*turn*, and the transcript file grows through the session, so a Stop-wired
capture would store a fresh, larger superset every turn — `history.Capture`'s
sha256 dedup only collapses byte-identical re-captures, which never happens on
a live transcript. `SessionEnd` fires once at orderly termination, and by the
harness contract its exit code and stdout are ignored — the exact shape a
fail-closed, non-blocking side-effect hook needs. Documented trade-off: a hard
crash (`SIGKILL`) fires no `SessionEnd`, so an uncleanly killed session is not
captured; the intent records this as deliberate (per-turn crash-resilience
would need session-keyed dedup in shipped core).

### Default: the 64 MiB transcript cap

`maxTranscriptBytes = 64 << 20`. Generous for a JSONL session log and bounded
so a pathological file cannot stall the redaction pass inside a session hook.
An over-cap transcript is refused *whole* — capturing a truncated prefix would
break the sha256 idempotency key — and the refusal is logged on stderr, so an
over-cap session is visible rather than silently absent.

## Milestones as delivered

1. `abcd hook session-end` verb, fail-closed on every input path
   (`internal/surface/cli/cli.go`), with `readTranscript` defensive open.
2. `hooks/hooks.json` `SessionEnd` wiring (verb reachable from the plugin
   surface — wired-or-it-isn't-done).
3. Test corpus in `internal/surface/cli/hook_session_end_test.go` covering
   every acceptance criterion end-to-end on the hook path (see mapping below),
   over the core coverage in `internal/core/history/history_test.go`.

## Acceptance-criteria satisfaction

AC as numbered in itd-89 → covering evidence (all in
`internal/surface/cli/hook_session_end_test.go` unless noted):

1. **Hook fires → redacted + stored, `history list` shows it, no user
   action** — `TestHookSessionEndCapturesTranscript` (drives the verb via
   stdin payload, asserts the record via `history.List`); the `SessionEnd`
   entry in `hooks/hooks.json` is the no-user-action wiring.
2. **Long session → exactly one record (SessionEnd, not Stop)** —
   `TestHookSessionEndOneRecordPerSession`; `hooks/hooks.json` carries a
   `SessionEnd` entry and no `Stop`-wired capture.
3. **Same transcript offered twice → one record** —
   `TestHookSessionEndIsIdempotent`; core:
   `TestCaptureIdempotentOnSourceSHA` and
   `TestCaptureIdenticalSourceDistinctSessionsWritesBoth` (bounds the dedup to
   session+kind+source, so a distinct session is never swallowed).
4. **Malformed payload / absent path / not a regular file / hostile session
   id / non-repo cwd → writes nothing, exit 0, reason on stderr** —
   `TestHookSessionEndNeverBlocksTheHost` (seven payload subcases, each
   asserting exit 0, zero records, and a non-empty stderr reason); core:
   `TestCaptureRejectsBadInput`.
5. **FIFO or device node → the open does not block** —
   `TestHookSessionEndDoesNotBlockOnIrregularFiles` (a FIFO with no writer and
   `/dev/null`, bounded by a watchdog so a regression fails in seconds).
6. **Nothing on stdout, diagnostics stderr-only** —
   `TestHookSessionEndWritesNothingToStdout`.
7. **Secrets/home paths redacted; surviving hard-fail span → no file at
   all** — `TestHookSessionEndRedactsOnThisPath` (redaction demonstrably runs
   on the hook path), `TestHistoryCaptureFromSubdirHonoursRepoPiiConfig`
   (per-repo `pii.json` honoured from a subdirectory),
   `TestHookSessionEndRefusesResidualHardFail` (a hard_fail span built to
   survive stage-one masking is refused whole: no file, reason on stderr, exit
   0); cap default: `TestHookSessionEndRefusesOverCapTranscript`. Core:
   `TestCaptureRedactsSecretsAndHomePaths`, `TestSurvivingCallerHomeBackstop`,
   `TestRedactionEngineStripsUsersHomePath`.

## Out of scope (unchanged from the intent)

Retrieval/search over the corpus (Pass B, itd-36), importing historical
sessions (nothing exists to import), changes to the redaction engine, any
user-facing surface, and `SubagentStop` wiring (deliberately undecided — it
shares Stop's per-turn growing-superset problem).
