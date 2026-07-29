---
schema_version: 1
id: "iss-125"
slug: "stage1-redaction-misses-hostnames"
severity: "minor"
category: "security"
source: "agent-finding"
found_during: "2026-07-25 three-document SOTA/adversarial review"
found_at: "internal/adapter/scanner/patterns.go"
blocked_by: [iss-96]
resolution: "Stage-1 redaction strips LAN hostnames and device names via the scanner's canonical network set"
impact: fix
---

Hostnames are absent from the Stage-1 redaction pattern list: itd-28's write-time sanitiser enumerates home-dir paths and provider secrets (AWS/GCP/Azure/Cloudflare/GitHub/Anthropic/OpenAI/Stripe/Slack/JWT/PEM) but no hostname patterns, and the transcript scanner's coverage gaps are already tracked as iss-96 — yet the PII discipline forbids posting real hostnames anywhere. Live specimen: a saved third-party PR page (reviewed 2026-07-25) shows two real machine hostnames leaked into public PR comments by a resident review fleet, including a '.fritz.box' home-router LAN name — exactly the class a Stage-1 net must catch before any abcd-produced artefact (review receipt, rendered PR comment, transcript) leaves the machine. Fix belongs in the scanner seam's canonical pattern set (one-canonical-primitive), not a per-surface copy; detector first: add hostname fixtures (LAN suffixes like .local/.fritz.box/.lan, machine-name shapes) to the scanner's acceptance corpus and watch them flag before widening the patterns.