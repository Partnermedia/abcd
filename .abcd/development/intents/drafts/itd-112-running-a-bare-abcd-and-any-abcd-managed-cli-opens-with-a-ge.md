---
id: itd-112
slug: running-a-bare-abcd-and-any-abcd-managed-cli-opens-with-a-ge
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
---

# Running a bare abcd (and any abcd-managed CLI) opens with a generated banner in the style of ferry's: a small colour logo — an obvious object or a colourful text-logo — plus name, tagline, and version, rendered from the repo's canonical identity block (itd-102) rather than hand-drawn, with a boilerplate generator so every abcd-managed CLI repo gets one; honours --no-color and quiet modes

## Press Release

> _Seeded from a quoted-text intent capture. Expand into the full press-release narrative before planning._

## Why This Matters

Running a bare abcd (and any abcd-managed CLI) opens with a generated banner in the style of ferry's: a small colour logo — an obvious object or a colourful text-logo — plus name, tagline, and version, rendered from the repo's canonical identity block (itd-102) rather than hand-drawn, with a boilerplate generator so every abcd-managed CLI repo gets one; honours --no-color and quiet modes

## Acceptance Criteria

> _Required (the itd-1 discipline): add at least one Given-When-Then bullet describing the verifiable bar for "shipped" before this draft can be planned._

## Reference Designs (observed 2026-08-15, maintainer's terminal)

Two live exemplars set the bar; describe, don't copy:

- **Text-logo style (`ragd`)**: a block-ASCII wordmark with a per-letter
  colour gradient (magenta → purple → green → blue), followed by a plain
  monochrome pitch paragraph and standard Cobra help. Effect: identity in the
  wordmark alone; everything below stays quiet.
- **Object style (`ferry`)**: a small pictorial ASCII object (~6 lines, ~18
  columns — a ship with coloured hull and cargo) in the left column; right
  column carries `name vX.Y.Z` in a warm accent colour, a prompt-glyph plus
  two-line tagline in muted grey, then two closing hint lines with the
  runnable commands (`ferry --help`, `ferry init`) colour-highlighted inline.
  Effect: object as brand, tagline as orientation, hints as the next action.
- Common properties to preserve: whole banner fits ~10 rows × ~50 columns,
  reads correctly with colour stripped (`--no-color` and non-TTY degrade to
  plain glyphs, never blank), and the bare invocation stays useful — the
  banner tops the existing status/help output, it does not replace it.

## Open Questions

- Object vs text-logo per repo: maintainer's pick at generation time, stored
  in the identity block? (An "obvious object" needs a human choice; a
  text-logo is always derivable.)
- Where the generator lives: `abcd identity` render surface vs the
  prepare-this-repo path for non-Go CLIs.
- Colour palette: fixed house palette vs per-repo accent declared alongside
  the identity block.

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
