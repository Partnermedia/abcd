---
schema_version: 1
id: "iss-96"
slug: "now-that-transcripts-are-captured-automatically-on-every-ses"
severity: "minor"
category: "security"
source: "manual-test"
found_during: "itd-89-m1"
found_at: "internal/adapter/scanner/patterns.go"
---

Now that transcripts are captured automatically on every session end, the scanner's secret-pattern coverage becomes load-bearing in a way it was not when capture was a manual verb nobody ran. Verified by live test: the bundled patterns DO catch anchored tokens (AKIA... access key IDs, ghp_/gho_/sk-ant- style prefixes) and absolute home paths, but they do NOT catch unanchored high-entropy values — an AWS SECRET access key (the 40-char base64 value, no prefix), a bare password, or a generic API token with no recognisable prefix all pass through into the store verbatim. This is the standard prefix-matching limitation and is pre-existing, not a regression; the point is that the blast radius changed. Consider entropy-based detection or the opt-in gitleaks adapter for the transcript path specifically, where the input is unstructured prose rather than curated source.