---
schema_version: 1
id: "iss-2609012039102035"
slug: "ghsa-qc3w-origin-url-userinfo-persisted-and-echoed"
severity: "minor"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/ahoy/store.go"
resolution: "originURL scrubs the remote's userinfo through scrubRemoteUserinfo before the value enters RepoIdentity, so index.json, meta.json, ahoy --json, ahoy dry-run and ahoy doctor --json inherit a credential-free URL and registerRepo's refresh self-heals a legacy credentialed entry on the next install. TestScrubRemoteUserinfo pins the rule (a password under any scheme and a bare user under http(s) are dropped; an ssh login and a scp-like git@host stay), TestDeriveIdentityStripsRemoteUserinfo pins the derivation site, and TestInstallPersistsNoRemoteUserinfo drives a placeholder credential through a real install into an isolated home and finds it in neither registry file. Sibling readers: remote.go:resolveGitHubRepo shares originURL and now accepts a credentialed https origin it previously refused as not-github (benign); scanner identity.go keeps only the owner segment; no other package reads or persists a remote URL. File modes are unchanged: nothing secret is stored."
impact: fix
---

GHSA-qc3w-8pv5-crc3 (CWE-312, advisory severity low): the origin remote URL is stored verbatim, credentials included. `internal/core/ahoy/store.go:originURL` returns `git remote get-url origin` trimmed and `deriveIdentity` places it in `RepoIdentity.Github`; `apply.go:stepHistory` persists it into `~/.abcd/history/<sha>/meta.json` and `registerRepo` into `~/.abcd/history/index.json` (new files 0644 through `fsutil.WriteFileAtomicPreserveMode`), and `abcd ahoy --json`, `abcd ahoy dry-run` and `abcd ahoy doctor --json` echo it. A remote of the form `https://user:token@github.com/owner/repo.git` therefore lands a credential at rest in two files and on stdout. Reproduced at v0.7.0 with a placeholder token. The fix must scrub userinfo at the one derivation site (`originURL`) so nothing downstream ever sees it — index.json, meta.json and the three JSON surfaces inherit the scrub, and `registerRepo`'s refresh self-heals a legacy credentialed entry on the next install — proven by a test that installs into an isolated home with a placeholder credential and finds it in neither file. Sibling readers: `remote.go:resolveGitHubRepo` shares `originURL` and today refuses a credentialed URL as "not a github.com URL"; after the scrub it accepts it, which is benign. `internal/adapter/scanner/identity.go:ProbeIdentity` keeps only the owner segment; not a sink.

## Grounds

- pursued: the single derivation site is the choke point every sink reads, so scrubbing there makes a downstream leak impossible rather than merely unlikely; a credential surviving in either registry file after the install test would show the choice wrong
