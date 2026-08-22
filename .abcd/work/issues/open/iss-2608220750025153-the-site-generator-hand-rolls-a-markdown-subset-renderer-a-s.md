---
schema_version: 1
id: "iss-2608220750025153"
slug: "the-site-generator-hand-rolls-a-markdown-subset-renderer-a-s"
severity: "minor"
category: "future-work-seed"
source: "user-observation"
found_during: "agent-observation"
found_at: "internal/core/site"
---

the site generator hand-rolls a Markdown-subset renderer, a strict HTML tokenizer and a CSL formatter to stay dependency-free; adopting goldmark and x/net/html instead is a maintainer dependency decision