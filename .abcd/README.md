# .abcd/

abcd's development namespace. Everything here is dev material — it stays in the
repo (transparent) and is present in every repository checkout, marketplace
installs and release source archives included, but never in the released
binaries. `docs/` (not here) holds user-facing documentation and is the only
dev-adjacent tree the release artifact carries.

Three tiers, on two axes (durability × sharing):

- **`development/`** — durable record (committed): brief, roadmap, intents,
  decisions/adrs, research. The specification for the build.
- **`work/`** — shared working (committed): `CONTEXT.md`, `DECISIONS.md`.
- **`.work.local/`** — local ephemeral (gitignored): `NEXT.md`, `scratch/`,
  `logs/`.

See [`../AGENTS.md`](../AGENTS.md) for the full layout and boundaries.
