---
schema_version: 1
id: "iss-132"
slug: "marker-scan-quoted-delimiter-doc-fps"
severity: "nitpick"
category: "tech-debt"
source: "impl-review"
found_during: "iss-111 fix (2026-07-24 run queue, burst 9)"
found_at: "internal/core/lifeboat/sources_conventions.go"
---

three documentation lines that literally quote a marker with its delimiter (TODO followed by colon) still self-match after the iss-111 fix and are irreducible under the issue's two bounded options; closing them needs a different mechanism, e.g. skipping fenced or inline-code marker examples in Markdown