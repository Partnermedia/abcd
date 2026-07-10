# Retire the name

**The rule.** The change that retires a path, identifier, term, or flag adds
its old name to `record-lint`'s `banned_tokens` (with an `allow_context` for
genuinely historical passages). Renames and bans travel together; a retirement
without a ban is half a retirement.

**Why.** Stale names are the drift that discipline demonstrably cannot hold:
the 2026-07-08 review found roughly fifty references to retired locations and
identifiers across the record — thirty-eight to `development/activity` (historical) in the
brief alone — and that drift survived a dedicated same-day consistency pass.
The banlist engine already exists; each ban is configuration, not code, and
once added it gates every future edit through the already-armed preflight and
CI hooks.

**Bounds.**

- The ban targets the *retired* spelling, not the concept — prose discussing
  the rename itself earns an `allow_context`, not an exemption from the rule.
- Applies to conceptual identifiers (lint names, flag names, workflow
  filenames) as much as to paths.

**Promotion.** This principle is born adjacent to its gate: the moment adding
a ban becomes a checked part of the rename workflow (rather than a
remembered one), it is a discipline.
