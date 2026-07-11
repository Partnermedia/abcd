# itd-3 rules-loader build — learnings

End-of-run note from the autonomous build of the modular rules loader (itd-3),
recording what shaped the design and the non-obvious decisions a future session
would otherwise re-derive. The design of record is
[`../../plans/2026-07-11-itd-3-rules-loader.md`](../../plans/2026-07-11-itd-3-rules-loader.md).

## What landed

Five phases, TDD, each a commit on `auto/itd-3-rules-loader-design`:

1. `internal/core/rules` — transport-free core: binary-embedded default domains,
   per-repo `.abcd/rules.json` merge, recall matching, dedup signatures.
2. The prompt-router hook (`abcd hook prompt-router` / `-reset`) — security-reviewed.
3. The `abcd rules [domain]` verb — the vendor-neutral front door.
4. ahoy wiring — `hooks/hooks.json`, verifier reconcile, marker rewrite.
5. The `OPINIONS` default domain — pointers into `principles/`, not copies.

Plus a review-fix commit wiring `force_refresh_every_n` to config.

## Non-obvious decisions (why it is shaped this way)

- **Core capability, hook as one front door — not an adapter seam.** The
  host-agnosticism fork resolved by keeping merge/match/dedup/render in a
  vendor-neutral core exercised by `abcd rules`, and quarantining the Claude Code
  coupling to two thin `abcd hook` shims. This is the repo's existing
  core-plus-front-doors shape (like CLI/MCP), not the adr-22 seam model. Rule
  loading is one core verb, not a swappable backend.

- **The per-repo `rules.json` is an override skeleton, not a copy of the
  defaults.** `stepRules` writes `{schema_version,disabled,domains:{}}`; empty
  `domains` inherits every bundled default. Writing the full defaults per-repo
  would duplicate the canonical primitive (the binary-embedded defaults) into
  every repo — the exact one-canonical-primitive failure. Merge is per-field, so
  `{"ROADMAP":{"state":"dormant"}}` silences a domain while keeping its rules.

- **Event-driven refresh beats fixed-N (D1).** The real threat to rule
  persistence is compaction, not prompt count. A `SessionStart`/`PreCompact`
  reset clears the dedup ledger so the next prompt re-injects; the fixed-N
  counter is only a large backstop (default 15) for always-relevant domains.
  This came from the SOTA verdict, which cited the `SessionStart(source=compact)`
  hook as the surgical post-compaction signal.

- **Zero model-facing tokens on no-match (D3).** `Render(nil)` is empty; the hook
  writes nothing to stdout when nothing matches. Observability is out-of-band: a
  stderr diagnostic every invocation (turn/injected/bytes). A silently-broken
  hook is distinguished from a healthy no-match by the hook exit-code contract +
  the diagnostic, not by an always-on context header — which reframed the
  intent's "<200-token header" acceptance criterion.

- **Fail-closed but non-blocking at the trust boundary.** A malformed payload,
  unreadable/oversized/symlinked `rules.json`, or state error injects nothing and
  logs — but always exits 0, so the loader can never wedge a session. The
  injected text is abcd's own rule content only; the prompt is matched, never
  reflected. Session ids are sha256-hashed into the state filename, so a hostile
  id cannot traverse the state dir.

- **Star-command boundary without RE2 lookahead.** The pinned
  `(?:^|\s)\*([A-Z][A-Z0-9_]*)(?=$|\s)` uses a lookahead Go's RE2 lacks; it is
  enforced by matching `\*([A-Z][A-Z0-9_]*)` and hand-checking the surrounding
  bytes. Uppercase + boundary is what stops `* bullet`, `*.py`, and `path/*X`
  from parsing as commands.

## Process learnings

- **The prefer-sota order paid off.** The adversary's fit-challenge surfaced the
  duplication and global-source collisions *before* the verdict, so the SOTA read
  was fit-aware — and it materially improved the design over CARL in four ways
  (event-driven refresh, per-domain content dedup, keyword+alias recall, zero
  no-match tokens) rather than rubber-stamping the prior art.

- **The shipped code carried latent expectations.** `store.go`,
  `detect_test.go`, and `config.json`'s `rules` key already encoded a
  Python-script hook name and every-N semantics. Reconciling them (Go subcommand
  substrings; backstop meaning) was as much of the work as the new code — a
  reminder to grep for what the tree already believes before building.

## Open follow-ups (not blocking)

- Suffix stemming for recall is deferred behind an eval (false-positive risk).
- Stale session-state files under the temp dir accumulate; a periodic prune could
  be added if it ever matters.
- `Load` guards the `rules.json` leaf but not a symlinked `.abcd` directory
  component (same trust level as writing the file; noted by the security review).
