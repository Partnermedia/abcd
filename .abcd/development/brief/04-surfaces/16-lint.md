# `/abcd:lint` — Check Repo Conformance

`/abcd:lint` reports whether a repository follows the working conventions — it applies rules about form, which adr-40's vocabulary names a lint; `/abcd:audit` is reserved for itd-16's hash-chain fidelity surface. It is
**strictly read-only** — it performs zero writes, and remediation stays with
`/abcd:prepare-this-repo` and the maintainer.

It answers a different question from `/abcd:ahoy`: `ahoy` reports whether the
*tool* is installed and configured for a repo (environment/setup health);
`lint` reports whether the *repo* conforms to the conventions. Two questions,
two verbs.

## Sub-verbs

> _Machine-checked (`surface_coverage`, spc-27): each row records the verb's
> adr-40 bucket (`lint` / `review` / `audit` / `gate`, or `—` for a
> non-assessment verb) and its existence (`shipped` / `staged`), verified
> against the committed command-tree snapshot in both directions._

`abcd lint` registers no sub-verbs. The staged `chain` and `lifeboat`
verbs belong to the **reserved** `/abcd:audit` surface (itd-16's hash-chain
fidelity checks, registered in [`02-constraints/04-naming.md`](../02-constraints/04-naming.md)),
not to the conformance lint — they left this file's table with the rename.


## Behaviour

```bash
abcd lint --json
```

emits `{ "findings": [ … ], "skipped": [ … ] }`. Each finding carries a stable
`ruleId`, a `severity` (`error` or `warn`), a `file`, a `message`, a `fix`, and a
`policyInfo` rationale; content-scanning rules (`docs-currency`, `privacy-hygiene`,
`identity-positioning`) additionally carry a `line`, while path-presence findings omit it. `skipped` names
rules whose enablement condition was not met (e.g. `docs-currency` where there is
no `docs/`), so a not-applicable rule reads as skipped, not failed. Without
`--json`, `abcd lint` prints a grouped, doctor-style human report (severity glyph,
rule id, `file:line`, message, indented fix) and a summary tail. The `--root` flag
lints a repo other than the current working directory.

The exit code is Conftest's tri-state: `0` clean, `1` warnings only, `2` any
error — so `abcd lint` gates a repo's CI as well as backing onboarding.

## The v1 rule set

| id | severity | checks |
|---|---|---|
| `three-tier-layout` | error | `.abcd/development/` and committed `.abcd/work/` present; `.abcd/.work.local/`, when present, gitignored; no local-tier artefacts (`NEXT.md`, `scratch/`, `logs/`) directly in a committed tier |
| `conventions-router` | error | `AGENTS.md` present at the repo root |
| `decision-durability` | warn | a committed `.abcd/work/DECISIONS.md`; decisions not living only in the gitignored layer |
| `docs-currency` | warn | reuses the docs-lint engine where `docs/` exists |
| `privacy-hygiene` | error (network-identifier findings mapped from a scanner `warn`/`info` land as `warn`) | no absolute local paths in committed files, and no real network identifiers on any tracked text line — the scanner's canonical merged pattern set (per-repo severities from `.abcd/config/pii.json` honoured), the fix naming reserved documentation values (RFC 5737/3849/2606/7042, or a persona-derived device name) — honouring an `abcd-lint:allow` line waiver (the `abcd-audit:allow` spelling is honoured too) |
| `identity-positioning` | warn | every registered surface still carries the canonical identity block's tagline (and pitch, where required); Where-gated on a committed `.abcd/positioning.json`, and per-repo upgradeable to `error` — see [`19-identity.md`](19-identity.md) |

## How it is built

The engine (`internal/core/repolint`) adapts its id/severity/where/fix/policy
vocabulary from repolinter's rule-object schema — severities `error|warn|off`,
chosen over the record-lint engine's (`internal/core/lint`) `blocker|warn` and
mapped only at the docs-currency rule's boundary — defines the
`abcd-lint:allow` / `abcd-audit:allow` line waiver natively in its privacy
rule, and adds path-presence and gitignore primitives. Rules are declarative data behind a
rule-loader seam, and output is serialised behind a seam that makes a later SARIF
export additive. No new dependency.

## References

- Plugin command: [`commands/lint.md`](../../../../commands/lint.md)
- Design record: [`plans/2026-07-13-abcd-audit-verb.md`](../../plans/2026-07-13-abcd-audit-verb.md)
- Intent: [`itd-85`](../../intents/drafts/itd-85-audit-verb.md)
- Onboarding consumer: [`15-prepare-this-repo.md`](15-prepare-this-repo.md)
