---
name: site
description: Render this repository's website — the landing page composed from repository text under the single-source rule, and the record export derived from the record, git history and the changelog — by invoking the abcd binary. The bare form performs zero writes; build writes only inside its output directory.
argument-hint: "[build]"
---

# `/abcd:site` the website as a surface of the record

A project's website drifts from the project. Copy is written for the site,
lives only there, and slowly stops being true; the record that says what the
project actually decided stays invisible in frontmatter nobody reads. This
command renders a site that cannot drift, because it contains no words of its
own: every sentence is a span of a file in this repository, selected by path and
heading through `.abcd/site.json`.

## Bare — what this repo has declared

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" site --json
```

emits `{ "manifest": …, "ui_strings": …, "baseline": …, "out_dir": … }`:

- `manifest` — whether `.abcd/site.json` is present. Absent means this repo
  declares no composition, and there is nothing to render.
- `chapters` — how many chapters the manifest composes the landing page from.
- `issue_ledger` — whether the working-tier issue ledger is published. It is an
  explicit per-repo opt-in; the default renders the durable record only.
- `ui_strings` — whether the interface-string allowlist the manifest names is
  present. It is the complete list of words the generator may add.
- `baseline` and `baseline_entries` — the committed unresolved-reference
  ratchet and its size.
- `version`, `commit` — what a render would stamp the footer with.
- `out_dir`, `out_exists`, `out_files` — where a render writes, and what is
  there now.

Report the declared inputs first, then the output directory's state. It writes
nothing and exits `0` whatever it finds.

## `build` — render the site

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" site build --out site
```

reads exactly this set — `.abcd/site.json`, `site-src/ui.json`,
`.abcd/record-lint.json` (where the record stores are), the record under
`.abcd/development/` and the opted-in issue ledger, git history,
`CHANGELOG.md`, the composed pages and assets under `docs/`, the static inputs
`site-src/{site.css,site.js,redirects,headers}`, `.abcd/site-baseline.json`
(the ratchet the health block counts against), and `.claude-plugin/plugin.json`
(the forge URL, licence and author the links and footer use) — and writes the
landing page, the record export, the redirect and header maps, the stylesheet,
the script, and every referenced raster into the output directory, and nowhere
else. It reaches no network. The default output directory is `site`, which the
repository does not track.

The last two are declared deviations from the generic input contract: a repo
without a package manifest renders without the forge links rather than failing,
and the baseline is per-repo site configuration on the same opt-in footing as
`.abcd/site.json` itself — the record data proper stays record-format, git and
`CHANGELOG.md`.

Three flags exist so a release build can pin what the footer says rather than
reading it from the working tree: `--version`, `--commit` and `--date`. Left
unset, the version and date come from the newest dated `CHANGELOG.md` heading
and the commit from git `HEAD`.

Report the files written, then the four measurements the render prints: the
record's size (records, links, mentions), the unresolved references against the
committed baseline, the chart packing's overlap count (which is zero or the
picture is wrong), and the version and commit stamped into the footer. An
unresolved-reference count above the baseline is worth naming to the maintainer
even though this verb does not gate on it.

A failure names its cause and its place: a markdown construct outside the
rendered subset is reported as `file:line`, and so is an image the page names
that the repository does not carry. Neither is a rendering bug to work around —
the fix is an edit to the page.

**Binary resolution.** Run `"${CLAUDE_PLUGIN_ROOT}/abcd"` — a plugin install
provisions the binary into the plugin root, so this is the rung that fires for a
plugin user. If that path does not exist, try `abcd` on `PATH`; if that fails
too, you are in a source checkout of this repo, where — and only there —
`go run ./cmd/abcd` works, the published payload carrying no `cmd/`. To put a
binary on `PATH`, run `ahoy install` through whichever rung just resolved:
`"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install`, `abcd ahoy install`, or
`go run ./cmd/abcd ahoy install` in a source checkout.

**User input:** $ARGUMENTS
