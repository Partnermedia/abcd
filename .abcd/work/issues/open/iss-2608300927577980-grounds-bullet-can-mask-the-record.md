---
schema_version: 1
id: "iss-2608300927577980"
slug: "grounds-bullet-can-mask-the-record"
severity: "minor"
category: "bug"
source: "impl-review"
found_during: "itd-179 adversarial security review, 2026-08-30"
found_at: "internal/core/intent/grounds.go (RecordGrounds)"
---

RecordGrounds accepts text carrying an unclosed HTML comment opener, writes the bullet, and the comment mask then hides that entry and every line after it from the grounds reader and the claims readers, while the result reports success and the CLI says nothing; refuse before the write when the appended bullet does not raise the parsed entry count or leaves a comment open. Notes: the grounds grammar error echoes the raw operand before redaction (terminal only); RecordGrounds writes without the mint lock, like the existing in-place stampers.
