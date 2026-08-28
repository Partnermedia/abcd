---
schema_version: 1
id: "iss-259"
slug: "intent-render-paths-unsanitised"
severity: "minor"
category: "security"
source: "impl-review"
found_during: "spc-24 build, security-reviewer finding"
found_at: "internal/surface/cli/cli.go"
resolution: "the intent plan/link/create renders route .Path through termsafe.Sanitize, and the sweep covers every .Path render site"
impact: fix
resolved_by:
  commit: "ff60708f"
---

intent-store paths render to the terminal unsanitised in several intent verbs: intent plan prints res.Intent.Path and intent create prints the minted path without termsafe.Sanitize, and a hostile clone can commit a drafts/itd-N-<slug>.md whose filename tail carries control bytes (intentFileRe admits anything but a newline after the id), reaching the owner's terminal as escape sequences. The promote render sanitises its two paths as of spc-24; close the class once by sweeping the sibling intent renders through termsafe.Sanitize.