---
id: adr-39
slug: host-tier-policy
status: accepted
date: 2026-08-16
supersedes: null
superseded_by: null
related_intents: [itd-22, itd-2]
related_rfcs: []
related_adrs: [adr-23, adr-24, adr-25]
---

# ADR-39: Host-tier policy — MCP floor, reference host, open-source default

## Context

The 2026-08-15 decision put multi-harness support in scope, reversing the
out-of-scope annotation on itd-22's lineage. adr-23 supplies the mechanism —
a transport-agnostic core behind thin front doors — but the record held no
policy for which hosts get which investment, what "supported" means, or what
the shipping default is. Harness surveys (2026-08-15) showed why a policy is
needed: extension surfaces differ radically between harnesses — one exposes
an ES-module hook API, another compiles Go plugins against a versioned
capability registry — so above the core, per-host work shares nothing except
the machinery this policy names.

## Decision

Six rules, together the host-tier policy:

1. **The MCP front door is the universal floor.** Any MCP-capable harness
   can drive abcd's core verbs with no per-host work. The floor is always
   available; no adaptor may require more than it to reach core behaviour.
2. **The current plugin host's integration is the reference
   implementation** — the assumed SOTA surface for the time being, kept
   current with the core.
3. **The shipping default ultimately moves to an open-source harness.** The
   default flips only when that host passes the same parity conformance
   suite the reference host passes, and the maintainer flips it deliberately
   (verifier selects, gates decide). The previous default remains a
   supported host after the flip.
4. **Per-host delivery climbs the adaptor ladder, most-native-first:** the
   host's own plugin format, then any other native seam (lifecycle wiring,
   config), then the MCP floor. A rung a host cannot support degrades to the
   next rung and says so (loud staging) — never a silent no-op. Behaviour
   lives once in the core and is never double-written for one host (one
   canonical primitive).
5. **Where a host exposes a native work-plane** — goal or work-item
   tracking, a canonical knowledge base, staged verdict gauntlets — the
   adaptor maps abcd's concepts onto the host's own rather than running a
   parallel artefact set. The mapping is per-host design work, owned by that
   host's adoption intent.
6. **Each host's adoption is its own intent**, filed at an explicit
   maintainer decision. Commitment records name no harness before its
   adoption is decided; until then a host is described by capability class.

## Consequences

- itd-22 carries the shared machinery only — the host profile seam, the
  ladder semantics, the parity conformance suite — and per-host adoption
  intents come one at a time on top of it.
- The MCP front door gets its own intent: under this policy the floor is the
  first build, not a later convenience.
- The parity suite is defined once and versioned; "supports host X" may be
  written in present tense only when the suite demonstrably passes on that
  host (enforcement claims are facts).
- The default-host flip is a gated, reversible event with a rehearsed way
  back, never a gradual drift.
