---
schema_version: 1
id: "iss-2608300258512411"
slug: "itd-183-second-round-residue"
severity: "minor"
category: "inconsistency"
source: "impl-review"
found_during: "itd-183 second-round reviews, 2026-08-30"
found_at: "internal/core/reading/assemble.go, .abcd/development/readings/README.md, internal/surface/cli/reading.go"
---

itd-183 second-round residue: --out and store checks compare lexically so a symlink into an admitted root passes refuseSelfAdmittingOutDir and a symlinked store directory passes os.Stat then enumerates nothing (Lstat and EvalSymlinks); the two artefact writes are each atomic but not jointly so, and a failed manifest write leaves a bundle that then makes the directory refuse forever; a walk row whose source directory is absent enumerates nothing silently (a brief chapter or glossary), while an empty bucket is legitimately absent and must stay a disclosed residue; the charter and projection comment say three fields travel today but two drafts already carry Scope Conditions and Mechanism so the entailment position projects five; a relative --out resolves against the repo root not the working directory and the render echoes the operator's string; the floor prose does not state that the heading signal is scoped to markdown files.
