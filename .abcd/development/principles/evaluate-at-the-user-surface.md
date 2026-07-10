# Evaluate at the user surface

**The rule.** An evaluator consumes the system through the same surface its
users do — the binary's arguments and rendered output, the plugin markdown,
the published docs. Any privileged path (an internal API, parsed internals,
direct core calls) makes the evaluator structurally blind to exactly the
failures users hit, so a verdict obtained through privileged access is not
evidence for a user-facing claim.

**Why.** The UX-evaluation study of Tan et al. (UCL, 2026; arXiv 2604.09581)
demonstrates the
blindness empirically: a GUI-grounded evaluator is susceptible to — and
therefore able to detect — the same pitfalls as human users, where agents
operating on simplified internal representations bypass the clutter and
ambiguity real users face. Its decisive case is a page where visual clarity
masks a functional defect: the internal element is present and correctly
identified, the flow is broken, and a privileged evaluator passes it. Weng's
harness-engineering essay (Lil'Log, 2026) gives the loop-shaped version of
the same failure: a system optimised against a proxy signal overfits the
proxy — a verdict from unit tests transfers to users only as far as the
tests model the user surface.

**Bounds.**

- Unit tests legitimately use privileged access; the rule binds evaluation
  that grounds user-facing or done claims, not the interior test pyramid.
- This is the evaluation face of the "wired or it isn't done" boundary:
  "demonstrably executes there" means demonstrated *there*, at the surface a
  user reaches, not at the core function the surface wraps.

**Live instance.** The wiring tests in `internal/surface/cli` drive verbs as
a user does — arguments in, text or JSON out — and assert on the same
rendered output a shell sees, rather than calling `internal/core` directly.

**Promotion.** The MVP is a placement convention: any evaluation that
grounds a user-facing claim lives at a surface package and takes only inputs
a user could supply. The tool is a surface-evaluation harness — a dogfood
run of the shipped verbs against this repo wired into CI — that makes the
privileged shortcut the harder path; as a maintained gate it carries the
usual lifecycle (false-positive budget, kill criterion, demotion to advisory
on stale calibration).
