# Personas

Customer quotes use placeholder personas from `.abcd/development/personas.json` — Alice, Bob, Carol, Dave, Eve, Frank, Grace, Henry, Iris, Jack, Kira, Liam, Maya, Nia — each with role hints (solo founder, staff engineer, product manager, etc.); the registry is canonical, and the roster continues that sequence as it grows. Selection is by role, never by name: an intent's audience picks the role, and the role determines the registered name.

**Codified abcd principle:** never use real names in press releases (PII), never use generic "a hypothetical user" phrasing (loses voice). Named placeholders keep quotes grounded without leaking identifiers.

See [`04-surfaces/05-intent.md` § 4](../04-surfaces/05-intent.md#4-persona-registry) for the surface-level integration: `/abcd:intent "<text>"` picks a persona for the customer quote in each press release.
