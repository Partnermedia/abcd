---
id: itd-137
slug: maya-sees-how-every-decision-connects-one-relationship-chart
spec_id: spc-39
kind: standalone
suggested_kind: null
reclassification_history: []
builds_on: [itd-135, itd-136]
severity: minor
impact: additive
---

# Maya sees how every decision connects — one relationship chart with a date coil and a links-only layout, and a five-lane genealogy of the record over time

## Press Release

> **Maya sees how every decision connects — in one chart, not a hairball.**
> Maya studies how software projects actually evolve, and a typed record
> graph of six hundred entries is exactly the artefact they want to read —
> if only anything rendered it. Now `/record/graph/` draws every record as a
> circle — never a shape code — in one chart with two build-time
> arrangements: a date coil that winds outward from the first record with
> concentric month zones like fare zones on a tube map, and a links-only
> force layout where typed cross-references pull connected work into islands.
> Arrowheads appear only where the relation is directed — `builds_on`,
> `supersedes`, `implements` — and mirrored references are collapsed so each
> distinct link draws once. Tapping a record rings it, fades everything
> unlinked, and opens a card: type and state as pills in GitHub's palette,
> every date of the record on one continuum, a pull-out of linked records
> phrased from the focused record, and back and forward buttons that walk the
> viewing history the way a browser does. Controls keep to the corners; the
> middle belongs to the chart; a "Stand by…" overlay covers the settling
> pass; and "Browse as a list" beneath the stage is the keyboard path and the
> accessible twin. `/record/timeline/` is the genealogy: five lanes —
> releases, decisions, intents, specs, issues — over one axis, crowded days
> fanned into capsules with counts, supersession arcs drawn only where both
> ends exist and dashed stubs ending in × where a target has left the tree.
> On a phone the card becomes a bottom sheet and the chart pans to keep the
> focused record visible — perfect at 390 px, or the element is removed from
> the phone view. "Every arrangement I could dismiss as decoration turned
> out to be a rule," said Maya. "The coil is the capture order; the islands
> are the typed links; the stubs are the retirements. The chart taught me
> the record's own conventions."

## Why This Matters

A typed graph nobody can see might as well be untyped. The two arrangements
are the two honest questions — when did work happen, and what does it build
on — and computing both at build time (deterministic, seeded, shipped in
`record.json`) means every visitor reads the same picture, the page runs no
simulation, and the layout cost is paid once per release rather than per
view. Drawing dashed stubs for retired targets rather than inventing
positions keeps the chart inside the record's own truth: the tree does not
carry those files, and the chart says so.

## Acceptance Criteria

- Given `/record/graph/`, then exactly two arrangements are offered by an
  explicit two-state control — the date coil with month zones and the
  links-only force layout — both precomputed at build time into
  `record.json`, deterministic across rebuilds of the same tree. The coil
  packer and the seeded spring embedding are written in-repo against
  `compose.py`/`build_data.py` as the reference implementation; any
  external layout library is a new dependency under the sign-off gate and
  is not assumed.
- Given the stored typed references, then mirrored pairs (intent `spec_id` ↔
  spec `implements`; `related` recorded in both files) collapse to one drawn
  link each, with arrowheads only on `builds_on`, `supersedes` and
  `implements`; body mentions stay off by default.
- Given a focused record, then the card shows the title, type and state
  pills in the GitHub palette, every date of the record on one continuum
  (frontmatter date, in-tree since, state since, last touched — from git),
  a linked-records pull-out phrased from the focused record, and the GitHub
  link; the last-touched date links the record's commit history on GitHub,
  so an amendment is traceable from the card, not just visible as a date;
  back and forward (and Alt+arrows) walk the viewing history.
- Given a keyboard-only visitor, then the list twin under the stage reaches
  every record and every link the chart shows.
- Given `/record/timeline/`, then five lanes render as one static SVG at
  build time; intents, specs and issues take their dates from one
  `git log --reverse --name-status` pass; a day too crowded to show singly
  becomes a capsule with its count; supersession arcs draw only where both
  ends exist, and a target absent from the tree gets a dashed stub ending in
  ×, never an arc to an invented position.
- Given `prefers-reduced-motion`, then all animation is removed, nothing
  drifts at rest, and the list view is offered in place of the chart.
- Given a 390 px viewport, then the card is a full-width bottom sheet, the
  chart pans so the focused record stays above it, and no route scrolls
  horizontally — verified by the static checks in `abcd site check` plus
  the screenshot audit in CI.

## Open Questions

- Whether the retired-ADR stubs graduate to tombstone files (shared with
  itd-136's baseline question).

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._
