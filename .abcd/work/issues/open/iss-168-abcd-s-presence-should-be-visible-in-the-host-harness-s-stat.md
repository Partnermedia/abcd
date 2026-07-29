---
schema_version: 1
id: "iss-168"
slug: "abcd-s-presence-should-be-visible-in-the-host-harness-s-stat"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "itd-105 grill session"
found_at: "commands/abcd"
---

abcd's presence should be visible in the host harness's status line / command line: a managed repo shows an 'abcd-managed' indicator (and possibly guard health) so the user can tell at a glance whether the current session is under abcd management without running a command. Needs a per-host adapter (e.g. a statusline hook) with the usual basics-built-in stance.