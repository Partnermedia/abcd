---
id: adr-21
slug: rebuild-in-go
status: accepted
date: 2026-07-06
supersedes: null
superseded_by: null
related_intents: []
related_rfcs: []
related_adrs: [adr-22, adr-23]
---

# ADR-21: Rebuild abcd as a Go binary

## Context

abcd is being rebuilt from scratch. Its implementation is a tree of Python
scripts (`scripts/abcd/` (historical), `ahoy.py`, the doctor and hook modules), driven
in-session and glued to the host through slash-commands and pre-commit hooks.
The design record — brief, intents, ADRs — describes that Python machinery in
detail, and much of it (the directory-as-truth lifecycle, the intent/spec/phase
layering, the redaction model) is architecture, not language. The rebuild has
to answer one question before any other: what does the new abcd run as?

The requirements the rebuild locks in point the same way. abcd must ship as a
single self-contained artifact an operator installs and runs (no interpreter to
provision, no virtualenv), it must present a deterministic core callable from
several front doors ([ADR-23](0023-transport-agnostic-core.md)), and it must
drop its bundled tools as hard dependencies
([ADR-22](0022-bundled-deps-as-pluggable-adapters.md)) — which removes the
Python-process coupling those tools relied on. A scripting runtime is a poor
fit for a distributable, statically-linked, concurrency-bearing CLI.

## Decision

We reimplement abcd as a **Go binary**. Go gives a single statically-linked
artifact per platform, a standard-library HTTP/JSON/process surface sufficient
for the adapters, first-class concurrency for parallel oracle and review runs,
and a fast test toolchain.

The record's Python mechanics are **port targets, not deletions**. Each
settled behaviour — the lifecycle store, the lints, the redaction stages, the
launch packaging — is re-expressed in Go against the same decisions; the ADRs
that describe them stay canonical for *what* the behaviour is, and this rebuild
changes only *how* it is realised.

## Alternatives Considered

- **Continue in Python.** Keeps every line already written and the in-session
  Python glue. Rejected: it cannot produce a single dependency-free artifact,
  fights the transport-agnostic core (an importable Python package is not a
  neutral core the way an internal Go package is), and the interpreter/venv
  provisioning is exactly the install friction the rebuild exists to remove.
- **Rewrite in another language (Rust, TypeScript).** Rust buys more than the
  problem needs and slows iteration; a TypeScript/Node binary reintroduces a
  runtime to provision. Rejected: Go sits at the fit point — compiled, single
  artifact, fast to write and test — without Rust's ceremony or Node's runtime.
- **Chosen: Go.** Single artifact, deterministic core, cheap concurrency, the
  distribution story the install/launch milestone needs.

## Consequences

- The rebuild is a reimplementation, not a refactor: the Python tree is
  replaced wholesale, and every ported behaviour needs a Go test that pins the
  same guarantee the Python code held.
- The toolchain shifts to Go conventions (`gofmt`/`goimports`, `go test`);
  Python-specific tooling and the interpreter-version pins leave the surface.
- Distribution becomes per-platform binaries, enabling the single-repo curated
  release ([ADR-28](0028-single-repo-curated-release.md)) to ship an artifact
  rather than a source tree.
- The design record's Python-implementation references now read as historical
  port sources; they are reconciled to the Go architecture as each behaviour is
  re-landed, not rewritten pre-emptively.
