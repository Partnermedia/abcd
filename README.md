<div align="center">

  <img src="docs/assets/img/logo.png" alt="abcd logo" width="150">

  <h1>Agent-Based Configuration for Development</h1>

  <p>An opinionated, intent-driven development framework for <a href="https://x.com/signulll/status/2030404483897815089">product thinkers</a>.</p>

  <img src="https://img.shields.io/badge/status-experimental-orange" alt="Status: experimental">
  <a href="https://github.com/REPPL/abcd-cli/releases"><img src="https://img.shields.io/github/v/release/REPPL/abcd-cli?cacheSeconds=300" alt="Release"></a>
  <img src="https://img.shields.io/github/last-commit/REPPL/abcd-cli?cacheSeconds=300" alt="Last commit">
  <br />
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white" alt="Go 1.25">
  <a href="https://claude.ai/claude-code"><img src="https://img.shields.io/badge/Built_with-Claude_Code-3B5CE7?logo=anthropic&logoColor=white" alt="Built with Claude Code"></a> <!-- docs-lint: allow -->
  <br />
  <img src="https://img.shields.io/badge/macOS-000000?logo=apple&logoColor=white" alt="macOS">
  <img src="https://img.shields.io/badge/Linux-core%20CI--tested-FCC624?logo=linux&logoColor=black" alt="Linux: core CI-tested">

</div>

---

<div align="center">
  <p><a href="#roles">Roles</a> — <a href="#process">Process</a> — <a href="#resources">Resources</a></p>
</div>

---

> *"This book has only one major purpose—to trigger the beginning of a new field of study: computer programming as a human activity"*
>
> —Gerald M. Weinberg, *The Psychology of Computer Programming*, 1971

AI coding is getting more powerful, yes one important group of people who most need that power — the domain experts, the [product thinkers](https://x.com/signulll/status/2030404483897815089), the people who actually know what should be built — are currently underserved to be able to use AI coding directly. `abcd`'s bet is that a small team of two roles can close that gap: A **product thinker** who holds the *why*, and a **facilitator** who translates the why into work that AI coding agents can act on.

<div align="center">
  <img src="docs/assets/img/intro.png"/>
</div>


# How `abcd` works

## Roles

`abcd` is being shaped by what real two-person teams discover as they use it. In the initial version, both roles — the product thinker and the facilitator — are human. In a later version, `abcd` aims to offer an automated facilitator, so a product thinker can run the framework with an AI translator alongside their agentic team of AI-engineers.

The product thinker and facilitator collaborate on artefacts that are sharedly owned, with others being autonomously generated and consumed by a team of AI-engineers. Two essential artefacts — an initial *briefing* document and set of articulated *intents* — are familiar teritory for product thinkers. `abcd` builds on those artefacts to *carry intent through to delivered reality*:

1. a **brief** that always-current as the project's living canvas *(sharedly owned by the product thinker and the facilitator)*
2. **press release** shaped intents *(user-facing and thus the domain of the product thinker)*
3. **automated reviews** that grades delivered reality against the original promise *(owned by the AI-engineering team)*

<div align="center">
  <img src="docs/assets/img/roles.png"/>
</div>

As a **product thinker**, you know who the user is. You know what *done* looks like when you see it. You know which trade-offs are acceptable and which would betray the point of the project. `abcd` is built around two moments where that judgement is decisive. First, at the start of a piece of work (when you set the *why* as an intent), and at the end (when you read the verdict on whether the *why* was delivered). What happens in between — turning your why into engineering work AI agents can act on — is the facilitator's job.

The **facilitator** is a *translator*, not an engineer-on-the-team in the traditional sense. Their work is to take what you wrote, shape it into plans an AI coding agent can execute well, run the framework's audit and review machinery, and tell you when the work didn't match the promise *(and what to do about it)*.

| | Product thinker | Facilitator |
|--|--|--|
| The brief — what is this project about? | bring the substance | shape it into the brief structure |
| Capturing an intent — *why does this change matter?* | write the press release | sharpen the acceptance criteria |
| Turning the why into engineering work | — | drive |
| Cross-cutting concerns the brief implies | — | derive and encode |
| Reading the verdict when work ships | read; decide what to do next | investigate any *not delivered* findings |

Contributing to the brief and writing intents-as-press-releases are the artefacts the product thinker writes.

1. The **brief** answers *what is this whole thing about?* — purpose, scope, the vocabulary the project uses, what "good" looks like. It's edited in place as the project evolves, never re-versioned, so it stays alive instead of going stale in a folder.
2. **Intents** answer *why does each user-facing change matter?* They're individually portable: Each one stands on its own, can be reordered, deferred, bundled, or dropped without rewriting the bigger picture.

Some things the project needs aren't user-facing. Often, these are cross-cutting rules every feature must satisfy *(e.g., a privacy review, an accessibility checklist)*, or background plumbing that enables other capability. As a product thinker, you don't have to recognise or label those. That's the work of your facilitator who derives them from the brief and handles them as part of the engineering work.

## Process

You sit down with your facilitator and whatever discovery material you have — recordings, notes, a shared workspace, a half-finished slide deck, a transcript of yesterday's stakeholder call. `abcd` has a skill that ingests that material and produces a plain-language draft of your project's brief. You read it together. The parts that feel fuzzy, you sharpen with a Socratic interview the framework provides. By the end of the session you have a brief that says — in language a stakeholder would recognise — what this project is about.

Once both of you have agreed on the brief, the facilitator begins to plan implementation while you continue to think of additional ideas and/or features. Capturing intents is as simple as typing:

```bash
/abcd:intent "<one-line idea>"
```

Each captured intent is a press release, written as if the change has already shipped, with a named user feeling the difference. Your facilitator handles the rest of the lifecycle — turning your *why* into engineering work, surfacing cross-cutting concerns, and running the fidelity reviewer when the work lands. You stay in the seat where your judgement matters most: Setting the why at the start, and reading the verdict at the end.

An idea is captured as a **press release** — written in present tense as if the change has already shipped, with a named user feeling the difference. Every intent declares **acceptance criteria** in plain "Given / When / Then" language — that's a hard gate, not a suggestion. Once you're ready to build it, your facilitator turns it into engineering work, and AI coding agents do the building. When the work lands, an automated reviewer reads each acceptance bullet against the actual shipped repository and writes the verdict back onto the intent itself.

```text
        ╭─────────────────────╮
        │  Half-formed idea   │
        ╰─────────────────────╯
                  │
                  ▼
        ┌─────────────────────┐
        │  Capture as a       │
        │  press release —    │
        │  what does the      │
        │  user feel after    │
        │  this ships?        │
        └─────────────────────┘
                  │
                  ▼
        ┌─────────────────────┐
        │  Add acceptance     │
        │  criteria — how     │
        │  will we tell, on   │
        │  the day, whether   │
        │  it was delivered?  │
        └─────────────────────┘
                  │
                  ▼
            ╱───────────╲
           ╱  Ready to   ╲ ── No ──┐
           ╲  build it?  ╱         │  refine, or grill
            ╲───────────╱          │  to stress-test
                  │ Yes            │
                  │ ◄──────────────┘
                  ▼
        ┌─────────────────────┐
        │  Facilitator turns  │
        │  the intent into    │
        │  engineering work;  │
        │  AI agents build it │
        └─────────────────────┘
                  │
                  ▼
        ┌─────────────────────┐
        │  Fidelity review    │
        │  reads each         │
        │  acceptance bullet  │
        │  against the actual │
        │  shipped repo       │
        └─────────────────────┘
                  │
                  ▼
        ┌─────────────────────┐
        │  Shipped — verdict  │
        │  written back onto  │
        │  the intent itself  │
        └─────────────────────┘
```

If something crosses your mind mid-flight that you don't want to lose — a half-formed observation, a question for the team, a doubt about the brief, a behaviour you'd expect a user to notice — abcd has a fast hatch for capturing it. You don't have to decide what kind of thing it is. Your facilitator triages those captures later.

Acceptance criteria use three words to describe a checkable outcome.

|  | What it pins down |
|------|-------------------|
| **Given** | The starting state — what's already true before anything happens. |
| **When** | The trigger — a single user or system action. |
| **Then** | The observable outcome — something a human (or the fidelity reviewer) can check by *looking at the result*, not by reading the author's intent. |

When the engineering work lands, `abcd` reads each acceptance bullet against the actual repository — the code, the configs, the tests, the docs — and grades them. The verdicts are written back onto the intent itself, in the same file as the press release. Your *why* and the *did-we-deliver-it* live in one place, side by side, for as long as the project does.

The reviewer is allowed to fail honestly. If a promise wasn't kept, it says so. If something was delivered but with a wrinkle worth your attention, it flags the wrinkle rather than glossing it. And if it genuinely couldn't tell from the repo, it says *that* — which is different from saying the promise wasn't met, and `abcd` insists on the distinction.


# Resources

## Install

One line, checksum-verified. It detects your OS/architecture, downloads the
binary and the `checksums.txt` manifest from the latest release, verifies the
binary's SHA-256 against the manifest (and refuses to install on any
mismatch — or if the manifest doesn't list the binary at all), then installs
to `/usr/local/bin`:

```sh
sh -c 'set -eu; cd "$(mktemp -d)"; os=$(uname -s | tr "[:upper:]" "[:lower:]"); arch=$(uname -m); case "$arch" in x86_64) arch=amd64;; aarch64) arch=arm64;; esac; b="abcd-$os-$arch"; curl -fsSLO "https://github.com/REPPL/abcd-cli/releases/latest/download/$b"; curl -fsSLO "https://github.com/REPPL/abcd-cli/releases/latest/download/checksums.txt"; grep " $b$" checksums.txt | if command -v sha256sum >/dev/null; then sha256sum -c -; else shasum -a 256 -c -; fi; sudo install -m 0755 "$b" /usr/local/bin/abcd; abcd version'
```

Prefer to inspect before running? The command is exactly what it says: two
downloads from [the latest release](https://github.com/REPPL/abcd-cli/releases/latest),
a checksum verification, and a `sudo install`. You can do the same by hand —
grab the binary for your platform plus `checksums.txt` from the releases
page, run `shasum -a 256 -c` (or `sha256sum -c`) against the matching line,
and copy the binary anywhere on your `PATH`. Every release is built and
published by CI from the exact tagged commit, with the checksums generated
over the same bytes that are uploaded.

## Build

```bash
make preflight   # build + vet + test + race (the pre-push gate)
go run ./cmd/abcd            # bare status board for the current directory
go run ./cmd/abcd version    # print the version
make build                   # cross-compile bin/abcd-<goos>-<arch>
```

## Layout

- [`cmd/abcd/`](cmd/abcd/) — CLI entry point.
- [`internal/`](internal/) — the engine (`core/`) and front doors (`surface/`);
  see [`internal/README.md`](internal/README.md).
- [`commands/`](commands/), [`.claude-plugin/`](.claude-plugin/) — the plugin
  surface (auto-loaded).
- [`.abcd/`](.abcd/) — the development record and working files (never shipped).

Contributor guidance: [`AGENTS.md`](AGENTS.md).
