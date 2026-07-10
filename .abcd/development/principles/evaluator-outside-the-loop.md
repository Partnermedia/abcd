# Evaluator outside the loop

**The rule.** The gate that judges a change and the permission layer that
bounds it live outside the surface the change may edit. A change that
touches its own gate — the lint config, the banlist, the ratchet baseline,
the detector's tests — is structurally suspect regardless of its content and
needs a decision from outside the loop: a separate change, judged by
whoever owns the gate, never bundled with the work the gate judges.

**Why.** An adaptive producer optimises whatever signal bounds it, and given
write access to the signal it optimises the signal. The Red Queen Gödel
Machine (Iacob et al., Cambridge/NVIDIA and others, 2026) is built on this
separation: feedback that creates candidates is isolated from the evidence
that selects them, selection runs on a held-out split the producer never
sees, and an evaluator is replaced only by beating the incumbent on a fixed
anchor corpus the contestants cannot touch — without which its ablations
show producers trivially saturating a gate they can game. Weng's
harness-engineering essay (Lil'Log, 2026) states the same as a design rule
for self-improving loops: the evaluator and permission control sit outside
the loop that evolves, with held-out tests, trace audits, and human review
at the decision points that matter. The lineage review already carries the
management-canon form as gatekeeper independence — the bar raiser sits
outside the pressured team (Bryar and Carr, *Working Backwards*).

**Bounds.**

- Gates do evolve; anchored succession is the sanctioned path — a challenger
  proves itself against the corpus and the incumbent before replacing it.
  The rule governs *who decides*, not immutability.
- Composes with [ratchet-not-big-bang](ratchet-not-big-bang.md): the
  baseline shrinks freely from inside the loop, but grows only by a
  reviewed, outside decision — silent regeneration defeats the ratchet.
- "Outside" is relative to the change, not the repo: the gate's owner may be
  the same human wearing a different hat, provided the gate edit lands as
  its own reviewable change naming why the gate moved.

**Promotion.** The MVP is a diff-shape convention: no change simultaneously
alters gated content and the gate that judges it — baseline regenerations,
banlist edits, and detector-test changes each land alone, stating their
reason. The tool is a hook or CI check that flags any diff touching both
sides of a gate it can map; as a maintained detector it is demotable to
advisory when its gate map goes stale and archivable to regression-only on
saturation.
