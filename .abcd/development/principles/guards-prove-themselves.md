# Guards prove themselves

**The rule.** Any code path whose job is refusal — secret scan, deny rule,
SSRF check, symlink guard, quotation budget, licence detection — ships with a
test that watches it refuse. An untested guard is decoration, and a false
sense of guard is a liability.

**Why.** A guard fails silent: when ordinary code regresses, someone notices
the missing behaviour; when a guard regresses, the system keeps working and
simply stops refusing, which is invisible until the thing it guarded against
happens. The 2026-07-08 review found the launch bundle's symlink-dereference
and scripts-deny guards and memory's quotation-budget and licence-detection
compliance checks all at zero coverage — precisely the paths where a silent
regression costs most. A sibling project enumerates each non-negotiable
invariant alongside a paired test and a mandatory review trigger, which is the
mature form of this rule.

**Bounds.**

- The test exercises the *refusal*, not just the happy path around it: it
  presents the forbidden input and asserts the rejection, its error shape, and
  that no side effect occurred.
- Applies with full force to guards that exist for compliance or trust
  reasons even when no exploit is obvious — those are the ones nobody
  re-checks by hand.

**Promotion.** Pairing each declared invariant with a named test (and a lint
that checks the pairing) promotes this to a discipline.
