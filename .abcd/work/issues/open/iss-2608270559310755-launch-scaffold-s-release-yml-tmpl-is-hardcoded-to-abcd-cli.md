---
schema_version: 1
id: "iss-2608270559310755"
slug: "launch-scaffold-s-release-yml-tmpl-is-hardcoded-to-abcd-cli"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "testimony-launch-scaffold-2026-08-27"
found_at: "internal/core/launch/scaffold/templates/release.yml.tmpl"
---

launch scaffold's release.yml.tmpl is hardcoded to abcd-cli's OWN artifact shape, so it calls itself a 'generic bare-repo' workflow but is not generic: it runs 'make build' to cross-compile 'the four binaries', 'sha256sum abcd-* > checksums.txt', and 'gh release create <tag> bin/abcd-* bin/checksums.txt' as literals (only the verify/Run step is a <% %> substitution). A scaffolded repo that is not a Go CLI producing four abcd-* binaries via make build therefore gets a release workflow that builds/publishes the wrong (or no matching) assets, and its install path expects artifacts the release never produces. This is the confirmed 'asset gap' seen in a scaffolded non-abcd-cli repo (its install expected four tarballs + SHA256SUMS the workflow does not publish). Fix requires a product decision on how a non-abcd-cli repo declares/derives its build+artifact shape (parameterise the build command / artifact glob / checksum name via render.go substitutions with a repo-declared shape; or detect Go-CLI vs other; or scaffold only the generic release plumbing by default and make binary build/publish opt-in). Neighbour iss-2608261041218890 (release.yml tag not shape-checked before make build) is a different defect in the same template. Do NOT touch abcd-cli's own .github/workflows/release.yml.