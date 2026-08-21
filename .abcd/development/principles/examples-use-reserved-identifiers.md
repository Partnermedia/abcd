# Examples use reserved identifiers

**The rule.** Every illustrative machine identifier in anything committed or
published — docs, tests, fixtures, intent scenarios, rule text, error-message
examples — comes from a reserved documentation range, never from a real
environment. The persona rule (itd-79) applied to infrastructure: just as a
quoted persona is always a registry name, an example host is always a fixture
host.

- **IPv4:** RFC 5737 documentation blocks — `192.0.2.0/24`, `198.51.100.0/24`,
  `203.0.113.0/24`.
- **IPv6:** RFC 3849 — `2001:db8::/32`.
- **Domains:** RFC 2606/6761 — `example.com`, `example.org`, `.example`,
  `.test`, `.invalid`.
- **MAC addresses:** RFC 7042 — `00:00:5E:00:53:00`–`FF`.
- **Hostnames / device names:** derive from the persona registry
  (`alice-laptop`, `bob-desktop`, `carol-server`) — role picks the persona,
  persona names the machine.

**Why.** A real identifier in an example is a leak that outlives its file:
history rewrites cannot recall cached views or merged-PR refs, so prevention
at authoring time is the only cheap point of control. Reserved values are also
*mechanically distinguishable*: when every legitimate example sits inside a
known range, any identifier outside it is flaggable with near-zero false
positives — the convention is what makes a network-identifier lint feasible at
all (iss-154). The 2026-07-29 incident demonstrated the failure mode: a
machine investigation journaled into a committed working file put a real
tailnet IP, two real device names, and the machine's presence patterns on a
public tip, and no gate could object because nothing distinguished a real
identifier from an illustrative one.

**Bounds.**

- The rule covers *illustrative* values. Real identifiers that must be handled
  (a user's actual config, a probe result) are runtime data — they belong in
  the gitignored local tier or the itd-74 private banlist's protection, never
  in committed prose; that boundary is itd-74's job, not this rule's.
- Loopback (`127.0.0.1`, `::1`) and unspecified (`0.0.0.0`) addresses are
  standard protocol values, not leaks; documenting a bind default is in
  bounds.

**Live instance.** The persona registry (`.abcd/development/personas.json`)
already implements the name half; no fixture-identifier equivalent exists yet
for hosts and addresses.

**Enforcement.** The `privacy-hygiene` rule flags any committed network
identifier *outside* the reserved ranges (the allowlist inversion recorded in
iss-154): `scanner/network.go` encodes the RFC 5737/3849/2606/7042 ranges and
cites this principle by name, and `repolint/rule_privacy.go` arms the rule.
Because that mechanical gate now exists, the itd-79 path calls for a
discipline-kind intent to carry the principle; minting it is tracked in iss-390.
