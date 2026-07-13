---
schema_version: 1
id: "iss-88"
slug: "managed-repo-label-overclaims"
severity: "minor"
category: "observation"
source: "agent-finding"
found_during: "2026-07-13 B1 dogfood: prepare-this-repo audit of Manuscripts"
found_at: "internal/core/ahoy/detect.go"
---

abcd ahoy classifies a repo with a partial or stray .abcd/ as folder_kind managed-repo even when adopted is null and index_registered is false -- managed reads as abcd manages this when it only means abcd could. No folder_kind value distinguishes a stray .abcd/ built by another workflow from an abcd-adopted repo. Observed on Manuscripts, whose .abcd/development/ was hand-built with zero abcd involvement yet is reported managed. Related: iss-62 managed-repo-identity-gate. Detector: a folder_kind vocabulary that separates adopted from merely-abcd-shaped; acceptance is ahoy --json on a hand-built .abcd/ not reporting managed.