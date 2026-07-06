---
id: adr-19
slug: plugin-json-version-carve-out
status: accepted
date: 2026-07-01
supersedes: null
superseded_by: null
related_intents: [itd-67]
related_rfcs: []
related_adrs: [adr-5, adr-18]
---

# ADR-19: The plugin version lives ONLY in the published snapshot; dev files stay unversioned, and the version location is chosen by a schema-validated decision, not hard-coded

## Context

fn-77 makes the public `abcd/` repo an installable, versioned Claude Code plugin.
That raised two coupled questions the framework had to settle before any manifest
was written:

1. **Where does the version live?** Claude Code discovers a plugin by directory
   convention and reads its manifest metadata. A published plugin needs a version
   so `/plugin update abcd` can compare installed-vs-available. But
   [adr-5](0005-brief-is-current-state.md) already established that the dev
   record carries **no** version label — versions are an *output* of promotion,
   not a sequencing input on abcdDev. So the dev repo's committed
   `.claude-plugin/plugin.json` must stay unversioned while the *published*
   snapshot carries one.

2. **Is `version` even a legal field?** Whether the Claude Code plugin-manifest
   schema *accepts* a `version` property was an external unknown at planning time.
   Hard-coding `plugin.json.version` as the answer would have been a guess. fn-77.1
   resolved it against the pinned schemastore fixtures and recorded the outcome in a
   machine-readable decision artifact
   ([`.abcd/config/version-location.json`](../../../config/version-location.json)),
   with three fully-specified terminal outcomes: **ACCEPT** (`version` is an
   explicit schema property → write `plugin.json.version`), **FALLBACK** (a
   different schema-valid explicit-property location), or **BLOCKED** (no
   schema-valid explicit version location → escalate).

The recorded fn-77.1 outcome is **ACCEPT**: `version` is an explicit property of
the pinned plugin-manifest schema, so the version location is
`.claude-plugin/plugin.json` at pointer `/version`, seed `0.1.0`. The marketplace
`source` resolves from the repo root (`marketplace_source_resolution_base:
repo_root`, `marketplace_source_to_root: ./`), unblocked.

## Decision

We will carry the plugin version **only** in the published public snapshot, at the
location the fn-77.1 decision artifact selects — today `plugin.json.version` in the
public repo. The dev repo's committed manifests stay **unversioned**. The version
is written during publishing (`/abcd:launch ship`) from the render handoff
(`scripts/abcd/public_manifest.render_public_manifests`), which reads
`version-location.json` and writes the version at the recorded
`manifest_path` + `json_pointer` — never by parsing a location string, and never
from a hard-coded field name. When the decision records `blocked: true`, the render
refuses: there is no schema-valid location to write, so a versioned public manifest
cannot be produced and the escalation stands.

This is compatible with the Claude Code plugin schema: fn-77.1 validated `version`
as an explicit property of the pinned manifest schema (not merely permitted by
`additionalProperties`), so a version-carrying public manifest is schema-valid. The
decision artifact records the exact accepting schema clause in its
`version_property_clause` field —
`claude-code-plugin-manifest.schema.json#/properties/version` — so this
compatibility claim traces to the pinned schema pointer, not to prose.

## Alternatives Considered

- **Version the dev `plugin.json` too, and let publishing copy it.** Rejected: it
  contradicts adr-5 (the dev record carries no version) and would make the dev
  repo's committed state drift with every routine snapshot. The version is an
  output of promotion; it belongs to the snapshot, not the source.
- **Hard-code `plugin.json.version` as the version location.** Rejected: whether
  the schema accepts `version` was an external unknown, and a hard-code would have
  silently shipped a schema-invalid manifest in the REJECT world. The
  decision-artifact indirection lets every downstream consumer (render, public
  writes, smoke) read a validated location without re-deriving it, and keeps the
  framework honest if a future schema revision moves or removes the field.
- **Store the version in `marketplace.json` only.** This is the FALLBACK branch,
  kept live in the decision schema but not selected — `plugin.json.version` is the
  more conventional, install-visible location and it validated as an explicit
  property, so ACCEPT wins.

## Consequences

- The version is single-sourced in the published snapshot; dev files never carry a
  version, so there is nothing to keep in sync on the dev side (honours adr-5).
- Every version-writing capability (`render_public_manifests`, the Phase-5 `launch
  ship` writer, the fn-77.5 smoke) is **version-neutral**: it reads
  `version-location.json` rather than assuming a field name, and **refuses** when
  `blocked: true`. A future schema change that relocates the version is absorbed by
  re-running fn-77.1's decision, not by editing call sites.
- The public repo becomes installable+versioned; `/plugin update abcd` gains a
  version to compare against. The dev/public lifecycle is explicit: dev =
  unversioned source of truth; public = versioned published output.
- A new obligation: the fn-77.5 no-half-state lint asserts the terminology and
  docs describe the *selected* location (or the escalation text under BLOCKED),
  never a stale hard-coded `plugin.json.version` claim.

fn-77.4 copies this ADR into the public repo and links it from the public docs.
