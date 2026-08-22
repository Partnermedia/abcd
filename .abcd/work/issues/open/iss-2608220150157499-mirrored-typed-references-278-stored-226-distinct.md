---
schema_version: 1
id: "iss-2608220150157499"
slug: "mirrored-typed-references-278-stored-226-distinct"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "abcdev-site-plan investigation 2026-08-21"
found_at: ".abcd/development"
---

The spec link is recorded from both ends (intent spec_id and spec implements, 30+ pairs) and 20 related_* pairs are listed in both files, so the record's 278 stored typed references are 226 distinct links. Harmless today, but every consumer that renders or counts links must collapse mirrored references or double-count; the site build is the first such consumer