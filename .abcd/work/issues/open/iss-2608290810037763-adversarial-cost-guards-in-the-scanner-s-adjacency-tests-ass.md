---
schema_version: 1
id: "iss-2608290810037763"
slug: "adversarial-cost-guards-in-the-scanner-s-adjacency-tests-ass"
severity: "minor"
category: "tech-debt"
source: "impl-review"
found_during: "intent-implementation-run"
found_at: "internal/adapter/scanner/adjacency_test.go"
---

Adversarial cost guards in the scanner's adjacency tests assert wall-clock seconds, which is a property of the machine rather than of the code, so they pass locally and fail on slower CI hardware under the race detector. One failed on both CI runners at 22.5 seconds against a 15 second bar while passing locally at 11.0 seconds on the same commit; the branch that introduced the probe and the merged tree measured identically, so it was not a regression. It was fixed for the costliest shape by shrinking that case's input, which is the lever the test file's own helper documents, but three sibling bars carry the same fragility and the same thin headroom. The machine-independent cost-CLASS guard beside them is the better model: it doubles the input and asserts the growth ratio, which no hardware difference can flip. Worth converting the remaining wall-clock ceilings to ratio assertions, or at least measuring a per-machine baseline and asserting a multiple of it.