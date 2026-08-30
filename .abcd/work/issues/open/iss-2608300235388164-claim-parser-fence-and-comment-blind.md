---
schema_version: 1
id: "iss-2608300235388164"
slug: "claim-parser-fence-and-comment-blind"
severity: "major"
category: "bug"
source: "impl-review"
found_during: "itd-177 adversarial reviews, 2026-08-30"
found_at: "internal/core/intent/claims.go (sectionLineRange, stampScopeConditions), internal/core/intent/ready.go, internal/core/intent/lifecycle.go (stampPlanned)"
---

The claim parser and stamper are fence-blind and comment-blind: a Scope Conditions heading inside a fenced code block shadows the real section (first match wins) so intent plan writes a marker into the fence and intent ready reports READY on the example while the real conditions go unchecked; a fenced example bullet under the real heading is stamped and counted; a commented-out block is read as live and the stamp's closing sequence ends the outer comment early; a hand-typed marker of the wrong shape is silent prose, gets a real marker glued beside it, and the bogus comment leaks into the JSON text; a second Scope Conditions heading is ignored without report; stampPlanned is a read-modify-write with no lock. Track fence state in sectionLineRange, refuse to stamp a section containing a fence or a duplicate heading, fault a malformed marker by name with a delete-it remedy.
