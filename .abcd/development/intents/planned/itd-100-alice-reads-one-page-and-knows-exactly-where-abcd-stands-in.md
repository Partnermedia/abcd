---
id: itd-100
slug: alice-reads-one-page-and-knows-exactly-where-abcd-stands-in
spec_id: spc-15
kind: standalone
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
impact: additive
---

# Alice Reads One Page and Knows Exactly Where abcd Stands in the Agentic-AI Landscape

## Press Release

> **Alice reads one page and knows exactly where abcd stands in the agentic-AI landscape.** Alice evaluates developer tooling for her team. Every tool describes itself in its own invented vocabulary, and she has to reverse-engineer how it relates to the terms the industry actually uses — MCP, the agent loop, context engineering, guardrails, policy-as-code, evaluations. abcd gives her a single reference page: a terminology crosswalk mapping each established agentic-AI term to how abcd incorporates it (naming the native verb or principle), adapts it under a sharper name, deliberately rejects it for a recorded reason, or is still watching it via a ledger entry she can follow. Every established definition cites a primary source — a spec, a standards body, or a peer-reviewed paper — and every claim about abcd is verified against the committed record, never memory. "I stopped reverse-engineering the invented names," said Alice. "One page told me, in the vocabulary I already use, what this tool does — and, rarer still, what it deliberately does not."

## Why This Matters

A tool that invents its own vocabulary asks every evaluator to do translation work it should have done itself, and the translation is where trust is won or lost: an unmapped term reads as either ignorance of the field or marketing evasion. The crosswalk makes the mapping a maintained, cited artefact instead of folklore — and it is deliberately honest in both directions, recording not only what abcd incorporates but what it rejects and why, with every rejection traceable to the committed record. A REJECTS row with a recorded reason is a stronger credibility signal than any feature claim.

## Acceptance Criteria

- Given Alice opens `docs/reference/terminology.md`, when she reads any row, then it carries the term, a one-line established definition with a footnote citation, and abcd's position as exactly one of USES / ADAPTS / REJECTS / WATCHING.
- Given any footnote is followed, then it resolves to a primary source — a specification, a standards body, a peer-reviewed paper with DOI, or the engineering document where the term originated — never an aggregator or secondary write-up.
- Given any abcd-position claim in the table, then it is verifiable against a committed record artefact (a verb, principle, ADR, intent, or ledger id); WATCHING rows cite an open ledger entry by id.
- Given `abcd docs` (docs-currency lint) runs over the repo, then the page passes, with host and vendor names confined to citation footnote lines under the sanctioned allow escape.
- Given the page links to abcd-native concepts, then every cross-reference targets `docs/**` only; native terms carry an inline gloss rather than a deep link into the development record.

## Open Questions

_None recorded yet._

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
