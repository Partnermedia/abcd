---
schema_version: 1
id: "iss-22"
slug: "itd-6-lifecycle-look"
severity: "minor"
category: "inconsistency"
source: "agent-finding"
found_during: "intent-dependency-sweep"
found_at: ".abcd/development/intents/planned/itd-6-rp-mcp-only-integration.md"
resolution: "Lifecycle look concluded: itd-6 stays in planned/. It is scheduled — Phase 0's Scope carries it as the oracle adapter seam under the ADR-25 framing, and phase-3's dependency rationale leans on it — so a superseded/ move would break adr-34's scheduled-implies-planned invariant. The banner was the defect: it asserted whole-intent supersession and historical-record status. Rewritten to say the RP-only-single-integration FRAMING is superseded (ADR-25/ADR-22) while the intent remains the live, scheduled vehicle whose RP specifics become the RP adapter's contract at spec time."
---

itd-6 sits in planned/ while its own banner records supersession-in-framing by ADR-25 (one optional adapter among many); lifecycle look needed