---
schema_version: 1
id: "iss-202"
slug: "scanner-pii-config-unguarded-read"
severity: "critical"
category: "bug"
source: "agent-finding"
found_during: "bug-hunt loop round 9 (state issue #197), security + resource/error-handling hunt angles + independent adversarial verification"
found_at: "internal/adapter/scanner/scanner.go:103 (New, config read)"
---

scanner.New reads the per-repo .abcd/config/pii.json override with a bare os.ReadFile — no O_NOFOLLOW, no size cap, no regular-file check — unlike every sibling trust-boundary config reader in this codebase (guard.Load, rules.Load, banlist private store) which goes through fsutil.ReadGuarded; a FIFO planted at that path hangs the read forever and a symlink to a device file (git-committable, mode 120000, no local write access needed) grows the read unbounded toward OOM; reachable automatically via the SessionEnd hook's history capture, which also holds an unbounded history.repoLock flock across the hang, permanently wedging transcript capture for that repo; this is the same bug class already fixed once for ahoy.Detect under iss-97. Independently verified: FIFO hangs past 3s deadline (control: no-config-file case returns in 0.01s), /dev/zero symlink accumulates ~137 MiB/s, causal fix test (swap to fsutil.ReadGuarded) confirmed both cases resolve