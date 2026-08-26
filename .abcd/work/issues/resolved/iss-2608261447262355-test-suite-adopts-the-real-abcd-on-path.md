---
schema_version: 1
id: "iss-2608261447262355"
slug: "test-suite-adopts-the-real-abcd-on-path"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "handover review, verified against origin/main after the name-guard hardening merged"
found_at: "internal/core/ahoy/store.go"
resolution: "effectiveBinTarget honours an explicitly set ABCD_BIN_TARGET over adoption, so a caller that names a sandbox target gets it. Adoption is unchanged when the variable is unset, which is production. Two tests pin both halves, the first watched failing against the old order."
impact: internal
---

The test suite can adopt and overwrite a developer's real abcd on PATH, because hermeticRepo does not scrub it and effectiveBinTarget prefers an owned PATH entry over an explicitly set ABCD_BIN_TARGET.

internal/surface/cli/cli_test.go hermeticRepo sets HOME, ABCD_PLUGIN_ROOT and ABCD_BIN_TARGET to temporary directories, and its own docstring says it "redirects HOME, the plugin root and the PATH symlink target to temp locations" — so sandboxing the install target is its stated contract. internal/core/ahoy/store.go effectiveBinTarget breaks that contract: it returns ownedPathEntry(pluginRoot) whenever one exists and only falls back to binTarget(), so the sandbox target is ignored. An ahoy install reached from such a test adopts the entry in place and rewrites it to point inside a test tempdir the test then deletes, so a plain go test ./... mutates state outside the sandbox.

PRECONDITION, measured rather than assumed, because the naive statement of this bug overclaims. Adoption needs an OWNED entry, and classifyBinTarget calls an entry owned only when it is a symlink into the plugin root (binTargetOwnedSymlink), a dev shim (binTargetDevShim), or a regular file whose provenance is recorded at ~/.abcd/path-entry AND whose bytes still hash to the recorded sha (binTargetOwnedCopy). A plain release binary with no provenance record classifies binTargetForeign and is never adopted. So the population at risk is a dev-shim install (`abcd ahoy install --dev`) or an ahoy-installed owned copy, not every machine with abcd on PATH.

Evidence for the precondition: on this machine the release-only account carries a real v0.6.6 binary at ~/.local/bin/abcd with no ~/.abcd/path-entry, so it classifies foreign; a peer session confirmed two full `go test ./...` runs on 2026-08-26 left it byte-identical, mtime unchanged. The same machine's development account runs a dev shim and is in the at-risk population. `abcd update` does not write the provenance record, so a release-only install does not drift into the at-risk set by updating; `abcd ahoy install` does.

Why it is worth more than the inconvenience: this repository is worked from two accounts on purpose, one running the freshly built binary and one running only the latest cut release so the released surface can be observed honestly. Converting the release-only account's binary into a dev shim would destroy that signal with no visible failure. The account is currently outside the at-risk set by classification alone, which is luck rather than design.

The partial fix shipped with the name-guard hardening covers hermeticEnv, which now scrubs abcd off PATH and uses Lstat so a dangling symlink is dropped too. hermeticRepo is deliberately not covered, because closing it is a design decision: should an explicitly set ABCD_BIN_TARGET suppress adoption of PATH entries outside it? That keeps dependency-gap realism, since other tools stay on PATH, while making install sandbox-safe, and it is inert in production where ABCD_BIN_TARGET is unset.

Detector: effectiveBinTarget returns the explicit ABCD_BIN_TARGET when one is set even though an owned entry is on PATH, and still returns the owned entry when it is unset. Both watched failing and passing respectively against the current adoption order.
