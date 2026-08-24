---
schema_version: 1
id: "iss-2608241740093740"
slug: "commands-launch-md-presents-the-release-day-site-deploy-as-a"
severity: "major"
category: "documentation"
source: "agent-finding"
found_during: "v0.6.4 docs-currency release gate 2026-08-24"
found_at: "commands/launch.md"
---

commands/launch.md presents the release-day site deploy as automatic when it never succeeds. The step table row 6 reads 'Deploy the website from the tag | automatic, after step 5' and the prose adds 'the site deploy is invoked by the release workflow, so approving step 5 releases step 6 with it'. release.yml's own comment records the opposite: a called workflow's job cannot resolve environment secrets, this repository holds no repository-scoped secrets, so the inherited set is empty and the deploy fails through the release chain (proven at v0.6.2 and v0.6.3) while succeeding through workflow_dispatch on site.yml. The always-taken path is demoted to a 'When it goes wrong' bullet, so an operator following steps 1-7 verbatim is told a step is automatic that has never worked. Same defect class as iss-2608231912566984, one surface away. Also in the same section: 'Three failures are worth recognising' is followed by four bullets, the site/deploy one added by b6602672 without updating the count.