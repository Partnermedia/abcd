---
schema_version: 1
id: "iss-96"
slug: "now-that-transcripts-are-captured-automatically-on-every-ses"
severity: "minor"
category: "security"
source: "manual-test"
found_during: "itd-89-m1"
found_at: "internal/adapter/scanner/patterns.go"
---

Now that transcripts are captured automatically on every session end, the scanner's secret-pattern coverage becomes load-bearing in a way it was not when capture was a manual verb nobody ran. Verified by live test: the bundled patterns DO catch anchored tokens (AKIA... access key IDs, ghp_/gho_/sk-ant- style prefixes) and absolute home paths, but they do NOT catch unanchored high-entropy values — an AWS SECRET access key (the 40-char base64 value, no prefix), a bare password, or a generic API token with no recognisable prefix all pass through into the store verbatim. This is the standard prefix-matching limitation and is pre-existing, not a regression; the point is that the blast radius changed. Consider entropy-based detection or the opt-in gitleaks adapter for the transcript path specifically, where the input is unstructured prose rather than curated source.

---

**Verification (2026-08-01, item A6 of the v0.5.0 plan).** Re-checked against the
network-identifier pattern set landed by A2, which is folded into
`DefaultPatterns` and therefore inherited by the transcript path. Outcome: one
class newly covered, the three classes this entry names confirmed as residue by
test rather than by inspection. The entry stays open, re-scoped below.

**Newly covered since capture.** Nothing in the entropy classes, but the
transcript path gained a detection class it did not have when this was filed:
non-reserved IPv4/IPv6/MAC addresses (`net:ipv4`, `net:ipv6`, `net:mac`,
hard_fail) and LAN/device hostnames (`net:lan_hostname`, `net:device_hostname`,
warn), by allowlist inversion over the reserved documentation ranges. All five
kinds are exercised on the transcript path itself in the corpus below, not
inferred from their presence in `DefaultPatterns`. A leaked machine identifier is
no longer part of this entry's residue.

**Residue, verified empirically.** Synthetic specimens presented to the
transcript path's own entry point — `(*Scanner).ScanText`, the call
`internal/core/history.Capture` makes at stage one — produce zero findings:

- bare secret-access-key value (40 characters, no prefix), both with and without
  an `AWS_SECRET_ACCESS_KEY=` key name in front of it — passes through;
- bare password, both after `password:` and after `--password=` — passes through;
- prefix-less API token (32 hex characters after `api_key =`; 40 base64
  characters after `Authorization: Bearer`) — passes through;
- a genuinely high-entropy 40-character alphanumeric value (5.32 bits per
  character, assembled at run time from a fixed-seed shuffle), both bare and
  after an `api_key =` key name — passes through.

Anchored controls in the same corpus are caught (`token:aws_access_key`,
`token:github_pat`, `token:anthropic`, `token:pem_private_key`,
`home_path_self`, `home_path_other`, plus all five network kinds above), so the
specimens reach a working detector; one of those controls sits inside the
negative test too, so an emptied pattern set cannot satisfy it vacuously. The
TOKEN patterns are prefix-anchored apart from `rp_session_key`, which keys on one
literal JSON field name (`"sessionKey"`) rather than on a generic key-name class,
and no pattern measures entropy — so a `password:`, `api_key =` or
`Authorization: Bearer` key name beside a value does not help. The set as a whole
is not uniformly prefix-anchored, and the entry should not be read as saying so:
the network kinds are an allowlist inversion over the reserved documentation
ranges, the identity kinds key on the probed identity, `token:pem_private_key` is
a literal header, and `Pattern.SkipAt`
(`internal/adapter/scanner/patterns.go:19-24`) already lets a pattern accept or
reject a match by what SURROUNDS it on the line. None of those mechanisms reaches
an unlabelled value, which is why the residue stands.

**Pinned in the tree**, so the residue is evidence rather than a claim:
`TestTranscriptPathMissesUnanchoredEntropy`
(`internal/adapter/scanner/transcript_coverage_test.go`) at the pattern-set
boundary, and `TestCaptureStoresUnanchoredEntropyVerbatim`
(`internal/core/history/history_test.go`) at the store boundary, which captures a
six-line transcript carrying an anchored token on the first line and the
unanchored specimens on the lines below it, then asserts each specimen present
verbatim in the written record while the anchored token is absent AND its
`maskSecret` fingerprint present — absence alone would also be satisfied by a
store that dropped the line. The same capture asserts the contra case, so the
class A2 did close is evidence rather than prose: a flagged private address on
the last line is absent from the stored record and its `[redacted-address]`
placeholder present.

**What the pins can and cannot catch.** They assert CURRENT behaviour, and their
reach is the PATTERN SET this path runs. They alarm reliably for growth in that
set — a charset/length floor, key-name context matching, any new `DefaultPatterns`
entry, and now a genuine entropy floor. That last one needed fixing: the original
specimens all repeat, and measure 3.12 (`FAKEfake00`×4), 2.16 (`deadbeef`×4) and
2.75 (base64 run) bits per character, all BELOW the ~3.5-bit threshold an entropy
detector conventionally uses, so such a detector could have landed with every pin
still green. A high-entropy specimen was therefore added at both boundaries —
5.32 bits per character, assembled at run time from a fixed-seed shuffle of the
alphanumeric alphabet, with its measured entropy asserted in-test (at or above
4.5 bits per character) so it cannot quietly degrade into a repetitive value.
They CANNOT alarm for option (c) below: an opt-in external-scanner adapter never
consults `DefaultPatterns`, so `ScanText` keeps returning nothing and both pins
stay green while coverage has in fact grown. That route must be re-pointed by
hand, and the pins alone are therefore not grounds to close this entry.

**Re-scoped to a decision, not an implementation.** What remains is a choice
between three options that differ in REACH, not only in false-positive cost.
Stating them as equal-coverage-at-different-cost would misdescribe them:

- **(a) an entropy/charset-class detector with a length floor.** The only one of
  the three that reads the VALUE, so the only one that reaches an unlabelled
  value: it would cover every specimen above, including the bare 40-character
  secret-key shape with no key name beside it.
- **(b) key-name context matching** (`password`, `secret`, `token`, `api_key` to
  the left of a delimiter). Labelled values ONLY. It is structurally incapable of
  this entry's own first-named case — a bare secret-access-key value with no key
  name, pinned as the case "bare 40-char secret-key shape, no key name" — because
  there is no key for it to match. Its merit is elsewhere: it generalises
  mechanisms the set already ships rather than introducing one, since
  `rp_session_key` matches on a literal field name today and `Pattern.SkipAt`
  already decides a match by its surrounding line.
- **(c) the opt-in external-scanner adapter over the transcript path only.**
  Reach is whatever the adapter's ruleset delivers — a measurement, not an
  assumption. Measured 2026-08-01 with gitleaks 8.24.3's default rules over the
  store fixture's own specimen set: ONE finding, the keyword-adjacent passphrase
  (`generic-api-key`). The bare secret-key shape, the hex token, the base64 token
  and the high-entropy 40-character value all passed, the last despite a
  `session_token` key name beside it. (The synthetic anchored control passed too,
  its repeated characters falling under gitleaks' own entropy filter — a fact
  about the fixture, not about that ruleset's coverage of real prefixed tokens.)

False-positive cost is the second axis, and it lands hardest on (a): a transcript
is full of hashes, base64 blobs and identifiers that are not credentials, and a
redaction false positive corrupts the record redaction exists to preserve. Reach
and cost together are the open question, and where the bar sits is the
maintainer's to decide — grill-then-implement, not autonomous work.
