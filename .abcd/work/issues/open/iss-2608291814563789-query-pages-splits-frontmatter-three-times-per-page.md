---
schema_version: 1
id: "iss-2608291814563789"
slug: "query-pages-splits-frontmatter-three-times-per-page"
severity: "minor"
category: "tech-debt"
source: "impl-review"
found_during: "ultra-v0.6.8-followup"
found_at: "internal/core/memory/ask.go"
---

ultra-v0.6.8 C9: QueryPages in internal/core/memory/ask.go splits every page's frontmatter three times — pageInfoFrom, pageBody and pageSourceBlock each call splitFileFrontmatter (a full normaliseNewlines plus strings.Split of up to maxMemoryPageBytes), and two of them re-run parseFrontmatter on the same region. abcd memory ask does this over the whole store per question. Fix: split and parse once per page and thread (region, body) into the three consumers; measure before and after on a synthetic store.
