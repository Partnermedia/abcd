# abcdev-site — the website work cluster

The research cluster behind the abcdev.app decisions (adr-47, adr-48; intents
itd-135–itd-139; the itd-140 discipline). Three artefacts:

| File | Role |
|------|------|
| [`plan.md`](plan.md) | The 2026-08-21 investigation and plan — SOTA survey, decisions, URL map, build pipeline, `record.json` schema, pages, design system, rollout, open decisions, risks |
| [`abcdev-implementation-prompt.md`](abcdev-implementation-prompt.md) | The build-session prompt, with the adversarial review of the two architecture questions (one repo or two; deploy per release or per merge) |

The clickable prototype (`abcd-web.html` — the behavioural spec whose "notes
for the team" toggle carries each section's rationale, every rendered block
carrying `data-src="path#heading"` provenance) stays in the local tier at
`.abcd/.work.local/scratch/site-plan/`; see the header of
[`plan.md`](plan.md) for why it cannot be committed.

The README-migration bundle (docs pages, assets, `.abcd/site.json`,
`site-src/ui.json`, `MIGRATION.md`, `compose.py` + `build_data.py`) lives
unpacked in the shared working tier at `.abcd/work/site-plan/readme-migration/`
— see the header of [`plan.md`](plan.md) for why. Nothing in this cluster is
rendered or linted as docs; Phase 1 of the build promotes the bundle's files
to their real paths.
