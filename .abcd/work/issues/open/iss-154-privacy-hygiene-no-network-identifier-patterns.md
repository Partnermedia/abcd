---
schema_version: 1
id: "iss-154"
slug: "privacy-hygiene-no-network-identifier-patterns"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "managed-repo NEXT.md privacy-leak investigation 2026-07-29"
found_at: "internal/core/audit/rule_privacy.go"
---

privacy-hygiene detects only absolute home paths — no network-identifier patterns (IPv4 incl. CGNAT/tailnet 100.64/10, hostnames, device names, ports-in-context). Field incident: an agent committed a Tailscale investigation (tailnet IP, two device hostnames, firewall posture, service port, presence/away patterns) into a public repo's committed tier and audit passed it silently. The v1 deferral note names only emails and private-repo names; network identifiers are not even recorded as deferred.

Design note (2026-07-29): the detector should be an allowlist inversion — flag any network identifier NOT inside the reserved documentation ranges (RFC 5737 IPv4 blocks, RFC 3849 2001:db8::/32, RFC 2606 example domains, RFC 7042 doc MACs; loopback/unspecified exempt). The principle `examples-use-reserved-identifiers` (principles/) is what makes this feasible at near-zero false positives: every legitimate committed example uses a reserved value, so anything outside the ranges is a finding. Shipping this lint promotes that principle to a discipline per the itd-79 path.