---
schema_version: 1
id: "iss-2609012039102703"
slug: "ghsa-mchq-version-stamp-rebuilds-unparseable-config"
severity: "minor"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/ahoy/apply.go"
resolution: "No install step rewrites a config.json it cannot parse: stepVersionStamp and stepConfigValues keep readConfig's error and refuse through one shared note (refuseMalformedConfig, given once per run) naming the file, the parse error and the repair; detection raises a single required, non-resolvable config.malformed diagnostic (detectConfigIntegrity) instead of install_meta.missing or the three config.*_missing gaps, so nothing arms a rewrite under --yes, and a run that refused its config reports partial rather than clean. TestInstallNeverRebuildsMalformedConfig drives the advisory's merge-conflicted config through install --yes and asserts the file is byte-identical, unlisted in Writes, named in Notes, with status partial; TestDetectMalformedConfigIsOneDiagnosticGap pins the detection side. The stepMarker/detectMarkerDrift sibling is fixed under its own record (iss-2609012039108437). Readers already fail-safe and unchanged: nativeScanningOptedOut, recordedSetupVersion, attributionOptedIn, loadPersistedInstallConfig."
impact: fix
---

GHSA-mchq-gm34-3j34 (CWE-754, advisory severity low): `ahoy install` rebuilds an unparseable `.abcd/config.json` that `stepConfigValues` refused to touch. `internal/core/ahoy/apply.go:stepVersionStamp` discards `readConfig`'s error, treats the nil map as empty and republishes a meta-only file through `writeConfig`; `detect.go:detectVersion` discards the same error, so a malformed file raises `install_meta.missing`, which `--yes` arms. `stepConfigValues` refuses correctly (`loadPersistedInstallConfig` ok=false) but silently, so the run ends "partial" with `config.*_missing` remaining and no note naming the malformed file, and `stepVisibility` is skipped so the .gitignore fence is never reconciled. Reproduced at v0.7.0 with a merge-conflicted config: `repo.visibility` (the disclosure fence) and `scan.native_secret_scanning` (the remote-apply opt-out) were gone afterwards. The fix must establish that a config that cannot be parsed is never rewritten by any install step, that detection raises one non-resolvable `config.malformed` diagnostic instead of arming the stamp, and that a partial install says why in its notes — the regression test drives a merge-conflicted config through `install --yes` and asserts the file is byte-identical afterwards. Distinct from iss-127 (concurrent config RMW).

## Grounds

- pursued: the file is the user's data and the parse error is the signal, so every reader fails safe on it and one diagnostic replaces the gaps that armed the rewrite; a config that survives install --yes byte-identical is the proof, and any step that still rewrote it would show the posture incomplete
