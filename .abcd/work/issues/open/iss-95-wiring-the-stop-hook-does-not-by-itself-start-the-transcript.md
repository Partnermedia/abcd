---
schema_version: 1
id: "iss-95"
slug: "wiring-the-stop-hook-does-not-by-itself-start-the-transcript"
severity: "major"
category: "architectural-insight"
source: "manual-test"
found_during: "itd-89-m1"
found_at: "internal/surface/cli/cli.go"
---

Wiring the Stop hook does NOT by itself start the transcript clock: history.Capture requires ~/.abcd/history/<root-sha>/transcripts/ to already exist and deliberately never creates it (ownedDirsReal validates the store's dirs are real, not symlinks). That dir is bootstrapped by 'abcd ahoy install'. On a machine where install has not run — including this one, where ~/.abcd/ does not exist at all — 'hook session-end' fails closed, logs to stderr, exits 0, and captures NOTHING. Silently. That is precisely the failure mode itd-89 exists to prevent: a hook that appears wired while the corpus never accrues. Decide: (a) the hook bootstraps the store itself (changes Capture's stated precondition, and a hook creating dirs is a trust-boundary act the ownedDirsReal discipline deliberately avoids), or (b) 'ahoy install' stays the sanctioned bootstrap and the not-installed case is made LOUD rather than a stderr line nobody reads (e.g. ahoy doctor already flags history.bootstrap_missing as a required gap). Until this is settled, itd-89's acceptance is met in code but not on any machine that has not installed.