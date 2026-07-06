# Reviews

Commissioned reviews of this project — plan reviews, code reviews, external audits — conducted **outside abcd's own command machinery**. The discriminator for what belongs here is provenance, not who (or what) did the thinking: if no abcd verb invocation produced the artefact, it lands here.

## What does NOT belong here

- **Per-invocation artifacts from abcd surfaces** (oracle audits, grill reports, disembark audits) — those go to `.abcd/logbook/<verb>/<ts>/` as traces of the command run that produced them.
- **Distilled outcomes** — when a review changes course, the settled decision graduates to `../../development/decisions/` (an ADR or a decision note). The review folder is the evidence trail, not the decision record.
- **Individual open findings** — findings graduate into intents, issues, or ADRs. Reviews are not a shadow backlog.

## Conventions

- One directory per review: `<YYYY-MM-DD>-<scope>/` (the date is content here, as with ADR `date:` fields — it identifies the point-in-time snapshot the review describes).
- `00-summary.md` carries the consolidated verdict and ranked actions; numbered siblings carry the underlying reports.
- **Append-only.** A review is immutable once written — reality is never edited to match a review, and a review is never edited to match reality. Follow-up work gets a new dated directory.
- All paths in review documents are repo-relative.

## Related Documentation

- [`../CONTEXT.md`](../CONTEXT.md) — current working state
- [`../DECISIONS.md`](../DECISIONS.md) — decisions pending graduation to ADRs
- [`../../development/decisions/`](../../development/decisions/) — where review outcomes graduate
