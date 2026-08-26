---
schema_version: 1
id: "iss-2608261133210491"
slug: "memory-storelock-wrong-ifmt-mask"
severity: "nitpick"
category: "tech-debt"
source: "agent-finding"
found_during: "bughunt-round-8"
found_at: "internal/core/memory/writer.go:70"
---

the memory store-lock guard tests mode AND S_IFREG nonzero instead of masking with S_IFMT, so its regular-file assertion also accepts symlink and socket modes; dead defence shielded by O_NOFOLLOW, fold into the iss-129 flock consolidation