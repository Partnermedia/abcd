---
schema_version: 1
id: "iss-2608270559313719"
slug: "abcd-launch-dry-run-and-ship-in-a-non-plugin-payload-repo-di"
severity: "minor"
category: "ux"
source: "user-observation"
found_during: "testimony-launch-dryrun-2026-08-27"
found_at: "internal/core/launch/includes.go"
---

abcd launch --dry-run (and ship) in a NON-plugin-payload repo dies with a raw 'include config not found: .abcd/config/launch-payload.json' (LoadIncludes preflight in internal/core/launch/includes.go), giving the operator no idea WHY. launch preview/ship is a plugin-payload-repo feature (it needs launch-payload.json, and ship additionally needs .claude-plugin/plugin.json); a repo that ships no plugin bundle legitimately has neither and should be told so, not handed a missing-file error. Fix (loud-staging/legibility): when the launch config is absent AND the repo is not a plugin-payload repo, the dry-run should explain 'launch preview/ship applies to plugin-payload repos; this repo's release path is launch scaffold + the CHANGELOG roll + auto-release' rather than reporting a raw missing include config. Surfaced from a Testimony (non-plugin repo) onboarding session.