---
schema_version: 1
id: "iss-255"
slug: "ahoy-install-gitignore-block-hides-committed-tiers"
severity: "major"
category: "bug"
source: "agent-observation"
found_during: "v0.5.0 cycle ahoy install on abcd-cli"
found_at: "internal/core/ahoy"
---

ahoy install writes a .gitignore managed block that ignores /.abcd/ (and /memory/) wholesale — on any repo that commits the durable tiers, including abcd-cli itself, this hides every new record file from git status and makes git add refuse them, silently breaking the record workflow for every session in the clone. Observed live: install on abcd-cli appended the block; new intents/issues stopped appearing as untracked; the write was reverted by hand. Adjacent to iss-176 (config in an ignored path reported NOT ENFORCEABLE) but broader: the block contradicts the three-tier layout the installer itself scaffolds. Detector: ahoy install on a repo whose .abcd/development is tracked must not write an ignore rule covering it; acceptance: install on abcd-cli leaves git status of a new .abcd/work/issues file visible.