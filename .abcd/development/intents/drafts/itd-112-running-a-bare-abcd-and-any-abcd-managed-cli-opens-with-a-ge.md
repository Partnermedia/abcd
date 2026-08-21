---
id: itd-112
slug: running-a-bare-abcd-and-any-abcd-managed-cli-opens-with-a-ge
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: [itd-133]
related_intents: [itd-102, itd-134]
severity: minor
---

# Running a bare abcd in an interactive terminal opens with a generated banner: the flag-hoist logo in half-block pixels beside name, tagline, and version from the canonical identity block — full colour ladder, shade-block mono, never on non-TTY surfaces

Typed links: `builds_on` itd-133 — the livery grids and palette are the
banner's art source; `refines` itd-102 — the text column renders from the
canonical identity block; `refined-by` itd-134 — the managed-repo banner
generator, split out at the 2026-08-21 grill, carries the "every
abcd-managed CLI" half and the per-repo accent question.

## Press Release

> _Seeded from a quoted-text intent capture. Expand into the full
> press-release narrative before planning._

## Why This Matters

A bare `abcd` today opens straight into the status board — correct, but
faceless. The identity block (itd-102) keeps the words canonical and the
livery package (itd-133) holds the marks; nothing yet composes them into
the first thing a person actually sees. The banner is that composition:
identity rendered, never hand-drawn, topping the existing output without
replacing it.

## Reference Designs (observed 2026-08-15, maintainer's terminal)

Two live exemplars from the maintainer's private toolchain set the bar
(described generically; private tool names stay out of the record):

- **Wordmark style**: a block-ASCII wordmark with a per-letter colour
  gradient, followed by a plain monochrome pitch paragraph and standard
  CLI help. Identity lives in the wordmark alone; everything below stays
  quiet.
- **Object style**: a small pictorial colour object (~6 lines, ~18 columns)
  in the left column; the right column carries `name vX.Y.Z` in a warm
  accent, a two-line tagline in muted grey, then closing hint lines with
  runnable commands highlighted inline. Object as brand, tagline as
  orientation, hints as the next action.
- Common properties to preserve: the whole banner fits ~10 rows × ~50
  columns, reads correctly with colour stripped (never blank), and the bare
  invocation stays useful — the banner tops the existing status output, it
  does not replace it.

## Grill outcomes (2026-08-21, maintainer-ruled)

- **Scope split:** itd-112 ships abcd's own banner and the terminal colour
  stack only. The managed-repo generator is itd-134.
- **Trigger:** bare invocation on an interactive TTY only. Non-TTY, CI,
  hooks, `--json`, quiet modes, and every subcommand stay banner-free — the
  plugin/model surface never receives banner bytes.
- **Colour ladder:** full auto-detection — 256-colour where detected,
  16-colour fallback, mono floor; honours the NO_COLOR spec and a
  `--no-color` flag.
- **Layout:** the true-geometry flag strip rendered in half-block pixels
  (two pixel rows per text row: 23 columns × 3 rows), right column with
  `abcd vX.Y.Z`, the identity-block tagline, and next-action hint lines
  with runnable commands highlighted.
- **Palette:** the fixed livery house palette. Per-repo accents belong to
  itd-134.
- **Mono form:** the art survives colour-stripping as shade-block glyphs
  (the maintainer chose art-always over text-only degradation); "never
  blank" is met by glyphs, not by dropping the mark.
- **Object vs text-logo** (formerly an open question here): foreclosed by
  the itd-133 decision — the object is the flag hoist.

## Open Questions

- The termsafe carve-out: the codebase's standing rule is that ESC never
  reaches a render path. A banner is sanctioned ANSI emission of *trusted,
  static* art — that boundary (trusted-static allowed, untrusted text still
  sanitised) needs an ADR at planning, not an exception buried in code.
- Windows terminals: the colour ladder's detection is unverifiable on CI
  (macOS + Linux only) — the acceptance criteria must scope what "any
  terminal" provably means.
- Unicode floor: half-blocks and shade blocks assume UTF-8; whether a
  pure-ASCII environment gets the wordmark-only form needs a ruling at
  planning.

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
