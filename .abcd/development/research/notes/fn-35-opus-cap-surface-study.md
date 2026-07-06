# fn-35 — Opus cap surface study (R5b)

Empirical-probe note for fn-35-ralph-quota-window-resilience.3 / R5b. Records the
documented `claude` CLI fallback behaviour for `claude-opus-4-8` with no
`--fallback-model`, so the quota classifier (R5a, owned by `.1`/`.2`) can rely on
an Opus cap surfacing a catchable rate-limit/error signal rather than a silent
Sonnet downgrade.

## Question

When a Ralph run is pinned to `claude-opus-4-8` (the `FLOW_RALPH_CLAUDE_MODEL`
default in `scripts/ralph/config.env`) and the account hits an Opus usage cap,
does the CLI surface a classifiable signal — or does it silently fall back to a
cheaper model, hiding the cap from `classify_iteration.py`?

## `claude --help` — `--fallback-model` is opt-in

`claude --help` documents the fallback as an explicit opt-in flag:

```
--fallback-model <model>   Enable automatic fallback to specified model(s)
                           when the default model is overloaded.
--model <model>            Model for the current session. Provide an alias for
                           the latest model (e.g. 'sonnet' or 'opus') or a
                           model's full name (e.g. 'claude-opus-4-8').
```

Ralph invokes Claude with `--model claude-opus-4-8` (via
`FLOW_RALPH_CLAUDE_MODEL`) and does **not** pass `--fallback-model`. Because
fallback is opt-in and Ralph never opts in, there is **no silent Sonnet
downgrade**: an Opus cap surfaces through the normal overloaded/rate-limited
error path that `classify_iteration.py` already inspects. The gating
no-`--fallback-model` guardrail and the Opus-token classification are R5a, owned
by `.1`/`.2`; this note records the documented behaviour that makes that
guardrail sound.

## Live Opus-cap sample

No live Opus-cap sample available yet. An Opus usage cap cannot be induced on
demand, so a captured-from-the-wild transcript of the exact emitted
tokens/`resetsAt` shape is defined as **follow-up evidence** to append here when
an Opus cap naturally occurs during a real run. R5b is satisfied by this note in
its current state (documented fallback behaviour + the no-sample-yet record); the
captured sample is an additive append, not a blocker.

## Verification

`claude --help | grep -iE 'fallback|model'` (run at implementation time)
confirmed the two flags above; `--fallback-model` is listed as an opt-in flag,
not a default.

## Related

- `scripts/ralph/config.env` — `FLOW_RALPH_CLAUDE_MODEL=claude-opus-4-8`, no
  `--fallback-model`.
- `scripts/ralph/classify_iteration.py` — the window-aware classifier that reads
  the surfaced rate-limit signal (R5a).
- `.flow/tasks/fn-35-ralph-quota-window-resilience.3.md` — R5b acceptance.
