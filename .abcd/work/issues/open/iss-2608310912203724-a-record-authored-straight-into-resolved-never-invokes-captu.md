---
schema_version: 1
id: "iss-2608310912203724"
slug: "a-record-authored-straight-into-resolved-never-invokes-captu"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "itd-179-fidelity-audit"
origin: researcher-authored
production_mode: hand-written
found_at: ".abcd/work/issues"
---

a record authored straight into resolved never invokes capture resolve so the grounds obligation reaches only records that pass through the verb

Found by the itd-179 fidelity audit and verified at integration HEAD.

itd-179 makes `capture resolve|wontfix|promote` refuse without grounds. That
obligation reaches exactly the records that pass through those verbs.

Most do not. Measured at HEAD: **689 of 715 records in `resolved/` carry no
`## Grounds` section, and all 689 carry a `resolution:` frontmatter key** --
authored directly into the terminal folder rather than transitioned into it. The
auditor's scoped measurement is sharper: 95 records entered `resolved/` after
the refusal landed on this branch, and 69 of those carry no grounds.

The distinction that matters: **the gate is not bypassed, it is never invoked.**
A refusal on a verb cannot bind a workflow that does not call the verb, and
authoring a record straight into `resolved/` is a normal thing to do in this
repository -- the branch that shipped the obligation is itself the largest
producer of grounds-less terminal records.

Related but distinct from iss-2608301747006182, which observes that no GATE
requires a terminal record to carry grounds. This record is about why arming
that gate would be a bigger change than it looks: it would refuse the majority
of the existing corpus, and the population it would refuse was created by the
ordinary working practice of this repo.
