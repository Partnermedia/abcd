# Acknowledgements

abcd stands on ideas, tools, and writing from many sources. This file records them
in three parts: the **development** that built abcd, the **inspirations** that
shaped its design, and the **references** it draws on. Each entry is added in the
same change that lands what it records — the pull request that adopts a pattern,
cites a source in an ADR, or integrates a tool — so the list grows with the work
rather than being reconstructed later. Runtime dependencies are not listed here;
they live in `go.mod` and the licence notices they carry.

## Development

Development of abcd has been assisted by Claude Code (Anthropic). Per-commit
disclosure uses an `Assisted-by:` trailer; the human contributor is the author of
record and is responsible for all AI-assisted output — its correctness, licensing,
and fit for the project. See [`CONTRIBUTING.md`](CONTRIBUTING.md).

## Inspirations

Ideas and methodologies that shaped the design — not code abcd depends on.

- **Amazon "Working Backwards"** — the press-release format of abcd's intents.
- **Architecture Decision Records (MADR)** — the shape of the decision record.
- **Diátaxis** — the four-type model behind the user documentation.
- **Domain-Driven Design (bounded contexts)** — the surface boundaries.
- **The Linux kernel's coding-assistants policy** — the `Assisted-by:` attribution
  model abcd adopts for AI-assisted commits.

## References & sources

Books, articles, and papers cited in the design record.

- Eric Evans, *Domain-Driven Design* (Addison-Wesley, 2003).
- A. Mavin, P. Wilkinson, A. Harwood, M. Novak, "Easy Approach to Requirements
  Syntax (EARS)" (RE, 2009).
- U.S. Copyright Office, *Copyright and Artificial Intelligence, Part 2:
  Copyrightability* (2025).
