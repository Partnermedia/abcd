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

Two files still carry the blanket `.abcd/` exclusion claim that iss-43 corrected in the README. Recorded rather than fixed there, because the disposition scoped iss-43 to one README line.

1. **`.abcd/work/CONTEXT.md:42`** — "Single repo, curated release (no dev→public mirror). `.abcd/**` never ships." The blanket claim holds only for the released binaries. The plugin channel installs from `.claude-plugin/marketplace.json`, whose plugin `source` is `"./"`, so a marketplace install carries the whole repository; and GitHub attaches auto-generated source archives to every release, which carry it too. This one matters more than its severity suggests: CONTEXT.md is the first file a session reads, and a sharp-edges bullet is trusted before anything else is.
2. **`AGENTS.md:112-113`** — "**Single repo, curated release.** `.abcd/**` stays in-tree but is excluded from the release artefact by packaging; the repo is the plugin marketplace." No such packaging exists: neither `.claude-plugin/plugin.json` nor `marketplace.json` carries an exclusion key, `.gitattributes` declares no `export-ignore`, and `release.yml` performs no curation — it uploads the four binaries and `checksums.txt`. "By packaging" names a mechanism, so it reads as a description of how the exclusion is achieved rather than of what the release happens to contain.

Fix direction: the same channel-truthful wording the README now carries — what every repository checkout holds (marketplace installs and release source archives included) against what the released binaries hold — with no appeal to a packaging step that does not exist. Building actual curation is feature work and is not what this issue asks for; if it is ever built, this issue's wording changes with it.
