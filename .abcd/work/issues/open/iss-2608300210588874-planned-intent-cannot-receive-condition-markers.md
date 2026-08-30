---
schema_version: 1
id: "iss-2608300210588874"
slug: "planned-intent-cannot-receive-condition-markers"
severity: "major"
category: "bug"
source: "impl-review"
found_during: "itd-177 build review, 2026-08-30"
found_at: "internal/core/intent/ready.go (scope_conditions remedy), internal/core/intent/lifecycle.go (Plan refuses planned)"
---

A planned intent whose scope-condition bullet carries no identity marker is permanently NOT READY: the scope_conditions remedy names abcd intent plan, which exits 2 on a planned record, and the plugin surface forbids hand-typing markers. This is the default path (the scaffold leaves prose, plan stamps nothing, the maintainer writes the bullet after planning) and lands on every corpus record the moment a real condition replaces the nullity token. spc-55 states that Plan is idempotent and a re-run stamps only the unmarked bullets; the code refuses the re-run instead.
