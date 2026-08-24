---
schema_version: 1
id: "iss-2608231912566984"
slug: "the-release-deploy-declares-cloudflare-secrets-nothing-checks-exist"
severity: "major"
category: "drift"
source: "user-observation"
found_during: "v0.6.2 release 2026-08-23"
found_at: ".github/workflows/site.yml"
---

The v0.6.2 release published its binaries, then failed on its final job, and
v0.6.3 failed the same way after a fix that did not fix it:

```
[ERROR] In a non-interactive environment, it's necessary to set a
CLOUDFLARE_API_TOKEN environment variable for wrangler to work.
```

## What is established, empirically

The same job succeeds or fails depending only on how the workflow was entered:

| entry | run | deploy |
| --- | --- | --- |
| `workflow_call` from release.yml | 32657507210 (v0.6.2) | FAILED |
| `workflow_call` from release.yml | 32674659520 (v0.6.3) | FAILED |
| `workflow_dispatch` on site.yml | 32675272207 | **SUCCEEDED** |

Identical job, identical `environment: site`, identical secrets, identical
wrangler invocation. The dispatch deployed abcdev.app at v0.6.3, serving 200.

So the credentials are correct and always were. `CLOUDFLARE_API_TOKEN` and
`CLOUDFLARE_ACCOUNT_ID` are environment secrets on `site`; the deploy job
declares `environment: site`; the environment's branch policy admits `main` and
`v*`, so the tag was never excluded. None of that is the fault.

**A called workflow's job does not resolve environment secrets**, even when it
declares the environment and the caller passes `secrets: inherit`. `inherit`
conveys the CALLER's secrets, and the calling job cannot declare an environment
(GitHub forbids `environment:` on a job that uses `uses:`), so it has only
repository and organization secrets to pass. This repository has none of either.
The inherited set is therefore empty, and `secrets.CLOUDFLARE_API_TOKEN`
resolves to nothing inside the callee.

## Two wrong diagnoses, recorded because the pattern matters more than the bug

Both were reasoned from configuration and both were asserted as fixes:

1. "The production secrets were never created." False: they existed, scoped to
   `site`, hours before the first failure. This nearly caused a Cloudflare API
   token to be created that was never needed.
2. "`secrets: inherit` is missing from the call." Necessary but not sufficient,
   and shipped in v0.6.3 as a completed fix. The 0.6.3 CHANGELOG entry titled
   "The release chain's site deploy receives its credentials" is therefore false
   in a published release.

The experiment that settled it — dispatch the same workflow directly — takes one
minute and was available from the first failure. Reasoning about workflow
semantics from the files is exactly the class of claim this repository keeps
finding to be unverifiable without running it.

## The recovery that works today

`gh workflow run site.yml` — the dispatch path, which `site.yml` documents as the
emergency redeploy and treats as production by definition (adr-48). It resolves
the newest published release when given no tag. This is documented in
`commands/launch.md`'s release-day section so the next person spends thirty
seconds rather than an evening.

## Directions, none taken

- Move the two Cloudflare secrets to repository scope so `secrets: inherit`
  actually conveys them. Simple, and it weakens the environment gating the
  `site` environment exists to provide.
- Stop calling `site.yml` from `release.yml`: trigger it on the tag push instead,
  so it runs as its own workflow and resolves its environment secrets normally.
  Costs the release chain its ordering guarantee between publish and deploy.
- Accept the dispatch as the production path and remove the call.

Each trades something real, and the choice wants a decision rather than a third
patch applied at midnight after two wrong ones.
