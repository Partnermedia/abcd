---
id: adr-25
slug: host-delegated-llm-default
status: accepted
date: 2026-07-06
supersedes: [adr-8]
superseded_by: null
related_intents: []
related_rfcs: []
related_adrs: [adr-6, adr-22, adr-23]
---

# ADR-25: The LLM is host-delegated by default; oracles are opt-in adapters

## Context

abcd's review and oracle path was a fixed cascade: RepoPrompt → codex →
in-session, with a dual-backend review that ran RP and codex in parallel and
trusted their verdicts asymmetrically (ADR-8).
That cascade assumed those tools were always present and always the way abcd
reached a model — the same mandatory-bundling premise the rebuild removes
([ADR-22](0022-bundled-deps-as-pluggable-adapters.md)). It also put abcd in the
business of owning model access, credentials, and orchestration for backends it
no longer requires.

The rebuild's core does deterministic work and returns structured results
([ADR-23](0023-transport-agnostic-core.md)); the open question is how it reaches
a model when a step genuinely needs one.

## Decision

The LLM is **host-delegated by default**. abcd's core does the deterministic
work, then hands a **prompt** to the **host's subagent dispatch** (the agent
harness driving abcd) rather than calling a model itself. The host owns model
choice, credentials, and execution; abcd owns the prompt and consumes the
structured result.

Concrete oracle backends — **native, CLI, API, MCP** — are **opt-in adapters**
behind the same seam, selected when an operator wants abcd to reach a model
directly. The old RP→codex→in-session cascade is **replaced** by this
delegate-by-default-plus-adapters model, so this ADR **supersedes**
ADR-8.

The *principle* ADR-8 established survives as **adapter guidance**, not as a
hard-wired cascade: a **scoped reviewer** (seeing only a selection) and a
**broad reviewer** (reasoning over the whole repo) have complementary blind
spots, and the two are trusted **asymmetrically** — the scoped verdict gates,
the broad reviewer is mined for findings, and a review-fix loop declares a
stopping rule up front. When an operator wires two oracle adapters for a
high-stakes review, this is how they should be combined; it is advice the
adapter layer offers, not a cascade the core imposes. ADR-6's concern about the
oracle **capturing** review artifacts is likewise part-superseded here: capture
is now whatever adapter is wired, over the native default.

## Alternatives Considered

- **Keep the fixed RP→codex→in-session cascade.** Preserves the tuned
  dual-backend behaviour. Rejected: it hard-requires two external tools ADR-22
  makes optional, and makes abcd own model access it should delegate to the
  host.
- **Native-only LLM (abcd always calls the model).** One code path. Rejected:
  it forces abcd to carry credentials and provider logic even when the host
  already has a capable subagent dispatch sitting right there.
- **Chosen: host-delegated by default, oracle adapters opt-in, ADR-8's
  scoped/broad + asymmetric-trust principle preserved as adapter guidance.**
  Zero required model plumbing; full capability when an operator wants it.

## Consequences

- With no oracle adapter wired, abcd needs no API keys or model config: it
  emits prompts and the host runs them. This is the default install.
- The scoped-vs-broad reviewer pairing and asymmetric-trust stopping rule move
  from hard-coded pipeline to documented adapter guidance — available when two
  backends are wired, absent when they are not.
- The oracle seam is one interface with four adapter shapes (native, CLI, API,
  MCP); each front door ([ADR-23](0023-transport-agnostic-core.md)) inherits
  whichever is configured.
- Review-artifact capture and its redaction follow whichever adapter runs,
  over the native transcript/redaction default
  ([ADR-29](0029-native-transcript-corpus.md), ADR-6).
