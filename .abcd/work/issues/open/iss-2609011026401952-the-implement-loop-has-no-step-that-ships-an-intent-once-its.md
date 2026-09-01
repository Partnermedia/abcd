---
schema_version: 1
id: "iss-2609011026401952"
slug: "the-implement-loop-has-no-step-that-ships-an-intent-once-its"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "manual-capture"
origin: researcher-authored
production_mode: hand-written
---

The implement loop has no step that ships an intent once its work merges: abcd intent plan moves drafts to planned and abcd spec close moves planned to shipped, but nothing in the build-review-merge routine invokes the second, so an intent whose code is on main sits in planned indefinitely and launch ship composes the changelog only from terminal folders. Two intents delivering a BREAKING CLI change were about to be released with no changelog line, and the omission is invisible because the cut exits 0 - a planned intent is not a refusal, it is simply not seen
