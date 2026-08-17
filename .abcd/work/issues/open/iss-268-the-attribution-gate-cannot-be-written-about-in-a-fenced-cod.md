---
schema_version: 1
id: "iss-268"
slug: "the-attribution-gate-cannot-be-written-about-in-a-fenced-cod"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "manual-capture"
---

The attribution gate cannot be written about in a fenced code block. GENERATED_RE is anchored to line start, so any line inside a markdown code fence that begins with the banned footer shape is refused exactly as a real footer would be -- and a fence is the natural way to document a shape. Pre-existing, not introduced by iss-262: the plain 'Generated with [tool](url)' form was already refused inside a fence; iss-262 extends that to the emphasised forms. Hit live on the iss-262 PR itself, whose first body quoted the footer in a fence and failed the gate it tightens. The current convention is to quote the shape mid-sentence instead, which the corpus pins with two accept cases, but that makes the one document most likely to show the shape -- the change that tightens the rule -- the one least able to show it. Options: (a) keep the convention and document it in AGENTS.md so the next author does not rediscover it through a red CI leg; (b) teach the gate fence state, which costs real complexity in a bash security check and opens an evasion route, since a footer wrapped in a fence would go unflagged while still rendering as visible text; (c) accept as-is. Needs a deliberate ruling rather than a same-PR patch.