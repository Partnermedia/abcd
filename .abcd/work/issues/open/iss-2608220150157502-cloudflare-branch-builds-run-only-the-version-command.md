---
schema_version: 1
id: "iss-2608220150157502"
slug: "cloudflare-branch-builds-run-only-the-version-command"
severity: "minor"
category: "drift"
source: "user-observation"
found_during: "abcdev-site-plan investigation 2026-08-21"
found_at: "wrangler.jsonc"
---

Cloudflare non-production branch builds run only the version command, so branch previews repeat the production build instead of building the branch — observed and recorded as comments in wrangler.jsonc. The adr-48 deploy design replaces these with a labelled preview deployed from Actions on push to main and turns Cloudflare's automatic production builds off