---
schema_version: 1
id: "iss-171"
slug: "the-path-install-detector-and-installer-assume-one-blessed-l"
severity: "major"
category: "observation"
source: "user-observation"
found_during: "second-repo install session"
found_at: "internal/core/ahoy/store.go"
---

The PATH-install detector and installer assume one blessed layout: binTarget defaults to /usr/local/bin/abcd (sudo territory abcd cannot assume) and symlink detection only recognises that exact target, so a working ~/.local/bin/abcd symlink install — the field-standard single-user location used by uv/pipx/rustup-class tools — reports symlink.missing, a false gap; the detector is itself running as abcd from PATH while reporting abcd not installed. Letting ahoy install 'fix' it would write a symlink to <plugin-root>/abcd without validating the target exists — a dangling symlink shadowing the working install. Redesign per fix-the-detector and no-sudo: detection scans PATH for abcd, resolves symlinks (EvalSymlinks, same seam as iss-170), classifies dev-shim/owned/foreign; install defaults to ~/.local/bin (create if needed), adopts an existing owned install in place, does system-wide dirs only behind an explicit --bin-dir flag failing loudly when unwritable, never escalates privileges, refuses to create a symlink whose target does not exist; '~/.local/bin not on PATH' becomes its own loud gap with the printed one-line fix (script-first: print the export line, do not auto-patch shell profiles as the first rung).