---
schema_version: 1
id: "iss-27"
slug: "corpus-tooling-to-go-core"
severity: "minor"
category: "future-work-seed"
source: "user-observation"
found_during: "sources-ingest session 2026-07-08"
---

corpus tooling graduates into the Go core: the sources-corpus contract (per-source folders with location-as-classification, CSL-JSON metadata, append-only provenance ledger, banlist projection) is proven by the user-tier script MVP; absorb it as a native abcd sources/ingest verb family, converging with the itd-36 provenance substrate (internal/core/provenance: licence detection, citation generation, source-hash registry) rather than porting standalone. The scripts remain the reference implementation until the native verbs are wired; /abcd:consult and /abcd:ingest keep one surface across the swap.