---
schema_version: 1
id: "iss-270"
slug: "the-attribution-gate-still-cannot-be-documented-inside-a-lis"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "manual-capture"
---

The attribution gate still cannot be documented inside a list item. iss-268 lets a fenced code block quote the banned footer shape, but strip_fenced_blocks applies markdown's ABSOLUTE three-space indent limit for a fence opener. CommonMark measures that limit relative to the enclosing container, so a fence nested under a list bullet -- indented four or more spaces to align with the item's content -- is not recognised, is not stripped, and the gate refuses the example exactly as iss-268 described. This repo's pull-request bodies are heavily bulleted, so the shape is common: a bullet explaining a rule, with a fenced example under it. Also unhandled for the same reason: a fence inside a blockquote. Fixing it properly means tracking container context (list item indent, blockquote markers) rather than absolute columns, which is a redesign of the strip semantics rather than a patch -- a bash gate should not grow a markdown block parser. Options: (a) relax the opener indent limit to any depth, on the ground that four-space-indented content is itself a code block and therefore also quotation, so treating it as quoted matches how it renders; (b) track list/blockquote context properly; (c) accept and document the limitation. Note that (a) is a widening of a security gate and needs the same subset property iss-268 settled on: the stripped region must stay inside what markdown renders as code. Found by the ruthless review of the iss-268 fix.