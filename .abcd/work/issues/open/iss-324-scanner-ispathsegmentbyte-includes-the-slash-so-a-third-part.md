---
schema_version: 1
id: "iss-324"
slug: "scanner-ispathsegmentbyte-includes-the-slash-so-a-third-part"
severity: "minor"
category: "security"
source: "agent-finding"
found_during: "bughunt-round-1"
found_at: "internal/adapter/scanner/identity.go"
---

scanner isPathSegmentByte includes the slash so a third partys home path inside a file URL escapes stage-one redaction
## Evidence
`internal/adapter/scanner/identity.go:454` — `isPathSegmentByte` includes `'/'` in its class, so `leadingBoundaryOK` suppresses a `genericHomeRe` match preceded by a slash. Empirically `open file:///home/alice/notes.md` with a different caller identity → 0 findings (`home_path_other` missed); `see file:///Users/bob/secret.txt` → 0. A third party's home path inside a `file://` URL passes Stage-1 redaction unredacted — a privacy false NEGATIVE on the redaction path.

## Adversarial verdict: CONFIRMED (minor) — RECORD-ONLY
Surfaced as a spin-off of the iss-305 repolint fix (which deliberately uses repolint's slash-EXCLUDING isPathSegmentChar to avoid importing this FN). Not fixed this round: changing the scanner's isPathSegmentByte to exclude `/` has its own false-POSITIVE surface on the redaction path that needs independent measurement before it is safe — recording it is the correct outcome. Distinct from iss-305 (repolint over-detection); this is scanner under-detection.
