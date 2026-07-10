# Verifier selects, gates decide

**The rule.** An LLM verifier slots *above* the deterministic gate, never in
place of it. Its province is graded selection where no deterministic check
exists — ranking candidates, best-of-N picking, monitoring a long run's
progress. Admission to the record is decided by deterministic gates (tests,
preflight, lint); a probabilistic verdict is a selector and an early-warning
signal, not an authority.

**Why.** Graded selection is where LLM judgment measurably helps, and gate
replacement is where it measurably fails. The LLM-as-a-Verifier study (Kwok
et al., Stanford/Berkeley/NVIDIA, 2026) ranks agent trajectories with an LLM
but keeps a hidden deterministic grader as ground truth throughout: verifier
pairwise accuracy tops out near 78%, so a gate built on it alone silently
passes roughly one wrong candidate in five, while oracle best-of-K reaches
near 99% — the selector, not the generator, is where the judgment pays. The same
study's graded score over a run's prefix separates healthy from stalled runs,
explicitly as advisory telemetry rather than a kill switch. The GLM-5 report
(Zhipu AI and Tsinghua, 2026) draws the complementary line: its agentic judge
counts only after measuring 94% agreement against human labels, and when a
learned reward was gamed the fix went to the measuring instrument's
grounding, never the outputs — a gate handed to a probabilistic judge
inherits gaming surface and calibration drift as standing liabilities.

**Bounds.**

- An LLM verdict may advise a gate — flag, rank, annotate, propose — but the
  gate's blocking decision stays deterministic. Where no deterministic check
  exists yet, the verifier is the interim and building the check is the debt.
- Claims backed only by LLM judgment are marked judged-only and never
  reported in the vocabulary of enforcement (composes with
  [enforcement-claims-are-facts](enforcement-claims-are-facts.md)).
- Discrete verdicts are for gating a single artefact; ranking uses graded
  scores — the study's discrete judge tied on 88 of 100 comparisons where a
  graded expectation ranked correctly.

**Live instance.** The 2026-07-08 multi-agent record review used independent
LLM refuters to grade raw findings before capture — 3 of 46 refuted and
dropped, the rest clustered into ledger issues — while admission of the
resulting changes stayed with `make preflight` and the record's deterministic
gates: selection above, gates deciding.

**Promotion.** The MVP is a convention marking judged-only verdicts as such
wherever they persist (ledger entries, review notes), plus a labelled
fixture set of known-good and known-bad artefacts attached to any oracle
whose scores feed selection. The tool is an opt-in oracle adapter tier whose
measured accuracy and tie rate are recorded per configuration — demotable to
advisory on stale calibration, archivable when a deterministic check
supersedes it.
