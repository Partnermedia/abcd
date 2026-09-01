---
schema_version: 1
id: "iss-2609012039108437"
slug: "step-marker-plants-both-markers-when-config-is-unparseable"
severity: "minor"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/ahoy/apply.go"
---

Sibling of GHSA-mchq-gm34-3j34 found during its assessment: `stepMarker` and `detectMarkerDrift` discard `readConfig`'s error the same way `stepVersionStamp` and `detectVersion` do, and plant marker blocks ignoring the user's `docs.target`. `internal/core/ahoy/apply.go:stepMarker` falls back to `docsTargetDefault` ("both") when the persisted config is unreadable, and `detect.go:detectMarkerDrift` raises `marker.missing` for the default targets, which arms that step. In the merge-conflict reproduction at v0.7.0 `install --yes` wrote a marker block into BOTH CLAUDE.md and AGENTS.md although the repo had chosen one target — a write into two user-facing files the user never asked for. The fix must establish that a malformed config plants no marker and raises no marker gap; the same `config.malformed` diagnostic covers it, and `detectConfigValues` stops reporting three misleading `config.*_missing` gaps for a file that is present but unparseable. Handled correctly already, no change: `remote.go:nativeScanningOptedOut` (fail-closed), `vintage.go:recordedSetupVersion`, `attribution_hook.go:attributionOptedIn`, `apply.go:loadPersistedInstallConfig`.
