---
schema_version: 1
id: "iss-2608311502474022"
slug: "the-amnesia-eval-s-timestamp-scan-rests-on-a-premise-that-is"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "itd-187 fidelity audit rcp-d3041aa2b510"
origin: researcher-authored
production_mode: hand-written
found_at: "evals/coldreading_determinism_test.go"
---

The amnesia eval's timestamp scan rests on a premise that is already false for paths. spc-65 confines the scan to the manifest on the ground that the manifest carries paths, field names and hashes only, so a timestamp-shaped token there is unambiguously a defect. But the packed-digit rule is an unanchored run of eight or more digits and item paths are scanned as plain scalars, so a manifest naming a record path that carries a long numeric id is reported as a packed run of digits, which is how a moment travels. Fed a path holding a real nineteen-digit capture id, the scan fires. It is latent today because no admitted record family carries such a path, but the justification the scoping decision rests on is wrong now rather than at some future point, and the scan will produce a false finding the first time such a family is admitted.
