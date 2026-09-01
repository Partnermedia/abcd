---
schema_version: 1
id: "iss-2609012111150528"
slug: "itd-130-says-the-persistent-data-dir-was-rejected-but-itd-132-adopted-it"
severity: "minor"
category: "inconsistency"
source: "agent-finding"
found_during: "ship-audit-itd-130-itd-132-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: ".abcd/development/intents/shipped/itd-130-abcd-update-completes-a-chosen-update-in-one-verb-it-fetches.md"
---

itd-130's Decisions section records that 'the persistent-data-dir alternative is rejected, not deferred' (grilled 2026-08-20), and one day later itd-132/spc-35 adopted exactly that shape as the download cache with a copied PATH file, superseding spc-21's re-fetch criterion. itd-130's text was never revised, so the shipped record now carries a decision its successor reversed without a typed link between them. Surfaced by the itd-130 fidelity audit (receipt rcp-264f7b144576) as a diverged item. The fix is a record change, not code: itd-132 should carry a typed 'reverses' link to the itd-130 decision, or itd-130's Decisions should note the supersession with the date, and the reversal is a human's to confirm.
