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
---

GHSA-mchq-gm34-3j34 (CWE-754, advisory severity low): `ahoy install` rebuilds an unparseable `.abcd/config.json` that `stepConfigValues` refused to touch. `internal/core/ahoy/apply.go:stepVersionStamp` discards `readConfig`'s error, treats the nil map as empty and republishes a meta-only file through `writeConfig`; `detect.go:detectVersion` discards the same error, so a malformed file raises `install_meta.missing`, which `--yes` arms. `stepConfigValues` refuses correctly (`loadPersistedInstallConfig` ok=false) but silently, so the run ends "partial" with `config.*_missing` remaining and no note naming the malformed file, and `stepVisibility` is skipped so the .gitignore fence is never reconciled. Reproduced at v0.7.0 with a merge-conflicted config: `repo.visibility` (the disclosure fence) and `scan.native_secret_scanning` (the remote-apply opt-out) were gone afterwards. The fix must establish that a config that cannot be parsed is never rewritten by any install step, that detection raises one non-resolvable `config.malformed` diagnostic instead of arming the stamp, and that a partial install says why in its notes — the regression test drives a merge-conflicted config through `install --yes` and asserts the file is byte-identical afterwards. Distinct from iss-127 (concurrent config RMW).
