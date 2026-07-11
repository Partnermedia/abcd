# itd-3 modular rules loader — SOTA-verified design

**Status:** design recorded 2026-07-11, **awaiting maintainer sign-off** before
any implementation. Reached via the [prefer-sota](../principles/prefer-sota.md)
process: an adversary challenged generic just-in-time rule-injection SOTA (CARL,
AGENTS.md, Cursor rules, Skills) for THIS repo's fit, then a `sota-researcher`
verdict was taken on the fit-surviving hypothesis. This doc is the design of
record for [itd-3](../intents/planned/itd-3-modular-rules-loader.md); it does not
authorise a build. Adopting a planned intent is a maintainer decision — the four
deltas from the intent (§7) are the sign-off surface.

## The decision

Ship a **modular rules loader**: default rule domains bundled in the Go binary +
a per-repo `.abcd/rules.json` override, recall-matched against each prompt and
injected just-in-time so the plugin's discipline never force-loads a monolithic
CLAUDE.md. Rule-loading is a **transport-agnostic core capability** with two
front doors — the vendor-neutral `abcd rules [domain]` CLI verb, and a Claude
Code hook (`abcd hook prompt-router`, wired via `hooks/hooks.json` on the
`UserPromptSubmit` event) that shells to the same core and injects its rendered
result. The last piece is a new `OPINIONS`/`CONVENTIONS` default domain whose
rules **point at** `.abcd/development/principles/` rather than copying them.

## Why this form (adversary → SOTA verdict)

### What the adversary rejected, each for a named repo preference

- **CARL's "the hook owns the logic" architecture** — collides with the
  transport-agnostic core (AGENTS.md boundary). The hook must be a dumb shim over
  `internal/core`; the capability lives in the core and renders through the
  vendor-neutral `abcd rules` verb.
- **CARL's global `~/.carl/carl.json` source** — collides with per-repo-only
  ([the intent's "What's Out of Scope"](../intents/planned/itd-3-modular-rules-loader.md);
  `~/.abcd/` holds only machine-wide state). No `~/.abcd/rules.json`. Ever.
- **CARL's MCP server + `add_rule`/`toggle_domain` verbs, and the Python hook** —
  each is a new runtime/dependency (ask-first). Rule *mutation* is a file-edit of
  `.abcd/rules.json` via standard tools; the hook is a Go subcommand; recall is
  Go-stdlib keyword matching. Net new dependencies: **zero**.
- **Bundling principle *text* in the binary for the OPINIONS domain** — collides
  with [one-canonical-primitive](../principles/one-canonical-primitive.md). The
  domain carries short directives + a pointer to the canonical principle file,
  never the principle body.
- **CARL's MCP verb-set as the CLI shape** — collides with the
  bare-command-as-render discipline (`02-constraints/04-naming.md`). Bare
  `abcd rules` renders; `abcd rules <DOMAIN>` is the sanctioned positional scope;
  no `show` sub-verb.
- **Best-effort silent injection** — collides with fail-closed input parsing and
  loud-staging. The hook reuses the in-repo `verifyHookManifest` fail-closed
  template (`internal/core/ahoy/store.go`): size caps, symlink refusal, schema
  validation, inject-nothing-and-surface-loudly on any parse failure.

### The surviving-fit hypothesis (the shape that survived every collision)

**A core capability with the Claude Code hook as one thin front door — not an
adapter seam.** `internal/core/rules` owns `merge(bundled defaults, repo
rules.json) → schema-validate → recall-match(prompt) → dedup → render`. That core
is transport-free and exercised directly by `abcd rules`, testable with zero
harness. The Claude Code coupling is quarantined to two shim entrypoints
(`prompt-router`, `prompt-router-reset`) plus `hooks/hooks.json`. This is the
**same core-plus-front-doors reconciliation the repo already uses for CLI-and-
later-MCP** — rule-loading is not a swappable backend behind an interface (adr-22
seams), it is one core verb surfaced through whichever front door is present. The
host-agnosticism fork resolves as: the *capability* stays host-agnostic in the
core and renders through `abcd rules`; only the *injection transport* is
Claude-Code-specific, exactly as only the *CLI transport* is Cobra-specific.

### The SOTA verdict (fit-aware)

The `sota-researcher` verdict on the fit-surviving hypothesis: it **is** current
SOTA — *just-in-time context loading via progressive disclosure* (Anthropic's
Agent Skills three-stage model and "effective context engineering" guidance are
the primary-source statements of exactly this shape: a small always-on index,
the bulk loaded only on relevance). The capability-in-core / hook-as-front-door
split matches the field's convergence (AGENTS.md as the cross-vendor instruction
standard under the Linux Foundation's Agentic AI Foundation alongside MCP).

**One load-bearing refinement the verdict adds over CARL:** Skills/MCP relevance
selection is *model-driven* (the agent elects to load). For **behavioural rules**
that is the wrong instrument — the evidence is that models silently ignore rules
they were trusted to load. **Deterministic keyword pre-filtering at the prompt
boundary is the better-fit variant for rules specifically, and is not superseded
by Skills.** Keyword-recall is the right primitive here. (Additive future note:
domains could later be *exported* as Skills to ride the cross-vendor standard —
not a redesign.)

## The design

Seven parts. Parts 1–6 are the loader; part 7 is the OPINIONS domain.

### 1. `internal/core/rules` — the transport-agnostic core

One package per capability (adr-23), returning structured results, no stdout, no
harness knowledge. Public surface (indicative):

- `Defaults() RuleSet` — the binary-embedded default domains (Go `embed`).
- `Load(repoRoot) (RuleSet, []Finding, error)` — merge defaults with
  `<repoRoot>/.abcd/rules.json`, schema-validate, return findings on drift.
- `Match(rs RuleSet, prompt string) []Domain` — deterministic recall.
- `Render(domains []Domain) string` — the injected text (and the `abcd rules`
  output). One renderer, both front doors.
- Dedup lives in the hook front door (session state), not the pure core.

### 2. `rules.json` schema (embedded, JSON Schema 2020-12)

**Keep the shipped `stepRules` shape** (`apply.go`), reject the legacy-harvest
`extends`/`overrides` sketch (it re-imports CARL's mergeable-global complexity
the per-repo decision rejected). Custom domains are additional keys under
`domains`, not a separate `custom_domains` block:

```json
{
  "schema_version": 1,
  "disabled": false,
  "domains": {
    "COMMITTING": {
      "state": "active",
      "recall": ["commit", "pr", "git add", "push"],
      "aliases": ["pull request", "committing"],
      "rules": ["..."]
    }
  }
}
```

- **`disabled`** (top-level) — repo-wide kill switch. A `*<DOMAIN>` star-command
  must **not** bypass it (a kill switch a star-command can defeat is not a kill
  switch).
- **`state`** per domain — `active` | `dormant`. `dormant` silences one domain;
  `*<DOMAIN>` overrides `dormant` but not top-level `disabled`.
- **`recall`** + **`aliases`** — the author-curated trigger set (see part 4).
- Domain keys constrained to `[A-Z][A-Z0-9_]*`; rejected before any use.

### 3. Recall matching (part 4 detail) — deterministic, zero-dep

Tokenize the prompt (split on non-alphanumeric → `strings.ToLower` → set), then
boolean membership against each active domain's `recall` + `aliases`. Case-fold
always. **Author-declared aliases are the highest-value dep-free lever** — recall
quality moves to the person who knows the domain's vocabulary. No TF/BM25/
embeddings (this is a curated boolean trigger, not corpus retrieval — ranking is
meaningless and adds false-positive bloat). Suffix stemming is **deferred behind
an eval**, not shipped on spec (`test`→`testing`→`attestation` false positives).

### 4. Injection + refresh discipline (the hook front door)

- On **`UserPromptSubmit`**: recall-match, dedup, inject matched domains.
- On **`SessionStart`** (`source` = `startup|resume|compact`): re-inject header +
  currently-live domains. **The `compact` source is the surgical post-compaction
  refresh** — strictly better than a fixed prompt counter, because compaction
  (not prompt count) is the real threat to rule persistence.
- **`PreCompact`**: record what was live; do **not** re-inject there (it fires
  before the summary erases the text). Re-injection happens on the following
  `SessionStart(compact)`.
- **Fixed-N re-inject is demoted to a large backstop (N≈15–20)** for
  always-relevant domains that never recall-match — not the primary mechanism.
  (This is delta D1 in §7.)

### 5. Dedup signature

**Per-domain FNV-1a (or sha256) content hash of the *rendered* block** — Go
stdlib, zero deps. Hash rendered text, not rule-id sets (id-only misses a
mid-session `rules.json` edit). Per-domain (not one whole-payload hash) so one
changed domain re-injects without re-sending the rest.

**Critical B1×B2 coupling:** the "already-injected" ledger **must be cleared on
`SessionStart(compact|clear)`** — after compaction the injected text is gone from
context, so a surviving signature would wrongly suppress the re-injection you
need most. Reset ledger → re-inject → repopulate. This is the single most
important correctness point in the design.

### 6. `abcd rules [domain]` CLI verb + `abcd hook` entrypoints

- Bare `abcd rules` → render the active rule set (read-only, never mutates).
- `abcd rules <DOMAIN>` → scope to one domain (positional, no `show` sub-verb).
- `abcd hook prompt-router` / `abcd hook prompt-router-reset` — the Claude Code
  hook entrypoints (read hook JSON on stdin, emit `additionalContext`,
  fail-closed). **Operator-internal transport, not a user command surface.**

### 7. The OPINIONS / CONVENTIONS default domain

The payoff of the `opinions.md` decision (rejected as a file; the win is a
default domain here). Recall keywords on convention/opinion/principle/"how do we"
triggers; each rule is a **short directive + a pointer** to the canonical file
under `.abcd/development/principles/` — never a copy (one-canonical-primitive).
Pointer resolution uses a **fixed allowlist of principle paths**, never an
id-derived filesystem path (path-traversal defence). Wired so managed repos
inherit it via `ahoy` (it ships in the binary-embedded defaults).

## Forks resolved (the intent's Open Questions)

| Fork | Resolution | Basis |
|---|---|---|
| N-refresh value | Event-driven primary (`SessionStart(compact)`); N≈15–20 backstop only | SOTA B1 — **delta D1** |
| `.abcdignore` | **Reject for v1** — a second override surface competing with `rules.json`; `dormant`/no-match already cover it | fit |
| `rules.json` shape | Keep shipped `{schema_version,disabled,domains{}}`; reject `extends/overrides` | fit + shipped code |
| Star-command parse | `(?:^|\s)\*([A-Z][A-Z0-9_]*)(?=$|\s)` — uppercase+boundary avoids `* bullet` / `*.py` collisions | fit |
| Dedup signature | Per-domain rendered-content hash; ledger cleared on reset | SOTA B2 |
| `state` vs `disabled` | Both kept — different scopes; `*DOMAIN` overrides `dormant`, not `disabled` | fit |
| No-match budget | **Zero model-facing tokens**; observability out-of-band | SOTA B4 — **delta D3** |

## Trust boundary (security posture, reused from `verifyHookManifest`)

The hook parses untrusted prompt text + injects context. Fail-closed:

- **Prompt-injection into the injected message** — the hook emits abcd's *own*
  rule text only; the prompt is matched, **never reflected** into the system
  message.
- **Path traversal** via a custom domain/rule id — ids constrained to
  `[A-Z][A-Z0-9_]*`; OPINIONS pointers resolve against a fixed allowlist.
- **Resource blowup** on a huge prompt — cap the scanned prompt length (mirror
  the 256KB `hooks.json` cap + irregular-file refusal in `store.go`).
- **Malicious `rules.json`** — schema-validate before use, size-cap, refuse a
  symlinked leaf, fail closed (inject nothing + surface loudly) on any failure.

The `security-reviewer` agent runs before the hook + injection logic are
presented (Phase 2); a BLOCK verdict stops the change.

## Gate + surface interactions

- **`surface_coverage` record-lint rule** — `abcd rules` needs a registry entry
  in `04-surfaces/README.md` (+ a `commands/abcd/` markdown if surfaced under
  `/abcd:`). `abcd hook *` is a transport entrypoint, not a user surface —
  confirm the rule treats it as operator-internal (like other non-surface verbs).
- **Hook-manifest verify (`store.go`)** — `requiredHookCommand` currently expects
  the substrings `prompt_router_hook` / `prompt_router_reset`. With the Go
  subcommand spelled `abcd hook prompt-router`, Phase 4 must reconcile: update
  the verify substrings to `hook prompt-router` / `hook prompt-router-reset` (and
  `detect_test.go`). This is a **behaviour-preserving reconciliation of a stale
  manifest expectation**, kept in its own commit.
- **Marker block drift** — `defaults/claude-md-marker-block.md` describes the
  loader in present tense as if it works. It stays factually wrong until the
  loader exists; Phase 4 rewrites it to be present-tense-accurate. Until then it
  is known drift, not a new claim.

## Implementation phasing (only after sign-off — TDD, mirrors task 5)

Each phase: a test watched fail then pass; `make preflight` + `make record-lint`
green before each commit; `Assisted-by: Claude:claude-opus-4-8` trailer; work on
an `auto/*` branch; no push/merge/PR.

1. **Core + schema + merge** — `internal/core/rules`: embedded defaults,
   `rules.json` schema, `Load`/`Match`/`Render`, dedup helper. Pure, no harness.
2. **Hook front door** — `abcd hook prompt-router` / `-reset`: stdin hook-JSON
   parse, recall→dedup→inject, `SessionStart(compact)` refresh, star-command
   bypass, fail-closed posture. `security-reviewer` before presenting.
3. **`abcd rules [domain]` verb** — bare-render + positional scope, wired into
   `internal/surface/cli`, `--json` parity.
4. **ahoy wiring** — `stepRules` writes the real default `rules.json` (not the
   empty `domains:{}` skeleton); `hooks/hooks.json` manifest; reconcile
   `requiredHookCommand`; rewrite the marker block present-tense-accurate.
5. **OPINIONS domain** — recall keywords + pointer-rules into
   `.abcd/development/principles/`; inherited by managed repos via ahoy.

## STOP — maintainer sign-off required

Per the design gate, implementation does not start until the maintainer signs off
**these four deltas from the intent** (each changes an intent commitment, so it is
a fidelity decision, not a design detail):

- **D1 — N-refresh.** Intent + `config.json` say re-inject every N=5. Design
  demotes fixed-N to a ~15–20 backstop and makes `SessionStart(compact)` the
  primary refresh. *Config key `rules.force_refresh_every_n` stays but changes
  meaning (backstop, not primary) and default.*
- **D2 — schema shape.** Design locks the shipped
  `{schema_version,disabled,domains{}}` shape and **rejects** the legacy-harvest
  `extends`/`overrides`/`custom_domains` sketch. Confirm the harvest sketch is
  superseded.
- **D3 — no-match budget.** Intent's acceptance criterion caps no-match overhead
  at "<200 tokens (header only)". Design reframes it to **0 model-facing tokens on
  no-match + a diagnostic log line every run**. This edits an acceptance
  criterion — maintainer's call.
- **D4 — `.abcdignore` rejected for v1.** Intent lists it as an open question;
  design declines it. Confirm.

On sign-off, this doc becomes the build contract and Phase 1 begins.
