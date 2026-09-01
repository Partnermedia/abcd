---
schema_version: 1
id: "iss-2609012111168872"
slug: "version-check-and-the-upgrade-abcd-error-sites-still-do-not-name-the-update-verb"
severity: "minor"
category: "ux"
source: "agent-finding"
found_during: "ship-audit-itd-130-itd-132-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/surface/cli/version.go"
---

itd-130's press release says 'now the next line is abcd update' and that the eight schema-too-new error sites 'now have a verb they can name'. Delivered: version --check still prints only 'update available: X -> Y' with no next step, and the schema-too-new sites (for example lifeboat/embark.go and release/ingest.go) still say 'upgrade abcd' without naming the verb. The verb exists; the surfaces that motivated it were not re-pointed. Surfaced by the itd-130 fidelity audit (receipt rcp-264f7b144576) as a diverged item. Fix is one string per site plus a test that greps the tree for 'upgrade abcd' outside the update verb's own remedy text.
