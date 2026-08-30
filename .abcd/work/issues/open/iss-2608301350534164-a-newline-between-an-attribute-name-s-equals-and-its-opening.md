---
schema_version: 1
id: "iss-2608301350534164"
slug: "a-newline-between-an-attribute-name-s-equals-and-its-opening"
severity: "major"
category: "security"
source: "user-observation"
found_during: "itd-183-round-10-security"
found_at: "internal/core/reading/project.go"
---

a newline between an attribute name's equals and its opening quote declines the markup mask on both readings so a value carrying a greater-than truncates the heading opener

Found by the round-10 adversarial security review of build/itd-183. NOT
introduced by this branch -- present on the round-9 parent too, and recorded
rather than chased under the ruling that round 10 closes only what the branch
introduced.

`maskMarkupData` finds an attribute value by skipping blanks after the `=`, and
a blank is space, tab, carriage return or form feed -- never a newline, because
the same helper reads YAML lines where a newline ends the line. So a tag written
across the break is not recognised as carrying a value at all, on EITHER
masking, and the unmasked opener again stops at the `>` inside the value:

```
<h2 title=
"a>b">Audit Notes</h2>
```

Every renderer reads that as the excluded heading. Verified leaking on this
round and on the parent.

This is the same root cause as iss-2608301350527962 -- the opener parse depends
on the mask, and a declined mask leaves it mis-parsed -- reached by a different
route. Remedy: an HTML-whitespace skip local to the markup mask, which a YAML
line scan has no business sharing. Seed material for the exclusion floor's own
intent.
