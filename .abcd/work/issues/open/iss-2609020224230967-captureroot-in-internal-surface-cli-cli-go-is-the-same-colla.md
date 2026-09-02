---
schema_version: 1
id: "iss-2609020224230967"
slug: "captureroot-in-internal-surface-cli-cli-go-is-the-same-colla"
severity: "major"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/surface/cli/cli.go"
---

captureRoot in internal/surface/cli/cli.go is the same collapse the rules root just had: it asks git for the toplevel and falls back to cwd on any error, so a repository git will not answer for (an ownership refusal under the isolated env, git absent from the host's PATH, a corrupt .git) is treated as no repository at all. History capture run from a subdirectory then hands that subdirectory to scanner.New, which finds no redaction override at <sub>/.abcd/config/pii.json and redacts with defaults only, silently — the exact B12 failure the function was written to close, reappearing whenever git cannot answer. The sibling walk already exists as gitutil.RepoShapedRoot; capture.discoverRepoRoot and site.checkoutRoot both three-state this correctly already. Found during the rules-root fix round's sibling sweep and left unfixed there: this is the PII plane, with its own tests and its own fail-open/fail-closed choice to make.
