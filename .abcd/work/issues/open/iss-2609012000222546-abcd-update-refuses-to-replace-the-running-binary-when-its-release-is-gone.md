---
schema_version: 1
id: "iss-2609012000222546"
slug: "abcd-update-refuses-to-replace-the-running-binary-when-its-release-is-gone"
severity: "major"
category: "process"
source: "user-observation"
found_during: "abcd-update-invocation-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/update/update.go"
---

abcd update refuses to replace the very binary that is running it when that binary's release no longer exists on the forge, and its remedy sends the user to remove the file, which removes the only tool that could reinstall. Observed 2026-09-01: ~/.local/bin/abcd was v0.6.6 (byte-identical to the plugin-cache binary the bootstrap provisioned); abcd update resolved v0.7.0, then refused with shape unprovenanced-file because the file's digest appears in no published checksums.txt. The 2026-08-30 decision deleted every release object older than v0.6.9 and notes that v0.6.0-v0.6.6 never had one, and it claims removing old assets affects only a deliberate pin; that is wrong for abcd update, whose ownership proof (spc-32, the OWNED BY PROVENANCE row) walks exactly those deleted manifests, so every v0.6.7, v0.6.8 and plugin-provisioned install now refuses the same way. The remedy text ('remove it and reinstall') names no reinstall command, and once the file is removed the update verb cannot run at all; the user was left with no abcd and recovered only by the README one-liner. Two directions, neither adopted: (1) when the target resolves to the running executable itself (os.Executable after symlink resolution, os.SameFile), the file is abcd by construction and cannot be a foreign binary, so provenance should only derive the old version, reporting it as an unpublished build with its digest and proceeding after the existing TTY confirmation; (2) at minimum the refusal remedy carries the reinstall one-liner or the docs path and states that update cannot run once the file is gone. Part of the ease-of-update design work: users must be able to move the CLI and the plugin route forward without a dead end.
