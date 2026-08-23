---
schema_version: 1
id: "iss-2608231912566984"
slug: "the-release-deploy-declares-cloudflare-secrets-nothing-checks-exist"
severity: "major"
category: "drift"
source: "user-observation"
found_during: "v0.6.2 release 2026-08-23"
found_at: ".github/workflows/site.yml"
resolution: "release.yml now passes secrets: inherit to the site.yml call, so the callee's deploy job can resolve the site environment's Cloudflare credentials. The named-secrets alternative is structurally unavailable: a job calling a reusable workflow may not declare an environment, so the caller can never resolve one to pass on. Fixed in the workflow and in the scaffold template it is regenerated from, with the reasoning beside the zizmor suppression. The production deploy path is verified empirically by the 0.6.3 release this resolution enables — v0.6.2 was its first attempt and it failed there."
impact: fix
resolved_by:
  commit: "0eb69ac5"
---

The v0.6.2 release published its binaries, checksums and attestation, then
failed on its final job:

```
[ERROR] In a non-interactive environment, it's necessary to set a
CLOUDFLARE_API_TOKEN environment variable for wrangler to work.
```

**The secrets exist and are correctly scoped.** `CLOUDFLARE_API_TOKEN` and
`CLOUDFLARE_ACCOUNT_ID` are environment secrets on `site`, the preview pair is on
`site-preview`, and the repository holds no repository-scoped secrets at all. The
deploy job declares `environment: site`. Nothing about the configuration is
wrong.

The cause is that `release.yml` invokes `site.yml` as a REUSABLE workflow and
passes no secrets:

```yaml
uses: ./.github/workflows/site.yml
with:
  tag: ...
  mode: production
# no `secrets:` key
```

A called workflow receives no secrets unless the caller passes them explicitly or
declares `secrets: inherit`. Declaring `environment:` on the callee's job is NOT
sufficient: the environment still gates the job, but `secrets.*` resolves empty
inside it. So wrangler ran with no token.

## Why it went unnoticed

Two independent reasons, and both are the interesting part.

The preview path never exercises it. `site.yml`'s preview deploy is reached by a
push to the default branch, so it runs as its own workflow and its secrets
resolve normally. Four preview deploys succeeded the same afternoon, including
one from the release merge itself. The production path had never run before —
v0.6.2 was this repository's first production deploy — so the defect could not
have been observed earlier.

And every other job that reaches production through the same `workflow_call`
passes, because none of them references a secret. `site.yml`'s render job says so
in its own comment: "NO SECRETS live here, and none are referenced. The
environment is a GATE, not a ...". That comment is accurate about render and is
exactly why render's success proves nothing about deploy.

## The class

Third instance in one day of the same shape, after iss-2608231607594913 and
iss-2608220150157502: a claim about state that nothing local can check. Here the
workflow declares the secrets it needs in `${{ secrets.* }}`, and no gate
verifies that the reference will resolve. `abcd launch --dry-run` cannot see it
either; it reports the gates it can run locally and has no view of a repository's
Actions wiring.

The failure is therefore maximally late: after a human approval, after the
binaries are public, at the last step of the last job.

## Resolution

Fixed by adding `secrets: inherit` to the call, in `.github/workflows/release.yml`
and in the scaffold template it is regenerated from
(`internal/core/launch/scaffold/templates/release.yml.tmpl`), so a managed repo
receiving this machinery does not inherit the defect. Self-scaffold parity holds.

The narrower alternative — naming the two secrets explicitly in the `secrets:`
block — does NOT work here: the caller job in `release.yml` declares no
environment, so it cannot resolve them to pass on. `secrets: inherit` is the form
that functions.

Left open deliberately: nothing yet FAILS EARLY when a workflow references a
secret that will not resolve. This issue records the defect and its repair; the
detector that would have caught it before a release is not built, and a preflight
asserting the reference resolves before anything is published is the obvious next
rung.
