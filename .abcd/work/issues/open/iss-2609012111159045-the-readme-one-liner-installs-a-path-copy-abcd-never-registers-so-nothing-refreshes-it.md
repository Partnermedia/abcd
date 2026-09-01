---
schema_version: 1
id: "iss-2609012111159045"
slug: "the-readme-one-liner-installs-a-path-copy-abcd-never-registers-so-nothing-refreshes-it"
severity: "major"
category: "drift"
source: "agent-finding"
found_during: "ship-audit-itd-130-itd-132-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "hooks/bootstrap.sh"
---

A PATH copy installed by the README one-liner (install -m 0755 into ~/.local/bin) writes no ~/.abcd/path-entry provenance record. The bootstrap refreshes only a registered copy that still matches its recorded hash, so a one-liner copy is never refreshed on plugin update; ahoy classifies it foreign; RefreshPathEntryDigest 'refreshes provenance, never creates it'; and abcd update can only vouch for it while its release manifest is still published. itd-132's press release promises 'the abcd command on PATH keeps working across any number of updates' as 'a regular file abcd owns and refreshes', but its ac-4 Given is ahoy install, so the one-liner route is outside what was delivered (audit receipt rcp-acde3e9ce729, diverged and missing items). Directions, none adopted: the one-liner registers the entry (a post-install 'abcd ahoy adopt', or the binary self-registering on first run of an unregistered regular-file PATH copy after proving it against the release manifest); or abcd update creates the record when it proves a file; or the docs stop promising refresh for that route. Part of the ease-of-update design work.
