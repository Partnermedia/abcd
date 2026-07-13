# `/abcd:audit` — Check Repo Conformance

`/abcd:audit` reports whether a repository follows the working conventions. It is
**strictly read-only** — it performs zero writes, and remediation stays with
`/abcd:prepare-this-repo` and the maintainer.

It answers a different question from `/abcd:ahoy`: `ahoy` reports whether the
*tool* is installed and configured for a repo (environment/setup health);
`audit` reports whether the *repo* conforms to the conventions. Two questions,
two verbs.

## Behaviour

```bash
abcd audit --json
```

emits `{ "findings": [ … ], "skipped": [ … ] }`. Each finding carries a stable
`ruleId`, a `severity` (`error` or `warn`), a `file` and `line`, a `message`, and
a `fix`. `skipped` names rules whose enablement condition was not met (e.g.
`docs-currency` where there is no `docs/`), so a not-applicable rule reads as
skipped, not failed. Without `--json`, `abcd audit` prints a grouped,
doctor-style human report (severity glyph, rule id, `file:line`, message, indented
fix) and a summary tail.

The exit code is Conftest's tri-state: `0` clean, `1` warnings only, `2` any
error — so `abcd audit` gates a repo's CI as well as backing onboarding.

## The v1 rule set

| id | severity | checks |
|---|---|---|
| `three-tier-layout` | error | `.abcd/development/` and committed `.abcd/work/` present; `.abcd/.work.local/`, when present, gitignored |
| `conventions-router` | error | `AGENTS.md` present at the repo root |
| `decision-durability` | warn | a committed `.abcd/work/DECISIONS.md`; decisions not living only in the gitignored layer |
| `docs-currency` | warn | reuses the docs-lint engine where `docs/` exists |
| `privacy-hygiene` | error | no absolute local paths in committed files, honouring an `abcd-audit:allow` line waiver |

## How it is built

The engine reuses `internal/core/lint`'s severity/fix/waiver vocabulary and adds
path-presence and gitignore primitives. Rules are declarative data behind a
rule-loader seam, and output is serialised behind a seam that makes a later SARIF
export additive. No new dependency.

## References

- Plugin command: [`commands/abcd/audit.md`](../../../../commands/abcd/audit.md)
- Design record: [`plans/2026-07-13-abcd-audit-verb.md`](../../plans/2026-07-13-abcd-audit-verb.md)
- Intent: [`itd-85`](../../intents/drafts/itd-85-audit-verb.md)
- Onboarding consumer: [`15-prepare-this-repo.md`](15-prepare-this-repo.md)
