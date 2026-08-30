---
schema_version: 1
id: "iss-2608300930057882"
slug: "intent-ready-grounds-json-omits-the-promised-members"
severity: "major"
category: "bug"
source: "impl-review"
found_during: "itd-179 adversarial ruthless review, 2026-08-30"
found_at: "commands/intent.md, internal/surface/cli/cli.go (intent ready --grounds), internal/core/intent/grounds.go (ParseGrounds, RecordGrounds)"
---

itd-179 ruthless findings: the plugin surface instructs the host to report a redacted member from intent ready --grounds --json that the verb never emits — the redaction note goes to stderr and the JSON is the unchanged readiness result, so the redaction-is-never-silent promise is broken on the one plane the feature is wired for (render an envelope carrying the grounds result beside the readiness result); the readiness reader admits a hand-typed entry below the writer's substance floor, so the seventh check under-enforces its own claim (skip sub-floor bullets in ParseGrounds); a successful write followed by a readiness fault exits 2 with no receipt, so a retry appends a duplicate entry; RecordGrounds enforces no bucket rule while the shipped-bucket exemption says never backfilled and the branch itself backfilled three shipped intents (refuse terminal buckets, or word the exemption honestly); missing versus malformed --grounds exit 2 versus 1 unevenly; the 36-record NOT READY consequence and the forward-only choice are recorded nowhere in the decision log.
