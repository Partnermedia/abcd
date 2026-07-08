# Unrecognised input never writes

**The rule.** A verb that mutates state fails closed: anything not exactly
recognised is an error, never a fallback interpretation. Machine-readable
output contracts (`--json`) hold on the error path exactly as on success.

**Why.** Fallback interpretation turns typos into mutations. The 2026-07-08
review demonstrated it live: a misspelled `capture` subcommand was swallowed
as capture *text* and filed a new issue in the ledger — the user asked to
resolve an issue and instead created one. The same review found `--json`
errors emitted as raw Go text, so the one consumer least able to improvise
(a machine) gets the least structured failure.

**Bounds.**

- Read-only verbs may be forgiving; the rule binds at the point of mutation.
  A bare status invocation guessing generously is a convenience — a write
  verb guessing is a defect.
- "Exactly recognised" includes near-misses: an argument one edit away from a
  sub-verb name deserves a did-you-mean error, not silent acceptance as
  payload.

**Promotion.** A surface-level test convention — every mutating verb has a
malformed-input case asserting no write occurred — would make this a
discipline.
