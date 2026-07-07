# evals/data — fixtures (reserved for v2)

Empty by design. This folder will hold user-specified and synthetic fixtures that
the smoke harness auto-discovers to drive richer, per-command scenarios beyond the
v1 structural smoke (help renders, no panic, read-only verbs run).

Shape is deliberately undecided until v2 (see intent **itd-75**): likely one
subdirectory per command with an input corpus and an expected-shape assertion, so
adding a scenario is dropping a folder here — never editing Go.
