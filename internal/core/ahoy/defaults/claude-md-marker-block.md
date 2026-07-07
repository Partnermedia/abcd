<!--
  Managed by abcd (Agent-Based Configuration for Development).
  Do NOT hand-edit content inside the abcd-managed fences — `/abcd:ahoy`
  silently overwrites this block on drift (per itd-3). Per-repo rule
  customisation goes in <repo>/.abcd/rules.json instead.
-->

## abcd rule loader

This repository uses the abcd modular rules loader. A `UserPromptSubmit` hook
inspects each prompt, recall-matches it against keyword triggers declared in
`<repo>/.abcd/rules.json` (and the plugin-bundled defaults), and injects the
matched domain rules into context — instead of force-loading the full ruleset
on every turn.

- Active rules: `abcd rules` (or `abcd rules <DOMAIN>` to scope).
- Per-repo overrides: edit `<repo>/.abcd/rules.json` (validated against
  `scripts/abcd/schemas/rules.schema.json`).
- Kill switch: set `"disabled": true` at the top of `.abcd/rules.json`.
- Explicit activation: start a prompt with `*<DOMAIN>` (e.g. `*COMMITTING`,
  `*PII`) to inject that domain unconditionally — overrides `dormant`.

### Default domains

`COMMITTING`, `DOCUMENTATION`, `ROADMAP`, `ISSUES`, `INTENTS`, `LIFEBOAT`,
`PII`. Each ships with a placeholder rule; the legacy harvest from
`~/ABCDevelopment/.claude/CLAUDE.md` is a follow-up phase (per itd-3).

### Reset triggers

`SessionStart` and `PreCompact` clear the dedup state so a fresh session sees
the rules again. Compaction-aware: the hook does not double-inject within a
single session, and resets cleanly across compactions.

For internals see `.abcd/development/brief/05-internals/03-configuration.md`
and the itd-3 / fn-14 spec.
