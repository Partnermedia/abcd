---
schema_version: 1
id: "iss-337"
slug: "probe-go-s-inner-directory-descent-dirroot-openroot-name-is"
severity: "minor"
category: "tech-debt"
source: "agent-finding"
found_during: "bughunt-round-2"
found_at: "internal/core/lifeboat/probe.go"
---

probe.go's inner directory descent (dirRoot.OpenRoot(name)) is guarded only by the parent ReadDir's IsDir, so a directory swapped for a FIFO in that window blocks — same TOCTOU shape as readLifeboatFile but race-only and unfixable via the nonBlock const (os.Root.OpenRoot takes no flags); needs a raw openat restructure, recorded not fixed

## Evidence

- `internal/core/lifeboat/probe.go:406` -- `dirRoot.OpenRoot(name)` guarded only by the parent ReadDir's IsDir; `os.Root.OpenRoot` opens with O_NOFOLLOW|O_CLOEXEC, no O_DIRECTORY/O_NONBLOCK, so a directory swapped for a FIFO in that window blocks.

## Verifier verdict -- CONFIRMED, recorded not fixed

Real TOCTOU, same shape as iss-328, but race-only (needs an active writer in the probed tree) and unfixable with the nonBlock const because os.Root.OpenRoot takes no flags -- a fix means a raw openat(O_DIRECTORY|O_NOFOLLOW|O_NONBLOCK) or an OpenInRoot-style restructure. Filed separately from the iss-329/iss-328 same-sweep fixes by the verifier's own recommendation; larger than this round should bundle.
