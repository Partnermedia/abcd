---
schema_version: 1
id: "iss-2608231005025233"
slug: "site-dashboard-tiles-not-linked"
severity: "minor"
category: "ux"
source: "user-observation"
found_during: "user-observation"
found_at: "internal/core/site/explorer.go"
---

The dashboard's stat tiles are inert: a reader who sees '27 principles' or '525 issues' has no way through to what the number counts. Each tile should be a link to where that store is read — principles and disciplines to the foundations page, issues/intents/specs/decisions to their listing, releases to the release history — with a hover frame so the affordance is visible before the click. Generic-side under itd-140: the destination is derived from the record's own store, and a store with no page of its own renders an unlinked tile rather than a dead link (report C of the 2026-08-23 second pass).