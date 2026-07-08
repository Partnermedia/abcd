# Personas come from the registry; the role picks the name

Every persona in a user story, press-release quote, worked example, or scenario
comes from the registry at [`personas.json`](../personas.json) — the single
source of truth for the roster. Selection is **by role, never by name**: the
scenario's audience determines the role, and the role's registered name is
used. Names form a fixed alphabetical sequence (Alice, Bob, Carol, Dave, …);
when the roster grows, new personas continue the sequence. No invented names,
no real people's names as stand-ins.

Every persona is referred to as **they/them**. The real user, when referred to
at all, is also they/them — never he or she.

**Why.** Persona names are load-bearing noise: a bespoke name per document
makes readers wonder whether the name carries meaning, and a real-sounding
name risks colliding with an actual person or project. Binding names to roles
makes them carry exactly one bit of meaning — the audience — consistently
across the whole corpus: Frank is always the DevOps voice, Kira always the
maintainer's. The alphabetical sequence keeps the roster auditable and
extension mechanical.

**Applies to.** Intents (press-release quotes and scenarios), brief surface
docs, ADR examples, plans, and any user-facing prose that stages a scenario.
Enforcement: a registry-membership lint is the intended gate (names outside
the roster fail); until it exists, this principle is review-enforced.
Documents that predate the rule are corrected opportunistically when
otherwise edited, not swept.
