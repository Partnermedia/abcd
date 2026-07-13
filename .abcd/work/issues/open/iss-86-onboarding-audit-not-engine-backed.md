---
schema_version: 1
id: "iss-86"
slug: "onboarding-audit-not-engine-backed"
severity: "major"
category: "future-work-seed"
source: "agent-finding"
found_during: "2026-07-13 B1 dogfood: prepare-this-repo audit of Manuscripts"
found_at: "commands/abcd/prepare-this-repo.md"
---

abcd has no read-only convention-audit surface, and the binary gap model is disjoint from the prepare-this-repo skill audit. abcd ahoy --json on a partially-scaffolded repo reports only abcd own install-plumbing gaps (config.json, rules.json, marker blocks, history store) and is silent on the gaps the onboarding skill exists to fix: the missing committed .abcd/work/ tier, the absent AGENTS.md conventions router, and durable decisions leaking into the gitignored layer. The entire Phase 2 gap report had to be produced by hand. Detector: a read-only abcd audit (or extended ahoy doctor) verb that checks the three-tier layout, decision-durability, docs Diataxis/tense, and privacy -- every convention the skill audits by prose. Acceptance corpus (Manuscripts dogfood): (a) missing .abcd/work/ tier undetected, (b) decisions/ vs decisions/adrs/ layout drift undetected, (c) whole convention audit done manually. This is the single highest-value onboarding gap: it is what would make prepare-this-repo engine-backed instead of agent-glue.