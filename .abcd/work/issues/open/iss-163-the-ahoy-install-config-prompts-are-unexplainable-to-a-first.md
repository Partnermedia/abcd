---
schema_version: 1
id: "iss-163"
slug: "the-ahoy-install-config-prompts-are-unexplainable-to-a-first"
severity: "major"
category: "observation"
source: "user-observation"
found_during: "marketplace-install smoke test"
found_at: "internal/core/ahoy/ahoy.go"
---

The ahoy install config prompts are unexplainable to a first-time user: core's Prompter interface passes bare enum values only (Prompt(key, choices, def) — e.g. oracle_backend: host-delegated|native|cli|api|mcp) with no per-choice help text, no definition of what an 'oracle' is, and no consequence statement (cost, credentials, tools needed). A host agent rendering the prompt has to invent descriptions, which can be wrong or circular ('native — uses the native backend'). Canonical plain-language descriptions for every choice belong in core, surfaced by every front door.