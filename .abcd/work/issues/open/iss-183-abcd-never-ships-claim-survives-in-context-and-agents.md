---
schema_version: 1
id: "iss-183"
slug: "abcd-never-ships-claim-survives-in-context-and-agents"
severity: "minor"
category: "documentation"
source: "impl-review"
found_during: "2026-08-05 iss-43 review"
found_at: ".abcd/work/CONTEXT.md"
---

The blanket `.abcd/` exclusion claim that iss-43 corrected in the README survives across the record. Recorded rather than fixed there, because the disposition scoped iss-43 to one README line.

What is true: the exclusion is **implemented but unwired**. `internal/core/launch/bundle.go:29-31` declares `DenyNamespaces` — "first-path-segment names that never ship", `.abcd` among them — as a structural deny no allowlist overrides, and `bundle_test.go` pins it (`TestAbcdNamespaceStructurallyExcluded`, `TestDefaultDenyNewTopLevel`). It is reachable: `abcd launch ship` resolves a bundle through it. But `Ship` stops at `WouldPublish` (`internal/core/launch/ship.go:36-41`, "it stops HERE and returns WouldPublish=true with NO network call"), and `.github/workflows/release.yml` never invokes the verb — it builds the four binaries and uploads them with `checksums.txt` directly. So no release the project has cut has passed through the filter, and two channels carry `.abcd/` regardless: a marketplace install takes the repository root (`.claude-plugin/marketplace.json`, plugin `source: "./"`), and GitHub attaches an auto-generated source archive to every release, `.gitattributes` declaring no `export-ignore`.

The defect is therefore narrower than "the claim is false" and sharper than "the docs are stale": prose written in the present tense describes a mechanism that exists in code but has never run.

Instances — descriptive and orientation documents asserting the claim as present-tense operating fact. Line refs rot; the quotes anchor.

1. `.abcd/work/CONTEXT.md:42` — "Single repo, curated release (no dev→public mirror). `.abcd/**` never ships." Sharpest of the set: CONTEXT.md is the first file a session reads, and a sharp-edges bullet is trusted before anything else is.
2. `AGENTS.md:112-113` — "**Single repo, curated release.** `.abcd/**` stays in-tree but is excluded from the release artifact by packaging; the repo is the plugin marketplace."
3. `.abcd/README.md:4` — "repo (transparent) but is excluded from the release artifact."
4. `.abcd/development/README.md:4` — "repo (transparent) but excluded from the release artifact".
5. `.abcd/development/brief/02-constraints/01-platform.md:9` — "**`.abcd/**` stays in-tree but is excluded from the release artifact by packaging** — exclusion is a build-time filter over the one tree, not a copy between two repos."
6. `.abcd/development/brief/05-internals/03-configuration.md:209` — the record row's middle cell: "committed — excluded from the release artefact by packaging".
7. `.abcd/development/roadmap/phases/phase-1-ahoy.md:11` — "`.abcd/**` excluded by packaging so the design record never ships in the" (the sentence continues on the following line).

Exempt, deliberately left (iss-42 and iss-44 precedent — a decision record and a dated snapshot are supposed to say what was decided then, and rewriting them to match today falsifies the record): adr-0028, which ratifies the wording; the dated plans under `.abcd/development/plans/`; and the bodies of intents under `planned/` and `shipped/`, which specify the mechanism they were written to deliver.

This list is not exhaustive and must not be trusted as one. The implementing round re-enumerates first, over the whole tree, with `never ships`, `excluded from the release`, and `by packaging` — the survey behind this issue already surfaces further candidates in `brief/01-product/`, `brief/04-surfaces/04-launch.md`, `brief/glossary/distribution/release.md` and `principles/script-first-mvp.md`, unclassified here — and sorts each hit into the descriptive set or the exempt set before touching anything.

Fix direction, one of two, chosen once and applied consistently: wire the existing publish path so the filter runs on a real release, which makes the prose true as written; or reword the descriptive instances to the current gap, on the same channel-truthful pattern the README now carries — what every repository checkout holds against what the released binaries hold — describing the launch bundler as the implemented mechanism it is rather than an operating one. Wiring is feature work with its own design questions and is not assumed here.
