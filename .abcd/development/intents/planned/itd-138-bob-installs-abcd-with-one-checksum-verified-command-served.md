---
id: itd-138
slug: bob-installs-abcd-with-one-checksum-verified-command-served
spec_id: spc-40
kind: standalone
suggested_kind: null
reclassification_history: []
builds_on: [itd-135]
severity: minor
impact: additive
---

# Bob installs abcd with one checksum-verified command served from the project's own domain

## Press Release

> **Bob installs abcd with one command served from the project's own
> domain.** Bob rolls tools out to a platform team, and every install path
> they adopt is a trust decision. `curl -fsSL https://abcdev.app/install.sh | sh`
> now does exactly what the README's one-liner does — detect the OS and
> architecture, refuse unsupported platforms in plain language, download the
> binary and `checksums.txt` from GitHub's permanent latest-release redirect,
> verify the SHA-256 and refuse on a mismatch or a manifest that does not
> list the binary, install to `~/.local/bin` without ever escalating
> privileges, print the PATH fix when needed, and finish with `abcd version`
> — because the script and the one-liner are rendered from the same
> template. The whole script runs through `main "$@"`, so a truncated
> download executes nothing. A visible link beside the command reads the
> script before running it, and the by-hand paragraph stays for anyone who
> will not pipe to a shell. That is the entire endpoint: no Homebrew tap, no
> redirect service, no second trust surface. "I read the script, checked the
> checksum step, and rolled it out," said Bob. "The one-liner on their README
> and the script on their domain could not disagree with each other, and
> that is what I am really buying."

## Why This Matters

An install command served from the project's own domain is the norm among
single-binary CLI tools, and its absence reads as immaturity; but every
additional distribution channel is a standing trust surface. Rendering the
script and the README one-liner from one template makes agreement structural
— a test asserts the same URLs and the same checksum step, with only the
platform detection resolved — and GitHub's version-free release asset names
plus the permanent `releases/latest/download/` redirect mean no redirect
infrastructure is needed at all. The deliberate refusals (no Homebrew, no
`/latest` redirects, no attestation page) are recorded with the plan and
stay one-line "later" items rather than silent gaps.

## Acceptance Criteria

- Given `https://abcdev.app/install.sh`, then it is served as a static asset
  with `Content-Type: text/plain; charset=utf-8`, generated from a template
  under `site-src/` that also renders README's one-liner, and a test asserts
  the script, the README form and the per-OS forms in
  `docs/how-to/install.md` share the same URLs and checksum step.
- Given a supported platform, when the script runs, then it downloads the
  platform binary and `checksums.txt` from
  `https://github.com/Partnermedia/abcd/releases/latest/download/`, verifies
  the SHA-256, refuses on a mismatch or a manifest that does not list the
  binary, installs to `~/.local/bin` (overridable with `--bin-dir`), never
  escalates privileges, prints the PATH fix when the directory is not on
  PATH, and ends by running `abcd version`.
- Given an unsupported platform, then the script refuses with a
  plain-language message naming the manual path (the releases page).
- Given a truncated download, then nothing executes, because the script's
  body runs only through a final `main "$@"`.
- Given the landing page's install strip, then a visible read-the-script
  link sits beside the command, and the manual-install paragraph with the
  releases link remains.

## Open Questions

_None recorded yet._

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._
