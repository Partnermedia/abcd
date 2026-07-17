---
schema_version: 1
id: "iss-107"
slug: "ahoy-install-visibility-v-yes-reports-already-up-to-date-wit"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "manual-capture"
---

ahoy install --visibility <v> --yes reports 'already up to date' without applying the explicitly requested visibility on an already-configured repo: skip-if-set beats an explicit flag — stepConfigValues short-circuits when repo.visibility is validly set (internal/core/ahoy/apply.go:196) and ValueOverrides are only consulted on the config-gap path (apply.go:179->220), so the flag silently no-ops; an explicit flag value should either apply as an update or error loudly, never silently skip (maintainer hit this from another repo and hand-edited its config)