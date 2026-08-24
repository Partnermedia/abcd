---
schema_version: 1
id: "iss-2608242043243131"
slug: "the-preflight-gate-list-is-prose-no-test-derives-from-the-re"
severity: "major"
category: "process"
source: "agent-finding"
found_during: "v0.6.5 docs-currency release gate 2026-08-24"
found_at: "Makefile"
---

the preflight gate list is prose no test derives from the recipe, so it drifts every time the recipe changes. Makefile's preflight target names its prerequisites once; docs/how-to/install.md, CONTRIBUTING.md and the Makefile's own comment each restate that list by hand. The restatements have now drifted twice in two releases: v0.6.4 corrected three surfaces that said three gates after lint-issues made it four (2852c095 updated AGENTS.md alone), and v0.6.5 re-introduced the same defect by adding site-render as a fifth prerequisite and leaving those surfaces saying four. Both times only the host-run docs-currency reviewer noticed, and both times it refused a release cut over it. The class is a derived-value restatement with no deriving test: the same shape as the marketplace slug, which TestInstallGuideDocumentsTheInstallAndUpdatePath solved by reading go.mod and asserting the documented string contains it. The detector is the same move here — parse preflight's prerequisites out of the Makefile and assert each enumerating surface names exactly that set. Acceptance corpus: the two instances above, both of which the test must flag when replayed against their pre-fix trees.