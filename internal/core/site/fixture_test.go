package site

// A miniature managed repository, built in-process, that exercises every path
// the landing page and record.json take: all four layouts, both asset kinds, a
// shipped intent with a MET audit, a dangling supersession, a two-release
// changelog, and a history with dated commits and attribution trailers.
//
// It is built rather than committed because half of what the build reads is git
// history, and a fixture whose history is a fixture is the only way to pin
// dates. Every commit carries a fixed date and a fixed author, so the whole
// export is a function of this file.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Partnermedia/abcd/internal/gittest"
)

// onePixelPNG is a 1×1 PNG. The build reads its IHDR for the width and height it
// puts on the rendered <img>, so the bytes have to be a real PNG.
var onePixelPNG = []byte{
	0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 'I', 'H', 'D', 'R',
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89,
	0x00, 0x00, 0x00, 0x0a, 'I', 'D', 'A', 'T',
	0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00, 0x05, 0x00, 0x01,
	0x0d, 0x0a, 0x2d, 0xb4,
	0x00, 0x00, 0x00, 0x00, 'I', 'E', 'N', 'D', 0xae, 0x42, 0x60, 0x82,
}

// fixture is the built repository.
type fixture struct {
	t    *testing.T
	repo *gittest.Repo
	root string
}

// newFixture builds the whole repository and returns it at HEAD.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	r := gittest.NewRepo(t)
	f := &fixture{t: t, repo: r, root: r.Root()}
	f.writeSources()
	f.commitAt("2026-01-05T09:00:00+00:00", "feat: the record and the pages", "")
	f.shipTheIntent()
	f.writeChangelog()
	f.commitAt("2026-03-02T09:00:00+00:00", "docs: two releases", "None")
	f.mergeInARecord()
	return f
}

// mergeInARecord gives the fixture the history shape a linear one cannot have:
// a record whose content enters the trunk ONLY as part of a merge commit, the
// way a conflict resolution or a rename settled while merging does.
//
// It is here because a strictly linear fixture cannot see the defect it guards.
// `git log --name-status` prints no file lines for a merge, so such a record is
// invisible to the whole history walk and comes out with no dates at all —
// which then seats it at the centre of a chronological chart.
func (f *fixture) mergeInARecord() {
	f.t.Helper()
	const date = "2026-03-05T09:00:00+00:00"
	trunk := f.gitOut("rev-parse", "--abbrev-ref", "HEAD")

	f.git(date, "checkout", "-b", "side")
	f.write("docs/explanation/aside.md", "# An aside\n\nWritten on a branch.\n")
	f.commitAt(date, "docs: an aside on a branch", "None")

	f.git(date, "checkout", trunk)
	// Merge without committing, then add the record as part of the resolution,
	// so its only appearance in history is inside the merge commit itself.
	f.git(date, "merge", "--no-ff", "--no-commit", "side")
	f.write(".abcd/development/decisions/adrs/0003-settled-while-merging.md", `---
id: adr-3
slug: settled-while-merging
status: accepted
date: 2026-03-05
supersedes: null
superseded_by: null
related_intents: []
related_rfcs: []
related_adrs: []
---

# ADR-3: Settled while merging

This decision's file entered the trunk as part of a merge commit.
`)
	f.commitAt(date, "merge: side", "None")
}

// gitOut runs one git command and returns its trimmed stdout.
func (f *fixture) gitOut(args ...string) string {
	f.t.Helper()
	full := append([]string{"-C", f.root}, args...)
	cmd := exec.Command("git", full...)
	cmd.Env = f.repo.Env()
	out, err := cmd.Output()
	if err != nil {
		f.t.Fatalf("git %v: %v", args, err)
	}
	return strings.TrimSpace(string(out))
}

// Root is the repository root.
func (f *fixture) Root() string { return f.root }

// write puts one file in the tree.
func (f *fixture) write(rel, body string) {
	f.t.Helper()
	abs := filepath.Join(f.root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fixture) writeBytes(rel string, body []byte) {
	f.t.Helper()
	abs := filepath.Join(f.root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(abs, body, 0o644); err != nil {
		f.t.Fatal(err)
	}
}

// git runs one git command at a fixed date, so the history the build reads is
// the same on every run.
func (f *fixture) git(date string, args ...string) {
	f.t.Helper()
	env := append(append([]string{}, f.repo.Env()...),
		"GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date)
	full := append([]string{
		"-C", f.root,
		"-c", "user.email=fixture@example.invalid",
		"-c", "user.name=Fixture",
		"-c", "commit.gpgsign=false",
	}, args...)
	cmd := exec.Command("git", full...)
	cmd.Env = env
	if out, err := cmd.CombinedOutput(); err != nil {
		f.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// commitAt stages everything and commits it, optionally with an attribution
// trailer.
func (f *fixture) commitAt(date, subject, assistedBy string) {
	f.t.Helper()
	msg := subject
	if assistedBy != "" {
		msg += "\n\nAssisted-by: " + assistedBy
	}
	f.git(date, "add", "-A")
	f.git(date, "commit", "--allow-empty", "-m", msg)
}

// shipTheIntent moves an intent from planned/ into shipped/ in its own commit,
// which is how the record says a promise was delivered — and the only way the
// build can date it.
func (f *fixture) shipTheIntent() {
	f.t.Helper()
	if err := os.MkdirAll(filepath.Join(f.root, ".abcd", "development", "intents", "shipped"), 0o755); err != nil {
		f.t.Fatal(err)
	}
	f.git("2026-02-10T09:00:00+00:00",
		"mv",
		".abcd/development/intents/planned/itd-2-the-shipped-one.md",
		".abcd/development/intents/shipped/itd-2-the-shipped-one.md")
	f.commitAt("2026-02-10T09:00:00+00:00", "feat: ship itd-2", "Assistant:assistant-model-1")
}

func (f *fixture) writeChangelog() {
	f.write("CHANGELOG.md", strings.Join([]string{
		"# Changelog",
		"",
		"## [Unreleased]",
		"",
		"## [0.2.0] - 2026-02-11",
		"",
		"### Added",
		"",
		"- The shipped one. (itd-2)",
		"",
		"## [0.1.0] - 2026-01-06",
		"",
		"### Added",
		"",
		"- The first cut.",
		"",
	}, "\n"))
}

// writeSources writes every committed input the build reads.
func (f *fixture) writeSources() {
	f.write(".claude-plugin/plugin.json", `{
  "name": "fixture",
  "description": "A fixture package.",
  "author": {"name": "Fixture Author"},
  "homepage": "https://example.invalid/fixture/repo",
  "repository": "https://example.invalid/fixture/repo",
  "license": "MIT"
}
`)

	f.write(".abcd/record-lint.json", `{
  "roots": [".abcd/development"],
  "banned_tokens": [],
  "rules": {
    "record_schema": {"enabled": true, "severity": "blocker", "record_stores": {
      "adr": ".abcd/development/decisions/adrs",
      "itd": ".abcd/development/intents",
      "spc": ".abcd/development/specs",
      "iss": ".abcd/work/issues"
    }}
  }
}
`)

	f.write(".abcd/site.json", `{
  "schema_version": 1,
  "purpose": "Composition manifest for the fixture.",
  "identity": {"file": ".abcd/development/brief/01-product/README.md", "heading": "Identity (canonical)"},
  "ui_strings": "site-src/ui.json",
  "home": {
    "hero": {"page": "docs/explanation/rationale.md", "figure": "first-image"},
    "chapters": [
      {"letter": "a", "page": "docs/explanation/roles.md", "layout": "cards-from-h2"},
      {"letter": "b", "page": "docs/explanation/artefacts.md", "layout": "lead-in-cards",
       "feature": {"kind": "shipped-intent-press-release", "pick": "newest-with-audit-MET",
                   "parts": ["press-release", "first-acceptance-criterion"]},
       "icons": "image-before-lead-in", "table_portraits": "roles"},
      {"letter": "c", "page": "docs/explanation/process.md", "layout": "prose",
       "figure": {"kind": "first-image", "labels-from-page": true}},
      {"letter": "d", "page": "docs/how-to/install.md", "layout": "install",
       "lead": "CLI", "after": "Afterwards", "left": ["Plugin"],
       "release": {"from": "CHANGELOG.md", "assets": "release.yml"},
       "tabs": "left-h2s, then lead-h3s and remaining-h2s as a labelled group"}
    ]
  },
  "record": {"issue_ledger": true},
  "docs": {"index": "docs/README.md", "cli": "docs/reference/cli/commands.md"},
  "record_pages": {"contributors": {"policy": {"file": "CONTRIBUTING.md", "heading": "Attribution", "part": "first-bullet"}}},
  "checks": {
    "every_text_node_has_source": true,
    "docs_lint_on_rendered_text": true,
    "command_snippets_pinned_to_cli_reference": true,
    "unresolved_reference_baseline": ".abcd/site-baseline.json"
  }
}
`)

	f.write("site-src/ui.json", `{
  "_purpose": "The complete list of words the fixture generator may add.",
  "nav_story": "Story",
  "nav_install": "Install",
  "nav_docs": "Docs",
  "nav_record": "Record",
  "nav_references": "References",
  "cta_roles": "Roles",
  "cta_install": "Install",
  "cta_docs": "Documentation",
  "record_link": "built in the open",
  "from_the_record": "from the record",
  "latest_release": "Latest release",
  "all_releases": "all releases",
  "copy": "copy",
  "copied": "copied",
  "search_docs": "Search the docs",
  "platform": {
    "darwin-arm64": "macOS · Apple silicon",
    "darwin-amd64": "macOS · Intel",
    "linux-arm64": "Linux · arm64",
    "linux-amd64": "Linux · amd64"
  },
  "tiles": {"releases": "releases", "adr": "decisions", "intent": "intents", "issue": "issues", "commits": "commits"},
  "more": "more",
  "standby": "Stand by…",
  "cli_group": "CLI",
  "matches_system": "matches this computer",
  "beta": "Beta"
}
`)

	f.write("site-src/redirects", "/old/  /docs/old/  301\n")
	f.write("site-src/headers", "/install.sh\n  Content-Type: text/plain; charset=utf-8\n")
	f.write("site-src/site.css", ":root{--ink:#000}\n")
	f.write("site-src/site.js", "/* fixture */\n")

	f.write(".abcd/site-baseline.json", `{
  "schema_version": 1,
  "unresolved_references": [
    {"from": "adr-2", "to": "adr-9"}
  ]
}
`)

	f.write(".abcd/development/brief/01-product/README.md", `# Product

## Identity (canonical)

- **Title:** fixture — A Fixture Product
- **Tagline:** A fixture tagline for a fixture product.
- **Pitch:** A fixture pitch, two clauses long, so the hero has something to
  carry in its aside.
`)

	// --- assets -------------------------------------------------------------
	f.writeBytes("docs/assets/img/intro.png", onePixelPNG)
	f.writeBytes("docs/assets/img/logo.png", onePixelPNG)
	f.writeBytes("docs/assets/img/role-thinker.png", onePixelPNG)
	f.writeBytes("docs/assets/img/role-facilitator.png", onePixelPNG)
	f.write("docs/assets/img/artefact-brief.svg",
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10" width="10" height="10"><rect width="10" height="10" fill="var(--ink, #000)"/></svg>`)
	f.write("docs/assets/img/loop.svg",
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10" width="10" height="10"><circle cx="5" cy="5" r="4" fill="var(--accent, #06c)"/></svg>`)

	// --- composed pages -----------------------------------------------------
	f.write("docs/explanation/rationale.md", `# Who this is for

A fixture lede with a [link](https://example.invalid/), some *emphasis*, some
**strength**, and a `+"`code span`"+`.

![A fixture cartoon](../assets/img/intro.png)

A second paragraph that the hero does not use.
`)

	f.write("docs/explanation/roles.md", `# Roles

The fixture has two roles, and this paragraph introduces them.

## Product thinker

![The product thinker](../assets/img/role-thinker.png)

The one who holds the *why*.

## Technical facilitator

![The facilitator](../assets/img/role-facilitator.png)

The one who translates it.
`)

	f.write("docs/explanation/artefacts.md", `# Artefacts

A fixture opening paragraph about the artefacts.

| | Product thinker | Facilitator |
|--|--|--|
| The brief | bring the substance | shape it |
| An intent | write the press release | sharpen the criteria |

![The brief](../assets/img/artefact-brief.svg)

The **brief** *(owned jointly)* answers what this is about.

**Intents** answer why each change matters.

A closing paragraph that is not a card.
`)

	f.write("docs/explanation/process.md", `# Process

![The loop](../assets/img/loop.svg)

**It starts with the brief.** A fixture paragraph opening the process.

## Capturing intents

One line, deliberately shaped like a capture:

`+"```sh"+`
fixture intent "<one-line idea>"
`+"```"+`

| | What it pins down |
|------|-------------------|
| **Given** | The starting state. |
| **When** | The trigger. |
| **Then** | The observable outcome. |

## Reading the verdict

The verdict is written back onto the intent.
`)

	f.write("docs/how-to/install.md", `# Install

You can use it as a [plugin](#plugin), a [binary](#cli), or by [building](#build) it.

## Plugin

Add the marketplace, then install the plugin:

`+"```text"+`
/plugin marketplace add fixture/repo
`+"```"+`

A second paragraph about the plugin.

A third paragraph about the plugin.

A fourth paragraph about the plugin.

A fifth paragraph, which folds away behind the disclosure.

## CLI

One line, checksum-verified, no administrator rights.

### macOS

`+"```sh"+`
sh -c 'echo fixture-darwin'
`+"```"+`

### Linux

`+"```sh"+`
sh -c 'echo fixture-linux'
`+"```"+`

### Windows

Not yet. This page will carry its command when it ships.

### Afterwards

Add this line to your shell profile:

`+"```sh"+`
export PATH="$HOME/.local/bin:$PATH"
`+"```"+`

A closing paragraph about what to do next.

## Build

`+"```bash"+`
go build ./cmd/fixture
`+"```"+`
`)

	f.write("docs/README.md", "# Documentation\n\nThe fixture's documentation index.\n")

	// --- the record ---------------------------------------------------------
	f.write(".abcd/development/decisions/adrs/0001-the-first-decision.md", `---
id: adr-1
slug: the-first-decision
status: accepted
date: 2026-01-02
supersedes: null
superseded_by: null
related_intents: [itd-1]
related_rfcs: []
related_adrs: [adr-2]
---

# ADR-1: The first decision

The first decision's body, which mentions iss-1 in prose.
`)

	f.write(".abcd/development/decisions/adrs/0002-the-second-decision.md", `---
id: adr-2
slug: the-second-decision
status: accepted
date: 2026-01-03
supersedes:
- adr-9
superseded_by: null
related_intents: []
related_rfcs: []
related_adrs: [adr-1]
---

# ADR-2: The second decision

The second decision's body.
`)

	f.write(".abcd/development/intents/drafts/itd-1-the-drafted-one.md", `---
id: itd-1
slug: the-drafted-one
spec_id: null
kind: standalone
builds_on: []
severity: minor
impact: additive
---

# The Drafted One

## Press Release

> **A fixture user gets a fixture thing.** It has not been built.

## Acceptance Criteria

- Given nothing, when nothing, then nothing.
`)

	f.write(".abcd/development/intents/planned/itd-2-the-shipped-one.md", `---
id: itd-2
slug: the-shipped-one
spec_id: spc-1
kind: standalone
builds_on: []
severity: minor
impact: additive
---

# The Shipped One

## Press Release

> **A fixture user opens the page and sees the record.** They had been reading
> frontmatter on a forge; now one page says what the project decided and why.
> "I stopped guessing," said the fixture user.

## Acceptance Criteria

- Given the fixture repository, when the build runs, then the landing page
  renders every chapter from its own page.
- Given a second criterion, then it does not appear in the quote.

## Audit Notes

Acceptance rollup: MET 2 · MET_WITH_CONCERNS 0 · NOT_MET 0 · INCONCLUSIVE 0
`)

	f.write(".abcd/development/specs/open/spc-1-the-spec.md", `---
id: spc-1
slug: the-spec
intent: itd-2
---
# The Spec

The spec's body, which mentions adr-1.
`)

	f.write(".abcd/work/issues/open/iss-1-a-fixture-issue.md", `---
schema_version: 1
id: "iss-1"
slug: "a-fixture-issue"
severity: "minor"
category: "tech-debt"
source: "agent-finding"
---

a fixture issue with no heading, mentioning adr-2 in its text
`)

	f.write("SECURITY.md", "# Security\n")
	f.write("ACKNOWLEDGEMENTS.md", "# Acknowledgements\n")
}
