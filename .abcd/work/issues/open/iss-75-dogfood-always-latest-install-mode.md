---
schema_version: 1
id: "iss-75"
slug: "dogfood-always-latest-install-mode"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "abcd-plugin-dogfooding"
---

abcd lacks an always-latest / dev install mode for dogfooding: 'abcd ahoy install' symlinks a pinned built binary and there is no plain bin/abcd, so tracking live development required a hand-rolled ~/.local/bin/abcd wrapper that runs 'go build -C <repo> && exec' on each call (rebuild-from-source, fail loudly on a broken build). This manual workaround means abcd's own install path does not cover the dogfooding case. Consider 'abcd ahoy install --dev' / a track-latest mode, or bless+document the wrapper as the sanctioned dogfood path. Recorded per the golden rule: a manual workaround must be captured, never silently bypassed, so abcd can be fixed.