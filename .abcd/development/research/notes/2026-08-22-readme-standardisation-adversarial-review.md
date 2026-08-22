# Adversarial review — a standardised README split for every managed repo

Dated 2026-08-22. A maintainer idea raised during the abcdev.app facilitation
session, put through a single adversarial leg at the maintainer's request —
not a full idea-admission gauntlet (no research or grill leg ran, so there is
no `abcd ideate record` verdict; this note is the durable record of the
review, and the graduation path runs through a full gauntlet later if
wanted). Ledger pointer: the `readme-standardisation-offered-never-imposed`
capture.

## The idea

Elevate the split abcd's own repository is about to make — README as a slim
contributor/facilitator page, the product narrative in `docs/` pages a
generated site renders for product thinkers — into a standing principle and
a standardised README structure for every abcd-managed repo. The maintainer
signal was explicit: open to completely restructuring README.md in
preparation, with abcd standardising the shape so the product thinker's
natural path is the web and the facilitator reads GitHub.

## The review

Six kill attempts; four landed as stated-idea-fatal, two partial. The fatal
hits, condensed:

1. **Audience colonisation.** "Facilitator" and "product thinker" are abcd's
   in-house surface taxonomy (the facilitator-default-thinker-optional
   principle scopes itself to abcd surfaces); a managed repo's README fronts
   that product's users and developers. Mandating abcd's audience model on
   every front page is the tool colonising the product.
2. **Contradicts the same-day reframing.** itd-140 rule 2 and adr-47
   decision 6 put the product-facing landing page on the specific side:
   opt-in, never a default. A structure-only retreat saves nothing — the
   product narrative is content, and the generalisation gauntlet verified
   the landing inputs exist only in this repository.
3. **Inflation past the shipped mechanism.** itd-102's positioning surfaces
   already hold a README to the repo's own declared identity, warn-tier,
   proposed-diff-only ("autonomous rewriting is permanently out"). A
   whole-README structural standard is the same idea inflated past its
   evidence and against its settlements; one-canonical-primitive routes any
   extension into the existing surface registry, not a second mechanism.
4. **Zero shipped instances.** abcd's own README→docs split is still
   planned plumbing (the site intents are drafts); itd-140 rule 3 forbids
   calling a pattern generic before a second, sparse instance demonstrates
   it. A zero-instance "standing principle for every repo" is the exact
   move the repo codified a gate against the day before.

Partial hits: the slim-README standard couples two opt-ins (a repo that
adopts the README shape but not the site evicts its product story to pages
nobody renders, losing GitHub above-the-fold conventions); and a universal
convention is a big-bang imposition with no enforcement mechanism
(ratchet-not-big-bang, enforcement-claims-are-facts) — though both admit a
compliant opt-in path.

## Verdict: reframed — offered, never imposed

The surviving shape:

1. The only cross-repo standard remains the **Identity surface registry** —
   a repo that wants its README held to structure registers additional
   surfaces in its own `.abcd/positioning.json`, checked warn-tier as
   proposed diffs per itd-102's settlements.
2. `prepare-this-repo`/scaffold may **propose a contributor-README template
   at onboarding** as a human-adopted diff (itd-93-style scaffold, per-repo
   editorial afterwards) — never a lint default.
3. The "website for product thinkers" stays **opt-in composition** per
   itd-140 and adr-47 decision 6, and no prose may call the split a
   convention until abcd's own instance has shipped **and** a second,
   distinct managed repo has demonstrably adopted it.
4. "Facilitator" and "product thinker" name abcd's own surface layers and
   never appear as the mandated audiences of a managed repo's front page.

## What this means for the session's record

Nothing already filed changes: abcd's own split stays recorded as this
repository's editorial choice (adr-47, brief `05-internals/10-site.md`),
and the two-instance gate is already itd-140's rule 3. The reframed offer
(scaffold template + registered surfaces) is a future-work seed held in the
ledger; graduating it to an intent or principle takes the full gauntlet,
with abcd's shipped instance as its first evidence.
