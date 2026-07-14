---
schema_version: 1
id: "iss-97"
slug: "detect-unbounded-marker-reads"
severity: "minor"
category: "security"
source: "agent-finding"
found_during: "iss-95 adversarial security review"
found_at: "internal/core/ahoy/detect.go"
---

ahoy.Detect reads CLAUDE.md/AGENTS.md/config via plain os.ReadFile with no size cap or O_NONBLOCK, unlike the hardened session-end transcript read; a FIFO or multi-GB file at cwd could hang or slurp any hook that calls Detect (session-start, session-end, ahoy verbs). Harden Detect's marker/config reads to match maxTranscriptBytes plus a non-blocking open.