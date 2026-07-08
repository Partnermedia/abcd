# Ratchet, not big bang

**The rule.** A new gate arms immediately against a frozen baseline of the
violations that exist on the day it lands; the baseline may only shrink. Never
delay arming a gate until the corpus it checks is clean.

**Why.** Gates that wait on cleanup never arm, and cleanup without a gate
re-drifts — the brief's dedicated issue-drain pass demonstrably re-drifted the
same day it ran. A baseline ratchet decouples the two: new violations fail
from day one while the inherited debt burns down on its own schedule, and the
shrinking baseline is itself a visible progress measure. A sibling project
runs its architectural-invariant lints exactly this way, freezing existing
violations in a committed baseline diff.

**Bounds.**

- The baseline is committed and reviewed like any other artefact; silently
  regenerating it to admit a new violation defeats the ratchet.
- A baseline entry is debt, not permission: each one should trace to an issue
  or intent that retires it.

**Promotion.** Adopting baseline support in `record-lint` (or any gate) makes
this mechanical for that gate; the principle graduates per-gate as baselines
become checked artefacts.
