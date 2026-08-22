# Artefacts

The product thinker and facilitator collaborate on artefacts that are jointly owned, with others generated and consumed autonomously by the AI-engineering team. Two of them — an initial *briefing* document and a set of articulated *intents* — are familiar territory for product thinkers; `abcd` builds on them, and adds a third, to *carry intent through to delivered reality*.

| | Product thinker | Facilitator |
|--|--|--|
| The brief — what is this project about? | bring the substance | shape it into the brief structure |
| Capturing an intent — *why does this change matter?* | write the press release | sharpen the acceptance criteria |
| Turning the why into engineering work | — | drive |
| Cross-cutting concerns the brief implies | — | derive and encode |
| Reading the verdict when work ships | read; decide what to do next | investigate any *not delivered* findings |

![The brief](../assets/img/artefact-brief.svg)

The **brief** *(owned jointly by the product thinker and the facilitator)* answers *what is this whole thing about?* — purpose, scope, the vocabulary the project uses, what "good" looks like. It makes one hard promise: **Everything it says reads true right now.** On day one, when the team has agreed a design but built nothing, most of the brief is ambition — and every ambitious passage is visibly marked as not yet real. As work ships, those markings come off one by one: the change that ships a capability also rewrites its passage in the brief to describe what actually exists (which is rarely word-for-word what was planned). The brief is never re-versioned and keeps no history — version control does that — so it earns its role as the project's living canvas one shipped change at a time.

![Intents](../assets/img/artefact-intents.svg)

**Intents** *(user-facing, and thus the product thinker's domain)* answer *why does each user-facing change matter?* Each is a one-page press release written as if the change had already shipped, with a named user feeling the difference, plus acceptance criteria in plain *Given / When / Then* language. Intents are how ambition travels into the brief: An intent is drafted, planned into engineering work, and built — and the same change that ships it updates the brief. Once its acceptance criteria are verifiably met, the intent is filed as shipped and becomes the permanent record of the *why*; the brief carries the *what is*; the engineering spec carries the *how*. Intents are individually portable: Each stands on its own and can be reordered, deferred, bundled, or dropped without rewriting the bigger picture.

![Automated audits](../assets/img/artefact-audits.svg)

**Automated audits** *(run by the AI-engineering team, read by you)* grade delivered reality against the original promise. When work lands, the intent audit reads each acceptance bullet against the actual repository and writes its verdict back onto the intent itself — so the *why* and the *did-we-deliver-it* live side by side, in one file, for as long as the project does.

Some things the project needs aren't user-facing — cross-cutting rules every feature must satisfy *(e.g., a privacy-hygiene lint, an accessibility checklist)*, or background plumbing that enables other capability. Those skip the press-release treatment — a cross-cutting rule is filed as a *discipline* rather than a press release, and plumbing goes straight into the brief — under the same promise: Real, or visibly marked as not yet real. As a product thinker you don't have to recognise or label them — that's your facilitator's job.
