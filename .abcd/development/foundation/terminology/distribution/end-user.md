---
term: end-user
bounded_context: distribution
definition: A person who installs and runs the published abcd plugin from the public marketplace — the consumer of a release, distinct from a persona (a modelled archetype used in intents and briefs).
aliases: ["installer", "plugin consumer", "user"]
forbidden_synonyms: []
status: stable
introduced_in: itd-67
starts_when: null
ends_when: null
not_to_be_confused_with: core/persona
versions: null
---

# end-user (distribution)

An **end-user** is a person who installs and runs the published abcd plugin from
the public marketplace — the real consumer of a [release](release.md).

[persona](../core/persona.md) forbids "user" as a synonym *in the core context*,
because abcd's intents and briefs model archetypes (personas), not literal users.
That is correct for design work. But the distribution context genuinely has
literal users: the people who run `/plugin install abcd@abcd-marketplace`. When
discussing the install/update experience, "end-user" is the precise term; when
discussing a modelled archetype in an intent's press release, "persona" remains
canonical.

## When to use

Use "end-user" for the real person installing/updating/running the published
plugin — in the install docs, the update path, and the smoke-test framing.

## When NOT to use

Do not use "end-user" for a modelled archetype in an intent or brief — that is a
[persona](../core/persona.md).

## Examples

- "An end-user adds the marketplace and installs the plugin in two commands."
- "The update path lets an end-user pull the latest release with one command."

## Related terms

- [persona](../core/persona.md) — the modelled archetype used in intents/briefs
- [release](release.md) — what an end-user installs
