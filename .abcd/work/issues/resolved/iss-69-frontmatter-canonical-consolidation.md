---
schema_version: 1
id: "iss-69"
slug: "frontmatter-canonical-consolidation"
severity: "minor"
category: "tech-debt"
source: "agent-finding"
found_during: "clean-slate-sweep"
found_at: "internal/core/frontmatter/frontmatter.go"
resolution: "frontmatter delimiter trailing-whitespace bug fixed (C5); lint's duplicate scanner consolidated onto internal/core/frontmatter via a thin adapter (inherits the fix). seed1b (memory nested-YAML parser) deferred — a genuinely different primitive, not a behaviour-preserving refactor. Also flagged: intent.go's writer-path scanner is a future consolidation candidate. ruthless SHIP."
---

frontmatter canonical consolidation + delimiter bug: internal/core/lint (frontmatterFields) and internal/core/memory (parseFrontmatter/splitFileFrontmatter) each carry their own frontmatter scanner instead of internal/core/frontmatter — migrate both, behaviour-preserving, separate refactor commit (seed1); note memory needs nested-map values (source:) the flat scanner lacks — widen or scope. Bug: the canonical scanner closing-delimiter check compares the CR-trimmed line to exactly triple-dash, so a closing triple-dash with a trailing space/tab is not recognised and body lines leak as fields — TrimRight before both delimiter checks (frontmatter.go:37, C5). Corpus: the two lint/memory copies + the delimiter case.